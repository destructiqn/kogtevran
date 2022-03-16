package aura

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft"
	"github.com/destructiqn/kogtevran/modules"
)

type MobAura struct {
	GenericAura
}

func (m *MobAura) GetDescription() []string {
	return []string{
		"Убивай всех мобов в радиусе 6 блоков",
		"",
		"§nПараметры",
		"§7maxDistance§f - максимальное расстояние до цели",
		"§7hitAnimation§f - показывать анимацию удара",
	}
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
