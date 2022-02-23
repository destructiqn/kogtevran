package unlimitedcps

import (
	"time"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
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

func HandleJoinGame(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleUnlimitedCPS) {
		go func(conn generic.Tunnel) {
			time.Sleep(100 * time.Millisecond)
			module, ok := conn.GetModuleHandler().GetModule(modules.ModuleUnlimitedCPS)
			if ok {
				err = module.(*UnlimitedCPS).Update()
			}
		}(tunnel)
	}

	return packet.Marshal(), true, err
}