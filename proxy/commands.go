package proxy

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/generic"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
	"github.com/ruscalworld/vimeinterceptor/protocol"
)

type CommandHandler func(args []string, tunnel generic.Tunnel) error

var Commands = map[string]CommandHandler{
	"toggle": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 1 {
			return errors.New("not enough args")
		}

		module, ok := tunnel.GetModuleHandler().GetModule(args[0])
		if !ok {
			return errors.New("unknown module")
		}

		status, err := module.Toggle()
		if err != nil {
			return err
		}

		statusText := "enabled"
		if !status {
			statusText = "disabled"
		}

		return tunnel.GetChatHandler().SendMessage(chat.Text(fmt.Sprintf("%s is now %s", module.GetIdentifier(), statusText)), protocol.ChatPositionAboveHotbar)
	},

	"effect": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 2 {
			return errors.New("not enough args")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		amplifier, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		return tunnel.WriteClient(pk.Marshal(0x1D, pk.VarInt(tunnel.GetPlayerHandler().GetEntityID()), pk.Byte(id), pk.Byte(amplifier), pk.VarInt(100000), pk.Boolean(true)))
	},

	"entities": func(args []string, tunnel generic.Tunnel) error {
		log.Println(tunnel.GetEntityHandler().GetEntities())
		return nil
	},

	"location": func(args []string, tunnel generic.Tunnel) error {
		log.Println(*tunnel.GetPlayerHandler().GetLocation())
		return nil
	},

	"speed": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 1 {
			return errors.New("not enough args")
		}

		speed, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return err
		}

		return tunnel.WriteClient(pk.Marshal(0x20, pk.VarInt(tunnel.GetPlayerHandler().GetEntityID()), pk.Int(1), pk.String("generic.movementSpeed"), pk.Double(0.699999988079071*speed), pk.VarInt(0)))
	},

	"open": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 1 {
			return errors.New("not enough args")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		data := []pk.FieldEncoder{pk.String("MiniDot"), pk.String("%"), pk.Byte(4), pk.String("available"), pk.String("list"), pk.Byte(16), pk.VarInt(id)}
		for i := 1; i <= id; i++ {
			data = append(data, pk.VarInt(i))
		}

		packet := pk.Marshal(0x3F, data...)
		return tunnel.WriteClient(packet)
	},
}

func HandleCommand(message string, tunnel generic.Tunnel) bool {
	if !strings.HasPrefix(message, "/") {
		return false
	}
	parts := strings.Split(message, " ")
	command := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
	handler, ok := Commands[command]
	if !ok {
		return false
	}

	err := handler(parts[1:], tunnel)
	if err != nil {
		_ = tunnel.GetChatHandler().SendMessage(chat.Text(err.Error()), protocol.ChatPositionSystemMessage)
	}

	return true
}

func HandleChatMessage(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	chatMessage := packet.(*protocol.ChatMessage)
	handled := HandleCommand(string(chatMessage.Message), tunnel)
	return packet.Marshal(), !handled, nil
}
