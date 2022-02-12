package aura

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft"
	"github.com/destructiqn/kogtevran/modules"
)

type MobAura struct {
	GenericAura
}

func (m *MobAura) Register(tunnel generic.Tunnel) {
	m.GenericAura.Filter = m.Predicate
	m.GenericAura.Register(tunnel)
}

func (m *MobAura) GetIdentifier() string {
	return modules.ModuleMobAura
}

func (m *MobAura) Predicate(entity minecraft.Entity) bool {
	_, ok := entity.(*minecraft.Player)
	return !ok
}
