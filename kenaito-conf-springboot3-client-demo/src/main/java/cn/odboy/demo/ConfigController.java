package cn.odboy.demo;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.Map;

@RestController
public class ConfigController {

    @Value("${app.name:unknown}")
    private String appName;

    @Value("${app.version:unknown}")
    private String appVersion;

    @Value("${db.url:unknown}")
    private String dbUrl;

    @Value("${db.username:unknown}")
    private String dbUsername;

    @Value("${custom.config1:unknown}")
    private String customConfig1;

    @GetMapping("/")
    public Map<String, Object> index() {
        Map<String, Object> result = new HashMap<>();
        result.put("appName", appName);
        result.put("appVersion", appVersion);
        result.put("dbUrl", dbUrl);
        result.put("dbUsername", dbUsername);
        result.put("customConfig1", customConfig1);
        return result;
    }

    @GetMapping("/health")
    public String health() {
        return "OK";
    }
}
