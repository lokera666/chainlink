package mercurytransmitter

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jpillora/backoff"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	commonconfig "github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	llotypes "github.com/smartcontractkit/chainlink-common/pkg/types/llo"
	"github.com/smartcontractkit/chainlink-data-streams/llo"

	corelogger "github.com/smartcontractkit/chainlink/v2/core/logger"
	"github.com/smartcontractkit/chainlink/v2/core/services/llo/evm"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay/evm/mercury/wsrpc"
	"github.com/smartcontractkit/chainlink/v2/core/services/relay/evm/mercury/wsrpc/pb"
	"github.com/smartcontractkit/chainlink/v2/core/utils"
)

var (
	promTransmitQueueDeleteErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "llo",
		Subsystem: "mercurytransmitter",
		Name:      "transmit_queue_delete_error_count",
		Help:      "Running count of DB errors when trying to delete an item from the queue DB",
	},
		[]string{"donID", "serverURL"},
	)
	promTransmitQueueInsertErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "llo",
		Subsystem: "mercurytransmitter",
		Name:      "transmit_queue_insert_error_count",
		Help:      "Running count of DB errors when trying to insert an item into the queue DB",
	},
		[]string{"donID", "serverURL"},
	)
	promTransmitQueuePushErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "llo",
		Subsystem: "mercurytransmitter",
		Name:      "transmit_queue_push_error_count",
		Help:      "Running count of DB errors when trying to push an item onto the queue",
	},
		[]string{"donID", "serverURL"},
	)
	promTransmitServerErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "llo",
		Subsystem: "mercurytransmitter",
		Name:      "transmit_server_error_count",
		Help:      "Number of errored transmissions that failed due to an error returned by the mercury server",
	},
		[]string{"donID", "serverURL", "code"},
	)
	promTransmitConcurrentTransmitGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "llo",
		Subsystem: "mercurytransmitter",
		Name:      "concurrent_transmit_gauge",
		Help:      "Gauge that measures the number of transmit threads currently waiting on a remote transmit call. You may wish to alert if this exceeds some number for a given period of time, or if it ever reaches its max.",
	},
		[]string{"donID", "serverURL"},
	)
	promTransmitConcurrentDeleteGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "llo",
		Subsystem: "mercurytransmitter",
		Name:      "concurrent_delete_gauge",
		Help:      "Gauge that measures the number of delete threads currently waiting on a delete call to the DB. You may wish to alert if this exceeds some number for a given period of time, or if it ever reaches its max.",
	},
		[]string{"donID", "serverURL"},
	)
)

type ReportPacker interface {
	Pack(digest types.ConfigDigest, seqNr uint64, report ocr2types.Report, sigs []ocr2types.AttributedOnchainSignature) ([]byte, error)
}

// A server handles the queue for a given mercury server

type server struct {
	lggr           logger.SugaredLogger
	verboseLogging bool

	transmitTimeout time.Duration

	c  wsrpc.Client
	pm *persistenceManager
	q  TransmitQueue

	deleteQueue chan [32]byte

	url string

	evmPremiumLegacyPacker ReportPacker
	jsonPacker             ReportPacker

	transmitSuccessCount            prometheus.Counter
	transmitDuplicateCount          prometheus.Counter
	transmitConnectionErrorCount    prometheus.Counter
	transmitQueueDeleteErrorCount   prometheus.Counter
	transmitQueueInsertErrorCount   prometheus.Counter
	transmitQueuePushErrorCount     prometheus.Counter
	transmitConcurrentTransmitGauge prometheus.Gauge
	transmitConcurrentDeleteGauge   prometheus.Gauge

	transmitThreadBusyCount atomic.Int32
	deleteThreadBusyCount   atomic.Int32
}

type QueueConfig interface {
	TransmitQueueMaxSize() uint32
	TransmitTimeout() commonconfig.Duration
}

func newServer(lggr logger.Logger, verboseLogging bool, cfg QueueConfig, client wsrpc.Client, orm ORM, serverURL string) *server {
	pm := NewPersistenceManager(lggr, orm, serverURL, int(cfg.TransmitQueueMaxSize()), flushDeletesFrequency, pruneFrequency)
	donIDStr := fmt.Sprintf("%d", pm.DonID())
	var codecLggr logger.Logger
	if verboseLogging {
		codecLggr = lggr
	} else {
		codecLggr = corelogger.NullLogger
	}

	s := &server{
		logger.Sugared(lggr),
		verboseLogging,
		cfg.TransmitTimeout().Duration(),
		client,
		pm,
		NewTransmitQueue(lggr, serverURL, int(cfg.TransmitQueueMaxSize()), pm),
		make(chan [32]byte, int(cfg.TransmitQueueMaxSize())),
		serverURL,
		evm.NewReportCodecPremiumLegacy(codecLggr, pm.DonID()),
		llo.JSONReportCodec{},
		promTransmitSuccessCount.WithLabelValues(donIDStr, serverURL),
		promTransmitDuplicateCount.WithLabelValues(donIDStr, serverURL),
		promTransmitConnectionErrorCount.WithLabelValues(donIDStr, serverURL),
		promTransmitQueueDeleteErrorCount.WithLabelValues(donIDStr, serverURL),
		promTransmitQueueInsertErrorCount.WithLabelValues(donIDStr, serverURL),
		promTransmitQueuePushErrorCount.WithLabelValues(donIDStr, serverURL),
		promTransmitConcurrentTransmitGauge.WithLabelValues(donIDStr, serverURL),
		promTransmitConcurrentDeleteGauge.WithLabelValues(donIDStr, serverURL),
		atomic.Int32{},
		atomic.Int32{},
	}

	return s
}

