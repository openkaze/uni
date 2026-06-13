package cmd

import (
	"fmt"

	"github.com/openkaze/uni"
)

var AppController uni.Controller

func Main() {
	// 你的目标测试 JSON
	configJSON := []byte(`
	{
		"log": {
			"level": "INFO"
		},
		"http_client": {
			"dns_provider": "google" 
		},
		"dns": {
			"cache": true,
			"forwarder": [
				{
					"tag": "google",
					"addr": "8.8.8.8"
				}
			]
		}
	}`)

	fmt.Println("=== 🛰️ 基于通用泛型 Map 开始加载 Runtime ===")
	if err := AppController.Reload(configJSON); err != nil {
		panic(err)
	}

	fmt.Println("\n=== 🔍 最终运行时打平拓扑校验 ===")
	rt := AppController.GetRuntime()
	for tag, inst := range rt.Instances {
		isPre := rt.PreDefineTags[tag]
		fmt.Printf("-> 全局可见 Tag: [%-12s] | 是否顶级隐式预留: %-5v | 内存指针: %p\n", tag, isPre, inst.Mod)
	}
}
