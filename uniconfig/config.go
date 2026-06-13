package uniconfig

import (
	"encoding/json"
	"fmt"

	"github.com/openkaze/uni"
)

// Configurable 允许模块动态参与自身的配置解析
type Configurable interface {
	Configure(ctx *ConfigContext, raw json.RawMessage) error
}

// ConfigContext 专用于第一阶段：动态解析和子模块加载
type ConfigContext struct {
	runtime *uni.Runtime
}

// Init 彻底动态化：不再有 RootConfig 结构体！
func (r *uni.Runtime) Init(configData []byte) error {
	// 💡 核心改变：将顶级配置直接视为一个泛型 Map
	var topLevelConfig map[string]json.RawMessage
	if err := json.Unmarshal(configData, &topLevelConfig); err != nil {
		return fmt.Errorf("invalid json format: %w", err)
	}

	cfgCtx := &ConfigContext{runtime: r}

	// 1. 第一轮循环：预先登记所有顶级的隐式 Tag，建立保护区
	for key := range topLevelConfig {
		implicitTag := ModuleTag(key)
		r.PreDefineTags[implicitTag] = true
	}

	// 2. 第二轮循环：动态创建顶级模块
	for key, raw := range topLevelConfig {
		modID := ModuleID(key)

		registryMu.RLock()
		info, exists := moduleRegistry[modID]
		registryMu.RUnlock()
		if !exists {
			return fmt.Errorf("unknown top-level module: %s", key)
		}

		mod := info.New()

		// 如果模块支持自己解析配置
		if cfg, ok := mod.(uni.Configurable); ok {
			if err := cfg.Configure(cfgCtx, raw); err != nil {
				return fmt.Errorf("failed to configure module '%s': %w", key, err)
			}
		}

		// 顶级字段没有写 tag，触发你的规则：隐式 Tag = ModuleID
		finalTag := ModuleTag(modID)
		r.Instances[finalTag] = &ModuleInstance{Tag: finalTag, Mod: mod}
	}

	return nil
}

func (r *uni.Runtime) InitAll() error {
	ctx := &Context{runtime: r}
	for _, inst := range r.Instances {
		if provisioner, ok := inst.Mod.(Provisioner); ok {
			if err := provisioner.Provision(ctx); err != nil {
				return err
			}
		}
	}
	return nil
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
