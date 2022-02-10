package aura

import (
	"time"

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

func (k *KillAura) GetInterval() time.Duration {
	return 50 * time.Millisecond
}
