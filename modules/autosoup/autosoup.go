package autosoup

import (
	"errors"
	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

var HotbarSlots = []int{36, 37, 38, 39, 40, 41, 42, 43, 44}

const (
	SoupID   = 282
	SoupSlot = 44
)

type AutoSoup struct {
	modules.SimpleModule
	MinHealth float64 `option:"minHealth"`
}

func (a *AutoSoup) GetIdentifier() string {
	return modules.ModuleAutoSoup
}

func (a *AutoSoup) GetDescription() []string {
	return []string{"Автоматически использует супы на KitPvP"}
}

func (a *AutoSoup) GetSlotWithSoup() int {
	inventory, ok := a.Tunnel.GetInventoryHandler().GetWindow(0)
	if !ok {
		return 0
	}

	for _, slot := range HotbarSlots {
		if inventory.GetItem(slot).BlockID == SoupID {
			return slot
		}
	}

	// If there was no slot with soup
	return a.PrepareSoup()
}

func (a *AutoSoup) PrepareSoup() int {
	inventory, ok := a.Tunnel.GetInventoryHandler().GetWindow(0)
	if !ok {
		return 0
	}

	soupSlot := a.FindSoup()
	if soupSlot == 0 {
		return 0
	}

	err := inventory.Move(soupSlot, SoupSlot)
	if err != nil {
		return 0
	}

	return SoupSlot
}

func (a *AutoSoup) FindSoup() int {
	inventory, ok := a.Tunnel.GetInventoryHandler().GetWindow(0)
	if !ok {
		return 0
	}

	for slot, item := range inventory.GetContents() {
		if item.BlockID == SoupID {
			return slot
		}
	}

	return 0
}

func (a *AutoSoup) UseSoup(slot int) error {
	playerHandler := a.Tunnel.GetPlayerHandler()
	err := playerHandler.ChangeSlot(slot - 36)
	if err != nil {
		return err
	}

	inventory, ok := a.Tunnel.GetInventoryHandler().GetWindow(0)
	if !ok {
		return errors.New("inventory is not available")
	}

	use := protocol.PlayerBlockPlacement{
		Location:        playerHandler.GetLocation().ToPosition(),
		HeldItem:        inventory.GetItem(slot),
		Face:            -1,
		CursorPositionX: -1,
		CursorPositionY: -1,
		CursorPositionZ: -1,
	}

	err = a.Tunnel.WriteServer(use.Marshal())
	if err != nil {
		return err
	}

	return playerHandler.ChangeSlot(playerHandler.GetCurrentSlot())
}

func HandleUpdateHealth(_ protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleAutoSoup) {
		module, _ := tunnel.GetModuleHandler().GetModule(modules.ModuleAutoSoup)
		autoSoup := module.(*AutoSoup)
		if tunnel.GetPlayerHandler().GetHealth() < autoSoup.MinHealth {
			slot := autoSoup.GetSlotWithSoup()

			err = autoSoup.UseSoup(slot)
			if err != nil {
				return nil, err
			}
		}
	}

	return generic.PassPacket(), nil
}
