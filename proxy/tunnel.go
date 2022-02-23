package proxy

import (
	"sync"

	"github.com/Tnze/go-mc/chat"
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/net"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

type TunnelPair struct {
	SessionID string
	Auxiliary *AuxiliaryChannel
	Primary   *MinecraftTunnel
}

type TunnelPairID struct {
	Username   string
	RemoteAddr string
}

var CurrentTunnelPool = &TunnelPool{
	pool: make(map[TunnelPairID]*TunnelPair),
}

type TunnelPool struct {
	pool  map[TunnelPairID]*TunnelPair
	mutex sync.Mutex
}

func (p *TunnelPool) RegisterPair(id TunnelPairID, pair *TunnelPair) {
	p.mutex.Lock()
	p.pool[id] = pair
	p.mutex.Unlock()
}

func (p *TunnelPool) UnregisterPair(id TunnelPairID) {
	tunnel, ok := p.GetPair(id)
	if !ok {
		return
	}

	p.mutex.Lock()
	delete(p.pool, id)
	p.mutex.Unlock()

	tunnel.SessionID = ""
	tunnel.Auxiliary.Conn.Close()
}

func (p *TunnelPool) GetPair(id TunnelPairID) (*TunnelPair, bool) {
	tunnel, ok := p.pool[id]
	return tunnel, ok
}

type MinecraftTunnel struct {
	PairID     TunnelPairID
	TunnelPair *TunnelPair

	Closed bool
	Server *net.Conn
	Client *net.Conn

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

func (t *MinecraftTunnel) Close() {
	for _, module := range t.GetModuleHandler().GetModules() {
		module.Close()
	}

	t.Closed = true
	_ = t.Server.Close()
	_ = t.Client.Close()
	CurrentTunnelPool.UnregisterPair(t.PairID)
}

func WrapConn(server, client *net.Conn) *MinecraftTunnel {
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
