package generic

import pk "github.com/destructiqn/kogtevran/net/packet"

type HandlerResult struct {
	ShouldPass bool
	IsModified bool
	Packet     pk.Packet
}

func PassPacket() *HandlerResult {
	return &HandlerResult{
		ShouldPass: true,
	}
}

func ModifyPacket(packet pk.Packet) *HandlerResult {
	return &HandlerResult{
		ShouldPass: true,
		IsModified: true,
		Packet:     packet,
	}
}

func RejectPacket() *HandlerResult {
	return &HandlerResult{}
}
