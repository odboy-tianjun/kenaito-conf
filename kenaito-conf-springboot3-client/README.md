# Kenaito Config Center Spring Boot 3 Client

[English](README_EN.md) | 中文

适用于 Spring Boot 3.x / JDK 17+ 应用的配置中心客户端，支持启动时拉取远程配置、定时同步、`@Value` 与 `@ConfigurationProperties` 字段动态刷新，以及配置中心不可用时的本地缓存降级。

## 功能特性

- 应用启动前从配置中心拉取全量配置，注入 Spring `Environment`
- 定时轮询增量变更（新增、修改、删除），无需重启应用
- 自动刷新 `@Value` 字段和 `@ConfigurationProperties` Bean
- 支持通过 `ConfigChangeHanlder` 监听配置变更
- 本地文件缓存（`config-cache.log`），配置中心不可用时自动降级
- 敏感配置日志脱敏
- 支持配置中心多地址（`base-url` 使用 `,` 或 `;` 分隔，自动故障转移）
- Spring Boot 3 自动配置（`AutoConfiguration.imports`）

## 环境要求

| 项目 | 版本 |
|------|------|
| JDK | 17+ |
| Spring Boot | 3.x |
| Maven | 3.6+ |

> Spring Boot 2.x / JDK 8 项目请使用 [`kenaito-conf-springboot2-client`](../kenaito-conf-springboot2-client/README.md)。

## 安装

### Maven 依赖

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot3-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

### 本地安装（开发调试）

```bash
cd kenaito-conf-springboot3-client
mvn clean install
```

## 快速开始

### 1. 添加配置

在 `application.properties` 中添加配置中心连接参数：

```properties
# 配置中心地址（多个地址用逗号分隔）
kenaito.config.base-url=http://localhost:8080

# 应用名称（必填）
kenaito.config.app-name=my-app

# 环境（dev/daily/prod）
kenaito.config.profile=daily

# 配置版本（默认 latest）
kenaito.config.version=latest

# 认证令牌（应用开启 Token 校验时必填）
kenaito.config.token=your-token

# 是否启用远程配置（默认 true）
kenaito.config.enabled=true

# 同步间隔（秒，默认 60）
kenaito.config.sync-interval-seconds=10

# 敏感配置键（用于日志脱敏，逗号分隔）
kenaito.config.sensitive-keys=password,secret,key,token
```

也可以使用 YAML：

```yaml
kenaito:
  config:
    base-url: http://localhost:8080
    app-name: my-app
    profile: daily
    version: latest
    sync-interval-seconds: 10
```

### 2. 使用 @Value 注入配置

```java
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ConfigController {

    @Value("${database.host:localhost}")
    private String databaseHost;

    @Value("${database.port:3306}")
    private Integer databasePort;

    @Value("${feature.enabled:false}")
    private Boolean featureEnabled;

    @GetMapping("/config")
    public String config() {
        return databaseHost + ":" + databasePort + ", feature=" + featureEnabled;
    }
}
```

配置中心的值会覆盖 `application.properties` 中的同名配置。

### 3. 使用 @ConfigurationProperties

```java
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "database")
public class DatabaseProperties {

    private String host;

    private Integer port;

    private String username;

    private String password;

    // getters and setters
}
```

### 4. 监听配置变更

实现 `ConfigChangeHanlder` 并注册为 Spring Bean：

```java
import cn.odboy.confcenter.ConfigChangeHanlder;
import org.springframework.stereotype.Component;

import java.util.Map;

@Component
public class MyConfigChangeListener implements ConfigChangeHanlder {

    @Override
    public void onConfigChanged(Map<String, Object> changedConfigs) {
        changedConfigs.forEach((key, value) ->
                System.out.printf("config changed: %s = %s%n", key, value));
    }
}
```

### 5. 自动刷新

配置中心的配置发生变更后，客户端会在下一个同步周期内：

1. 更新 `ConfigcenterContext` 中的配置值；
2. 更新 Spring `Environment` 中的 `kenaitoConfigPropertySource`；
3. 回调所有 `ConfigChangeHanlder` Bean；
4. 刷新受影响的 `@Value` 字段和 `@ConfigurationProperties` 字段。

刷新日志示例：

```
[RefreshScope] refreshed @Value field: cn.odboy.demo.ConfigController.host | old='10.0.0.1' | new='10.0.0.2'
[RefreshScope] refreshed @ConfigurationProperties field: cn.odboy.demo.DatabaseProperties.port = '13306'
```

> 与 Spring Cloud 的 `@RefreshScope` 不同，本客户端直接通过反射刷新字段，不需要 `spring-cloud-context` 依赖。

