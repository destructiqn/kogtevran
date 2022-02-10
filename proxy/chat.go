package proxy

import (
	"github.com/Tnze/go-mc/chat"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type ChatHandler struct {
	tunnel *MinecraftTunnel
}

func NewChatHandler(tunnel *MinecraftTunnel) *ChatHandler {
	return &ChatHandler{tunnel: tunnel}
}

func (c *ChatHandler) SendMessage(message chat.Message, position protocol.ChatPosition) error {
	return c.tunnel.WriteClient(pk.Marshal(protocol.ClientboundChatMessage, message, pk.Byte(position)))
}
