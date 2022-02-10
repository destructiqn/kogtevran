package aura

import (
	"github.com/ruscalworld/vimeinterceptor/minecraft"
	"github.com/ruscalworld/vimeinterceptor/modules"
)

type MobAura struct {
	GenericAura
}

func (m *MobAura) GetIdentifier() string {
	return modules.ModuleMobAura
}

func (m *MobAura) Filter(entity minecraft.Entity) bool {
	_, ok := entity.(*minecraft.Player)
	return !ok
}
