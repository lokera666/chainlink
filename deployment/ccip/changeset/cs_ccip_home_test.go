package changeset

import (
	"math/big"
	"regexp"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-ccip/chainconfig"
	cciptypes "github.com/smartcontractkit/chainlink-ccip/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-testing-framework/lib/utils/testcontext"

	"github.com/smartcontractkit/chainlink/deployment"
	"github.com/smartcontractkit/chainlink/deployment/ccip/changeset/internal"
	"github.com/smartcontractkit/chainlink/v2/core/capabilities/ccip/types"
	"github.com/smartcontractkit/chainlink/v2/core/logger"

	"github.com/stretchr/testify/require"

	commonchangeset "github.com/smartcontractkit/chainlink/deployment/common/changeset"
	"github.com/smartcontractkit/chainlink/deployment/common/proposalutils"
)

func TestInvalidOCR3Params(t *testing.T) {
	e, _ := NewMemoryEnvironment(t,
		WithPrerequisiteDeployment(nil))
	chain1 := e.Env.AllChainSelectors()[0]
	envNodes, err := deployment.NodeInfo(e.Env.NodeIDs, e.Env.Offchain)
	require.NoError(t, err)
	// Need to deploy prerequisites first so that we can form the USDC config
	// no proposals to be made, timelock can be passed as nil here
	e.Env, err = commonchangeset.ApplyChangesets(t, e.Env, nil, []commonchangeset.ChangesetApplication{
		{
			Changeset: commonchangeset.WrapChangeSet(DeployHomeChain),
			Config: DeployHomeChainConfig{
				HomeChainSel:     e.HomeChainSel,
				RMNDynamicConfig: NewTestRMNDynamicConfig(),
				RMNStaticConfig:  NewTestRMNStaticConfig(),
				NodeOperators:    NewTestNodeOperator(e.Env.Chains[e.HomeChainSel].DeployerKey.From),
				NodeP2PIDsPerNodeOpAdmin: map[string][][32]byte{
					TestNodeOperator: envNodes.NonBootstraps().PeerIDs(),
				},
			},
		},
		{
			Changeset: commonchangeset.WrapChangeSet(DeployChainContracts),
			Config: DeployChainContractsConfig{
				ChainSelectors:    []uint64{chain1},
				HomeChainSelector: e.HomeChainSel,
			},
		},
	})
	require.NoError(t, err)

	state, err := LoadOnchainState(e.Env)
	require.NoError(t, err)
	nodes, err := deployment.NodeInfo(e.Env.NodeIDs, e.Env.Offchain)
	require.NoError(t, err)
	params := DefaultOCRParams(e.FeedChainSel, nil, nil, true, true)
	// tweak params to have invalid config
	// make DeltaRound greater than DeltaProgress
	params.OCRParameters.DeltaRound = params.OCRParameters.DeltaProgress + time.Duration(1)
	_, err = internal.BuildOCR3ConfigForCCIPHome(
		e.Env.OCRSecrets,
		state.Chains[chain1].OffRamp,
		e.Env.Chains[chain1],
		nodes.NonBootstraps(),
		state.Chains[e.HomeChainSel].RMNHome.Address(),
		params.OCRParameters,
		params.CommitOffChainConfig,
		params.ExecuteOffChainConfig,
	)
	require.Errorf(t, err, "expected error")
	pattern := `DeltaRound \(\d+\.\d+s\) must be less than DeltaProgress \(\d+s\)`
	matched, err1 := regexp.MatchString(pattern, err.Error())
	require.NoError(t, err1)
	require.True(t, matched)
}

