# kenaito-conf-springboot3-client

配置中心 Spring Boot 3 客户端（当前使用 Spring Boot 2.7.18 + JDK 11 编译）

## 重要说明

**当前状态：**
- 项目名称为 springboot3-client，但实际使用 Spring Boot 2.7.18 + JDK 11 编译
- 代码已重构，移除了 mendmix 依赖，使用标准 Java/Spring 工具类
- 为未来升级到 Spring Boot 3 做好准备

**升级到 Spring Boot 3 的步骤：**
1. 安装 JDK 17 或更高版本
2. 修改 `pom.xml`：
   ```xml
   <maven.compiler.source>17</maven.compiler.source>
   <maven.compiler.target>17</maven.compiler.target>
   <spring-boot.version>3.2.5</spring-boot.version>
   ```
3. 修改 Spring Framework 版本：
   ```xml
   <version>6.1.6</version>
   ```
4. 重新编译：`mvn clean install`

## 与 springboot2-client 的区别

1. **移除 mendmix 依赖** - 使用标准 Java/Spring 工具类替代
2. **使用 Java HttpClient** - 替代 mendmix 的 HttpUtils
3. **使用 Jackson** - 替代 mendmix 的 JsonUtils
4. **内置加密工具** - 使用 Java 原生 javax.crypto 实现 AES/DES 加密
5. **更清晰的代码结构** - 便于维护和扩展

## 配置说明

在 `application.properties` 中配置：

```properties
# 启用配置中心
kenaito.config.enabled=true

# 配置中心服务地址（支持多个，逗号分隔）
kenaito.config.baseUrl=http://localhost:8080

# 应用名称
kenaito.config.appName=myapp

# 环境标识
kenaito.config.profile=dev

# 版本号
kenaito.config.version=latest

# 同步间隔（秒）
kenaito.config.sync-interval-seconds=60

# 认证 token（如果配置中心启用了认证）
kenaito.config.token=your-token-here
```

## 使用方式

### Spring Boot 应用

自动生效，无需额外配置。`KenaitoEnvironmentInitializer` 会在 Spring Boot 启动时自动加载远程配置。

### 传统 Spring 应用

在 Spring 配置文件中添加：

```xml
<bean class="cn.odboy.confcenter.spring.CCPropertyPlaceholderConfigurer">
    <property name="remoteEnabled" value="true"/>
</bean>
```

## 安装到本地 Maven 仓库

```bash
mvn clean install -DskipTests
```

## 依赖说明

- Spring Boot 2.7.18（optional）
- Spring Context 5.3.31（optional）
- Jackson 2.17.0（JSON 处理）
- SLF4J 2.0.12（日志）
- Java HttpClient（HTTP 请求，JDK 11+ 内置）

## 升级路径

当你的项目升级到 JDK 17 后，可以修改 pom.xml 中的版本配置，切换到 Spring Boot 3：

```xml
<properties>
    <maven.compiler.source>17</maven.compiler.source>
    <maven.compiler.target>17</maven.compiler.target>
    <spring-boot.version>3.2.5</spring-boot.version>
</properties>

<dependency>
    <groupId>org.springframework</groupId>
    <artifactId>spring-context</artifactId>
    <version>6.1.6</version>
</dependency>
```
