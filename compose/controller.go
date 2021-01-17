package compose

import (
	"log"
	"sync"
)

type Controller struct {
	Cmds []*Cmd

	wg      sync.WaitGroup
	mu      sync.Mutex
	killing bool
}

func NewController(configs []*CmdConfig) *Controller {
	cmds := make([]*Cmd, 0)
	for _, config := range configs {
		cmd := NewCmd(config)
		cmds = append(cmds, cmd)
	}

	return &Controller{
		Cmds: cmds,
	}
}

func (ct *Controller) Run() {
	for _, cmd := range ct.Cmds {
		ct.mu.Lock()
		if ct.killing {
			ct.mu.Unlock()
			break
		}
		ct.mu.Unlock()

		if err := cmd.Start(); err != nil {
			log.Printf("%v cmd.Start error: %v\n", cmd.Config.Name, err)
			ct.Kill()
			break
		}

		if err := cmd.Ready(); err != nil {
			log.Printf("%v cmd.Ready error: %v\n", cmd.Config.Name, err)
			ct.Kill()
			break
		}

		ct.wg.Add(1)
		go func(cmd *Cmd) {
			cmd.Wait()
			ct.wg.Done()
			ct.Kill()
		}(cmd)
	}

	ct.wg.Wait()
}

func (ct *Controller) Kill() {
	ct.mu.Lock()
	if ct.killing {
		ct.mu.Unlock()
		return
	}
	ct.killing = true
	ct.mu.Unlock()

	for i := len(ct.Cmds) - 1; i >= 0; i-- {
		ct.Cmds[i].Kill()
	}
}
