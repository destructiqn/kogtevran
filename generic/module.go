package generic

import "time"

type Module interface {
	Register(conn Tunnel)
	GetIdentifier() string
	Toggle() (bool, error)
	IsEnabled() bool
	Update() error
	Close()
}

type TickingModule interface {
	Module
	Tick() error
	GetInterval() time.Duration
	StopTicker()
	GetInterruptChannel() chan bool
}

type ModuleHandler interface {
	IsModuleEnabled(identifier string) bool
	GetModule(identifier string) (Module, bool)
	GetModules() []Module
	RegisterModule(module Module)
}
