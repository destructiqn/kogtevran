package nofall

import (
	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/modules"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
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

func HandleServerPlayerPositionAndLook(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	playerPosition := packet.(*protocol.ServerPlayerPositionAndLook)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoFall) {
		playerPosition.OnGround = true
	}

	return playerPosition.Marshal(), true, nil
}
