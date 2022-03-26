package speedhack

import (
	"github.com/destructiqn/kogtevran/generic"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

type SpeedHack struct {
	modules.SimpleModule
	Speed float64 `option:"speed"`
}

func (s *SpeedHack) GetIdentifier() string {
	return modules.ModuleSpeedHack
}

func (s *SpeedHack) GetDescription() []string {
	return []string{
		"Увеличивает скорость передвижения",
		"",
		"§nПараметры",
		"§7speed§f - множитель скорости",
	}
}

func HandleEntityProperties(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	entityProperties := packet.(*protocol.EntityProperties)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleSpeedHack) {
		module, _ := tunnel.GetModuleHandler().GetModule(modules.ModuleSpeedHack)

		modified := false
		fakeSpeed := pk.Property{
			Key:   "generic.movementSpeed",
			Value: 0.699999988079071 * pk.Double(module.(*SpeedHack).Speed) / 2,
		}

		for i, property := range entityProperties.Properties {
			if property.Key == "generic.movementSpeed" {
				entityProperties.Properties[i] = fakeSpeed
				modified = true
				break
			}
		}

		if !modified {
			entityProperties.Properties = append(entityProperties.Properties, fakeSpeed)
		}

		return generic.ModifyPacket(entityProperties.Marshal()), nil
	}

	return generic.PassPacket(), nil
}
