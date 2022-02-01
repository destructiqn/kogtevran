package main

import (
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

const (
	ModuleFlight = "flight"
)

type Module interface {
	Register(conn *WrappedConn)
	GetIdentifier() string
	Toggle() (bool, error)
	IsEnabled() bool
}

type DefaultModule struct {
	Conn    *WrappedConn
	Enabled bool
}

func (m *DefaultModule) Register(conn *WrappedConn) {
	m.Conn = conn
}

func (m *DefaultModule) IsEnabled() bool {
	return m.Enabled
}

func RegisterDefaultModules(conn *WrappedConn) {
	conn.RegisterModule(&Flight{})
}

type Flight struct {
	DefaultModule
}

func (f *Flight) GetIdentifier() string {
	return ModuleFlight
}

func (f *Flight) Toggle() (bool, error) {
	flags := 0x04
	if f.Enabled {
		flags = 0
	}

	err := f.Conn.WriteClient(pk.Marshal(0x39, pk.Byte(flags), pk.Float(0.1), pk.Float(0.1)))
	if err != nil {
		return f.Enabled, err
	}

	f.Enabled = !f.Enabled
	return f.Enabled, nil
}
