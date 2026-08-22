package cn.odboy.demo;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.env.Environment;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * Demo endpoints showing {@code @Value} and {@code @ConfigurationProperties}
 * values refresh when configs change in the config center.
 *
 * @author odboy
 * @date 2026-10-21
 */
@RestController
public class ConfigController {

    @Value("${database.host:unknown}")
    private String host;

    @Value("${database.port:unknown}")
    private String port;

    @Value("${database.name:unknown}")
    private String name;

    @Value("${database.user:unknown}")
    private String user;

    @Value("${database.password:unknown}")
    private String password;

    @Value("${app.name:unknown}")
    private String appName;

    private final DatabaseProperties databaseProperties;

    private final ConfigChangeLogger configChangeLogger;

    private final Environment environment;

    @Autowired
    public ConfigController(DatabaseProperties databaseProperties,
                            ConfigChangeLogger configChangeLogger,
                            Environment environment) {
        this.databaseProperties = databaseProperties;
        this.configChangeLogger = configChangeLogger;
        this.environment = environment;
    }

    @GetMapping("/")
    public Map<String, Object> index() {
        return Map.of(
                "app", appName,
                "profiles", String.join(",", environment.getActiveProfiles()),
                "configCenter", environment.getProperty("kenaito.config.base-url", ""),
                "database.host", host);
    }

    @GetMapping("/config")
    public Map<String, Object> config() {
        Map<String, Object> result = new HashMap<>();
        result.put("host", host);
        result.put("port", port);
        result.put("name", name);
        result.put("user", user);
        result.put("password", password);
        return result;
    }

    @GetMapping("/configprops")
    public Map<String, Object> configprops() {
        Map<String, Object> result = new HashMap<>();
        result.put("host", databaseProperties.getHost());
        result.put("port", databaseProperties.getPort());
        result.put("name", databaseProperties.getName());
        result.put("user", databaseProperties.getUser());
        result.put("password", databaseProperties.getPassword());
        return result;
    }

    @GetMapping("/changes")
    public List<Map<String, Object>> changes() {
        return configChangeLogger.getHistory();
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "UP");
    }
}
