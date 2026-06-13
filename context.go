package uni

// Context 传递给组件的上下文，避免组件直接操作 Runtime 的锁
type Context struct {
	runtime *Runtime
}
