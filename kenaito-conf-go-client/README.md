# Kenaito Config Center Go Client

Go 语言的配置中心客户端 SDK。

## 功能特性

- 从配置中心拉取应用配置
- 定时同步配置变更
- 支持配置变更回调监听
- AES 解密支持（与 Java/Python 客户端兼容）
- 本地文件缓存（配置中心不可用时降级）

## 安装

### 方式一：使用 go mod（推荐）

在你的项目根目录执行：

```bash
go get kenaito-conf/go-client
```

### 方式二：本地开发

如果你在本地开发或修改客户端代码，可以使用 `replace` 指令：

```bash
# 在你的 go.mod 中添加
require kenaito-conf/go-client v0.0.0

replace kenaito-conf/go-client => ../kenaito-conf-go-client
```

然后执行：

```bash
go mod tidy
```

## 快速开始

```go
package main

import (
    "fmt"
    "kenaito-conf/go-client"
    "log"
)

func main() {
    // 创建配置
    cfg := &client.Config{
        BaseURL:      "http://localhost:8080",
        AppName:      "my-app",
        Env:          "daily",
        Version:      "latest",
        Token:        "", // 可选，如果应用配置了 Token 校验
        SyncInterval: 30, // 同步间隔（秒）
        CacheDir:     ".", // 缓存目录
    }

    // 创建客户端
    c := client.New(cfg)

    // 注册配置变更回调
    c.OnChange(func(changed map[string]string) {
        fmt.Println("配置变更:")
        for k, v := range changed {
            fmt.Printf("  %s = %s\n", k, v)
        }
    })

    // 启动客户端（拉取配置 + 开启定时同步）
    if err := c.Start(); err != nil {
        log.Fatalf("启动失败: %v", err)
    }
    defer c.Stop()

    // 读取配置
    value := c.Get("database.host")
    fmt.Printf("database.host = %s\n", value)

    // 获取所有配置
    all := c.GetAll()
    for k, v := range all {
        fmt.Printf("%s = %s\n", k, v)
    }
}
```

## 配置参数

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `BaseURL` | string | 配置中心地址 | 必填 |
| `AppName` | string | 应用名称 | 必填 |
| `Env` | string | 环境（dev/daily/prod） | `dev` |
| `Version` | string | 配置版本 | `latest` |
| `Token` | string | 应用 Token（如果启用了 Token 校验） | 空 |
| `SyncInterval` | int | 同步间隔（秒） | `60` |
| `IgnoreGlobal` | bool | 是否忽略全局配置 | `false` |
| `CacheDir` | string | 本地缓存目录 | `.` |

## API 说明

### `client.New(cfg *Config) *Client`

创建客户端实例。

### `(*Client).Start() error`

启动客户端，拉取配置并开启定时同步。

### `(*Client).Stop()`

停止客户端，关闭定时同步。

### `(*Client).Get(key string) string`

获取指定配置项的值。

### `(*Client).GetAll() map[string]string`

获取所有配置项。

### `(*Client).OnChange(handler func(map[string]string))`

注册配置变更回调函数。当配置发生变更时，会调用该函数，参数为变更的配置项（key-value 对）。

## 示例项目

完整的示例代码请参考：[kenaito-conf-go-client-demo](../kenaito-conf-go-client-demo)

运行示例：

```bash
cd kenaito-conf-go-client-demo
go run main.go
```

## 注意事项

1. **Token 校验**：如果应用配置了 Token，需要在 `Config.Token` 中传入正确的 Token 值
2. **缓存文件**：客户端会在 `CacheDir` 目录下创建 `config-cache.json` 文件，用于配置中心不可用时降级
3. **AES 解密**：配置值如果以 `{Cipher}` 开头，会自动使用 AES 解密，密钥由配置中心统一管理
