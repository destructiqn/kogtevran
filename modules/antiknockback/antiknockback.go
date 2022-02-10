package antiknockback

import (
	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/modules"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type AntiKnockback struct {
	modules.SimpleModule
	X, Y, Z int
}

func (a *AntiKnockback) GetIdentifier() string {
	return modules.ModuleAntiKnockback
}

func HandleEntityVelocity(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	entityVelocity := packet.(*protocol.EntityVelocity)
	moduleHandler := tunnel.GetModuleHandler()

	if moduleHandler.IsModuleEnabled(modules.ModuleAntiKnockback) && int32(entityVelocity.EntityID) == tunnel.GetPlayerHandler().GetEntityID() {
		module, _ := moduleHandler.GetModule(modules.ModuleAntiKnockback)
		antiKnockback := module.(*AntiKnockback)

		entityVelocity.VX = pk.Short(antiKnockback.X)
		entityVelocity.VY = pk.Short(antiKnockback.Y)
		entityVelocity.VZ = pk.Short(antiKnockback.Z)
	}

	return entityVelocity.Marshal(), true, nil
}
