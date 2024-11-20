package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespaceSync       = "synchronizer"
	namespaceTxSelector = "txselector"
)

var (
	// Reorgs block reorg count
	Reorgs = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespaceSync,
			Name:      "reorgs",
			Help:      "",
		})

	// LastBlockNum last block synced
	LastBlockNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "synced_last_block_num",
			Help:      "",
		})

	// EthLastBlockNum last eth block synced
	EthLastBlockNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "eth_last_block_num",
			Help:      "",
		})
	// LastBatchNum last batch synced
	LastBatchNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "synced_last_batch_num",
			Help:      "",
		})

	// EthLastBatchNum last eth batch synced
	EthLastBatchNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "eth_last_batch_num",
			Help:      "",
		})

	// GetL1L2TxSelection L1L2 tx selection count
	GetL1L2TxSelection = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespaceTxSelector,
			Name:      "get_l1_l2_txselection_total",
			Help:      "",
		})

	// GetL2TxSelection L2 tx selection count
	GetL2TxSelection = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespaceTxSelector,
			Name:      "get_l2_txselection_total",
			Help:      "",
		})

	// SelectedL1UserTxs selected L1 user tx count
	SelectedL1UserTxs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceTxSelector,
			Name:      "selected_l1_user_txs",
			Help:      "",
		})

	// WaitServerProof duration time to get the calculated
	// proof from the server.
	WaitServerProof = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceSync,
			Name:      "wait_server_proof",
			Help:      "",
		}, []string{"batch_number", "pipeline_number"})
)

// MeasureDuration measure the method execution duration
// and save it into a histogram metric
func MeasureDuration(histogram *prometheus.HistogramVec, start time.Time, lvs ...string) {
	duration := time.Since(start)
	histogram.WithLabelValues(lvs...).Observe(float64(duration.Milliseconds()))
}
