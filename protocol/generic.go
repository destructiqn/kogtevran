package protocol

import pk "github.com/ruscalworld/vimeinterceptor/net/packet"

type Packet interface {
	Read(packet pk.Packet) error
	Marshal() pk.Packet
}
