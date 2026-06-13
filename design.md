这是我设计的 module 框架雏形

# Overview

- 参考 caddy 的 module 框架
- caddy 的 module 框架，没有运行时 ID ，同一个 Module 如果有多个实例，几乎不可在配置文件中引用
- 因为 caddy 是直接从 config json 生成对应对象，配置文件不允许引用
- 我们讨论设计了 Runtime 这种东西来解决问题

# 设计约束

## 1 配置文件不可引用的问题

参见 Overview

## 2 允许 热重载

- 实现无锁、高性能、并发安全的热重载
- Go 语言提供了 sync/atomic 包
- 实现

## 3 灵活的配置文件

- 配置文件是 JSON 格式
- 允许不同的 Module 自定义解析器

# 当前任务 

我们要设计三个 Module 用于测试

我们不实现 Module 的具体功能，或者说提供假功能

- 第一种：配置文件不写 Tag
- 第二种：配置文件写Tag，引用配置文件的 Tag
- 第三种：父 Module 不写 tag 但是子 Module 写 Tag

```json
{
    "log": {
        "level": "INFO"
    },
    "http_client": {
        "dns_provider": "google" // 🌲 跨越树枝，直接引用！
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
}
```
## 配置文件不写Tag 并不意味着没有 Tag

我的设计是用 Module ID 当做 Tag

这里会引出一个问题，如果出现这些情况该怎么办？

我们直接把这些当做错误的例子，让开发者自适应

```json
"resource": [
    { "filepath": "./a.db" }, // 隐式 Tag 被当作 "resource"
    { "filepath": "./b.db" }  // 隐式 Tag 又是 "resource"，直接把上面的覆盖了！
]
```

但必须加上一些约定

- 每个 Module 在代码层面必须有 Tag，可以不在配置文件中写，建议用 Module ID 当做 TAG
- 同时我们引入一个变量 PreDefineTags 存放这种不在配置文件写 Tag 的 Module 的 Tag
- JSON 配置文件的顶级字段代表一个 Module，不建议两个顶级字段代表一个 Module

# P1

你给我的代码，我做了一些文件位置上的调整和 module 的加载方式的调整，还有一些细微之处

原有的代码已经无法支持新的文件位置

我们需要做一些修改

等一下，我会发给你当前的代码结构

注意：我们只是在验证 uni 框架的形式，允许任何你觉得“好”的重构，允许你完全重构

我们只是在快速验证，只写必要的结构代码