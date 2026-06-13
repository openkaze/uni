# runtime.go

```go
package uni

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Runtime struct {
	mu        sync.RWMutex
	Instances map[ModuleTag]*ModuleInstance
}

// Controller 负责管理运行时的生命周期和热重载
type Controller struct {
	// 使用原子指针，保证并发读写安全
	current atomic.Pointer[Runtime]
}

// GetRuntime 供业务协程调用，获取当前的运行时（无锁，极快）
func (c *Controller) GetRuntime() *Runtime {
	return c.current.Load()
}

// Reload 触发热重载
func (c *Controller) Reload(configData []byte) error {
	// 1. 创建全新的运行时（局部变量）
	newRuntime := &Runtime{
		// 初始化新运行时的各项数据...
	}

	// 2. 初始化新运行时的组件（假设有这个方法）
	if err := newRuntime.Init(configData); err != nil {
		return err
	}

	// 3. 启动新运行时，开始监听流量
	newRuntime.Start()

	// 4. 【核心一步】利用原子操作，瞬间切换指针
	// 这之后，所有新来的请求调用 GetRuntime() 拿到的都是 newRuntime
	oldRuntime := c.current.Swap(newRuntime)

	// 5. 优雅关闭老运行时（局部变量）
	if oldRuntime != nil {
		go func() {
			// 给老运行时一段缓冲时间，处理完残余流量后关闭
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			oldRuntime.Stop(ctx)
			// 函数结束后，oldRuntime 没有任何引用，会被 GC 完美回收
		}()
	}

	return nil
}

```

# config_init.go

```go
package uni

import (
	"encoding/json"
	"fmt"
)

// 对应 JSON 的临时结构体
type ConfigBlock struct {
	Modules []struct {
		Tag    ModuleTag       `json:"tag"`
		ID     ModuleID        `json:"id"`
		Config json.RawMessage `json:"config"` // 留给各个组件自己解析
	} `json:"modules"`
}

func (r *Runtime) Init(configData []byte) error {
	r.Instances = make(map[ModuleTag]*ModuleInstance)

	// 1. 解析外层配置文件
	var cfg ConfigBlock
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return fmt.Errorf("parse config failed: %w", err)
	}

	// 2. 遍历配置，去全局注册表里找对应的 ModuleInfo 并实例化
	for _, block := range cfg.Modules {
		// 假设 globalRegistry 是我们在上一步定义的全局类型注册表
		info, exists := moduleRegistry[block.ID]
		if !exists {
			return fmt.Errorf("unknown module id: %s", block.ID)
		}

		// 调用你定义的 New() 函数，创建一个干净的组件对象
		newMod := info.New()

		// (可选) 如果组件有自己的 UnmarshalJSON，可以把 block.Config 塞给它
		// if jsonMod, ok := newMod.(json.Unmarshaler); ok { ... }

		// 3. 塞进当前这个 newRuntime 的 map 里
		r.Instances[block.Tag] = &ModuleInstance{
			Tag: block.Tag,
			Mod: newMod,
		}
	}

	return nil
}

```

# context.go

```go
package uni

// Context 传递给组件的上下文，避免组件直接操作 Runtime 的锁
type Context struct {
	runtime *Runtime
}

```

# modules.go

```go
package uni

import (
	"context"
	"fmt"
	"sync"
)

type Module interface {
	UniModule() ModuleInfo
}

type ModuleID string

type ModuleInfo struct {
	ID  ModuleID
	New func() Module
}

type ModuleTag string

type ModuleInstance struct {
	Tag ModuleTag
	Mod Module
}

// Provisioner 允许组件在启动时去容器里“捞”其他依赖
type Provisioner interface {
	Provision(ctx *Context) error
}

// 全局的模块注册表（只存类型定义，不存运行实例）
var moduleRegistry = make(map[ModuleID]*ModuleInfo)
var registryMu sync.RWMutex

func (c *Context) FindInstance(tag ModuleTag) (Module, error) {
	c.runtime.mu.RLock()
	defer c.runtime.mu.RUnlock()
	inst, ok := c.runtime.Instances[tag]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", tag)
	}
	return inst.Mod, nil
}

// InitAll 模拟两阶段加载中的第二阶段：依赖注入
func (r *Runtime) InitAll() error {
	ctx := &Context{runtime: r}

	for _, inst := range r.Instances {
		// 如果组件实现了 Provisioner 接口，就触发它去寻找依赖
		if provisioner, ok := inst.Mod.(Provisioner); ok {
			if err := provisioner.Provision(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Runtime) Start() {}

func (r *Runtime) Stop(ctx context.Context) {}

```

# cmd/main.go

```go
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"uni"
)

// 1. 定义一个全局的控制器（整个进程只有这一个 Controller）
var AppController uni.Controller

func Main() {
	// 2. 首次启动：读取配置文件并初始化
	firstConfig, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	// 这一步执行完，AppController 内部的 current 指针就指向了 Runtime v1
	if err := AppController.Reload(firstConfig); err != nil {
		panic(err)
	}

	// 3. 启动业务 HTTP 服务
	http.HandleFunc("/user/profile", UserProfileHandler)
	go http.ListenAndServe(":8080", nil)

	// 4. 监听外部信号（比如通过另一个控制 API，或者监听文件变动）
	// 当有人修改了 config.json 并触发重载时：
	http.HandleFunc("/-/reload", func(w http.ResponseWriter, r *http.Request) {
		newConfig, _ := os.ReadFile("config.json")

		// 核心：调用 Reload。里面会生成 Runtime v2，解析成功后，瞬间把指针切到 v2
		if err := AppController.Reload(newConfig); err != nil {
			w.Write([]byte("Reload failed: " + err.Error()))
			return
		}
		w.Write([]byte("Reload success!"))
	})

	select {} // 阻塞主线程
}

// 5. 具体的业务处理函数（每来一个 HTTP 请求，就会走这里）
func UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	// 【极其重要】每次处理请求，都去 Controller 拿“当前最新”的运行时！
	// 这样能保证如果刚刚发生了 Reload，这里拿到的立马就是最新的 Runtime v2
	rt := AppController.GetRuntime()

	// 从当前运行时中获取 main_db 实例
	dbInstance, exists := rt.Instances["main_db"]
	if !exists {
		http.Error(w, "Database not found", 500)
		return
	}

	// 使用 dbInstance 执行你的业务逻辑...
	fmt.Fprintf(w, "Using DB instance: %p\n", dbInstance.Mod)
}

```

# cmd/main/main.go

```go
package main

import "uni/cmd"

func main() {
	cmd.Main()
}

```
