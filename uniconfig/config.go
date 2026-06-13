package uniconfig

import (
	"encoding/json"

	"github.com/openkaze/uni"
)

// Configurable 允许模块动态参与自身的配置解析
type Configurable interface {
	Configure(ctx *ConfigContext, raw json.RawMessage) error
}

// ConfigContext 专用于第一阶段：动态解析和子模块加载
type ConfigContext struct {
	runtime *uni.Runtime
}