func Test_PromoteCandidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testcontext.Get(t)
			tenv, _ := NewMemoryEnvironment(t,
				WithChains(2),
				WithNodes(4))
			state, err := LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			// Deploy to all chains.
			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			var (
				capReg   = state.Chains[tenv.HomeChainSel].CapabilityRegistry
				ccipHome = state.Chains[tenv.HomeChainSel].CCIPHome
			)
			donID, err := internal.DonIDForChain(capReg, ccipHome, dest)
			require.NoError(t, err)
			require.NotEqual(t, uint32(0), donID)
			t.Logf("donID: %d", donID)
			candidateDigestCommitBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitBefore)
			ActiveDigestExecBefore, err := ccipHome.GetActiveDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, ActiveDigestExecBefore)

			var mcmsConfig *MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &MCMSConfig{
					MinDelay: 0,
				}
			}
			// promotes zero digest on commit and ensure exec is not affected
			_, err = commonchangeset.ApplyChangesets(t, tenv.Env, map[uint64]*proposalutils.TimelockExecutionContracts{
				tenv.HomeChainSel: {
					Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
					CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
				},
			}, []commonchangeset.ChangesetApplication{
				{
					Changeset: commonchangeset.WrapChangeSet(PromoteCandidateChangeset),
					Config: PromoteCandidateChangesetConfig{
						HomeChainSelector: tenv.HomeChainSel,
						PluginInfo: []PromoteCandidatePluginInfo{
							{
								RemoteChainSelectors: []uint64{dest},
								PluginType:           types.PluginTypeCCIPCommit,
							},
						},
						MCMS: mcmsConfig,
					},
				},
			})
			require.NoError(t, err)

			// after promoting the zero digest, active digest should also be zero
			activeDigestCommit, err := ccipHome.GetActiveDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, activeDigestCommit)

			activeDigestExec, err := ccipHome.GetActiveDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, ActiveDigestExecBefore, activeDigestExec)
		})
	}
}

func Test_SetCandidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testcontext.Get(t)
			tenv, _ := NewMemoryEnvironment(t,
				WithChains(2),
				WithNodes(4))
			state, err := LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			// Deploy to all chains.
			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			var (
				capReg   = state.Chains[tenv.HomeChainSel].CapabilityRegistry
				ccipHome = state.Chains[tenv.HomeChainSel].CCIPHome
			)
			donID, err := internal.DonIDForChain(capReg, ccipHome, dest)
			require.NoError(t, err)
			require.NotEqual(t, uint32(0), donID)
			candidateDigestCommitBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitBefore)
			candidateDigestExecBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestExecBefore)

			var mcmsConfig *MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &MCMSConfig{
					MinDelay: 0,
				}
			}
			tokenConfig := NewTestTokenConfig(state.Chains[tenv.FeedChainSel].USDFeeds)
			_, err = commonchangeset.ApplyChangesets(t, tenv.Env, map[uint64]*proposalutils.TimelockExecutionContracts{
				tenv.HomeChainSel: {
					Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
					CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
				},
			}, []commonchangeset.ChangesetApplication{
				{
					Changeset: commonchangeset.WrapChangeSet(SetCandidateChangeset),
					Config: SetCandidateChangesetConfig{
						SetCandidateConfigBase: SetCandidateConfigBase{
							HomeChainSelector: tenv.HomeChainSel,
							FeedChainSelector: tenv.FeedChainSel,
							MCMS:              mcmsConfig,
						},
						PluginInfo: []SetCandidatePluginInfo{
							{
								OCRConfigPerRemoteChainSelector: map[uint64]CCIPOCRParams{
									dest: DefaultOCRParams(tenv.FeedChainSel, tokenConfig.GetTokenInfo(logger.TestLogger(t), state.Chains[dest].LinkToken, state.Chains[dest].Weth9), nil, true, false),
								},
								PluginType: types.PluginTypeCCIPCommit,
							},
							{
								OCRConfigPerRemoteChainSelector: map[uint64]CCIPOCRParams{
									dest: DefaultOCRParams(tenv.FeedChainSel, tokenConfig.GetTokenInfo(logger.TestLogger(t), state.Chains[dest].LinkToken, state.Chains[dest].Weth9), nil, false, true),
								},
								PluginType: types.PluginTypeCCIPExec,
							},
						},
					},
				},
			})
			require.NoError(t, err)

			// after setting a new candidate on both plugins, the candidate config digest
			// should be nonzero.
			candidateDigestCommitAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestCommitAfter)
			require.NotEqual(t, candidateDigestCommitBefore, candidateDigestCommitAfter)

			candidateDigestExecAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestExecAfter)
			require.NotEqual(t, candidateDigestExecBefore, candidateDigestExecAfter)
		})
	}
}

