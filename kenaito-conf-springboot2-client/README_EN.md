# Kenaito Config Center Spring Boot 2 Client

Config center client for Spring Boot 2.x applications, supporting automatic config refresh and dynamic `@Value` annotation updates.

## Features

- Fetch application configs from the config center
- Periodic config sync
- Automatic `@Value` annotation refresh
- Local file cache (fallback when config center is unavailable)

## Installation

### Maven Dependency

```xml
<dependency>
    <groupId>cn.odboy</groupId>
    <artifactId>kenaito-conf-springboot2-client</artifactId>
    <version>1.0.3</version>
</dependency>
```

### Local Install (Development)

```bash
cd kenaito-conf-springboot2-client
mvn clean install
```

## Quick Start

### 1. Configuration File

Add config center connection parameters in `application.properties`:

```properties
# Config center URL
kenaito.config.baseUrl=http://localhost:8080

# Application name (required)
kenaito.config.appName=my-app

# Environment (dev/daily/prod)
kenaito.config.profile=daily

# Config version
kenaito.config.version=latest

# Auth token (optional)
kenaito.config.token=your-token

# Whether to enable remote config (default true)
kenaito.config.enabled=true

# Sync interval in seconds (default 60)
kenaito.config.sync-interval-seconds=10

# Ignore business group global configs (default false)
kenaito.config.global-ignore=false
```

### 2. Inject Configs with @Value

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

### 3. Automatic Config Refresh

When configs change in the config center, the client automatically refreshes `@Value` annotated fields without restarting the application.

**Config change listener log:**
```
[RefreshScope] Refreshed @Value field: ConfigController.databaseHost | old='localhost' | new='192.168.1.100' | converted='192.168.1.100' | verify='192.168.1.100' | success=true
```

### 4. Using @ConfigurationProperties

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

## Configuration Properties

| Property | Description | Default | Required |
|----------|-------------|---------|----------|
| `kenaito.config.baseUrl` | Config center URL | - | Yes |
| `kenaito.config.appName` | Application name | `spring.application.name` | Yes |
| `kenaito.config.profile` | Environment | `dev` | No |
| `kenaito.config.version` | Config version | `latest` | No |
| `kenaito.config.token` | Auth token | Empty | No |
| `kenaito.config.enabled` | Whether to enable | `true` | No |
| `kenaito.config.sync-interval-seconds` | Sync interval (seconds) | `60` | No |
| `kenaito.config.global-ignore` | Ignore business group global configs | `false` | No |
| `kenaito.configcenter.local-storage-dir` | Local cache directory | Application working directory | No |

> The Spring Boot 2 client reads the camel-case `baseUrl` / `appName` keys. Sensitive keys are fixed to `password`, `passwd`, `key`, `secret`, `token`, `credentials` and cannot be customized.

## Auto Configuration

The client works via Spring Boot auto-configuration:

- **KenaitoEnvironmentInitializer**: Initializes the config center connection
- **KenaitoRefreshAutoConfiguration**: Registers the config refresh listener

No manual Bean configuration needed — just add the dependency and configure the properties.

## Demo

Run the demo project:

```bash
cd kenaito-conf-springboot2-client-demo
mvn spring-boot:run
```

Demo defaults:
- Port: 8081
- App name: `kenaito-springboot-demo`
- Environment: `daily`
- Config center URL: `http://localhost:8080`

Visit `http://localhost:8081/` or `http://localhost:8081/configprops` to view current configs.

## Notes

1. **@Value Refresh Limitation**: Only supports primitive types (String, Integer, Boolean, Long, Double, etc.) and their wrapper classes
2. **Config Priority**: Remote config > Local config > Default value
3. **Log Masking**: Configs containing sensitive keys are automatically masked in logs (first half + `****`)
4. **Network Errors**: Sync failures are logged as errors without affecting normal application operation
5. **Local Fallback**: The client creates a `config-cache.log` file in the local cache directory, automatically loading configs from cache when the config center is unavailable

## Changelog

- **1.0.3** - Simplified config refresh logic, removed EnvironmentChangeEvent wrapper
- **1.0.2** - Added sensitive config masking support
- **1.0.1** - Added sync interval configuration
- **1.0.0** - Initial release
