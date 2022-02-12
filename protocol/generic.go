package protocol

import pk "github.com/destructiqn/kogtevran/net/packet"

type Packet interface {
	Read(packet pk.Packet) error
	Marshal() pk.Packet
}
