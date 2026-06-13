package uni

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Runtime 容器
type Runtime struct {
	Instances     map[ModuleTag]*ModuleInstance
	PreDefineTags map[ModuleTag]bool // 存放顶级无 Tag 模块的隐式 Tag
}

func (r *Runtime) FindInstance(tag ModuleTag) (Module, error) {
	inst, ok := r.Instances[tag]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", tag)
	}
	return inst.Mod, nil
}

func (r *Runtime) Start() {}

func (r *Runtime) Stop(ctx context.Context) {}

// Controller 负责管理运行时的生命周期和热重载
type Controller struct {
	// 使用原子指针，保证并发读写安全
	current atomic.Pointer[Runtime]
}

// GetRuntime 供业务协程调用，获取当前的运行时（无锁，极快）
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
