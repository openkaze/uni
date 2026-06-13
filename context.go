package uni

import (
	"encoding/json"
	"fmt"
)

type Context struct {
	runtime *Runtime
}

func (c *Context) FindInstance(tag ModuleTag) (Module, error) { return c.runtime.FindInstance(tag) }

// ConfigContext 移入 uni 包，专门负责第一阶段子模块的打平与“提升”
type ConfigContext struct {
	runtime *Runtime
}

func (c *ConfigContext) LoadSubModule(subModID ModuleID, raw json.RawMessage) (Module, ModuleTag, error) {
	registryMu.RLock()
	info, exists := moduleRegistry[subModID]
	registryMu.RUnlock()
	if !exists {
		return nil, "", fmt.Errorf("unknown sub-module id: %s", subModID)
	}

	var meta struct {
		Tag ModuleTag `json:"tag"`
	}
	_ = json.Unmarshal(raw, &meta)

	if meta.Tag == "" {
		return nil, "", fmt.Errorf("config error: sub-module of type '%s' must have an explicit 'tag'", subModID)
	}

	if c.runtime.PreDefineTags[meta.Tag] {
		return nil, "", fmt.Errorf("security error: sub-module tag '%s' collides with top-level pre-defined tag", meta.Tag)
	}

	subMod := info.New()
	if cfg, ok := subMod.(Configurable); ok {
		if err := cfg.Configure(c, raw); err != nil {
			return nil, "", err
		}
	}

	if _, dup := c.runtime.Instances[meta.Tag]; dup {
		return nil, "", fmt.Errorf("duplicate instance tag: %s", meta.Tag)
	}
	c.runtime.Instances[meta.Tag] = &ModuleInstance{Tag: meta.Tag, Mod: subMod}

	return subMod, meta.Tag, nil
}
