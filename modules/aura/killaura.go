package aura

import (
	"github.com/ruscalworld/vimeinterceptor/minecraft"
	"github.com/ruscalworld/vimeinterceptor/modules"
)

type KillAura struct {
	GenericAura
}

func (k *KillAura) GetIdentifier() string {
	return modules.ModuleKillAura
}

func (k *KillAura) Filter(entity minecraft.Entity) bool {
	_, ok := entity.(*minecraft.Player)
	return ok
}