## 配置项说明

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `kenaito.config.base-url` | 配置中心地址，多个地址用 `,` 或 `;` 分隔 | - | 是 |
| `kenaito.config.app-name` | 应用名称（配置中心中的应用编码） | `spring.application.name` | 是 |
| `kenaito.config.profile` | 环境标识 | `dev` | 否 |
| `kenaito.config.version` | 配置版本 | `latest` | 否 |
| `kenaito.config.token` | 认证令牌 | 空 | 否 |
| `kenaito.config.enabled` | 是否启用远程配置 | `true` | 否 |
| `kenaito.config.sync-interval-seconds` | 增量同步间隔（秒） | `60` | 否 |
| `kenaito.config.global-ignore` | 是否忽略业务组全局配置 | `false` | 否 |
| `kenaito.config.sensitive-keys` | 日志中需要脱敏的键（逗号分隔） | `password,passwd,key,secret,token,credentials` | 否 |
| `kenaito.config.local-storage-dir` | 本地缓存目录 | 应用工作目录 | 否 |

> `baseUrl`/`appName`/`base-url`/`app-name` 等驼峰写法和旧的 `kenaito.configcenter.*` 前缀同样兼容，推荐使用表中列出的中划线写法。

## 自动配置说明

客户端通过 Spring Boot 3 的自动配置机制工作，引入依赖并配置参数即可使用，无需手动注册 Bean：

- `KenaitoEnvironmentInitializer`
  通过 `META-INF/spring.factories` 注册的 `EnvironmentPostProcessor`，在 Spring 上下文创建前拉取远程配置，并将其作为最高优先级的 PropertySource 加入 `Environment`。
- `KenaitoRefreshAutoConfiguration`
  通过 `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports` 注册的自动配置，负责：
  - 将 `ApplicationContext` 交给客户端；
  - 注册 `RefreshScopeRegistry`（`BeanPostProcessor`），收集 `@Value` 字段与 `@ConfigurationProperties` Bean；
  - 在配置变更时刷新受影响的字段。

### 与 Spring Boot 2 客户端的差异

| 项目 | Spring Boot 2 客户端 | Spring Boot 3 客户端 |
|------|----------------------|----------------------|
| JDK | 8+ | 17+ |
| Spring Boot | 2.x | 3.x |
| 自动配置注册 | `META-INF/spring.factories` | `META-INF/spring/...AutoConfiguration.imports` |
| 公共 API | `ConfigChangeHanlder` 等 | 相同 |

## Demo

Demo 项目：[`kenaito-conf-springboot3-client-demo`](../kenaito-conf-springboot3-client-demo/)

```bash
# 先安装客户端到本地仓库
cd kenaito-conf-springboot3-client
mvn clean install

# 启动 Demo
cd ../kenaito-conf-springboot3-client-demo
mvn spring-boot:run
```

Demo 默认配置：

- 端口：`8084`
- 应用名称：`kenaito-springboot-demo`（可通过 `KENAITO_APP_NAME` 覆盖）
- 环境：`daily`（可通过 `KENAITO_ENV` 覆盖）
- 配置中心地址：`http://localhost:8080`（可通过 `KENAITO_BASE_URL` 覆盖）

可用接口：

| 接口 | 说明 |
|------|------|
| `GET /` | 应用信息与部分配置 |
| `GET /config` | `@Value` 注入的配置 |
| `GET /configprops` | `@ConfigurationProperties` 绑定的配置 |
| `GET /changes` | 最近 20 次配置变更记录 |
| `GET /health` | 健康检查 |

## 本地缓存与容灾

客户端在成功拉取远程配置后，会将配置写入本地缓存文件 `config-cache.log`（默认位于应用工作目录，可通过 `kenaito.config.local-storage-dir` 指定）。

当配置中心不可用且本地存在缓存时，客户端会从缓存加载配置，保证应用仍可启动；同时定时轮询会持续重试，配置中心恢复后自动继续同步。

## 注意事项

1. **刷新类型限制**：`@Value` 刷新支持 `String`、`int/Integer`、`long/Long`、`boolean/Boolean`、`double/Double`、`float/Float`、`short/Short`、`byte/Byte`、`char/Character`，其他类型按字符串尝试赋值。
2. **配置优先级**：远程配置优先于本地同名配置；配置中心连接参数（`kenaito.config.*`）以本地配置为准。
3. **启动失败**：未配置 `base-url` 或 `app-name` 时应用启动失败；配置中心暂时不可用不会阻止启动（若本地无缓存，则仅使用本地配置）。
4. **日志脱敏**：键名包含敏感词的配置在启动日志中只显示前一半字符 + `****`。
5. **多实例**：每个应用实例独立轮询配置中心，变更检测基于 `lastTime` 时间戳。

## 版本历史

- **1.0.0** - 初始版本，适配 Spring Boot 3 / JDK 17
