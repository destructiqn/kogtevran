package antiknockback

import (
	"github.com/destructiqn/kogtevran/generic"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

type AntiKnockback struct {
	modules.SimpleModule
	X, Y, Z int
}

func (a *AntiKnockback) GetDescription() []string {
	return []string{"Теперь тебя будет невозможно откинуть"}
}

func (a *AntiKnockback) GetIdentifier() string {
	return modules.ModuleAntiKnockback
}

func HandleEntityVelocity(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	entityVelocity := packet.(*protocol.EntityVelocity)
	moduleHandler := tunnel.GetModuleHandler()

	if moduleHandler.IsModuleEnabled(modules.ModuleAntiKnockback) && int32(entityVelocity.EntityID) == tunnel.GetPlayerHandler().GetEntityID() {
		module, _ := moduleHandler.GetModule(modules.ModuleAntiKnockback)
		antiKnockback := module.(*AntiKnockback)

		entityVelocity.VX = pk.Short(antiKnockback.X)
		entityVelocity.VY = pk.Short(antiKnockback.Y)
		entityVelocity.VZ = pk.Short(antiKnockback.Z)
	}

	return generic.ModifyPacket(entityVelocity.Marshal()), nil
}
