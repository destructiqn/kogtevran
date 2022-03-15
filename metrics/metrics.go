package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	TotalConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kogtevran",
		Subsystem: "server",
		Name:      "connections",
		Help:      "Amount of currently handled connections",
	}, []string{"type"})

	HandshakeCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kogtevran",
		Subsystem: "server",
		Name:      "handshakes",
		Help:      "Amount of handshake requests",
	}, []string{"state"})

	Disconnects = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "kogtevran",
		Subsystem: "server",
		Name:      "disconnects",
		Help:      "Amount of disconnects initiated by proxy",
	}, []string{"reason"})

	UsedModules = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "kogtevran",
		Subsystem: "server",
		Name:      "modules",
		Help:      "Amount of currently enabled module instances",
	}, []string{"identifier"})
)

func RegisterMetrics() {
	prometheus.MustRegister(TotalConnections)
	prometheus.MustRegister(HandshakeCount)
	prometheus.MustRegister(Disconnects)
	prometheus.MustRegister(UsedModules)
}
