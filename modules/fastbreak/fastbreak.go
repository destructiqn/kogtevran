package fastbreak

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	"github.com/destructiqn/kogtevran/protocol"
)

type FastBreak struct {
	modules.SimpleModule
}

func (f *FastBreak) GetIdentifier() string {
	return modules.ModuleFastBreak
}

func HandlePlayerDigging(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	playerDigging := packet.(*protocol.PlayerDigging)
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleFastBreak) {
		if playerDigging.Status == 0 {
			err := tunnel.WriteServer(playerDigging.Marshal())
			if err != nil {
				return nil, err
			}

			playerDigging.Status = 2
			return generic.ModifyPacket(playerDigging.Marshal()), nil
		}
	}

	return generic.PassPacket(), nil
}
