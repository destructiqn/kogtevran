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

type Entity struct {
	Location *Location
	isPlayer bool
}

type Location struct {
	X, Y, Z    float64
	Yaw, Pitch byte
}

func (l *Location) Distance(another *Location) float64 {
	dx, dy, dz := another.X-l.X, another.Y-l.Y, another.Z-l.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

type WrappedConn struct {
	Closed  bool
	Server  *net.Conn
	Client  *net.Conn
	Modules map[string]Module

	ServerWrite sync.Mutex
	ClientWrite sync.Mutex

	State            ConnectionState
	EnableEncryption chan []byte

	EntityID      int32
	Location      *Location
	IsFlying      bool
	Entities      map[int32]*Entity
	EntitiesMutex sync.Mutex
}

func (w *WrappedConn) initEntity(entityID int32) {
	w.EntitiesMutex.Lock()
	w.Entities[entityID] = &Entity{&Location{}, false}
	w.EntitiesMutex.Unlock()
}

func (w *WrappedConn) initPositionedEntity(entityID int32, x, y, z float64, yaw, pitch byte, isPlayer bool) {
	w.EntitiesMutex.Lock()
	w.Entities[entityID] = &Entity{
		Location: &Location{x, y, z, yaw, pitch},
		isPlayer: isPlayer,
	}
	w.EntitiesMutex.Unlock()
}

func (w *WrappedConn) entityRelativeMove(entityID int32, dx, dy, dz float64) {
	entity, ok := w.Entities[entityID]
	if !ok {
		return
	}

	entity.Location.X += dx
	entity.Location.Y += dy
	entity.Location.Z += dz
}

func (w *WrappedConn) entityTeleport(entityID int32, x, y, z float64, yaw, pitch byte) {
	entity, ok := w.Entities[entityID]
	if !ok {
		return
	}

	entity.Location.X, entity.Location.Y, entity.Location.Z = x, y, z
	entity.Location.Yaw, entity.Location.Pitch = yaw, pitch
}

func (w *WrappedConn) resetEntities() {
	w.EntitiesMutex.Lock()
	for id := range w.Entities {
		delete(w.Entities, id)
	}
	w.EntitiesMutex.Unlock()
}

func (w *WrappedConn) destroyEntities(entityIDs []int32) {
	w.EntitiesMutex.Lock()
	for _, id := range entityIDs {
		delete(w.Entities, id)
	}
	w.EntitiesMutex.Unlock()
}

func (w *WrappedConn) IsModuleEnabled(moduleID string) bool {
	module, ok := w.Modules[moduleID]
	if !ok {
		return false
	}

	return module.IsEnabled()
}

func (w *WrappedConn) WriteClient(packet pk.Packet) error {
	w.ClientWrite.Lock()
	err := w.Client.WritePacket(packet)
	w.ClientWrite.Unlock()
	return err
}

func (w *WrappedConn) WriteServer(packet pk.Packet) error {
	w.ServerWrite.Lock()
	err := w.Server.WritePacket(packet)
	w.ServerWrite.Unlock()
	return err
}

func (w *WrappedConn) Disconnect() {
	w.Closed = true
	_ = w.Server.Close()
	_ = w.Client.Close()
}

func WrapConn(server, client *net.Conn) *WrappedConn {
	return &WrappedConn{
		Server:           server,
		Client:           client,
		Modules:          make(map[string]Module),
		Entities:         make(map[int32]*Entity),
		Location:         &Location{},
		EnableEncryption: make(chan []byte),
	}
}

type ChatPosition byte

var (
	ChatPositionChat          ChatPosition = 0
	ChatPositionSystemMessage ChatPosition = 1
	ChatPositionAboveHotbar   ChatPosition = 2
)

func (w *WrappedConn) RegisterModule(module Module) {
	module.Register(w)
	w.Modules[module.GetIdentifier()] = module

	tickingModule, isTicking := module.(TickingModule)
	if isTicking {
		go func(module TickingModule) {
			ticker := time.NewTicker(module.GetInterval())
			for {
				_ = <-ticker.C
				if !module.IsEnabled() {
					continue
				}

				err := module.Tick()
				if err != nil {
					log.Println("error ticking", module.GetIdentifier(), err)
				}
			}
		}(tickingModule)
	}
}

func (w *WrappedConn) SendMessage(message chat.Message, position ChatPosition) error {
	return w.WriteClient(pk.Marshal(0x02, message, pk.Byte(position)))
}

func (w *WrappedConn) Attack(target int32) error {
	return w.WriteServer(pk.Marshal(0x02, pk.VarInt(target), pk.VarInt(1)))
}
