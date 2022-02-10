package aura

import (
	"github.com/ruscalworld/vimeinterceptor/minecraft"
	"github.com/ruscalworld/vimeinterceptor/modules"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type Predicate func(mob minecraft.Entity) bool

type GenericAura struct {
	modules.SimpleTickingModule
	MaxDistance  float64
	HitAnimation bool
}

func (a *GenericAura) Tick() error {
	a.Tunnel.GetEntityHandler().Lock()
	location := a.Tunnel.GetPlayerHandler().GetLocation()
	for id, entity := range a.Tunnel.GetEntityHandler().GetEntities() {
		if !a.Filter(entity) || entity.GetLocation().Distance(location) > a.MaxDistance {
			continue
		}

		err := a.Tunnel.GetPlayerHandler().Attack(id)
		if err != nil {
			return err
		}

		if a.HitAnimation {
			_ = a.Tunnel.WriteClient(pk.Marshal(protocol.ClientboundAnimation, pk.VarInt(a.Tunnel.GetPlayerHandler().GetEntityID()), pk.UnsignedByte(0x00)))
		}
	}

	a.Tunnel.GetEntityHandler().Unlock()
	return nil
}

func (a *GenericAura) Filter(_ minecraft.Entity) bool {
	return true
}
