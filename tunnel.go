package main

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/net"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

type ConnectionState int

const (
	ConnStateHandshake ConnectionState = 0
	ConnStateLogin     ConnectionState = 1
	ConnStatePlay      ConnectionState = 2
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

type Location struct {
	X, Y, Z    float64
	Yaw, Pitch byte
}

func (l *Location) Distance(another *Location) float64 {
	dx, dy, dz := another.X-l.X, another.Y-l.Y, another.Z-l.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

type MinecraftTunnel struct {
	SessionID                 string
	AuxiliaryChannel          *AuxiliaryChannel
	AuxiliaryChannelAvailable chan bool

	Closed  bool
	Server  *net.Conn
	Client  *net.Conn
	Modules map[string]Module

	ServerWrite sync.Mutex
	ClientWrite sync.Mutex

	State                  ConnectionState
	VerifyToken            []byte
	KeyBruteMode           bool
	SharedSecretCandidates [][]byte
	EnableEncryptionS2C    chan []byte
	EnableEncryptionC2S    chan [][]byte

	EntityID      int32
	Location      *Location
	IsFlying      bool
	Entities      map[int]Entity
	EntitiesMutex sync.Mutex
}

func (t *MinecraftTunnel) initPlayer(entityID int, player *Player) {
	t.EntitiesMutex.Lock()
	t.Entities[entityID] = player
	t.EntitiesMutex.Unlock()
}

func (t *MinecraftTunnel) initMob(entityID int, mob *Mob) {
	t.EntitiesMutex.Lock()
	t.Entities[entityID] = mob
	t.EntitiesMutex.Unlock()
}

func (t *MinecraftTunnel) entityRelativeMove(entityID int, dx, dy, dz float64) {
	entity, ok := t.Entities[entityID]
	if !ok {
		return
	}

	entity.GetLocation().X += dx
	entity.GetLocation().Y += dy
	entity.GetLocation().Z += dz
}

func (t *MinecraftTunnel) entityTeleport(entityID int, x, y, z float64, yaw, pitch byte) {
	entity, ok := t.Entities[entityID]
	if !ok {
		return
	}

	entity.GetLocation().X, entity.GetLocation().Y, entity.GetLocation().Z = x, y, z
	entity.GetLocation().Yaw, entity.GetLocation().Pitch = yaw, pitch
}

func (t *MinecraftTunnel) resetEntities() {
	t.EntitiesMutex.Lock()
	for id := range t.Entities {
		delete(t.Entities, id)
	}
	t.EntitiesMutex.Unlock()
}

func (t *MinecraftTunnel) destroyEntities(entityIDs []int) {
	t.EntitiesMutex.Lock()
	for _, id := range entityIDs {
		delete(t.Entities, id)
	}
	t.EntitiesMutex.Unlock()
}

func (t *MinecraftTunnel) IsModuleEnabled(moduleID string) bool {
	module, ok := t.Modules[moduleID]
	if !ok {
		return false
	}

	return module.IsEnabled()
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
	for _, module := range t.Modules {
		module.Close()
	}

	t.Closed = true
	_ = t.Server.Close()
	_ = t.Client.Close()
	CurrentTunnelPool.UnregisterTunnel(t.SessionID)
}

func WrapConn(server, client *net.Conn) *MinecraftTunnel {
	return &MinecraftTunnel{
		Server:                    server,
		Client:                    client,
		Modules:                   make(map[string]Module),
		Entities:                  make(map[int]Entity),
		Location:                  &Location{},
		AuxiliaryChannelAvailable: make(chan bool),
		EnableEncryptionS2C:       make(chan []byte),
		EnableEncryptionC2S:       make(chan [][]byte),
	}
}

type ChatPosition byte

const (
	ChatPositionChat ChatPosition = iota
	ChatPositionSystemMessage
	ChatPositionAboveHotbar
)

func (t *MinecraftTunnel) RegisterModule(module Module) {
	module.Register(t)
	t.Modules[module.GetIdentifier()] = module

	tickingModule, isTicking := module.(TickingModule)
	if isTicking {
		go func(module TickingModule) {
			ticker := time.NewTicker(module.GetInterval())
			for {
				select {
				case <-ticker.C:
					if !module.IsEnabled() {
						continue
					}

					err := module.Tick()
					if err != nil {
						log.Println("error ticking", module.GetIdentifier(), err)
					}
				case <-tickingModule.GetInterruptChannel():
					break
				}
			}
		}(tickingModule)
	}
}

func (t *MinecraftTunnel) SendMessage(message chat.Message, position ChatPosition) error {
	return t.WriteClient(pk.Marshal(0x02, message, pk.Byte(position)))
}

func (t *MinecraftTunnel) Attack(target int) error {
	return t.WriteServer(pk.Marshal(0x02, pk.VarInt(target), pk.VarInt(1)))
}
