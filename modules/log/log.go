package log

import (
	"encoding/json"

	"github.com/openkaze/uni"
)

func init() {
	uni.RegisterModule(new(LogModule))
}

type LogModule struct {
	Tag   string `json:"tag"`
	Level string `json:"level"`
}

func (l *LogModule) UniModule() uni.ModuleInfo {
	return uni.ModuleInfo{
		ID:  "log",
		New: func() uni.Module { return new(LogModule) },
	}
}

func (l *LogModule) Configure(ctx *uni.ConfigContext, raw json.RawMessage) error {
	return uni.ConfigureModule(ctx, l, raw)
}
