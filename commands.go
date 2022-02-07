package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Tnze/go-mc/chat"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

type CommandHandler func(args []string, conn *WrappedConn) error

var Commands = map[string]CommandHandler{
	"toggle": func(args []string, conn *WrappedConn) error {
		if len(args) < 1 {
			return errors.New("not enough args")
		}

		module, ok := conn.Modules[args[0]]
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

		return conn.SendMessage(chat.Text(fmt.Sprintf("%s is now %s", module.GetIdentifier(), statusText)), ChatPositionAboveHotbar)
	},

	"set": func(args []string, conn *WrappedConn) error {
		if len(args) < 2 {
			return errors.New("not enough args")
		}

		switch strings.ToLower(args[0]) {
		case "flightspeed", "fs":
			speed, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			flight := conn.Modules[ModuleFlight].(*Flight)
			flight.Speed = float64(speed)
			err = flight.Update()
			if err != nil {
				return err
			}

			_ = conn.SendMessage(chat.Text(fmt.Sprintf("flight speed amplifier is set to %.0f", flight.Speed)), ChatPositionAboveHotbar)
		case "msg":
			spammer := conn.Modules[ModuleSpammer].(*Spammer)
			spammer.Message = strings.Join(args[1:], " ")

			_ = conn.SendMessage(chat.Text(fmt.Sprintf("spam message is set %s", spammer.Message)), ChatPositionAboveHotbar)
		case "kb":
			if len(args) < 4 {
				return errors.New("not enough args")
			}

			antiKnockback := conn.Modules[ModuleAntiKnockback].(*AntiKnockback)

			var err error
			antiKnockback.X, err = strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			antiKnockback.Y, err = strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			antiKnockback.Z, err = strconv.Atoi(args[3])
			if err != nil {
				return err
			}

			if err != nil {
				return err
			}

			_ = conn.SendMessage(chat.Text(fmt.Sprintf("kb: x %d, y %d, z %d", antiKnockback.X, antiKnockback.Y, antiKnockback.Z)), ChatPositionAboveHotbar)
		}
		return nil
	},

	"effect": func(args []string, conn *WrappedConn) error {
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

		return conn.WriteClient(pk.Marshal(0x1D, pk.VarInt(conn.EntityID), pk.Byte(id), pk.Byte(amplifier), pk.VarInt(100000), pk.Boolean(true)))
	},

	"entities": func(args []string, conn *WrappedConn) error {
		log.Println(conn.Entities)
		return nil
	},

	"location": func(args []string, conn *WrappedConn) error {
		log.Println(*conn.Location)
		return nil
	},

	"speed": func(args []string, conn *WrappedConn) error {
		if len(args) < 1 {
			return errors.New("not enough args")
		}

		speed, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return err
		}

		return conn.WriteClient(pk.Marshal(0x20, pk.VarInt(conn.EntityID), pk.Int(1), pk.String("generic.movementSpeed"), pk.Double(0.699999988079071*speed), pk.VarInt(0)))
	},

	"open": func(args []string, conn *WrappedConn) error {
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
		return conn.WriteClient(packet)
	},
}

func HandleCommand(message string, conn *WrappedConn) bool {
	if !strings.HasPrefix(message, "/") {
		return false
	}
	parts := strings.Split(message, " ")
	command := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
	handler, ok := Commands[command]
	if !ok {
		return false
	}

	err := handler(parts[1:], conn)
	if err != nil {
		_ = conn.SendMessage(chat.Text(err.Error()), ChatPositionSystemMessage)
	}

	return true
}
