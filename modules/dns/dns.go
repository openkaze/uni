package dns

import (
	"encoding/json"

	"github.com/openkaze/uni"
)

// 3.2 父 Module 定义
type DnsModule struct {
	Cache      bool               `json:"cache"`
	Forwarders []*ForwarderModule // 依然保持对子组件的感知
}

func (m *DnsModule) UniModule() ModuleInfo {
	return uni.ModuleInfo{ID: "dns", New: func() Module { return &DnsModule{} }}
}

// 核心：由父模块利用通用的 ConfigContext 来解析并提升子模块
func (m *DnsModule) Configure(ctx *ConfigContext, raw json.RawMessage) error {
	var shadow struct {
		Cache     bool              `json:"cache"`
		Forwarder []json.RawMessage `json:"forwarder"`
	}
	if err := json.Unmarshal(raw, &shadow); err != nil {
		return err
	}
	m.Cache = shadow.Cache

	// 动态解析嵌套的子盲盒
	for _, fRaw := range shadow.Forwarder {
		// 动态通过框架加载子组件，子组件会在底层自动被“提升”到全局的 Instances 中
		subMod, _, err := ctx.LoadSubModule("dns.forwarder", fRaw)
		if err != nil {
			return err
		}
		m.Forwarders = append(m.Forwarders, subMod.(*ForwarderModule))
	}
	return nil
}
