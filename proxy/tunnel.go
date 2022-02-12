package proxy

import (
	"sync"

	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/net"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

var CurrentTunnelPool = &TunnelPool{
	pool: make(map[string]*MinecraftTunnel),
}

type TunnelPool struct {
	pool  map[string]*MinecraftTunnel
	mutex sync.Mutex
}

func (p *TunnelPool) RegisterTunnel(sessionID string, tunnel *MinecraftTunnel) {
	p.mutex.Lock()
	p.pool[sessionID] = tunnel
	p.mutex.Unlock()
	tunnel.SessionID = sessionID
}

func (p *TunnelPool) UnregisterTunnel(sessionID string) {
	tunnel, ok := p.GetTunnel(sessionID)
	if !ok {
		return
	}

	p.mutex.Lock()
	delete(p.pool, sessionID)
	p.mutex.Unlock()

	tunnel.SessionID = ""
	tunnel.AuxiliaryChannel.Conn.Close()
}

func (p *TunnelPool) GetTunnel(sessionID string) (*MinecraftTunnel, bool) {
	tunnel, ok := p.pool[sessionID]
	return tunnel, ok
}

type MinecraftTunnel struct {
	SessionID                 string
	AuxiliaryChannel          *AuxiliaryChannel
	AuxiliaryChannelAvailable chan bool

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
	CurrentTunnelPool.UnregisterTunnel(t.SessionID)
}

func WrapConn(server, client *net.Conn) *MinecraftTunnel {
	tunnel := &MinecraftTunnel{
		Server:                    server,
		Client:                    client,
		AuxiliaryChannelAvailable: make(chan bool),
		EnableEncryptionS2C:       make(chan []byte),
		EnableEncryptionC2S:       make(chan [][]byte),
	}

	tunnel.InventoryHandler = NewInventoryHandler()
	tunnel.TexteriaHandler = NewTexteriaHandler(tunnel)
	tunnel.ModuleHandler = NewModuleHandler(tunnel)
	tunnel.EntityHandler = NewEntityHandler()
	tunnel.PlayerHandler = NewPlayerHandler(tunnel)
	tunnel.ChatHandler = NewChatHandler(tunnel)

	return tunnel
}

func (t *MinecraftTunnel) Attack(target int) error {
	return t.WriteServer(pk.Marshal(0x02, pk.VarInt(target), pk.VarInt(1)))
}
