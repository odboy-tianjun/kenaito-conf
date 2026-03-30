# Kenaito Config Center Spring Boot 2 Client

适用于 Spring Boot 2.x 应用的配置中心客户端，支持配置自动刷新和 `@Value` 注解动态更新。

## 安装

### Maven 依赖

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot2-client</artifactId>
    <version>1.0.3</version>
</dependency>
```

### 本地安装（开发调试）

```bash
cd kenaito-conf-springboot2-client
mvn clean install
```

## 快速开始

### 1. 配置文件

在 `application.properties` 中添加配置中心连接参数：

```properties
# 配置中心地址
kenaito.configcenter.base-url=http://localhost:8080

# 应用名称（必填）
kenaito.configcenter.app-name=my-app

# 环境（dev/daily/prod）
kenaito.configcenter.env=daily

# 配置版本
kenaito.configcenter.version=latest

# 认证令牌（可选）
kenaito.configcenter.token=your-token

# 是否启用远程配置（默认 true）
kenaito.configcenter.enabled=true

# 同步间隔（秒，默认 10）
kenaito.configcenter.sync-interval=10

# 敏感配置键（用于日志脱敏，可选）
kenaito.configcenter.sensitive-keys=password,secret,key,token
```

### 2. 使用 @Value 注入配置

```java
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.GetMapping;

@RestController
public class ConfigController {

    @Value("${database.host:localhost}")
    private String databaseHost;

    @Value("${database.port:3306}")
    private Integer databasePort;

    @Value("${kenaito.feature.enabled:false}")
    private Boolean featureEnabled;

    @GetMapping("/config")
    public Map<String, Object> getConfig() {
        Map<String, Object> config = new HashMap<>();
        config.put("databaseHost", databaseHost);
        config.put("databasePort", databasePort);
        config.put("featureEnabled", featureEnabled);
        return config;
    }
}
```

### 3. 配置自动刷新

当配置中心的配置发生变更时，客户端会自动刷新 `@Value` 注解的字段值，无需重启应用。

**配置变更监听日志：**
```
[RefreshScope] Refreshed @Value field: ConfigController.databaseHost | old='localhost' | new='192.168.1.100' | converted='192.168.1.100' | verify='192.168.1.100' | success=true
```

### 4. 使用 @ConfigurationProperties

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

## 配置项说明

| 配置项 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `kenaito.configcenter.base-url` | 配置中心地址 | - | 是 |
| `kenaito.configcenter.app-name` | 应用名称 | - | 是 |
| `kenaito.configcenter.env` | 环境标识 | `dev` | 否 |
| `kenaito.configcenter.version` | 配置版本 | `latest` | 否 |
| `kenaito.configcenter.token` | 认证令牌 | 空 | 否 |
| `kenaito.configcenter.enabled` | 是否启用 | `true` | 否 |
| `kenaito.configcenter.sync-interval` | 同步间隔（秒） | `10` | 否 |
| `kenaito.configcenter.sensitive-keys` | 敏感键（日志脱敏） | 空 | 否 |

## 自动配置说明

客户端通过 Spring Boot 自动配置机制工作：

- **KenaitoEnvironmentInitializer**：初始化配置中心连接
- **KenaitoRefreshAutoConfiguration**：注册配置刷新监听器

无需手动配置任何 Bean，引入依赖并配置参数即可使用。

## Demo

运行示例项目：

```bash
cd kenaito-conf-springboot2-client-demo
mvn spring-boot:run
```

Demo 默认配置：
- 端口：8083
- 应用名称：`kenaito-springboot2-demo`
- 环境：`daily`
- 配置中心地址：`http://localhost:8080`

访问 `http://localhost:8083/config` 查看当前配置。

## 注意事项

1. **@Value 刷新限制**：仅支持基本类型（String、Integer、Boolean、Long、Double 等）和它们的包装类
2. **配置优先级**：远程配置 > 本地配置 > 默认值
3. **日志脱敏**：包含敏感键的配置在日志中会自动脱敏（显示前一半 + `****`）
4. **网络异常**：同步失败时会记录错误日志，不影响应用正常运行

## 版本历史

- **1.0.3** - 简化配置刷新逻辑，移除 EnvironmentChangeEvent 包装
- **1.0.2** - 添加敏感配置脱敏支持
- **1.0.1** - 添加配置同步间隔配置
- **1.0.0** - 初始版本
