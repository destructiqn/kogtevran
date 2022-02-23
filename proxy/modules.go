package proxy

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"

	"github.com/destructiqn/kogtevran/modules/antiknockback"
	"github.com/destructiqn/kogtevran/modules/aura"
	"github.com/destructiqn/kogtevran/modules/cmdcam"
	"github.com/destructiqn/kogtevran/modules/flight"
	"github.com/destructiqn/kogtevran/modules/longjump"
	"github.com/destructiqn/kogtevran/modules/nofall"
	"github.com/destructiqn/kogtevran/modules/spammer"
	"github.com/destructiqn/kogtevran/modules/tpaura"
	"github.com/destructiqn/kogtevran/modules/unlimitedcps"
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

	err = t.tunnel.GetTexteriaHandler().UpdateInterface()
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
	modulesList := ModuleList(t.GetModules())
	sort.Sort(modulesList)

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

		x += 90
		if x > 300 {
			y += 25
			x = 10
		}

		modulesControls = append(modulesControls, control)

		if module.IsEnabled() {
			modulesDisplay = append(modulesDisplay, module.GetIdentifier())
			control["tc"] = 0xFFFFFF
			control["color"] = -0xABAB01
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

func RegisterDefaultModules(tunnel generic.Tunnel) {
	moduleHandler := tunnel.GetModuleHandler()

	genericAura := aura.GenericAura{
		MaxDistance: 7, HitAnimation: false,
		SimpleTickingModule: modules.SimpleTickingModule{Interval: 35 * time.Millisecond},
	}

	tpAura := tpaura.TPAura{
		SimpleTickingModule: modules.SimpleTickingModule{Interval: 250 * time.Millisecond},
		SearchRadius:        20,
		TeleportRadius:      4,
	}

	moduleHandler.RegisterModule(&flight.Flight{Speed: 3})
	moduleHandler.RegisterModule(&antiknockback.AntiKnockback{})
	moduleHandler.RegisterModule(&nofall.NoFall{})
	moduleHandler.RegisterModule(&aura.KillAura{GenericAura: genericAura})
	moduleHandler.RegisterModule(&aura.MobAura{GenericAura: genericAura})
	moduleHandler.RegisterModule(&spammer.Spammer{SimpleTickingModule: modules.SimpleTickingModule{Interval: 20 * time.Second}})
	moduleHandler.RegisterModule(&cmdcam.CMDCam{})
	moduleHandler.RegisterModule(&longjump.LongJump{Power: 2})
	moduleHandler.RegisterModule(&unlimitedcps.UnlimitedCPS{})
	moduleHandler.RegisterModule(&tpAura)
}
