package dns

import (
	"encoding/json"

	"github.com/openkaze/uni"
)

type ForwarderModule struct {
	Tag  uni.ModuleTag `json:"tag"`
	Type string        `json:"type" uni:"namespace=uni.dns.forwarder inline_key=type"`
	Addr string        `json:"addr"`
}

func (m *ForwarderModule) UniModule() uni.ModuleInfo {
	return uni.ModuleInfo{ID: "dns.forwarder", New: func() uni.Module { return &ForwarderModule{} }}
}

func (m *ForwarderModule) Configure(ctx *uni.ConfigContext, raw json.RawMessage) error {
	return json.Unmarshal(raw, m)
}
