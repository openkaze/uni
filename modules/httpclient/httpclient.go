package httpclient

import (
	"encoding/json"
	"fmt"

	"github.com/openkaze/uni"
)

func init() {
	uni.RegisterModule(new(HttpClientModule))
}

type HttpClientModule struct {
	Tag         string `json:"tag"`
	DnsProvider string `json:"dns_provider"` // 存储方案 A 双引号关联的字符串别名
}

func (h *HttpClientModule) UniModule() uni.ModuleInfo {
	return uni.ModuleInfo{
		ID:  "http_client",
		New: func() uni.Module { return new(HttpClientModule) },
	}
}

func (h *HttpClientModule) Configure(ctx *uni.ConfigContext, raw json.RawMessage) error {
	return uni.ConfigureModule(ctx, h, raw)
}

// Provision Task 2 & Task 3: 在第二阶段执行强类型图状挂载安全门禁校验
func (h *HttpClientModule) Provision(ctx *uni.Context) error {
	if h.DnsProvider != "" {
		// 高亮执法：跨枝校验绑定的组件是否归属于 "dns.forwarder" 命名空间
		inst, err := ctx.FindInstance("dns.forwarder", h.DnsProvider)
		if err != nil {
			return fmt.Errorf("http_client [%s] dependency bind error: %w", h.Tag, err)
		}
		fmt.Printf("[HttpClient][%s] -> 延迟绑定方案 A 指针成功! 目标 [%s] 实际注册身份为: (%s)\n",
			h.Tag, h.DnsProvider, inst.UniModule().ID)
	}
	return nil
}
