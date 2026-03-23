package cn.odboy.demo;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.Map;

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

    @Autowired
    private DatabaseProperties databaseProperties;

    @GetMapping("/")
    public Map<String, Object> index() {
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

    @GetMapping("/health")
    public String health() {
        return "OK";
    }
}
