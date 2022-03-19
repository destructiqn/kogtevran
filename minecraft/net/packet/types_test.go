package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlot_ReadFrom(t *testing.T) {
	var slot Slot
	_, err := slot.ReadFrom(bytes.NewReader([]byte{0xFF, 0xFF}))
	if err != nil {
		t.Error(err)
		return
	}

	assert.True(t, slot.IsEmpty())

	slot = Slot{}
	_, err = slot.ReadFrom(bytes.NewReader([]byte{0x01, 0x16, 0x01, 0x00, 0x00, 0x00}))
	if err != nil {
		t.Error(err)
		return
	}

	assert.False(t, slot.IsEmpty())
	assert.Equal(t, int16(278), slot.BlockID)
	assert.Equal(t, int8(1), slot.ItemCount)
	assert.Equal(t, int16(0), slot.ItemDamage)
}
