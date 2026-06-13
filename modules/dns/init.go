package dns

import "github.com/openkaze/uni"

func init() {
	uni.RegisterModule(new(DnsModule).UniModule())
	uni.RegisterModule(new(ForwarderModule).UniModule())
}
