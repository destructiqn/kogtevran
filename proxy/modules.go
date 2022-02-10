package proxy

import (
	"log"
	"time"

	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/modules/antiknockback"
	"github.com/ruscalworld/vimeinterceptor/modules/aura"
	"github.com/ruscalworld/vimeinterceptor/modules/flight"
	"github.com/ruscalworld/vimeinterceptor/modules/nofall"
	"github.com/ruscalworld/vimeinterceptor/modules/spammer"
)

type ModuleHandler struct {
	tunnel  *MinecraftTunnel
	modules map[string]generic.Module
}

func NewModuleHandler(tunnel *MinecraftTunnel) *ModuleHandler {
	return &ModuleHandler{tunnel, make(map[string]generic.Module)}
}

func (t *ModuleHandler) RegisterModule(module generic.Module) {
	module.Register(t.tunnel)
	t.modules[module.GetIdentifier()] = module

	tickingModule, isTicking := module.(generic.TickingModule)
	if isTicking {
		go func(module generic.TickingModule) {
			ticker := time.NewTicker(module.GetInterval())
			for {
				select {
				case <-ticker.C:
					if !module.IsEnabled() {
						continue
					}

					err := module.Tick()
					if err != nil {
						log.Println("error ticking", module.GetIdentifier(), err)
					}
				case <-tickingModule.GetInterruptChannel():
					break
				}
			}
		}(tickingModule)
	}
}

func (t *ModuleHandler) GetModules() []generic.Module {
	modules := make([]generic.Module, 0)
	for _, module := range t.modules {
		modules = append(modules, module)
	}
	return modules
}

func (t *ModuleHandler) GetModule(identifier string) (generic.Module, bool) {
	module, ok := t.modules[identifier]
	return module, ok
}

func (t *ModuleHandler) IsModuleEnabled(moduleID string) bool {
	module, ok := t.modules[moduleID]
	if !ok {
		return false
	}

	return module.IsEnabled()
}

func RegisterDefaultModules(tunnel generic.Tunnel) {
	moduleHandler := tunnel.GetModuleHandler()
	genericAura := aura.GenericAura{MaxDistance: 7, HitAnimation: true}

	moduleHandler.RegisterModule(&flight.Flight{Speed: 1})
	moduleHandler.RegisterModule(&antiknockback.AntiKnockback{})
	moduleHandler.RegisterModule(&nofall.NoFall{})
	moduleHandler.RegisterModule(&aura.KillAura{GenericAura: genericAura})
	moduleHandler.RegisterModule(&aura.MobAura{GenericAura: genericAura})
	moduleHandler.RegisterModule(&spammer.Spammer{})
}
