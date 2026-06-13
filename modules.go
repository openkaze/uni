package uni

import (
	"encoding/json"
	"sync"
)

type Module interface {
	UniModule() ModuleInfo
}

type ModuleID string
type ModuleTag string

type ModuleInfo struct {
	ID  ModuleID
	New func() Module
}

type ModuleInstance struct {
	Tag ModuleTag
	Mod Module
}

type Configurable interface {
	Configure(ctx *ConfigContext, raw json.RawMessage) error
}

// JSON 配置 → 反序列化成模块结构体 → 调用 Provision() 建立引用和内部状态(Ready) → (安全替换后) 调用 Start() 模块正式开始工作
type Provisioner interface {
	Provision(ctx *Context) error
}

var (
	moduleRegistry = make(map[ModuleID]*ModuleInfo)
	registryMu     sync.RWMutex
)

func RegisterModule(info ModuleInfo) {
	registryMu.Lock()
	defer registryMu.Unlock()
	moduleRegistry[info.ID] = &info
}
