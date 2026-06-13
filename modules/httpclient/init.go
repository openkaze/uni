package httpclient

import "github.com/openkaze/uni"

func init() {
	uni.RegisterModule(new(HttpClientModule).UniModule())
}
