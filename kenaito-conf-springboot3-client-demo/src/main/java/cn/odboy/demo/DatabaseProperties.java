package cn.odboy.demo;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

/**
 * Database configuration bound with {@code @ConfigurationProperties}.
 *
 * @author odboy
 * @date 2026-10-21
 */
@Component
@ConfigurationProperties(prefix = "database")
public class DatabaseProperties {

    private String host;

    private Integer port;

    private String name;

    private String user;

    private String password;

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public Integer getPort() {
        return port;
    }

    public void setPort(Integer port) {
        this.port = port;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getUser() {
        return user;
    }

    public void setUser(String user) {
        this.user = user;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }
}
