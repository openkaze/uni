package dns

import (
	"encoding/json"

	"github.com/openkaze/uni"
)

type DnsModule struct {
	Cache      bool `json:"cache"`
	Forwarders []*ForwarderModule
}

func (m *DnsModule) UniModule() uni.ModuleInfo {
	return uni.ModuleInfo{ID: "dns", New: func() uni.Module { return &DnsModule{} }}
}

func (m *DnsModule) Configure(ctx *uni.ConfigContext, raw json.RawMessage) error {
	var shadow struct {
		Cache     bool              `json:"cache"`
		Forwarder []json.RawMessage `json:"forwarder"`
	}
	if err := json.Unmarshal(raw, &shadow); err != nil {
		return err
	}
	m.Cache = shadow.Cache

	for _, fRaw := range shadow.Forwarder {
		// 直接通过 uni 的上下文动态加载并提升子模块
		subMod, _, err := ctx.LoadSubModule("dns.forwarder", fRaw)
		if err != nil {
			return err
		}
		m.Forwarders = append(m.Forwarders, subMod.(*ForwarderModule))
	}
	return nil
}
