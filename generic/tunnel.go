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
	GetInventoryHandler() InventoryHandler
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

type Window interface {
	GetType() string
	GetSize() int
	GetTitle() chat.Message
	GetContents() map[int]pk.Slot
	GetItem(slot int) pk.Slot
	PutItem(slot int, item pk.Slot)
}

type InventoryHandler interface {
	GetWindows() []Window
	GetWindow(id int) (Window, bool)
	OpenWindow(id int, window Window)
	CloseWindow(id int)
	Reset()
}
