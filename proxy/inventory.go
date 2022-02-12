package proxy

import (
	"sync"

	"github.com/Tnze/go-mc/chat"
	"github.com/destructiqn/kogtevran/generic"
	pk "github.com/destructiqn/kogtevran/net/packet"
	"github.com/destructiqn/kogtevran/protocol"
)

type InventoryHandler struct {
	windows map[int]generic.Window
	sync.Mutex
}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{windows: map[int]generic.Window{
		0: NewWindow(44, "", chat.Text("Player Inventory")),
	}}
}

func (i *InventoryHandler) GetWindows() []generic.Window {
	i.Lock()
	defer i.Unlock()

	windows := make([]generic.Window, 0)
	for _, window := range i.windows {
		windows = append(windows, window)
	}

	return windows
}

func (i *InventoryHandler) GetWindow(id int) (generic.Window, bool) {
	window, ok := i.windows[id]
	return window, ok
}

func (i *InventoryHandler) OpenWindow(id int, window generic.Window) {
	i.Lock()
	defer i.Unlock()
	i.windows[id] = window
}

func (i *InventoryHandler) CloseWindow(id int) {
	if id != 0 {
		i.Lock()
		defer i.Unlock()
		delete(i.windows, id)
	}
}

func (i *InventoryHandler) Reset() {
	i.Lock()
	defer i.Unlock()
	for id := range i.windows {
		if id == 0 {
			continue
		}

		delete(i.windows, id)
	}
}

type Window struct {
	size  int
	wType string
	title chat.Message
	items map[int]pk.Slot
	sync.Mutex
}

func NewWindow(size int, wType string, title chat.Message) *Window {
	return &Window{size: size, wType: wType, title: title, items: make(map[int]pk.Slot)}
}

func (w *Window) GetType() string {
	return w.wType
}

func (w *Window) GetSize() int {
	return w.size
}

func (w *Window) GetTitle() chat.Message {
	return w.title
}

func (w *Window) GetContents() map[int]pk.Slot {
	return w.items
}

func (w *Window) PutItem(slot int, item pk.Slot) {
	w.Lock()
	defer w.Unlock()
	w.items[slot] = item
}

func (w *Window) GetItem(slot int) pk.Slot {
	return w.items[slot]
}

func HandleOpenWindow(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	openWindow := packet.(*protocol.OpenWindow)

	tunnel.GetInventoryHandler().OpenWindow(int(openWindow.WindowID), NewWindow(
		int(openWindow.NumberOfSlots), string(openWindow.WindowType), openWindow.WindowTitle),
	)

	return openWindow.Marshal(), true, nil
}

func HandleCloseWindow(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	switch packet.(type) {
	case *protocol.ServerCloseWindow:
		closeWindow := packet.(*protocol.ServerCloseWindow)
		tunnel.GetInventoryHandler().CloseWindow(int(closeWindow.WindowID))
	case *protocol.CloseWindow:
		closeWindow := packet.(*protocol.CloseWindow)
		tunnel.GetInventoryHandler().CloseWindow(int(closeWindow.WindowID))
	}

	return packet.Marshal(), true, nil
}

func HandleSetSlot(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	setSlot := packet.(*protocol.SetSlot)
	window, ok := tunnel.GetInventoryHandler().GetWindow(int(setSlot.WindowID))
	if ok {
		window.PutItem(int(setSlot.Slot), setSlot.SlotData)
	}

	return packet.Marshal(), true, nil
}

func HandleWindowItems(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	windowItems := packet.(*protocol.WindowItems)
	window, ok := tunnel.GetInventoryHandler().GetWindow(int(windowItems.WindowID))
	if ok {
		for slot, item := range windowItems.SlotData {
			window.PutItem(slot, item)
		}
	}

	return packet.Marshal(), true, nil
}
