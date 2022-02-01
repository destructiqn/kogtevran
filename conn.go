package main

import (
	"sync"

	"github.com/Tnze/go-mc/chat"
	"github.com/ruscalworld/vimeinterceptor/net"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

type WrappedConn struct {
	Closed  bool
	Server  *net.Conn
	Client  *net.Conn
	Modules map[string]Module

	ServerWrite sync.Mutex
	ClientWrite sync.Mutex
}

func (w *WrappedConn) IsModuleEnabled(moduleID string) bool {
	module, ok := w.Modules[moduleID]
	if !ok {
		return false
	}

	return module.IsEnabled()
}

func (w *WrappedConn) WriteClient(packet pk.Packet) error {
	w.ClientWrite.Lock()
	err := w.Client.WritePacket(packet)
	w.ClientWrite.Unlock()
	return err
}

func (w *WrappedConn) WriteServer(packet pk.Packet) error {
	w.ServerWrite.Lock()
	err := w.Server.WritePacket(packet)
	w.ServerWrite.Unlock()
	return err
}

func (w *WrappedConn) Disconnect() {
	w.Closed = true
	_ = w.Server.Close()
	_ = w.Client.Close()
}

func WrapConn(server, client *net.Conn) *WrappedConn {
	return &WrappedConn{
		Server:  server,
		Client:  client,
		Modules: make(map[string]Module),
	}
}

type ChatPosition byte

var (
	ChatPositionChat          ChatPosition = 0
	ChatPositionSystemMessage ChatPosition = 1
	ChatPositionAboveHotbar   ChatPosition = 2
)

func (w *WrappedConn) RegisterModule(module Module) {
	module.Register(w)
	w.Modules[module.GetIdentifier()] = module
}

func (w *WrappedConn) SendMessage(message chat.Message, position ChatPosition) error {
	return w.WriteClient(pk.Marshal(0x02, message, pk.Byte(position)))
}
