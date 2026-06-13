package log

import (
	"encoding/json"

	"github.com/openkaze/uni"
)

type LogModule struct {
	Level string `json:"level"`
}

func (m *LogModule) UniModule() uni.ModuleInfo {
	return uni.ModuleInfo{ID: "log", New: func() uni.Module { return &LogModule{} }}
}

func (m *LogModule) Configure(ctx *uni.ConfigContext, raw json.RawMessage) error {
	return json.Unmarshal(raw, m)
}
