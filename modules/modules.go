package modules

import (
	"github.com/ruscalworld/vimeinterceptor/generic"
)

const (
	ModuleFlight        = "Flight"
	ModuleAntiKnockback = "AntiKnockback"
	ModuleNoFall        = "NoFall"
	ModuleKillAura      = "KillAura"
	ModuleSpammer       = "Spammer"
	ModuleMobAura       = "MobAura"
)

type DefaultModule struct {
	Tunnel  generic.Tunnel
	Enabled bool
}

func (m *DefaultModule) Register(conn generic.Tunnel) {
	m.Tunnel = conn
}

func (m *DefaultModule) IsEnabled() bool {
	return m.Enabled
}

func (m *DefaultModule) Close() {
	// Do nothing by default
}

type SimpleModule struct {
	DefaultModule
}

func (s *SimpleModule) Toggle() (bool, error) {
	s.Enabled = !s.Enabled
	return s.Enabled, nil
}

type SimpleTickingModule struct {
	SimpleModule
	InterruptTicker chan bool
}

func (m *SimpleTickingModule) GetInterruptChannel() chan bool {
	return m.InterruptTicker
}

func (m *SimpleTickingModule) Register(tunnel generic.Tunnel) {
	m.SimpleModule.Register(tunnel)
	m.InterruptTicker = make(chan bool)
}

func (m *SimpleTickingModule) Close() {
	m.StopTicker()
	m.SimpleModule.Close()
}

func (m *SimpleTickingModule) StopTicker() {
	m.InterruptTicker <- true
}
