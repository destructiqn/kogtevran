package spammer

import (
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

type Spammer struct {
	modules.SimpleTickingModule
	Message string `option:"message"`
}

func (s *Spammer) GetDescription() []string {
	return []string{
		"Автоматически отправляет сообщения в чат",
		"",
		"§nПараметры",
		"§7message§f - сообщение, которое нужно отправлять в чат",
		"§7interval§f - интервал между сообщениями",
	}
}

func (s *Spammer) GetIdentifier() string {
	return modules.ModuleSpammer
}

func (s *Spammer) Tick() error {
	processedMsg := transliterate(s.Message)
	return s.Tunnel.WriteServer(pk.Marshal(protocol.ServerboundChatMessage, pk.String(processedMsg)))
}
