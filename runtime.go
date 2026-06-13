package uni

import (
	"encoding/json"
	"fmt"
)

type Runtime struct {
	Instances     map[ModuleTag]*ModuleInstance
	PreDefineTags map[ModuleTag]bool
}

func (r *Runtime) FindInstance(tag ModuleTag) (Module, error) {
	inst, ok := r.Instances[tag]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", tag)
	}
	return inst.Mod, nil
}

// Init 核心动态解析：将顶级配置转为泛型 Map
func (r *Runtime) Init(configData []byte) error {
	var topLevelConfig map[string]json.RawMessage
	if err := json.Unmarshal(configData, &topLevelConfig); err != nil {
		return fmt.Errorf("invalid json format: %w", err)
	}

	cfgCtx := &ConfigContext{runtime: r}

	// 1. 预先登记顶级的隐式 Tag
	for key := range topLevelConfig {
		r.PreDefineTags[ModuleTag(key)] = true
	}

	// 2. 动态创建并配置顶级模块
	for key, raw := range topLevelConfig {
		modID := ModuleID(key)

		registryMu.RLock()
		info, exists := moduleRegistry[modID]
		registryMu.RUnlock()
		if !exists {
			return fmt.Errorf("unknown top-level module: %s", key)
		}

		mod := info.New()
		if cfg, ok := mod.(Configurable); ok {
			if err := cfg.Configure(cfgCtx, raw); err != nil {
				return fmt.Errorf("failed to configure module '%s': %w", key, err)
			}
		}

		finalTag := ModuleTag(modID)
		r.Instances[finalTag] = &ModuleInstance{Tag: finalTag, Mod: mod}
	}
	return nil
}

func (r *Runtime) InitAll() error {
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

func (r *Runtime) Start() {}
func (r *Runtime) Stop()  {}
