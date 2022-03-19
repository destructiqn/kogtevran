package proxy

import (
	"bytes"
	"fmt"

	"github.com/destructiqn/kogtevran/generic"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
)

type TexteriaHandler struct {
	tunnel      *MinecraftTunnel
	initialized bool
}

func NewTexteriaHandler(tunnel *MinecraftTunnel) *TexteriaHandler {
	return &TexteriaHandler{tunnel: tunnel}
}

// HandleClientboundTexteriaPacket - метод, обрабатывавший пакеты Texteria, но из-за проблем с парсингом ByteMap'ов
// не используется
//
// Deprecated
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

		modified := minecraftTunnel.GetTexteriaHandler().(*TexteriaHandler).InterceptAction(byteMap)
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

func HandleKeyboardPacketCandidate(data []byte, tunnel generic.Tunnel) (result []byte, next bool, err error) {
	var dataMap map[string]interface{}
	dataMap, err = pk.ReadMap(data)
	if err != nil {
		return
	}

	if dataMap["%"] == "kv:module:toggle" {
		module, ok := tunnel.GetModuleHandler().GetModule(dataMap["module"].(string))
		if !ok {
			return data, true, nil
		}

		_, err = tunnel.GetModuleHandler().ToggleModule(module)
		return nil, false, err
	}

	return data, true, nil
}

func (t *TexteriaHandler) SendClient(data ...map[string]interface{}) error {
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
	if data["%"] == "reset" {
		t.initialized = false
	}

	return false
}

func GetBranding() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"%":      "add",
			"id":     "kv.mh",
			"pos":    "TOP_LEFT",
			"type":   "Rectangle",
			"width":  142,
			"height": 46,
			"color":  -0x80000000,

			"vis": []map[string]interface{}{
				{
					"type": "always",
					"show": true,
				},
				{
					"type": "f3",
					"show": false,
				},
			},

			"x": 7,
			"y": 25,
		},
		{
			"%":       "add",
			"id":      "kv.mj",
			"pos":     "TOP_LEFT",
			"type":    "Text",
			"scale.x": 2.5,
			"scale.z": 2.5,
			"scale.y": 2.5,
			"text":    []string{"§9K§rogtevra§9n"},

			"vis": []map[string]interface{}{
				{
					"type": "always",
					"show": true,
				},
				{
					"type": "f3",
					"show": false,
				},
			},

			"x": 14,
			"y": 32,
		},
		{
			"%":    "add",
			"id":   "kv.mk",
			"pos":  "TOP_LEFT",
			"type": "Text",
			"text": []string{"by §9kliri"},

			"vis": []map[string]interface{}{
				{
					"type": "always",
					"show": true,
				},
				{
					"type": "f3",
					"show": false,
				},
			},

			"x": 14,
			"y": 56,
		},
		{
			"%":    "add",
			"id":   "kv.mz",
			"al":   "RIGHT",
			"pos":  "TOP_RIGHT",
			"type": "Text",
			"text": []string{fmt.Sprintf("§9Kogtevran §7%s", generic.GetRevision())},

			"vis": []map[string]interface{}{
				{
					"type": "always",
					"show": true,
				},
				{
					"type": "f3",
					"show": false,
				},
			},

			"x": 2,
			"y": 2,
		},
		{
			"%":    "add",
			"id":   "kv.mi",
			"al":   "RIGHT",
			"pos":  "BOTTOM_RIGHT",
			"type": "Text",
			"text": []string{"§7vk.com§9/destructiqn"},

			"vis": []map[string]interface{}{
				{
					"type": "always",
					"show": true,
				},
				{
					"type": "chat",
					"show": false,
				},
			},

			"x": 2,
			"y": 2,
		},
	}
}

func (t *TexteriaHandler) UpdateInterface() error {
	modulesDetails := t.tunnel.GetModuleHandler().(*ModuleHandler).GetModulesDetails()
	branding := GetBranding()
	return t.tunnel.TexteriaHandler.SendClient(append(modulesDetails, branding...)...)
}
