package proxy

import (
	"bytes"
	"fmt"

	"github.com/ruscalworld/vimeinterceptor/generic"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type TexteriaHandler struct {
	tunnel *MinecraftTunnel
}

func NewTexteriaHandler(tunnel *MinecraftTunnel) *TexteriaHandler {
	return &TexteriaHandler{tunnel: tunnel}
}

func HandleClientboundTexteriaPacket(data []byte, tunnel generic.Tunnel) (result []byte, next bool, err error) {
	minecraftTunnel := tunnel.(*MinecraftTunnel)

	var amount pk.VarInt
	reader := bytes.NewReader(data)
	_, err = amount.ReadFrom(reader)
	if err != nil {
		return
	}

	buffer := &bytes.Buffer{}
	_, err = amount.WriteTo(buffer)
	if err != nil {
		return
	}

	for i := 0; i < int(amount); i++ {
		var actionData pk.ByteArray
		_, err = actionData.ReadFrom(reader)
		if err != nil {
			return
		}

		var byteMap map[string]interface{}
		byteMap, err = pk.ReadMap(actionData)
		if err != nil {
			return
		}

		modified := minecraftTunnel.GetTexteriaHandler().InterceptAction(byteMap)
		if modified {
			actionData, err = pk.EncodeMap(byteMap)
			if err != nil {
				return
			}
		}

		_, err = actionData.WriteTo(buffer)
		if err != nil {
			return
		}
	}

	return buffer.Bytes(), true, nil
}

func (t *TexteriaHandler) SendClient(data map[string]interface{}) error {
	byteMap, err := pk.EncodeMap(data)
	if err != nil {
		return err
	}

	packet := pk.Marshal(protocol.ClientboundPluginMessage, pk.VarInt(1), pk.ByteArray(byteMap))
	return t.tunnel.WriteClient(packet)
}

func (t *TexteriaHandler) InterceptAction(data map[string]interface{}) bool {
	if id, ok := data["id"]; ok && id == "vn.n" {
		data["text"] = []string{fmt.Sprintf("%s §r(§8%8s§r)", data["text"].([]string)[0], t.tunnel.SessionID)}
		return true
	}

	return false
}