func Test_RevokeCandidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testcontext.Get(t)
			tenv, _ := NewMemoryEnvironment(t,
				WithChains(2),
				WithNodes(4))
			state, err := LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			// Deploy to all chains.
			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			var (
				capReg   = state.Chains[tenv.HomeChainSel].CapabilityRegistry
				ccipHome = state.Chains[tenv.HomeChainSel].CCIPHome
			)
			donID, err := internal.DonIDForChain(capReg, ccipHome, dest)
			require.NoError(t, err)
			require.NotEqual(t, uint32(0), donID)
			candidateDigestCommitBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitBefore)
			candidateDigestExecBefore, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestExecBefore)

			var mcmsConfig *MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &MCMSConfig{
					MinDelay: 0,
				}
			}
			tokenConfig := NewTestTokenConfig(state.Chains[tenv.FeedChainSel].USDFeeds)
			_, err = commonchangeset.ApplyChangesets(t, tenv.Env, map[uint64]*proposalutils.TimelockExecutionContracts{
				tenv.HomeChainSel: {
					Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
					CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
				},
			}, []commonchangeset.ChangesetApplication{
				{
					Changeset: commonchangeset.WrapChangeSet(SetCandidateChangeset),
					Config: SetCandidateChangesetConfig{
						SetCandidateConfigBase: SetCandidateConfigBase{
							HomeChainSelector: tenv.HomeChainSel,
							FeedChainSelector: tenv.FeedChainSel,
							MCMS:              mcmsConfig,
						},
						PluginInfo: []SetCandidatePluginInfo{
							{
								OCRConfigPerRemoteChainSelector: map[uint64]CCIPOCRParams{
									dest: DefaultOCRParams(tenv.FeedChainSel, tokenConfig.GetTokenInfo(logger.TestLogger(t), state.Chains[dest].LinkToken, state.Chains[dest].Weth9), nil, true, true),
								},
								PluginType: types.PluginTypeCCIPCommit,
							},
							{
								OCRConfigPerRemoteChainSelector: map[uint64]CCIPOCRParams{
									dest: DefaultOCRParams(tenv.FeedChainSel, tokenConfig.GetTokenInfo(logger.TestLogger(t), state.Chains[dest].LinkToken, state.Chains[dest].Weth9), nil, true, true),
								},
								PluginType: types.PluginTypeCCIPExec,
							},
						},
					},
				},
			})
			require.NoError(t, err)

			// after setting a new candidate on both plugins, the candidate config digest
			// should be nonzero.
			candidateDigestCommitAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestCommitAfter)
			require.NotEqual(t, candidateDigestCommitBefore, candidateDigestCommitAfter)

			candidateDigestExecAfter, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.NotEqual(t, [32]byte{}, candidateDigestExecAfter)
			require.NotEqual(t, candidateDigestExecBefore, candidateDigestExecAfter)

			// next we can revoke candidate - this should set the candidate digest back to zero
			_, err = commonchangeset.ApplyChangesets(t, tenv.Env, map[uint64]*proposalutils.TimelockExecutionContracts{
				tenv.HomeChainSel: {
					Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
					CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
				},
			}, []commonchangeset.ChangesetApplication{
				{
					Changeset: commonchangeset.WrapChangeSet(RevokeCandidateChangeset),
					Config: RevokeCandidateChangesetConfig{
						HomeChainSelector:   tenv.HomeChainSel,
						RemoteChainSelector: dest,
						PluginType:          types.PluginTypeCCIPCommit,
						MCMS:                mcmsConfig,
					},
				},
				{
					Changeset: commonchangeset.WrapChangeSet(RevokeCandidateChangeset),
					Config: RevokeCandidateChangesetConfig{
						HomeChainSelector:   tenv.HomeChainSel,
						RemoteChainSelector: dest,
						PluginType:          types.PluginTypeCCIPExec,
						MCMS:                mcmsConfig,
					},
				},
			})
			require.NoError(t, err)

			// after revoking the candidate, the candidate digest should be zero
			candidateDigestCommitAfterRevoke, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPCommit))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestCommitAfterRevoke)

			candidateDigestExecAfterRevoke, err := ccipHome.GetCandidateDigest(&bind.CallOpts{
				Context: ctx,
			}, donID, uint8(types.PluginTypeCCIPExec))
			require.NoError(t, err)
			require.Equal(t, [32]byte{}, candidateDigestExecAfterRevoke)
		})
	}
}

