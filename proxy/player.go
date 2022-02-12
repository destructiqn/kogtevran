package proxy

import (
	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/minecraft"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
	"time"
)

type PlayerHandler struct {
	tunnel   *MinecraftTunnel
	entityID int32
	isFlying bool
	location *minecraft.Location
}

func NewPlayerHandler(tunnel *MinecraftTunnel) *PlayerHandler {
	return &PlayerHandler{
		tunnel:   tunnel,
		location: &minecraft.Location{},
	}
}

func (p *PlayerHandler) GetEntityID() int32 {
	return p.entityID
}

func (p *PlayerHandler) IsFlying() bool {
	return p.isFlying
}

func (p *PlayerHandler) SetFlying(isFlying bool) {
	p.isFlying = isFlying
}

func (p *PlayerHandler) GetLocation() *minecraft.Location {
	return p.location
}

func (p *PlayerHandler) Attack(target int) error {
	return p.tunnel.WriteServer(pk.Marshal(0x02, pk.VarInt(target), pk.VarInt(1)))
}

func HandleJoinGame(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	joinGame := packet.(*protocol.JoinGame)
	tunnel.GetPlayerHandler().(*PlayerHandler).entityID = int32(joinGame.EntityID)
	tunnel.GetEntityHandler().(*EntityHandler).ResetEntities()

	go func() {
		time.Sleep(time.Second)
		_ = tunnel.GetModuleHandler().(*ModuleHandler).UpdateModuleList()
	}()

	return packet.Marshal(), true, nil
}

func HandlePlayerPositionAndLook(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerPositionAndLook := packet.(*protocol.PlayerPositionAndLook)
	flags := playerPositionAndLook.Flags
	x, y, z := playerPositionAndLook.X, playerPositionAndLook.Y, playerPositionAndLook.Z
	yaw, pitch := playerPositionAndLook.Yaw, playerPositionAndLook.Pitch
	playerHandler := tunnel.GetPlayerHandler().(*PlayerHandler)

	if flags&0x01 > 0 {
		playerHandler.location.X += float64(x)
	} else {
		playerHandler.location.X = float64(x)
	}

	if flags&0x02 > 0 {
		playerHandler.location.Y += float64(y)
	} else {
		playerHandler.location.Y = float64(y)
	}

	if flags&0x04 > 0 {
		playerHandler.location.Z += float64(z)
	} else {
		playerHandler.location.Z = float64(z)
	}

	playerHandler.location.Yaw, playerHandler.location.Pitch = byte(yaw), byte(pitch)
	return packet.Marshal(), true, nil
}

func HandlePlayerPosition(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerPosition := packet.(*protocol.PlayerPosition)
	location := tunnel.GetPlayerHandler().GetLocation()
	location.X, location.Y, location.Z = float64(playerPosition.X), float64(playerPosition.Y), float64(playerPosition.Z)
	return playerPosition.Marshal(), true, nil
}

func HandleServerPlayerPositionAndLook(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerPosition := packet.(*protocol.ServerPlayerPositionAndLook)
	location := tunnel.GetPlayerHandler().GetLocation()
	location.X, location.Y, location.Z = float64(playerPosition.X), float64(playerPosition.Y), float64(playerPosition.Z)
	location.Yaw, location.Pitch = byte(playerPosition.Yaw), byte(playerPosition.Pitch)
	return playerPosition.Marshal(), true, nil
}

func HandleServerPlayerAbilities(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerAbilities := packet.(*protocol.ServerPlayerAbilities)
	tunnel.GetPlayerHandler().SetFlying(playerAbilities.Flags&0x02 > 0)
	return packet.Marshal(), true, nil
}
