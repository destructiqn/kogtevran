package proxy

import (
	"fmt"
	"github.com/destructiqn/kogtevran/license"
	"github.com/destructiqn/kogtevran/modules/fastbreak"
	"github.com/destructiqn/kogtevran/modules/longjump"
	"github.com/destructiqn/kogtevran/modules/nobadeffects"
	"github.com/destructiqn/kogtevran/modules/nuker"
	"github.com/destructiqn/kogtevran/modules/unlimitedcps"
	"log"
	"sort"
	"time"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"

	"github.com/destructiqn/kogtevran/modules/antiknockback"
	"github.com/destructiqn/kogtevran/modules/aura"
	"github.com/destructiqn/kogtevran/modules/cmdcam"
	"github.com/destructiqn/kogtevran/modules/flight"
	"github.com/destructiqn/kogtevran/modules/nofall"
	"github.com/destructiqn/kogtevran/modules/spammer"
	"github.com/destructiqn/kogtevran/modules/tpaura"
)

type ModuleHandler struct {
	tunnel  *MinecraftTunnel
	modules map[string]generic.Module
}

func NewModuleHandler(tunnel *MinecraftTunnel) *ModuleHandler {
	return &ModuleHandler{tunnel, make(map[string]generic.Module)}
}

func (m *ModuleHandler) RegisterModule(module generic.Module) {
	module.Register(m.tunnel)
	m.modules[module.GetIdentifier()] = module

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

type ModuleList []generic.Module

func (m ModuleList) Len() int {
	return len([]generic.Module(m))
}

func (m ModuleList) Less(i, j int) bool {
	return m[i].GetIdentifier() < m[j].GetIdentifier()
}

func (m ModuleList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m *ModuleHandler) GetModules() []generic.Module {
	aModules := make([]generic.Module, 0)
	for _, module := range m.modules {
		aModules = append(aModules, module)
	}
	return aModules
}

func (m *ModuleHandler) GetModule(identifier string) (generic.Module, bool) {
	module, ok := m.modules[identifier]
	return module, ok
}

func (m *ModuleHandler) IsModuleEnabled(moduleID string) bool {
	module, ok := m.modules[moduleID]
	if !ok {
		return false
	}

	return module.IsEnabled()
}

func (m *ModuleHandler) ToggleModule(module generic.Module) (bool, error) {
	defer UpdateModuleMetrics()

	if _, ok := module.(*modules.ClientModule); !ok {
		value, err := module.Toggle()
		if err != nil {
			return value, err
		}

		return module.IsEnabled(), m.UpdateModule(module)
	} else {
		auxiliary := m.tunnel.TunnelPair.Auxiliary
		err := auxiliary.SendMessage(ModuleToggle, AuxiliaryToggleModule{module.GetIdentifier()})
		if err != nil {
			return module.IsEnabled(), err
		}
	}

	return module.IsEnabled(), nil
}

func (m *ModuleHandler) UpdateModule(module generic.Module) error {
	err := m.tunnel.GetTexteriaHandler().UpdateInterface()
	if err != nil {
		return err
	}

	return module.Update()
}

func (m *ModuleHandler) GetModulesDetails() []map[string]interface{} {
	y := 25
	modulesDisplay := make([]string, 0)
	modulesControls := make([]map[string]interface{}, 0)

	categories := modules.GetCategoryList()
	sort.Sort(categories)

	for _, category := range categories {
		x := 10
		modulesList := make(ModuleList, 0)

		for _, moduleID := range category.ModuleIDs {
			if module, ok := m.GetModule(moduleID); ok {
				modulesList = append(modulesList, module)
			}
		}

		sort.Sort(modulesList)

		modulesControls = append(modulesControls, map[string]interface{}{
			"id":     fmt.Sprintf("kv.cc.%s.bg", category),
			"pos":    "BOTTOM_RIGHT",
			"type":   "Rectangle",
			"width":  len(modulesList)*90 + 8,
			"height": 48,
			"color":  -0x80000000,
			"x":      4,
			"y":      y - 6,
		})

		for _, module := range modulesList {
			control := map[string]interface{}{
				"id":    fmt.Sprintf("kv.mc.%s", module.GetIdentifier()),
				"type":  "Button",
				"pos":   "BOTTOM_RIGHT",
				"w":     85,
				"h":     20,
				"t":     module.GetIdentifier(),
				"tc":    0xAAAAAA,
				"hc":    0x5555FF,
				"x":     x,
				"y":     y,
				"color": -0x33000000,
				"click": map[string]interface{}{
					"act":  "CHAT",
					"data": fmt.Sprintf("/toggle %s", module.GetIdentifier()),
				},
			}

			modulesControls = append(modulesControls, control)

			if module.IsEnabled() {
				modulesDisplay = append(modulesDisplay, module.GetIdentifier())
				control["tc"] = 0xFFFFFF
				control["color"] = -0xABAB01
			}

			x += 90
		}

		modulesControls = append(modulesControls, map[string]interface{}{
			"id":   fmt.Sprintf("kv.cc.%s", category),
			"type": "Text",
			"al":   "RIGHT",
			"pos":  "BOTTOM_RIGHT",
			"text": []string{category.Name},
			"x":    10,
			"y":    y + 25,
		})

		y += 56
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
			"y": 12,
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

func RegisterDefaultModules(tunnel *MinecraftTunnel) {
	moduleHandler := tunnel.GetModuleHandler()
	tpAuraTicking := modules.SimpleTickingModule{Interval: 250 * time.Millisecond}

	if tunnel.HasFeature(license.FeatureFlight) {
		moduleHandler.RegisterModule(&flight.Flight{Speed: 3})
	}

	if tunnel.HasFeature(license.FeatureAntiKnockback) {
		moduleHandler.RegisterModule(&antiknockback.AntiKnockback{})
	}

	if tunnel.HasFeature(license.FeatureNoFall) {
		moduleHandler.RegisterModule(&nofall.NoFall{})
	}

	if tunnel.HasFeature(license.FeatureKillAura) {
		genericAura := aura.GenericAura{
			MaxDistance: 7, HitAnimation: false,
			SimpleTickingModule: modules.SimpleTickingModule{Interval: 35 * time.Millisecond},
		}

		moduleHandler.RegisterModule(&aura.KillAura{GenericAura: genericAura})
		moduleHandler.RegisterModule(&aura.MobAura{GenericAura: genericAura})
	}

	if tunnel.HasFeature(license.FeatureLongJump) {
		moduleHandler.RegisterModule(&longjump.LongJump{Power: 2})
	}

	if tunnel.HasFeature(license.FeatureUnlimitedCPS) {
		moduleHandler.RegisterModule(&unlimitedcps.UnlimitedCPS{})
	}

	if tunnel.HasFeature(license.FeatureTPAura) {
		moduleHandler.RegisterModule(&tpaura.TPAura{SearchRadius: 20, TeleportRadius: 4, SimpleTickingModule: tpAuraTicking})
	}

	if tunnel.HasFeature(license.FeaturePlayerESP) {
		moduleHandler.RegisterModule(&modules.ClientModule{Identifier: modules.ModulePlayerESP})
	}

	if tunnel.HasFeature(license.FeatureNuker) {
		moduleHandler.RegisterModule(&nuker.Nuker{Radius: 2, SimpleTickingModule: modules.SimpleTickingModule{Interval: time.Second}})
	}

	if tunnel.HasFeature(license.FeatureFastBreak) {
		moduleHandler.RegisterModule(&fastbreak.FastBreak{})
	}

	if tunnel.HasFeature(license.FeatureNoBadEffects) {
		moduleHandler.RegisterModule(&nobadeffects.NoBadEffects{})
	}

	moduleHandler.RegisterModule(&spammer.Spammer{SimpleTickingModule: modules.SimpleTickingModule{Interval: 20 * time.Second}})
	moduleHandler.RegisterModule(&cmdcam.CMDCam{})
}
