package proxy

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"text/template"
	"time"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/license"
	"github.com/destructiqn/kogtevran/modules"

	"github.com/destructiqn/kogtevran/modules/antiknockback"
	"github.com/destructiqn/kogtevran/modules/aura"
	"github.com/destructiqn/kogtevran/modules/cmdcam"
	"github.com/destructiqn/kogtevran/modules/fastbreak"
	"github.com/destructiqn/kogtevran/modules/flight"
	"github.com/destructiqn/kogtevran/modules/longjump"
	"github.com/destructiqn/kogtevran/modules/nobadeffects"
	"github.com/destructiqn/kogtevran/modules/nofall"
	"github.com/destructiqn/kogtevran/modules/nuker"
	"github.com/destructiqn/kogtevran/modules/spammer"
	"github.com/destructiqn/kogtevran/modules/tpaura"
	"github.com/destructiqn/kogtevran/modules/unlimitedcps"
)

type ModuleHandler struct {
	tunnel                *MinecraftTunnel
	modules               map[string]generic.Module
	initializedCategories map[string]bool
}

func NewModuleHandler(tunnel *MinecraftTunnel) *ModuleHandler {
	return &ModuleHandler{tunnel, make(map[string]generic.Module), make(map[string]bool)}
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
		if auxiliary == nil {
			return module.IsEnabled(), nil
		}

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

const (
	moduleMargin       = 8
	moduleButtonHeight = 20
	moduleButtonWidth  = 85

	categoryMargin = 16
	categoryWidth  = categoryMargin + moduleButtonWidth

	categoryTitleHeight = 16
	headerHeight        = moduleMargin * 3
)

func (m *ModuleHandler) GetModulesDetails() []map[string]interface{} {
	x := categoryMargin * 10
	modulesDisplay := make([]string, 0)
	elements := make([]map[string]interface{}, 0)

	categories := modules.GetCategoryList()
	sort.Sort(categories)

	script, err := template.ParseFiles("texteria/gui.js")
	if err != nil {
		return nil
	}

	scriptContext := map[string]string{}

	buffer := &bytes.Buffer{}
	err = script.Execute(buffer, scriptContext)
	if err != nil {
		return nil
	}

	for _, category := range categories {
		y := headerHeight + moduleMargin*3 + 4
		modulesList := make(ModuleList, 0)

		for _, moduleID := range category.ModuleIDs {
			if module, ok := m.GetModule(moduleID); ok {
				modulesList = append(modulesList, module)
			}
		}

		sort.Sort(modulesList)
		categoryControlID := fmt.Sprintf("kv.cc.%s.bg", category.Name)
		categoryHeight := moduleMargin*2 + categoryTitleHeight + len(modulesList)*(moduleButtonHeight+moduleMargin)

		if _, ok := m.initializedCategories[category.Name]; !ok {
			elements = append(elements, map[string]interface{}{
				/* Category Background */
				"id":     categoryControlID,
				"pos":    "TOP_LEFT",
				"type":   "Rectangle",
				"width":  categoryWidth,
				"height": categoryHeight,
				"color":  -0x80000000,
				"x":      x,
				"y":      categoryMargin * 4,
				"script": string(buffer.Bytes()),
				"click": map[string]interface{}{
					"act": "SCRIPT",
				},
			}, map[string]interface{}{
				/* Category Header */
				"id":         fmt.Sprintf("kv.cc.%s.head", category.Name),
				"pos":        "TOP_RIGHT",
				"type":       "Rectangle",
				"width":      categoryWidth,
				"height":     headerHeight,
				"color":      -0xABAB01,
				"x":          -categoryWidth,
				"y":          3 * moduleMargin,
				"attach.to":  categoryControlID,
				"attach.loc": "TOP_RIGHT",
			}, map[string]interface{}{
				/* Category Title */
				"id":         fmt.Sprintf("kv.cc.%s", category),
				"type":       "Text",
				"al":         "LEFT",
				"attach.to":  categoryControlID,
				"attach.loc": "TOP_RIGHT",
				"text":       []string{category.Name},
				"x":          moduleMargin - categoryWidth,
				"y":          2*moduleMargin + 1,
			})

			m.initializedCategories[category.Name] = true
		}

		for _, module := range modulesList {
			control := map[string]interface{}{
				/* Module Button */
				"id":         fmt.Sprintf("kv.mc.%s", module.GetIdentifier()),
				"type":       "Button",
				"w":          moduleButtonWidth,
				"h":          moduleButtonHeight,
				"t":          module.GetIdentifier(),
				"tc":         0xAAAAAA,
				"hc":         0x5555FF,
				"x":          moduleMargin - categoryWidth,
				"y":          y,
				"tooltip":    module.GetDescription(),
				"color":      -0x33000000,
				"attach.to":  categoryControlID,
				"attach.loc": "TOP_RIGHT",
				"click": map[string]interface{}{
					"act":  "CHAT",
					"data": fmt.Sprintf("/toggle %s", module.GetIdentifier()),
				},
			}

			elements = append(elements, control)

			if module.IsEnabled() {
				modulesDisplay = append(modulesDisplay, module.GetIdentifier())
				control["tc"] = 0xFFFFFF
				control["color"] = -0xABAB01
			}

			y += moduleMargin + moduleButtonHeight
		}

		x += categoryMargin + categoryWidth
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
			"e": elements,
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

	if tunnel.HasFeature(license.FeatureESP) {
		moduleHandler.RegisterModule(&modules.ClientModule{
			Identifier:  modules.ModulePlayerESP,
			Description: []string{"Отныне ты можешь видеть игроков через стены"},
		})

		moduleHandler.RegisterModule(&modules.ClientModule{
			Identifier:  modules.ModuleChestESP,
			Description: []string{"Отныне ты можешь видеть сундуки через стены"},
		})
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
