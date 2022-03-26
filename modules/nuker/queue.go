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
		select {
		case task := <-n.breakQueue:
			BreakBlock(task.Location, task.Delay, n.Tunnel)

			n.queueLock.Lock()
			delete(n.backlog, task.Location)
			n.queueLock.Unlock()
		case status := <-n.toggleQueue:
			if !status {
				// If disabled, wait for enable
				<-n.toggleQueue
			}
		}
	}
}

func (n *Nuker) enqueue(task *Task) {
    n.queueLock.Lock()
    if _, ok := n.backlog[task.Location]; !ok && n.backlog != nil {
        n.backlog[task.Location] = true
        n.queueLock.Unlock()

        n.breakQueue <- task
    } else {
        n.queueLock.Unlock()
    }
}
