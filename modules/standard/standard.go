// Package standard imports all the standard Uni modules.
//
// This file exists purely to register modules via their init() functions.
package standard

import (
	_ "github.com/openkaze/uni/modules/dns"
	_ "github.com/openkaze/uni/modules/httpclient"
	_ "github.com/openkaze/uni/modules/log"
)
