package log

import "github.com/openkaze/uni"

func init() {
	uni.RegisterModule(new(LogModule).UniModule())
}
