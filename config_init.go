package uni

import (
	"encoding/json"
	"fmt"
)

// 对应 JSON 的临时结构体
type ConfigBlock struct {
	Modules []struct {
		Tag    ModuleTag       `json:"tag"`
		ID     ModuleID        `json:"id"`
		Config json.RawMessage `json:"config"` // 留给各个组件自己解析
	} `json:"modules"`
}

func (r *Runtime) Init(configData []byte) error {
	r.Instances = make(map[ModuleTag]*ModuleInstance)

	// 1. 解析外层配置文件
	var cfg ConfigBlock
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return fmt.Errorf("parse config failed: %w", err)
	}

	// 2. 遍历配置，去全局注册表里找对应的 ModuleInfo 并实例化
	for _, block := range cfg.Modules {
		// 假设 globalRegistry 是我们在上一步定义的全局类型注册表
		info, exists := moduleRegistry[block.ID]
		if !exists {
			return fmt.Errorf("unknown module id: %s", block.ID)
		}

		// 调用你定义的 New() 函数，创建一个干净的组件对象
		newMod := info.New()

		// (可选) 如果组件有自己的 UnmarshalJSON，可以把 block.Config 塞给它
		// if jsonMod, ok := newMod.(json.Unmarshaler); ok { ... }

		// 3. 塞进当前这个 newRuntime 的 map 里
		r.Instances[block.Tag] = &ModuleInstance{
			Tag: block.Tag,
			Mod: newMod,
		}
	}

	return nil
}
