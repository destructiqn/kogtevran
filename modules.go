package main

import (
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"time"
)

const (
	ModuleFlight        = "flight"
	ModuleAntiKnockback = "antiKnockback"
	ModuleNoFall        = "noFall"
	ModuleKillAura      = "killAura"
	ModuleSpammer 		= "spammer"
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
	conn.RegisterModule(&Flight{})
	conn.RegisterModule(&AntiKnockback{})
	conn.RegisterModule(&NoFall{})
	conn.RegisterModule(&KillAura{})
	conn.RegisterModule(&Spammer{})
}

type Flight struct {
	DefaultModule
}

func (f *Flight) GetIdentifier() string {
	return ModuleFlight
}

func (f *Flight) Toggle() (bool, error) {
	flags := 0x04
	if f.Enabled {
		flags = 0
	}

	err := f.Conn.WriteClient(pk.Marshal(0x39, pk.Byte(flags), pk.Float(0.05), pk.Float(0.1)))
	if err != nil {
		return f.Enabled, err
	}

	f.Enabled = !f.Enabled
	return f.Enabled, nil
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

type Spammer struct {
	SimpleModule
}

func (s *Spammer) Tick() error {
	return s.Conn.WriteServer(pk.Marshal(0x01, pk.String("Я курочка")))	
}

func (s *Spammer) GetInterval() time.Duration {
	return 20 * time.Second
}

func (s *Spammer) GetIdentifier() string {
	return ModuleSpammer
}