# Kenaito Config Center Spring Boot 2 Client

Config center client for Spring Boot 2.x applications, supporting automatic config refresh and dynamic `@Value` annotation updates.

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
kenaito.configcenter.base-url=http://localhost:8080

# Application name (required)
kenaito.configcenter.app-name=my-app

# Environment (dev/daily/prod)
kenaito.configcenter.env=daily

# Config version
kenaito.configcenter.version=latest

# Auth token (optional)
kenaito.configcenter.token=your-token

# Whether to enable remote config (default true)
kenaito.configcenter.enabled=true

# Sync interval in seconds (default 10)
kenaito.configcenter.sync-interval=10

# Sensitive config keys (for log masking, optional)
kenaito.configcenter.sensitive-keys=password,secret,key,token
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
| `kenaito.configcenter.base-url` | Config center URL | - | Yes |
| `kenaito.configcenter.app-name` | Application name | - | Yes |
| `kenaito.configcenter.env` | Environment | `dev` | No |
| `kenaito.configcenter.version` | Config version | `latest` | No |
| `kenaito.configcenter.token` | Auth token | Empty | No |
| `kenaito.configcenter.enabled` | Whether to enable | `true` | No |
| `kenaito.configcenter.sync-interval` | Sync interval (seconds) | `10` | No |
| `kenaito.configcenter.sensitive-keys` | Sensitive keys (log masking) | Empty | No |

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
- Port: 8083
- App name: `kenaito-springboot2-demo`
- Environment: `daily`
- Config center URL: `http://localhost:8080`

Visit `http://localhost:8083/config` to view current configs.

## Notes

1. **@Value Refresh Limitation**: Only supports primitive types (String, Integer, Boolean, Long, Double, etc.) and their wrapper classes
2. **Config Priority**: Remote config > Local config > Default value
3. **Log Masking**: Configs containing sensitive keys are automatically masked in logs (first half + `****`)
4. **Network Errors**: Sync failures are logged as errors without affecting normal application operation

## Changelog

- **1.0.3** - Simplified config refresh logic, removed EnvironmentChangeEvent wrapper
- **1.0.2** - Added sensitive config masking support
- **1.0.1** - Added sync interval configuration
- **1.0.0** - Initial release
