package proxy

import (
	"github.com/Tnze/go-mc/chat"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
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
