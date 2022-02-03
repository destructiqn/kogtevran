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

		speed, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		return conn.WriteClient(pk.Marshal(0x20, pk.VarInt(conn.EntityID), pk.Int(1), pk.String("generic.movementSpeed"), pk.Double(speed), pk.VarInt(0)))
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
