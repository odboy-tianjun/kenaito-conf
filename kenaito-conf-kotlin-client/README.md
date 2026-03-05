# Kenaito Config Center Kotlin Client

Kotlin 语言的配置中心客户端 SDK。

## 依赖

- Kotlin 1.9+
- kotlinx-coroutines
- kotlinx-serialization

## 编译

```bash
cd kenaito-conf-kotlin-client
./gradlew build
```

## 使用

```kotlin
import kenait.conf.Config
import kenait.conf.ConfigClient

val config = Config(
    baseUrl = "http://localhost:8080",
    appName = "kenaito-springboot-demo",
    env = "daily",
    version = "latest",
    syncInterval = 30,
)

val client = ConfigClient(config)

client.onChange { changed ->
    changed.forEach { (k, v) ->
        println("config changed: $k = $v")
    }
}

client.start()

val value = client.get("database.host")
val all = client.getAll()

client.stop()
```

## 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `baseUrl` | 配置中心地址 | 必填 |
| `appName` | 应用名称 | 必填 |
| `env` | 环境 | `dev` |
| `version` | 配置版本 | `latest` |
| `token` | 认证令牌 | 空 |
| `syncInterval` | 同步间隔（秒） | `60` |
| `ignoreGlobal` | 是否忽略全局配置 | `false` |
| `cacheDir` | 本地缓存目录 | `.` |

## Demo

```bash
cd kenaito-conf-kotlin-client-demo
./gradlew run
```
