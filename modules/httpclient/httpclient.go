package httpclient

import (
	"encoding/json"
	"fmt"
)

// ===================================================
// 第二种：HttpClientModule（不写 Tag，但强引用配置文件的其他 Tag）
// ===================================================
type HttpClientModule struct {
	DnsProvider ModuleTag        `json:"dns_provider"`
	dnsRef      *ForwarderModule // 最终拉取到的依赖指针
}

func (m *HttpClientModule) UniModule() ModuleInfo {
	return ModuleInfo{ID: "http_client", New: func() Module { return &HttpClientModule{} }}
}

func (m *HttpClientModule) Configure(ctx *ConfigContext, raw json.RawMessage) error {
	return json.Unmarshal(raw, m)
}

// 在第二阶段实施强引用拉取
func (m *HttpClientModule) Provision(ctx *Context) error {
	ref, err := ctx.FindInstance(m.DnsProvider)
	if err != nil {
		return fmt.Errorf("http_client configuration error: referenced tag '%s' does not exist", m.DnsProvider)
	}
	forwarder, ok := ref.(*ForwarderModule)
	if !ok {
		return fmt.Errorf("http_client error: tag '%s' is not a dns.forwarder", m.DnsProvider)
	}
	m.dnsRef = forwarder
	fmt.Printf("[HttpClientModule] 🌲 跨越树枝引用成功！已精准绑定到全局提升的子模块 [%s] (地址: %s)\n", m.DnsProvider, forwarder.Addr)
	return nil
}
