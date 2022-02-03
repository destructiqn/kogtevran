package texteria

import (
	"bytes"
	pk "github.com/ruscalworld/vimeinterceptor/net/packet"
)

type ByteMap map[string]interface{}

func (b ByteMap) Bytes() []byte {
	buf := new(bytes.Buffer)
	for key, value := range b {
		buf.Write([]byte(key))

		switch value.(type) {
		case int:
			i := value.(int)
			if i >= 0 && i < 2097152 {
				buf.Write([]byte{byte(13)})
				_, _ = pk.VarInt(i).WriteTo(buf)
			} else if i < 0 && i > -1048576 {
				buf.Write([]byte{byte(17)})
				_, _ = pk.VarInt(i).WriteTo(buf)
			} else {
				buf.Write([]byte{byte(1), byte(i)})
			}
		}
	}

	return buf.Bytes()
}
