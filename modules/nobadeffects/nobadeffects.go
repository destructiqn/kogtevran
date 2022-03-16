package nobadeffects

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

var BadEffects = map[pk.Byte]bool{
	2:  true,
	4:  true,
	7:  true,
	9:  true,
	15: true,
	17: true,
	18: true,
	19: true,
	20: true,
	25: true,
	27: true,
}

type NoBadEffects struct {
	modules.SimpleModule
}

func (n *NoBadEffects) GetIdentifier() string {
	return modules.ModuleNoBadEffects
}

func (n *NoBadEffects) Toggle() (bool, error) {
	value, err := n.SimpleModule.Toggle()
	if err != nil {
		return false, err
	}

	if value {
		entityID := n.Tunnel.GetPlayerHandler().GetEntityID()

		for id := range BadEffects {
			packet := &protocol.RemoveEntityEffect{
				EntityID: pk.VarInt(entityID),
				EffectID: id,
			}

			err := n.Tunnel.WriteClient(packet.Marshal())
			if err != nil {
				return false, err
			}
		}
	}

	return value, nil
}

func HandleEntityEffect(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	entityEffect := packet.(*protocol.EntityEffect)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNoBadEffects) {
		if v, ok := BadEffects[entityEffect.EffectID]; ok && v {
			return generic.RejectPacket(), nil
		}
	}

	return generic.PassPacket(), nil
}
