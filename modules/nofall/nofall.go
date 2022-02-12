package nofall

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

type NoFall struct {
	modules.SimpleModule
}

func (n *NoFall) GetIdentifier() string {
	return modules.ModuleNoFall
}

func HandlePlayer(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	player := packet.(*protocol.Player)
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		player.OnGround = true
	}
	return player.Marshal(), true, nil
}

func HandlePlayerPosition(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerPosition := packet.(*protocol.PlayerPosition)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		playerPosition.OnGround = true
	}

	return playerPosition.Marshal(), true, nil
}

func HandlePlayerLook(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerLook := packet.(*protocol.PlayerLook)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		playerLook.OnGround = true
	}

	return playerLook.Marshal(), true, nil
}

func HandleServerPlayerPositionAndLook(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerPosition := packet.(*protocol.ServerPlayerPositionAndLook)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		playerPosition.OnGround = true
	}

	return playerPosition.Marshal(), true, nil
}
