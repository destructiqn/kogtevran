package proxy

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/generic"
	"github.com/ruscalworld/vimeinterceptor/modules"
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

		_, err := tunnel.GetModuleHandler().ToggleModule(module)
		return err
	},

	"set": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 1 {
			return errors.New("not enough args")
		}

		module, ok := tunnel.GetModuleHandler().GetModule(args[0])
		if !ok {
			return errors.New("unknown module")
		}

		value, ok := modules.GetOptionValue(module, args[1])
		if !ok {
			return errors.New("unknown option")
		}

		var (
			newValue interface{}
			err      error
		)

		switch value.(type) {
		case string:
			newValue = strings.Join(args[2:], " ")
		case bool:
			newValue = args[2] == "true" || args[2] == "1"
		case float64:
			newValue, err = strconv.ParseFloat(args[2], 64)
		case time.Duration:
			newValue, err = time.ParseDuration(args[2])
		default:
			newValue, err = strconv.Atoi(args[2])
		}

		if err != nil {
			return err
		}

		ok = modules.SetOptionValue(module, args[1], newValue)
		if !ok {
			return errors.New("unable to change value")
		}

		err = module.Update()
		if err != nil {
			return err
		}

		return tunnel.GetChatHandler().SendMessage(
			chat.Text(fmt.Sprintf("set %s of %s to %v", args[1], module.GetIdentifier(), newValue)), protocol.ChatPositionAboveHotbar,
		)
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

	"inventory": func(args []string, tunnel generic.Tunnel) error {
		for id, window := range tunnel.GetInventoryHandler().GetWindows() {
			fmt.Println(id, window, ":")
			for slot, item := range window.GetContents() {
				fmt.Println(slot, item)
			}
			fmt.Println("---")
		}
		return nil
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
