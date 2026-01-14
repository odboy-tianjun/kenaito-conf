# kenaito-conf-springboot2-client-demo

配置中心客户端示例项目。

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

1. 先安装 kenaito-conf-springboot2-client 到本地 Maven 仓库：
```bash
cd E:\SelfDev\Gitea2026\kenaito-conf\kenaito-conf-springboot2-client
mvn clean install -DskipTests
```

2. 运行示例项目：
```bash
cd E:\SelfDev\Gitea2026\kenaito-springboot-demo
mvn spring-boot:run
```

3. 访问测试：
```bash
curl http://localhost:8081/
```

## 配置优先级

远程配置 > 本地配置

在配置中心添加应用 `demo` 的配置后，远程配置会覆盖本地同名配置。
