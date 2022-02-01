package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Tnze/go-mc/chat"
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
