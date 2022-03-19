package cheststealer

import (
	"strings"

	"github.com/destructiqn/kogtevran/generic"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

type ChestStealer struct {
	modules.SimpleModule
}

func (c *ChestStealer) GetDescription() []string {
	return []string{"Автоматически забирай все предметы из сундуков"}
}

func (c *ChestStealer) GetIdentifier() string {
	return modules.ModuleChestStealer
}

func HandleOpenWindow(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	openWindow := packet.(*protocol.OpenWindow)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleChestStealer) {
		window, ok := tunnel.GetInventoryHandler().GetWindow(int(openWindow.WindowID))
		if ok && IsSuitable(window) {
			return pk.Packet{}, false, nil
		}
	}

	return packet.Marshal(), true, nil
}

func HandleSetSlot(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	setSlot := packet.(*protocol.SetSlot)
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleChestStealer) {
		window, ok := tunnel.GetInventoryHandler().GetWindow(int(setSlot.WindowID))
		if ok && IsSuitable(window) {
			err = tunnel.WriteServer(TakeItem(int(setSlot.WindowID), int(setSlot.Slot), setSlot.SlotData).Marshal())
			if err != nil {
				return
			}

			return pk.Packet{}, false, nil
		}
	}

	return setSlot.Marshal(), true, nil
}

func HandleWindowItems(packet protocol.Packet, tunnel generic.Tunnel) (result pk.Packet, next bool, err error) {
	windowItems := packet.(*protocol.WindowItems)
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleChestStealer) {
		window, ok := tunnel.GetInventoryHandler().GetWindow(int(windowItems.WindowID))
		if ok && IsSuitable(window) {
			for slot, item := range window.GetContents() {
				if slot >= window.GetSize()-36 {
					break
				}

				if item.BlockID == -1 {
					continue
				}

				err = tunnel.WriteServer(TakeItem(int(windowItems.WindowID), slot, item).Marshal())
				if err != nil {
					return
				}
			}

			return pk.Packet{}, false, nil
		}
	}

	return windowItems.Marshal(), true, nil
}

func TakeItem(windowID, slot int, item pk.Slot) protocol.Packet {
	return &protocol.ClickWindow{
		WindowID:     pk.UnsignedByte(windowID),
		Slot:         pk.Short(slot),
		ActionNumber: pk.Short(slot),
		Mode:         1,
		ClickedItem:  item,
	}
}

func IsSuitable(window generic.Window) bool {
	return window.GetType() == "minecraft:chest" && strings.Contains(strings.ToLower(window.GetTitle().String()), "chest")
}
