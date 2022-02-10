package spammer

import (
	"time"

	"github.com/ruscalworld/vimeinterceptor/modules"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type Spammer struct {
	modules.SimpleTickingModule
	Message string `option:"message"`
}

func (s *Spammer) GetIdentifier() string {
	return modules.ModuleSpammer
}

func (s *Spammer) Tick() error {
	processedMsg := transliterate(s.Message)
	return s.Tunnel.WriteServer(pk.Marshal(protocol.ServerboundChatMessage, pk.String(processedMsg)))
}

func (s *Spammer) GetInterval() time.Duration {
	return 20 * time.Second
}