func (s *server) HealthReport() map[string]error {
	report := map[string]error{}
	services.CopyHealth(report, s.c.HealthReport())
	services.CopyHealth(report, s.q.HealthReport())
	return report
}

func (s *server) runDeleteQueueLoop(stopCh services.StopChan, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := stopCh.NewCtx()
	defer cancel()

	// Exponential backoff for very rarely occurring errors (DB disconnect etc)
	b := backoff.Backoff{
		Min:    1 * time.Second,
		Max:    120 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	for {
		select {
		case hash := <-s.deleteQueue:
			for {
				s.deleteThreadBusyCountInc()
				if err := s.pm.orm.Delete(ctx, [][32]byte{hash}); err != nil {
					s.lggr.Errorw("Failed to delete transmission record", "err", err, "transmissionHash", hash)
					s.transmitQueueDeleteErrorCount.Inc()
					select {
					case <-time.After(b.Duration()):
						// Wait a backoff duration before trying to delete again
						continue
					case <-stopCh:
						s.deleteThreadBusyCountDec()
						// abort and return immediately on stop even if items remain in queue
						return
					}
				}
				break
			}
			// success
			b.Reset()
			s.deleteThreadBusyCountDec()
		case <-stopCh:
			// abort and return immediately on stop even if items remain in queue
			return
		}
	}
}

func (s *server) transmitThreadBusyCountInc() {
	val := s.transmitThreadBusyCount.Add(1)
	s.transmitConcurrentTransmitGauge.Set(float64(val))
}
func (s *server) transmitThreadBusyCountDec() {
	val := s.transmitThreadBusyCount.Add(-1)
	s.transmitConcurrentTransmitGauge.Set(float64(val))
}
func (s *server) deleteThreadBusyCountInc() {
	val := s.deleteThreadBusyCount.Add(1)
	s.transmitConcurrentDeleteGauge.Set(float64(val))
}
func (s *server) deleteThreadBusyCountDec() {
	val := s.deleteThreadBusyCount.Add(-1)
	s.transmitConcurrentDeleteGauge.Set(float64(val))
}

func (s *server) runQueueLoop(stopCh services.StopChan, wg *sync.WaitGroup, donIDStr string) {
	defer wg.Done()
	// Exponential backoff with very short retry interval (since latency is a priority)
	// 5ms, 10ms, 20ms, 40ms etc
	b := backoff.Backoff{
		Min:    5 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: true,
	}
	ctx, cancel := stopCh.NewCtx()
	defer cancel()
	cont := true
	for cont {
		cont = func() bool {
			t := s.q.BlockingPop()
			if t == nil {
				// queue was closed
				return false
			}

			s.transmitThreadBusyCountInc()
			defer s.transmitThreadBusyCountDec()

			req, res, err := func(ctx context.Context) (*pb.TransmitRequest, *pb.TransmitResponse, error) {
				ctx, cancelFn := context.WithTimeout(ctx, utils.WithJitter(s.transmitTimeout))
				defer cancelFn()
				return s.transmit(ctx, t)
			}(ctx)

			lggr := s.lggr.With("transmission", t, "response", res, "transmissionHash", fmt.Sprintf("%x", t.Hash()))
			if req != nil {
				lggr = s.lggr.With("req.Payload", req.Payload, "req.ReportFormat", req.ReportFormat)
			}

			if ctx.Err() != nil {
				// only canceled on transmitter close so we can exit
				return false
			} else if err != nil {
				s.transmitConnectionErrorCount.Inc()
				lggr.Errorw("Transmit report failed", "err", err)
				if ok := s.q.Push(t); !ok {
					s.lggr.Error("Failed to push report to transmit queue; queue is closed")
					return false
				}
				// Wait a backoff duration before pulling the most recent transmission
				// the heap
				select {
				case <-time.After(b.Duration()):
					return true
				case <-stopCh:
					return false
				}
			}

			b.Reset()
			if res.Error == "" {
				s.transmitSuccessCount.Inc()
				lggr.Debug("Transmit report success")
			} else {
				// We don't need to retry here because the mercury server
				// has confirmed it received the report. We only need to retry
				// on networking/unknown errors
				switch res.Code {
				case DuplicateReport:
					s.transmitSuccessCount.Inc()
					s.transmitDuplicateCount.Inc()
					lggr.Debug("Transmit report success; duplicate report")
				default:
					promTransmitServerErrorCount.WithLabelValues(donIDStr, s.url, strconv.FormatInt(int64(res.Code), 10)).Inc()
					lggr.Errorw("Transmit report failed; mercury server returned error", "err", res.Error, "code", res.Code)
				}
			}

			select {
			case s.deleteQueue <- t.Hash():
			default:
				lggr.Criticalw("Delete queue is full")
			}
			return true
		}()
	}
}

func (s *server) transmit(ctx context.Context, t *Transmission) (*pb.TransmitRequest, *pb.TransmitResponse, error) {
	var payload []byte
	var err error

	switch t.Report.Info.ReportFormat {
	case llotypes.ReportFormatJSON:
		payload, err = s.jsonPacker.Pack(t.ConfigDigest, t.SeqNr, t.Report.Report, t.Sigs)
	case llotypes.ReportFormatEVMPremiumLegacy, llotypes.ReportFormatEVMABIEncodeUnpacked:
		payload, err = s.evmPremiumLegacyPacker.Pack(t.ConfigDigest, t.SeqNr, t.Report.Report, t.Sigs)
	default:
		return nil, nil, fmt.Errorf("Transmit failed; don't know how to Pack unsupported report format: %q", t.Report.Info.ReportFormat)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("Transmit: encode failed; %w", err)
	}

	req := &pb.TransmitRequest{
		Payload:      payload,
		ReportFormat: uint32(t.Report.Info.ReportFormat),
	}

	resp, err := s.c.Transmit(ctx, req)
	return req, resp, err
}
