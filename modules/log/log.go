package log

import (
	"encoding/json"

	"github.com/openkaze/uni"
	"github.com/openkaze/uni/uniconfig"
)

type LogModule struct {
	Level string `json:"level"`
}

func (m *LogModule) UniModule() uni.ModuleInfo {
	return uni.ModuleInfo{ID: "log", New: func() uni.Module { return &LogModule{} }}
}

func (m *LogModule) Configure(ctx *uniconfig.ConfigContext, raw json.RawMessage) error {
	return json.Unmarshal(raw, m) // 自己管自己的解析
}
