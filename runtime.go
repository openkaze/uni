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
