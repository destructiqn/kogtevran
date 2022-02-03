package main

import (
	"time"

	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

const (
	ModuleFlight        = "flight"
	ModuleAntiKnockback = "antiKnockback"
	ModuleNoFall        = "noFall"
	ModuleKillAura      = "killAura"
)

type Module interface {
	Register(conn *WrappedConn)
	GetIdentifier() string
	Toggle() (bool, error)
	IsEnabled() bool
}

type TickingModule interface {
	Module
	Tick() error
	GetInterval() time.Duration
}

type DefaultModule struct {
	Conn    *WrappedConn
	Enabled bool
}

type SimpleModule struct {
	DefaultModule
}

func (s *SimpleModule) Toggle() (bool, error) {
	s.Enabled = !s.Enabled
	return s.Enabled, nil
}

func (m *DefaultModule) Register(conn *WrappedConn) {
	m.Conn = conn
}

func (m *DefaultModule) IsEnabled() bool {
	return m.Enabled
}

func RegisterDefaultModules(conn *WrappedConn) {
	conn.RegisterModule(&Flight{Speed: 1})
	conn.RegisterModule(&AntiKnockback{})
	conn.RegisterModule(&NoFall{})
	conn.RegisterModule(&KillAura{})
}

type Flight struct {
	DefaultModule
	Speed float64
}

func (f *Flight) GetIdentifier() string {
	return ModuleFlight
}

func (f *Flight) Toggle() (bool, error) {
	f.Enabled = !f.Enabled
	err := f.Update()
	if err != nil {
		return f.Enabled, err
	}
	return f.Enabled, nil
}

func (f *Flight) Update() error {
	flags := 0
	if f.Enabled {
		flags = 0x04
	}

	if f.Conn.IsFlying {
		flags |= 0x02
	}

	return f.Conn.WriteClient(pk.Marshal(0x39, pk.Byte(flags), pk.Float(0.05*f.Speed), pk.Float(0.1)))
}

type AntiKnockback struct {
	SimpleModule
}

func (a *AntiKnockback) GetIdentifier() string {
	return ModuleAntiKnockback
}

type NoFall struct {
	SimpleModule
}

func (s *SimpleModule) GetIdentifier() string {
	return ModuleNoFall
}

type KillAura struct {
	SimpleModule
}

func (k *KillAura) GetIdentifier() string {
	return ModuleKillAura
}

func (k *KillAura) Tick() error {
	k.Conn.EntitiesMutex.Lock()
	for id, entity := range k.Conn.Entities {
		if entity.Location.Distance(k.Conn.Location) > 5 {
			continue
		}

		err := k.Conn.Attack(id)
		if err != nil {
			return err
		}
	}

	k.Conn.EntitiesMutex.Unlock()
	return nil
}

func (k *KillAura) GetInterval() time.Duration {
	return 50 * time.Millisecond
}
