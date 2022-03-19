package proxy

import (
	"sync"

	"github.com/destructiqn/kogtevran/license"
	"github.com/destructiqn/kogtevran/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type TunnelPair struct {
	SessionID string
	Auxiliary *AuxiliaryChannel
	Primary   *MinecraftTunnel
	License   license.License
}

type TunnelPairID struct {
	Username   string
	RemoteAddr string
}

var CurrentTunnelPool = &TunnelPool{
	pool: make(map[TunnelPairID]*TunnelPair),
}

type TunnelPool struct {
	pool map[TunnelPairID]*TunnelPair
	sync.Mutex
}

func (p *TunnelPool) RegisterPair(id TunnelPairID, pair *TunnelPair) {
	p.Lock()
	p.pool[id] = pair
	p.Unlock()
	UpdateConnectionMetrics()
}

func (p *TunnelPool) UnregisterPair(id TunnelPairID) {
	pair, ok := p.GetPair(id)
	if !ok {
		return
	}

	p.Lock()
	delete(p.pool, id)
	p.Unlock()

	pair.SessionID = ""
	_ = pair.Auxiliary.Close()
	pair.Auxiliary = nil

	pair.Primary.Close()
	pair.Primary = nil

	UpdateConnectionMetrics()
}

func (p *TunnelPool) GetPair(id TunnelPairID) (*TunnelPair, bool) {
	tunnel, ok := p.pool[id]
	return tunnel, ok
}

func UpdateConnectionMetrics() {
	CurrentTunnelPool.Lock()
	defer CurrentTunnelPool.Unlock()

	var (
		auxiliaryConnections int
		minecraftConnections int
	)

	for _, pair := range CurrentTunnelPool.pool {
		if pair.Primary != nil {
			minecraftConnections++
		}

		if pair.Auxiliary != nil {
			auxiliaryConnections++
		}
	}

	metrics.TotalConnections.With(prometheus.Labels{"type": "auxiliary"}).Set(float64(auxiliaryConnections))
	metrics.TotalConnections.With(prometheus.Labels{"type": "minecraft"}).Set(float64(minecraftConnections))
}

func UpdateModuleMetrics() {
	CurrentTunnelPool.Lock()
	defer CurrentTunnelPool.Unlock()

	data := make(map[string]int)
	for _, pair := range CurrentTunnelPool.pool {
		if pair.Primary == nil {
			continue
		}

		for _, module := range pair.Primary.ModuleHandler.GetModules() {
			if module.IsEnabled() {
				if _, ok := data[module.GetIdentifier()]; !ok {
					data[module.GetIdentifier()] = 0
				}

				data[module.GetIdentifier()] += 1
			}
		}
	}

	metrics.UsedModules.Reset()
	for module, count := range data {
		metrics.UsedModules.With(prometheus.Labels{"identifier": module}).Set(float64(count))
	}
}
