# kenaito-conf-springboot3-client-demo

配置中心 Spring Boot 3 客户端示例项目。

## 环境要求

- **JDK 17 或更高版本**（Spring Boot 3 最低要求）
- Maven 3.6+

## 配置说明

在 `application.properties` 中配置：

```properties
# 配置中心地址
kenaito.config.server=http://localhost:8080
kenaito.config.appName=demo
kenaito.config.env=dev
kenaito.config.version=latest
kenaito.config.profile=dev
```

## 运行步骤

1. 确保已安装 JDK 17+：
```bash
java -version
```

2. 先安装 kenaito-conf-springboot3-client 到本地 Maven 仓库：
```bash
cd E:\SelfDev\Gitea2026\kenaito-conf\kenaito-conf-springboot3-client
mvn clean install -DskipTests
```

3. 运行示例项目：
```bash
cd E:\SelfDev\Gitea2026\kenaito-conf\kenaito-conf-springboot3-client-demo
mvn spring-boot:run
```

4. 访问测试：
```bash
curl http://localhost:8082/
```

## 配置优先级

远程配置 > 本地配置

在配置中心添加应用 `demo` 的配置后，远程配置会覆盖本地同名配置。

## 与 Spring Boot 2 Demo 的区别

| 项目 | Spring Boot 版本 | JDK 版本 | 端口 |
|------|------------------|----------|------|
| kenaito-conf-springboot2-client-demo | 2.7.18 | 1.8+ | 8081 |
| kenaito-conf-springboot3-client-demo | 3.2.5 | 17+ | 8082 |

## 注意事项

- 当前环境只有 JDK 11，无法运行此项目
- 需要安装 JDK 17+ 后才能编译和运行
- 如果需要在 JDK 11 环境测试，请使用 kenaito-conf-springboot2-client-demo
