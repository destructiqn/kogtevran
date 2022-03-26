package nuker

import (
	"sync"
	"time"

	"github.com/destructiqn/kogtevran/generic"
	"github.com/destructiqn/kogtevran/minecraft"
	"github.com/destructiqn/kogtevran/minecraft/blocks"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
	"github.com/destructiqn/kogtevran/minecraft/protocol"
	"github.com/destructiqn/kogtevran/modules"
)

type Nuker struct {
	modules.SimpleTickingModule
	Radius int     `option:"radius"`
	Delay  float64 `option:"delay"`

	queueLock   sync.Mutex
	backlog     map[pk.Position]bool
	breakQueue  chan *Task
	toggleQueue chan bool
}

func (n *Nuker) GetDescription() []string {
	return []string{
		"Все блоки вокруг тебя ломаются автоматически",
		"",
		"§nПараметры",
		"§7radius§f - радиус, в котором нужно ломать блоки",
		"§7interval§f - интервал между циклами ломания блоков",
		"§7delay§f - коэффициент скорости ломания блоков",
	}
}

func (n *Nuker) GetIdentifier() string {
	return modules.ModuleNuker
}

func (n *Nuker) Toggle() (bool, error) {
    v, err := n.SimpleTickingModule.Toggle()
    if n.toggleQueue != nil {
        n.toggleQueue <- v
    }
    return v, err
}

func (n *Nuker) Tick() error {
	center := n.Tunnel.GetPlayerHandler().GetLocation()

	if n.breakQueue == nil {
		n.breakQueue = make(chan *Task)
		n.toggleQueue = make(chan bool)
		n.backlog = make(map[pk.Position]bool)
		go n.handleQueue()
	}

	for x := int(center.X) - n.Radius; x <= int(center.X)+n.Radius; x++ {
		for y := int(center.Y) - n.Radius; y <= int(center.Y)+n.Radius; y++ {
			for z := int(center.Z) - n.Radius; z <= int(center.Z)+n.Radius; z++ {
				position := pk.Position{X: x, Y: y, Z: z}
				n.queueLock.Lock()
				
				if _, ok := n.backlog[position]; !ok {
				  BreakBlock(position, 0, n.Tunnel)
				}
								
				n.queueLock.Unlock()
			}
		}
	}

	return nil
}

func HandleBlockChange(packet protocol.Packet, tunnel generic.Tunnel) (result *generic.HandlerResult, err error) {
	blockChange := packet.(*protocol.BlockChange)

	if tunnel.GetModuleHandler().IsModuleEnabled(modules.ModuleNuker) {
		nuker, _ := tunnel.GetModuleHandler().GetModule(modules.ModuleNuker)

		location := blockChange.Location
		playerLocation := tunnel.GetPlayerHandler().GetLocation()
		blockLocation := &minecraft.Location{X: float64(location.X), Y: float64(location.Y), Z: float64(location.Z)}

		if playerLocation.Distance(blockLocation) > float64(nuker.(*Nuker).Radius) {
			return generic.PassPacket(), nil
		}

		id := blockChange.BlockID >> 4
		block, ok := blocks.ByID[blocks.ID(id)]
		if !ok || !block.Diggable {
			return generic.PassPacket(), nil
		}

		go nuker.(*Nuker).enqueue(&Task{
			Location: blockChange.Location,
			Delay:    time.Duration(block.Hardness*nuker.(*Nuker).Delay*1000) * time.Millisecond,
		})
	}

	return generic.PassPacket(), nil
}

func BreakBlock(location pk.Position, delay time.Duration, tunnel generic.Tunnel) {
	start := &protocol.PlayerDigging{
		Face:     1,
		Status:   0,
		Location: location,
	}

	finish := &protocol.PlayerDigging{
		Face:     1,
		Status:   2,
		Location: location,
	}

	_ = tunnel.WriteServer(start.Marshal())
	time.Sleep(delay)
	_ = tunnel.WriteServer(finish.Marshal())
}
