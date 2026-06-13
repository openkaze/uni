package uni

import "sync/atomic"

type Controller struct {
	current atomic.Pointer[Runtime]
}

func (c *Controller) GetRuntime() *Runtime {
	return c.current.Load()
}

func (c *Controller) Reload(configData []byte) error {
	newRuntime := &Runtime{
		Instances:     make(map[ModuleTag]*ModuleInstance),
		PreDefineTags: make(map[ModuleTag]bool),
	}
	if err := newRuntime.Init(configData); err != nil {
		return err
	}
	if err := newRuntime.InitAll(); err != nil {
		return err
	}
	c.current.Swap(newRuntime)
	return nil
}
