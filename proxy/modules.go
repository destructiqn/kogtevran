package proxy

import (
	"fmt"
	"log"
	"time"

	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/modules"
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
			for {
				select {
				case <-time.NewTimer(tickingModule.GetInterval()).C:
					if !module.IsEnabled() {
						continue
					}

					err := module.Tick()
					if err != nil {
						log.Println("error ticking", module.GetIdentifier(), err)
					}
				case <-tickingModule.GetInterruptChannel():
					return
				}
			}
		}(tickingModule)
	}
}

func (t *ModuleHandler) GetModules() []generic.Module {
	aModules := make([]generic.Module, 0)
	for _, module := range t.modules {
		aModules = append(aModules, module)
	}
	return aModules
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

func (t *ModuleHandler) ToggleModule(module generic.Module) (bool, error) {
	value, err := module.Toggle()
	if err != nil {
		return value, err
	}

	err = t.UpdateModuleList()
	if err != nil {
		return value, err
	}

	err = module.Update()
	if err != nil {
		return value, err
	}

	return value, nil
}

func (t *ModuleHandler) GetModulesDetails() []map[string]interface{} {
	modulesDisplay := make([]string, 0)
	modulesControls := make([]map[string]interface{}, 0)

	x, y := 10, 20
	for _, module := range t.GetModules() {
		control := map[string]interface{}{
			"id":    fmt.Sprintf("kv.mc.%s", module.GetIdentifier()),
			"type":  "Button",
			"pos":   "BOTTOM_RIGHT",
			"w":     85,
			"h":     20,
			"t":     module.GetIdentifier(),
			"tc":    0xFFFFFF,
			"x":     x,
			"y":     y,
			"color": -0x22BBBB,
			"click": map[string]interface{}{
				"act":  "CHAT",
				"data": fmt.Sprintf("/toggle %s", module.GetIdentifier()),
			},
		}

		x += 90
		if x > 300 {
			y += 25
			x = 10
		}

		modulesControls = append(modulesControls, control)

		if module.IsEnabled() {
			modulesDisplay = append(modulesDisplay, module.GetIdentifier())
			control["color"] = -0xBB22BB
		}
	}

	return []map[string]interface{}{
		{
			"%":    "add",
			"id":   "kv.ml",
			"al":   "RIGHT",
			"pos":  "TOP_RIGHT",
			"type": "Text",
			"text": modulesDisplay,

			"vis": []map[string]interface{}{
				{
					"type": "always",
					"show": true,
				},
				{
					"type": "f3",
					"show": false,
				},
			},

			"x": 2,
			"y": 2,
		},
		{
			"%": "add:group",
			"e": modulesControls,
			"vis": []map[string]interface{}{
				{
					"type": "chat",
					"show": true,
				},
			},
		},
	}
}

func (t *ModuleHandler) UpdateModuleList() error {
	details := t.GetModulesDetails()
	return t.tunnel.TexteriaHandler.SendClient(details)
}

func RegisterDefaultModules(tunnel generic.Tunnel) {
	moduleHandler := tunnel.GetModuleHandler()

	genericAura := aura.GenericAura{
		MaxDistance: 7, HitAnimation: true,
		SimpleTickingModule: modules.SimpleTickingModule{Interval: 50 * time.Millisecond},
	}

	moduleHandler.RegisterModule(&flight.Flight{Speed: 1})
	moduleHandler.RegisterModule(&antiknockback.AntiKnockback{})
	moduleHandler.RegisterModule(&nofall.NoFall{})
	moduleHandler.RegisterModule(&aura.KillAura{GenericAura: genericAura})
	moduleHandler.RegisterModule(&aura.MobAura{GenericAura: genericAura})
	moduleHandler.RegisterModule(&spammer.Spammer{SimpleTickingModule: modules.SimpleTickingModule{Interval: 20 * time.Second}})
}
