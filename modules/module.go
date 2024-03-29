package modules

import (
	"time"

	"github.com/destructiqn/kogtevran/generic"
)

const (
	ModuleFlight        = "Flight"
	ModuleAntiKnockback = "AntiKnockback"
	ModuleNoFall        = "NoFall"
	ModuleKillAura      = "KillAura"
	ModuleSpammer       = "Spammer"
	ModuleMobAura       = "MobAura"
	ModuleChestStealer  = "ChestStealer"
	ModuleCMDCam        = "CMDCam"
	ModuleLongJump      = "LongJump"
	ModuleUnlimitedCPS  = "UnlimitedCPS"
	ModuleTPAura        = "TPAura"
	ModulePlayerESP     = "PlayerESP"
	ModuleChestESP      = "ChestESP"
	ModuleNuker         = "Nuker"
	ModuleFastBreak     = "FastBreak"
	ModuleNoBadEffects  = "NoBadEffects"
	ModuleSpeedHack     = "SpeedHack"
	ModuleAutoSoup      = "AutoSoup"
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

func (m *DefaultModule) Update() error {
	// Do nothing by default
	return nil
}

func (m *DefaultModule) Close() {
	// Do nothing by default
}

func (m *DefaultModule) SetEnabled(enabled bool) {
	m.Enabled = enabled
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
	Interval        time.Duration `option:"interval"`
	InterruptTicker chan bool
}

func (m *SimpleTickingModule) GetInterval() time.Duration {
	return m.Interval
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

type ClientModule struct {
	SimpleModule
	Identifier  string
	Description []string
}

func (c *ClientModule) Close() {
	if c.IsEnabled() {
		_, _ = c.Tunnel.GetModuleHandler().ToggleModule(c)
		c.Enabled = false
	}

	c.SimpleModule.Close()
}

func (c *ClientModule) GetIdentifier() string {
	return c.Identifier
}

func (c *ClientModule) GetDescription() []string {
	return c.Description
}
