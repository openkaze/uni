package uni

import (
	"encoding/json"
	"fmt"
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

// Provisioner 允许组件在启动时去容器里“捞”其他依赖
type Provisioner interface {
	Provision(ctx *Context) error
}

func RegisterModule(info ModuleInfo) {
	registryMu.Lock()
	defer registryMu.Unlock()
	moduleRegistry[info.ID] = &info
}

// LoadSubModule 供父模块调用，动态解析并“提升”子模块到全局
func (c *ConfigContext) LoadSubModule(subModID ModuleID, raw json.RawMessage) (Module, ModuleTag, error) {
	registryMu.RLock()
	info, exists := moduleRegistry[subModID]
	registryMu.RUnlock()
	if !exists {
		return nil, "", fmt.Errorf("unknown sub-module id: %s", subModID)
	}

	// 1. 提取子模块的显示 Tag
	var meta struct {
		Tag ModuleTag `json:"tag"`
	}
	_ = json.Unmarshal(raw, &meta)

	if meta.Tag == "" {
		return nil, "", fmt.Errorf("config error: sub-module of type '%s' must have an explicit 'tag'", subModID)
	}

	// 🛑 触发你的约束：检查子模块是否霸占了顶级的隐式预留 Tag
	if c.runtime.PreDefineTags[meta.Tag] {
		return nil, "", fmt.Errorf("security error: sub-module tag '%s' collides with top-level pre-defined tag", meta.Tag)
	}

	// 2. 实例化并解析子模块自身
	subMod := info.New()
	if cfg, ok := subMod.(Configurable); ok {
		if err := cfg.Configure(c, raw); err != nil {
			return nil, "", err
		}
	}

	// 3. 【核心提升】直接打平塞入全局实例表
	if _, dup := c.runtime.Instances[meta.Tag]; dup {
		return nil, "", fmt.Errorf("duplicate instance tag: %s", meta.Tag)
	}
	c.runtime.Instances[meta.Tag] = &ModuleInstance{Tag: meta.Tag, Mod: subMod}

	return subMod, meta.Tag, nil
}

func (c *Context) FindInstance(tag ModuleTag) (Module, error) {
	c.runtime.mu.RLock()
	defer c.runtime.mu.RUnlock()
	inst, ok := c.runtime.Instances[tag]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", tag)
	}
	return inst.Mod, nil
}

// 全局的模块注册表（只存类型定义，不存运行实例）
var (
	moduleRegistry = make(map[ModuleID]*ModuleInfo)
	registryMu     sync.RWMutex
)
