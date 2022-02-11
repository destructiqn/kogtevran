package proxy

import (
	"bytes"
	"fmt"
	"github.com/ruscalworld/vimeinterceptor/generic"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type TexteriaHandler struct {
	tunnel      *MinecraftTunnel
	initialized bool
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

	var modulesDetails []map[string]interface{}
	if !minecraftTunnel.TexteriaHandler.initialized {
		modulesDetails = minecraftTunnel.ModuleHandler.GetModulesDetails()
		amount += pk.VarInt(len(modulesDetails))
	}

	finalBuffer := &bytes.Buffer{}
	_, err = (amount).WriteTo(finalBuffer)
	if err != nil {
		return
	}

	_, err = buffer.WriteTo(finalBuffer)
	if err != nil {
		return
	}

	for _, modulesDetailsFragment := range modulesDetails {
		var encodedFragment []byte
		encodedFragment, err = pk.EncodeMap(modulesDetailsFragment)
		if err != nil {
			return
		}

		_, err = pk.ByteArray(encodedFragment).WriteTo(finalBuffer)
		if err != nil {
			return
		}
	}

	minecraftTunnel.TexteriaHandler.initialized = true
	return finalBuffer.Bytes(), true, err
}

func (t *TexteriaHandler) SendClient(data []map[string]interface{}) error {
	buffer := &bytes.Buffer{}
	_, err := pk.VarInt(len(data)).WriteTo(buffer)
	if err != nil {
		return err
	}

	for _, fragment := range data {
		byteMap, err := pk.EncodeMap(fragment)
		if err != nil {
			return err
		}

		_, err = pk.ByteArray(byteMap).WriteTo(buffer)
		if err != nil {
			return err
		}
	}

	messageData := pk.PluginMessageData(buffer.Bytes())
	packet := pk.Marshal(protocol.ClientboundPluginMessage, pk.String("Texteria"), &messageData)
	return t.tunnel.WriteClient(packet)
}

func (t *TexteriaHandler) InterceptAction(data map[string]interface{}) bool {
	modified := false

	if data["%"] == "reset" {
		t.initialized = false
	}

	if id, ok := data["id"]; ok && id == "vn.n" {
		data["text"] = []string{fmt.Sprintf("%s §r(§8%8s§r)", data["text"].([]string)[0], t.tunnel.SessionID)}
		modified = true
	}

	return modified
}
