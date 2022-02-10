package generic

import (
	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/minecraft"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type Tunnel interface {
	SetState(state protocol.ConnectionState)
	WriteClient(packet pk.Packet) error
	WriteServer(packet pk.Packet) error
	GetPlayerHandler() PlayerHandler
	GetEntityHandler() EntityHandler
	GetModuleHandler() ModuleHandler
	GetChatHandler() ChatHandler
	Close()
}

type PlayerHandler interface {
	IsFlying() bool
	SetFlying(isFlying bool)
	GetLocation() *minecraft.Location
	GetEntityID() int32
	Attack(target int) error
}

type EntityHandler interface {
	InitPlayer(entityID int, player *minecraft.Player)
	InitMob(entityID int, mob *minecraft.Mob)
	EntityRelativeMove(entityID int, dx, dy, dz float64)
	EntityTeleport(entityID int, x, y, z float64, yaw, pitch byte)
	ResetEntities()
	DestroyEntities(entityIDs []int)
	GetEntity(entityID int) (minecraft.Entity, bool)
	GetEntities() map[int]minecraft.Entity
	Lock()
	Unlock()
}

type ChatHandler interface {
	SendMessage(message chat.Message, position protocol.ChatPosition) error
}
