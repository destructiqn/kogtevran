package unlimitedcps

import (
	"time"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	"github.com/destructiqn/kogtevran/protocol"
)

type UnlimitedCPS struct {
	modules.DefaultModule
}

func (u *UnlimitedCPS) GetIdentifier() string {
	return modules.ModuleUnlimitedCPS
}

func (u *UnlimitedCPS) Toggle() (bool, error) {
	u.Enabled = !u.Enabled
	return u.Enabled, u.Update()
}

func (u *UnlimitedCPS) Update() error {
	return u.Tunnel.GetTexteriaHandler().SendClient(map[string]interface{}{
		"%":     "option:set",
		"field": "disable-cps-limit",
		"value": u.Enabled,
	})
}

func HandleJoinGame(_ protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleUnlimitedCPS) {
		go func(conn generic.Tunnel) {
			time.Sleep(100 * time.Millisecond)
			module, ok := conn.GetModuleHandler().GetModule(modules.ModuleUnlimitedCPS)
			if ok {
				err = module.(*UnlimitedCPS).Update()
			}
		}(tunnel)
	}

	return generic.PassPacket(), nil
}
