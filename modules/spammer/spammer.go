package spammer

import (
	"time"

	"github.com/ruscalworld/vimeinterceptor/modules"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

type Spammer struct {
	modules.SimpleModule
	Message string
}

func (s *Spammer) GetIdentifier() string {
	return modules.ModuleSpammer
}

func (s *Spammer) Tick() error {
	processedMsg := transliterate(s.Message)
	return s.Tunnel.WriteServer(pk.Marshal(0x01, pk.String(processedMsg)))
}

func (s *Spammer) GetInterval() time.Duration {
	return 20 * time.Second
}
