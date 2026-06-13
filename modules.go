package uni

import (
	"sync"
)

type Module interface {
	UniModule() ModuleInfo
}

// module id
type ModuleID string

// module runtime tag
type ModuleTag string

type ModuleInfo struct {
	ID  ModuleID
	New func() Module
}

type ModuleInstance struct {
	Tag ModuleTag
	Mod Module
}

func RegisterModule(info ModuleInfo) {
	registryMu.Lock()
	defer registryMu.Unlock()
	moduleRegistry[info.ID] = &info
}

// Provisioner 允许组件在启动时去容器里“捞”其他依赖
type Provisioner interface {
	Provision(ctx *Context) error
}

type Validator interface {
	Validate() error
}

type CleanerUpper interface {
	Cleanup() error
}

// 全局的模块注册表（只存类型定义，不存运行实例）
var (
	moduleRegistry = make(map[ModuleID]*ModuleInfo)
	registryMu     sync.RWMutex
)
