# Kenaito Config Center Spring Boot 3 Client

English | [中文](README.md)

A config center client for Spring Boot 3.x / JDK 17+ applications. It fetches remote configuration at startup, polls for incremental changes, refreshes `@Value` and `@ConfigurationProperties` fields at runtime, and falls back to a local cache when the config center is unavailable.

## Features

- Fetches all configs before the Spring context starts and injects them into the `Environment`
- Polls for incremental changes (add, update, delete) without restarting the application
- Refreshes `@Value` fields and `@ConfigurationProperties` beans automatically
- Notifies `ConfigChangeHanlder` beans on every change
- Local file cache (`config-cache.log`) used as a fallback when the config center is down
- Sensitive values masked in logs
- Multiple config center addresses (`base-url` separated by `,` or `;`) with automatic failover
- Spring Boot 3 auto-configuration (`AutoConfiguration.imports`)

## Requirements

| Item | Version |
|------|---------|
| JDK | 17+ |
| Spring Boot | 3.x |
| Maven | 3.6+ |

> For Spring Boot 2.x / JDK 8 applications use [`kenaito-conf-springboot2-client`](../kenaito-conf-springboot2-client/README_EN.md).

## Installation

### Maven

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot3-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Local install (for development)

```bash
cd kenaito-conf-springboot3-client
mvn clean install
```

## Quick Start

### 1. Add configuration

Add the config center settings to `application.properties`:

```properties
# Config center address (multiple addresses can be separated by commas)
kenaito.config.base-url=http://localhost:8080

# Application name (required)
kenaito.config.app-name=my-app

# Environment (dev/daily/prod)
kenaito.config.profile=daily

# Config version (default: latest)
kenaito.config.version=latest

# Auth token (required when the application enables token validation)
kenaito.config.token=your-token

# Enable remote config (default: true)
kenaito.config.enabled=true

# Sync interval in seconds (default: 60)
kenaito.config.sync-interval-seconds=10

# Keys masked in logs (comma separated)
kenaito.config.sensitive-keys=password,secret,key,token
```

YAML is supported as well:

```yaml
kenaito:
  config:
    base-url: http://localhost:8080
    app-name: my-app
    profile: daily
    version: latest
    sync-interval-seconds: 10
```

### 2. Inject configs with @Value

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

Values from the config center override local properties with the same key.

### 3. Use @ConfigurationProperties

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

### 4. Listen for changes

Implement `ConfigChangeHanlder` and register it as a Spring bean:

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

### 5. Automatic refresh

When configuration changes in the config center, the client will, within the next sync interval:

1. Update the values in `ConfigcenterContext`;
2. Update the `kenaitoConfigPropertySource` in the Spring `Environment`;
3. Invoke every `ConfigChangeHanlder` bean;
4. Refresh the affected `@Value` fields and `@ConfigurationProperties` fields.

Example refresh log:

```
[RefreshScope] refreshed @Value field: cn.odboy.demo.ConfigController.host | old='10.0.0.1' | new='10.0.0.2'
[RefreshScope] refreshed @ConfigurationProperties field: cn.odboy.demo.DatabaseProperties.port = '13306'
```

> Unlike Spring Cloud's `@RefreshScope`, this client refreshes fields directly via reflection and does not require `spring-cloud-context`.

## Configuration Reference

| Property | Description | Default | Required |
|----------|-------------|---------|----------|
| `kenaito.config.base-url` | Config center address, multiple addresses separated by `,` or `;` | - | Yes |
| `kenaito.config.app-name` | Application name (application code in the config center) | `spring.application.name` | Yes |
| `kenaito.config.profile` | Environment identifier | `dev` | No |
| `kenaito.config.version` | Config version | `latest` | No |
| `kenaito.config.token` | Auth token | empty | No |
| `kenaito.config.enabled` | Enable remote config | `true` | No |
| `kenaito.config.sync-interval-seconds` | Incremental sync interval (seconds) | `60` | No |
| `kenaito.config.global-ignore` | Ignore business group global configs | `false` | No |
| `kenaito.config.sensitive-keys` | Keys masked in logs (comma separated) | `password,passwd,key,secret,token,credentials` | No |
| `kenaito.config.local-storage-dir` | Local cache directory | Application working directory | No |

> Camel case variants (`baseUrl`, `appName`) and the legacy `kenaito.configcenter.*` prefix are also accepted. The kebab-case keys above are recommended.

## Auto-configuration

The client relies on Spring Boot 3 auto-configuration. No manual bean registration is needed:

- `KenaitoEnvironmentInitializer`
  An `EnvironmentPostProcessor` registered via `META-INF/spring.factories`. It fetches remote configs before the Spring context starts and adds them as the highest-priority property source.
- `KenaitoRefreshAutoConfiguration`
  Auto-configuration registered via `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`. It:
  - gives the `ApplicationContext` to the client;
  - registers `RefreshScopeRegistry` (`BeanPostProcessor`) to collect `@Value` fields and `@ConfigurationProperties` beans;
  - refreshes affected fields when configs change.

### Differences from the Spring Boot 2 client

| Item | Spring Boot 2 client | Spring Boot 3 client |
|------|----------------------|----------------------|
| JDK | 8+ | 17+ |
| Spring Boot | 2.x | 3.x |
| Auto-configuration registration | `META-INF/spring.factories` | `META-INF/spring/...AutoConfiguration.imports` |
| Public API | `ConfigChangeHanlder`, etc. | Same |

## Demo

Demo project: [`kenaito-conf-springboot3-client-demo`](../kenaito-conf-springboot3-client-demo/)

```bash
# Install the client into the local repository first
cd kenaito-conf-springboot3-client
mvn clean install

# Start the demo
cd ../kenaito-conf-springboot3-client-demo
mvn spring-boot:run
```

Demo defaults:

- Port: `8084`
- Application name: `kenaito-springboot-demo` (override with `KENAITO_APP_NAME`)
- Environment: `daily` (override with `KENAITO_ENV`)
- Config center: `http://localhost:8080` (override with `KENAITO_BASE_URL`)

Endpoints:

| Endpoint | Description |
|----------|-------------|
| `GET /` | Application info and some configs |
| `GET /config` | Configs injected via `@Value` |
| `GET /configprops` | Configs bound via `@ConfigurationProperties` |
| `GET /changes` | The 20 most recent config changes |
| `GET /health` | Health check |

## Local Cache and Disaster Recovery

After a successful fetch, the client writes the remote configs to `config-cache.log` (in the application working directory by default, configurable via `kenaito.config.local-storage-dir`).

If the config center is unavailable and a cache exists, the client loads configs from the cache so the application can still start. Polling keeps retrying and resumes syncing once the config center is back.

## Notes

1. **Refresh type limits**: `@Value` refresh supports `String`, `int/Integer`, `long/Long`, `boolean/Boolean`, `double/Double`, `float/Float`, `short/Short`, `byte/Byte` and `char/Character`. Other types are assigned as strings.
2. **Priority**: remote configs override local properties with the same key; config center connection settings (`kenaito.config.*`) come from local configuration.
3. **Startup**: the application fails to start when `base-url` or `app-name` is missing. A temporarily unavailable config center does not block startup (without a cache only local configs are used).
4. **Masking**: values whose keys contain a sensitive word are logged as the first half of the value followed by `****`.
5. **Multiple instances**: every instance polls independently; change detection is based on the `lastTime` timestamp.

## Changelog

- **1.0.0** - Initial release for Spring Boot 3 / JDK 17
