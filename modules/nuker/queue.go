package nuker

import (
	"time"

	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
)

type Task struct {
	Location pk.Position
	Delay    time.Duration
}

func (n *Nuker) handleQueue() {
	for {
		task := <-n.breakQueue
		BreakBlock(task.Location, task.Delay, n.Tunnel)

		n.queueLock.Lock()
		delete(n.backlog, task.Location)
		n.queueLock.Unlock()
	}
}

func (n *Nuker) enqueue(task *Task) {
	if _, ok := n.backlog[task.Location]; !ok && n.backlog != nil {
		n.queueLock.Lock()
		n.backlog[task.Location] = true
		n.queueLock.Unlock()

		n.breakQueue <- task
	}
}