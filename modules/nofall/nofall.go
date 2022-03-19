package nofall

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

type NoFall struct {
	modules.SimpleModule
}

func (n *NoFall) GetDescription() []string {
	return []string{"Не получай урон от падения"}
}

func (n *NoFall) GetIdentifier() string {
	return modules.ModuleNoFall
}

func HandlePlayer(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	player := packet.(*protocol.Player)
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		player.OnGround = true
		return generic.ModifyPacket(player.Marshal()), nil
	}
	return generic.PassPacket(), nil
}

func HandlePlayerPosition(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	playerPosition := packet.(*protocol.PlayerPosition)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		playerPosition.OnGround = true
		return generic.ModifyPacket(playerPosition.Marshal()), nil
	}

	return generic.PassPacket(), nil
}

func HandlePlayerLook(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	playerLook := packet.(*protocol.PlayerLook)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		playerLook.OnGround = true
		return generic.ModifyPacket(playerLook.Marshal()), nil
	}

	return generic.PassPacket(), nil
}

func HandleServerPlayerPositionAndLook(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	playerPosition := packet.(*protocol.ServerPlayerPositionAndLook)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		playerPosition.OnGround = true
		return generic.ModifyPacket(playerPosition.Marshal()), nil
	}

	return generic.PassPacket(), nil
}
