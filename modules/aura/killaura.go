package aura

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft"
	"github.com/destructiqn/kogtevran/modules"
)

type KillAura struct {
	GenericAura
}

func (k *KillAura) Register(tunnel generic.Tunnel) {
	k.GenericAura.Filter = k.Predicate
	k.GenericAura.Register(tunnel)
}

func (k *KillAura) GetIdentifier() string {
	return modules.ModuleKillAura
}

func (k *KillAura) Predicate(entity minecraft.Entity) bool {
	_, ok := entity.(*minecraft.Player)
	return ok
}
