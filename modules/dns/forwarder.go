package dns

import (
	"encoding/json"

	"github.com/openkaze/uni"
)

type ForwarderModule struct {
	Tag  uni.ModuleTag `json:"tag"`
	Addr string        `json:"addr"`
}

func (m *ForwarderModule) UniModule() uni.ModuleInfo {
	return uni.ModuleInfo{ID: "dns.forwarder", New: func() uni.Module { return &ForwarderModule{} }}
}
func (m *ForwarderModule) Configure(ctx *ConfigContext, raw json.RawMessage) error {
	return json.Unmarshal(raw, m)
}
