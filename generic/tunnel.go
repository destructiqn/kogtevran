package generic

import (
	"github.com/Tnze/go-mc/chat"
	"github.com/destructiqn/kogtevran/minecraft"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
)

type Tunnel interface {
	SetState(state protocol.ConnectionState)
	WriteClient(packet pk.Packet) error
	WriteServer(packet pk.Packet) error
	GetInventoryHandler() InventoryHandler
	GetTexteriaHandler() TexteriaHandler
	GetPlayerHandler() PlayerHandler
	GetEntityHandler() EntityHandler
	GetModuleHandler() ModuleHandler
	GetChatHandler() ChatHandler
	Disconnect(message chat.Message)
	GetRemoteAddr() string
	Close()
}

type PlayerHandler interface {
	IsFlying() bool
	IsOnGround() bool
	SetFlying(isFlying bool)
	GetLocation() *minecraft.Location
	GetEntityID() int32
	GetHealth() float64
	GetPlayerName() string
	Attack(target int) error
	ChangeSlot(slot int) error
	GetCurrentSlot() int
}

type EntityHandler interface {
	InitPlayer(entityID int, player *minecraft.Player)
	InitMob(entityID int, mob *minecraft.Mob)
	EntityRelativeMove(entityID int, dx, dy, dz float64)
	EntityTeleport(entityID int, x, y, z, yaw, pitch float64)
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
	Click(slot int, mode, button byte) error
	Move(from, to int) error
}

type InventoryHandler interface {
	GetWindows() []Window
	GetWindow(id int) (Window, bool)
	OpenWindow(id int, window Window)
	CloseWindow(id int)
	Reset()
}

type TexteriaHandler interface {
	UpdateInterface() error
	SendClient(data ...map[string]interface{}) error
}
