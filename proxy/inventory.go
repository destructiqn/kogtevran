package proxy

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/destructiqn/kogtevran/generic"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
)

type InventoryHandler struct {
	tunnel  *MinecraftTunnel
	windows map[int]generic.Window
	sync.Mutex
}

func NewInventoryHandler(tunnel *MinecraftTunnel) *InventoryHandler {
	handler := &InventoryHandler{tunnel: tunnel}
	handler.windows = map[int]generic.Window{
		0: NewWindow(handler, 0, 44, "", chat.Text("Player Inventory")),
	}
	return handler
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
	handler *InventoryHandler
	id      byte
	size    int
	wType   string
	title   chat.Message
	items   map[int]pk.Slot
	sync.Mutex
}

func NewWindow(handler *InventoryHandler, id byte, size int, wType string, title chat.Message) *Window {
	return &Window{handler: handler, id: id, size: size, wType: wType, title: title, items: make(map[int]pk.Slot)}
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

func (w *Window) Click(slot int, mode, button byte) error {
	tunnel := w.handler.tunnel
	rand.Seed(time.Now().UnixNano())

	packet := protocol.ClickWindow{
		WindowID:     pk.UnsignedByte(w.id),
		Slot:         pk.Short(slot),
		Button:       pk.Byte(button),
		ActionNumber: pk.Short(rand.Intn(100000)),
		Mode:         pk.Byte(mode),
		ClickedItem:  w.GetItem(slot),
	}

	return tunnel.WriteServer(packet.Marshal())
}

func (w *Window) Move(from, to int) error {
	w.Lock()
	defer w.Unlock()

	if w.GetItem(from).BlockID == -1 {
		return nil
	}

	err := w.Click(from, 0, 0)
	if err != nil {
		return err
	}

	err = w.Click(to, 0, 0)
	if err != nil {
		return err
	}

	if w.GetItem(to).BlockID != -1 {
		err = w.Click(from, 0, 0)
		if err != nil {
			return err
		}
	}

	w.items[from], w.items[to] = w.items[to], w.items[from]

	tunnel := w.handler.tunnel
	updateSource := protocol.SetSlot{
		WindowID: pk.Byte(w.id),
		Slot:     pk.Short(from),
		SlotData: w.GetItem(from),
	}

	err = tunnel.WriteClient(updateSource.Marshal())
	if err != nil {
		return err
	}

	updateDestination := protocol.SetSlot{
		WindowID: pk.Byte(w.id),
		Slot:     pk.Short(to),
		SlotData: w.GetItem(to),
	}

	err = tunnel.WriteClient(updateDestination.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func HandleOpenWindow(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	openWindow := packet.(*protocol.OpenWindow)

	inventoryHandler := tunnel.GetInventoryHandler()
	inventoryHandler.OpenWindow(int(openWindow.WindowID), NewWindow(
		inventoryHandler.(*InventoryHandler), byte(openWindow.WindowID), int(openWindow.NumberOfSlots),
		string(openWindow.WindowType), openWindow.WindowTitle),
	)

	return generic.PassPacket(), nil
}

func HandleCloseWindow(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	switch packet.(type) {
	case *protocol.ServerCloseWindow:
		closeWindow := packet.(*protocol.ServerCloseWindow)
		tunnel.GetInventoryHandler().CloseWindow(int(closeWindow.WindowID))
	case *protocol.CloseWindow:
		closeWindow := packet.(*protocol.CloseWindow)
		tunnel.GetInventoryHandler().CloseWindow(int(closeWindow.WindowID))
	}

	return generic.PassPacket(), nil
}

func HandleSetSlot(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	setSlot := packet.(*protocol.SetSlot)
	window, ok := tunnel.GetInventoryHandler().GetWindow(int(setSlot.WindowID))
	if ok {
		window.PutItem(int(setSlot.Slot), setSlot.SlotData)
	}

	return generic.PassPacket(), nil
}

func HandleWindowItems(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	windowItems := packet.(*protocol.WindowItems)
	window, ok := tunnel.GetInventoryHandler().GetWindow(int(windowItems.WindowID))
	if ok {
		for slot, item := range windowItems.SlotData {
			window.PutItem(slot, item)
		}
	}

	return generic.PassPacket(), nil
}
