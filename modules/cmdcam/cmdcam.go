package cmdcam

import "github.com/destructiqn/kogtevran/modules"

type CMDCam struct {
	modules.SimpleModule
}

func (c *CMDCam) GetIdentifier() string {
	return modules.ModuleCMDCam
}

func (c *CMDCam) Toggle() (bool, error) {
	err := c.Tunnel.GetTexteriaHandler().SendClient(map[string]interface{}{
		"%":     "option:set",
		"field": "cmdcam",
		"value": !c.Enabled,
	})
	if err != nil {
		return c.Enabled, err
	}

	return c.SimpleModule.Toggle()
}
