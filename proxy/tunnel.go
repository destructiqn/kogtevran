package proxy

import (
	"log"
	"net"
	"sync"

	"github.com/Tnze/go-mc/chat"
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/license"
	"github.com/destructiqn/kogtevran/metrics"
	mcnet "github.com/destructiqn/kogtevran/net"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
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

type MinecraftTunnel struct {
	PairID     TunnelPairID
	TunnelPair *TunnelPair

	Closed bool
	Server *mcnet.Conn
	Client *mcnet.Conn

	ServerWrite sync.Mutex
	ClientWrite sync.Mutex

	State                  protocol.ConnectionState
	VerifyToken            []byte
	KeyBruteMode           bool
	SharedSecretCandidates [][]byte
	EnableEncryptionS2C    chan []byte
	EnableEncryptionC2S    chan [][]byte

	InventoryHandler *InventoryHandler
	TexteriaHandler  *TexteriaHandler
	ModuleHandler    *ModuleHandler
	PlayerHandler    *PlayerHandler
	EntityHandler    *EntityHandler
	ChatHandler      *ChatHandler
}

func (t *MinecraftTunnel) GetInventoryHandler() generic.InventoryHandler {
	return t.InventoryHandler
}

func (t *MinecraftTunnel) GetTexteriaHandler() generic.TexteriaHandler {
	return t.TexteriaHandler
}

func (t *MinecraftTunnel) GetChatHandler() generic.ChatHandler {
	return t.ChatHandler
}

func (t *MinecraftTunnel) SetState(state protocol.ConnectionState) {
	t.State = state
}

func (t *MinecraftTunnel) GetEntityHandler() generic.EntityHandler {
	return t.EntityHandler
}

func (t *MinecraftTunnel) GetPlayerHandler() generic.PlayerHandler {
	return t.PlayerHandler
}

func (t *MinecraftTunnel) GetModuleHandler() generic.ModuleHandler {
	return t.ModuleHandler
}

func (t *MinecraftTunnel) Disconnect(reason chat.Message) {
	metrics.Disconnects.With(prometheus.Labels{"reason": reason.String()}).Inc()
	_ = t.WriteClient(pk.Marshal(0x00, reason))
	t.Close()
}

func (t *MinecraftTunnel) WriteClient(packet pk.Packet) error {
	t.ClientWrite.Lock()
	err := t.Client.WritePacket(packet)
	t.ClientWrite.Unlock()
	return err
}

func (t *MinecraftTunnel) WriteServer(packet pk.Packet) error {
	t.ServerWrite.Lock()
	err := t.Server.WritePacket(packet)
	t.ServerWrite.Unlock()
	return err
}

func (t *MinecraftTunnel) GetRemoteAddr() string {
	host, _, err := net.SplitHostPort(t.Client.Socket.RemoteAddr().String())
	if err != nil {
		log.Println(err)
		return ""
	}

	return host
}

func (t *MinecraftTunnel) Close() {
	if t.Closed {
		return
	}

	for _, module := range t.GetModuleHandler().GetModules() {
		module.Close()
	}

	t.Closed = true
	_ = t.Server.Close()
	_ = t.Client.Close()
	CurrentTunnelPool.UnregisterPair(t.PairID)
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

func WrapConn(server, client *mcnet.Conn) *MinecraftTunnel {
	tunnel := &MinecraftTunnel{
		Server:              server,
		Client:              client,
		EnableEncryptionS2C: make(chan []byte),
		EnableEncryptionC2S: make(chan [][]byte),
	}

	tunnel.InventoryHandler = NewInventoryHandler()
	tunnel.TexteriaHandler = NewTexteriaHandler(tunnel)
	tunnel.ModuleHandler = NewModuleHandler(tunnel)
	tunnel.EntityHandler = NewEntityHandler(tunnel)
	tunnel.PlayerHandler = NewPlayerHandler(tunnel)
	tunnel.ChatHandler = NewChatHandler(tunnel)

	return tunnel
}

func (t *MinecraftTunnel) Attack(target int) error {
	return t.WriteServer(pk.Marshal(0x02, pk.VarInt(target), pk.VarInt(1)))
}

func (t *MinecraftTunnel) HasFeature(feature license.Feature) bool {
	return generic.IsDevelopmentEnvironment() || t.TunnelPair.License.HasFeature(feature)
}