func transferToTimelock(
	t *testing.T,
	tenv DeployedEnv,
	state CCIPOnChainState,
	source,
	dest uint64) {
	// Transfer ownership to timelock so that we can promote the zero digest later down the line.
	_, err := commonchangeset.ApplyChangesets(t, tenv.Env, map[uint64]*proposalutils.TimelockExecutionContracts{
		source: {
			Timelock:  state.Chains[source].Timelock,
			CallProxy: state.Chains[source].CallProxy,
		},
		dest: {
			Timelock:  state.Chains[dest].Timelock,
			CallProxy: state.Chains[dest].CallProxy,
		},
		tenv.HomeChainSel: {
			Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
			CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
		},
	}, []commonchangeset.ChangesetApplication{
		{
			Changeset: commonchangeset.WrapChangeSet(commonchangeset.TransferToMCMSWithTimelock),
			Config:    genTestTransferOwnershipConfig(tenv, []uint64{source, dest}, state),
		},
	})
	require.NoError(t, err)
	assertTimelockOwnership(t, tenv, []uint64{source, dest}, state)
}

func Test_UpdateChainConfigs(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mcmsEnabled bool
	}{
		{
			name:        "MCMS enabled",
			mcmsEnabled: true,
		},
		{
			name:        "MCMS disabled",
			mcmsEnabled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tenv, _ := NewMemoryEnvironment(t, WithChains(3))
			state, err := LoadOnchainState(tenv.Env)
			require.NoError(t, err)

			allChains := maps.Keys(tenv.Env.Chains)
			source := allChains[0]
			dest := allChains[1]
			otherChain := allChains[2]

			if tc.mcmsEnabled {
				// Transfer ownership to timelock so that we can promote the zero digest later down the line.
				transferToTimelock(t, tenv, state, source, dest)
			}

			ccipHome := state.Chains[tenv.HomeChainSel].CCIPHome
			otherChainConfig, err := ccipHome.GetChainConfig(nil, otherChain)
			require.NoError(t, err)
			assert.NotZero(t, otherChainConfig.FChain)

			var mcmsConfig *MCMSConfig
			if tc.mcmsEnabled {
				mcmsConfig = &MCMSConfig{
					MinDelay: 0,
				}
			}
			_, err = commonchangeset.ApplyChangesets(t, tenv.Env, map[uint64]*proposalutils.TimelockExecutionContracts{
				tenv.HomeChainSel: {
					Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
					CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
				},
			}, []commonchangeset.ChangesetApplication{
				{
					Changeset: commonchangeset.WrapChangeSet(UpdateChainConfig),
					Config: UpdateChainConfigConfig{
						HomeChainSelector:  tenv.HomeChainSel,
						RemoteChainRemoves: []uint64{otherChain},
						RemoteChainAdds:    make(map[uint64]ChainConfig),
						MCMS:               mcmsConfig,
					},
				},
			})
			require.NoError(t, err)

			// other chain should be gone
			chainConfigAfter, err := ccipHome.GetChainConfig(nil, otherChain)
			require.NoError(t, err)
			assert.Zero(t, chainConfigAfter.FChain)

			// Lets add it back now.
			_, err = commonchangeset.ApplyChangesets(t, tenv.Env, map[uint64]*proposalutils.TimelockExecutionContracts{
				tenv.HomeChainSel: {
					Timelock:  state.Chains[tenv.HomeChainSel].Timelock,
					CallProxy: state.Chains[tenv.HomeChainSel].CallProxy,
				},
			}, []commonchangeset.ChangesetApplication{
				{
					Changeset: commonchangeset.WrapChangeSet(UpdateChainConfig),
					Config: UpdateChainConfigConfig{
						HomeChainSelector:  tenv.HomeChainSel,
						RemoteChainRemoves: []uint64{},
						RemoteChainAdds: map[uint64]ChainConfig{
							otherChain: {
								EncodableChainConfig: chainconfig.ChainConfig{
									GasPriceDeviationPPB:    cciptypes.BigInt{Int: big.NewInt(internal.GasPriceDeviationPPB)},
									DAGasPriceDeviationPPB:  cciptypes.BigInt{Int: big.NewInt(internal.DAGasPriceDeviationPPB)},
									OptimisticConfirmations: internal.OptimisticConfirmations,
								},
								FChain:  otherChainConfig.FChain,
								Readers: otherChainConfig.Readers,
							},
						},
						MCMS: mcmsConfig,
					},
				},
			})
			require.NoError(t, err)

			chainConfigAfter2, err := ccipHome.GetChainConfig(nil, otherChain)
			require.NoError(t, err)
			assert.Equal(t, chainConfigAfter2.FChain, otherChainConfig.FChain)
			assert.Equal(t, chainConfigAfter2.Readers, otherChainConfig.Readers)
			assert.Equal(t, chainConfigAfter2.Config, otherChainConfig.Config)
		})
	}
}
