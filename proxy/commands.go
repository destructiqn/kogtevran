package proxy

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/modules"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
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

	"gamestate": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 1 {
			return errors.New("not enough args")
		}

		reason, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		value := 0
		if len(args) == 2 {
			value, err = strconv.Atoi(args[1])
			if err != nil {
				return err
			}
		}

		packet := &protocol.ChangeGameState{
			Reason: pk.UnsignedByte(reason),
			Value:  pk.Float(value),
		}

		return tunnel.WriteClient(packet.Marshal())
	},

	"dig": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 3 {
			return errors.New("not enough args")
		}

		x, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		y, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		z, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}

		position := pk.Position{X: x, Y: y, Z: z}
		start := &protocol.PlayerDigging{
			Face:     1,
			Status:   0,
			Location: position,
		}

		finish := &protocol.PlayerDigging{
			Face:     1,
			Status:   2,
			Location: position,
		}

		err = tunnel.WriteServer(start.Marshal())
		if err != nil {
			return err
		}

		return tunnel.WriteServer(finish.Marshal())
	},

	"gamemode": func(args []string, tunnel generic.Tunnel) error {
		if len(args) < 1 {
			return tunnel.GetChatHandler().SendMessage(chat.TranslateMsg("commands.gamemode.usage"), protocol.ChatPositionSystemMessage)
		}

		var gameMode pk.Float
		var translation string

		switch args[0] {
		case "0", "survival":
			gameMode = 0
			translation = "survival"
		case "1", "creative":
			gameMode = 1
			translation = "creative"
		case "2", "adventure":
			gameMode = 2
			translation = "adventure"
		case "3", "spectator":
			gameMode = 3
			translation = "spectator"
		}

		packet := &protocol.ChangeGameState{
			Reason: 3,
			Value:  gameMode,
		}

		err := tunnel.WriteClient(packet.Marshal())
		if err != nil {
			return err
		}

		return tunnel.GetChatHandler().SendMessage(
			chat.TranslateMsg("commands.gamemode.success.self",
				chat.TranslateMsg(fmt.Sprintf("gameMode.%s", translation)),
			), protocol.ChatPositionSystemMessage,
		)
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

func HandleChatMessage(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	chatMessage := packet.(*protocol.ChatMessage)
	handled := HandleCommand(string(chatMessage.Message), tunnel)

	if handled {
		return generic.RejectPacket(), nil
	} else {
		return generic.PassPacket(), nil
	}
}
