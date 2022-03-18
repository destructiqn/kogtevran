package aura

import (
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft"
	"github.com/destructiqn/kogtevran/modules"
)

type KillAura struct {
	GenericAura
}

func (k *KillAura) GetDescription() []string {
	return []string{
		"Убивай всех игроков в радиусе 6 блоков",
		"",
		"§nПараметры",
		"§7maxDistance§f - максимальное расстояние до цели",
		"§7hitAnimation§f - показывать анимацию удара",
	}
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
