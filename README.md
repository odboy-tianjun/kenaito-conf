# Kenaito Config Center

基于 Go 重构的配置中心。

## 技术栈

- **后端**: Go + Gin + xorm + MySQL
- **前端**: Vue 2 + Vite 5 + Element UI 2
- **客户端**: Go（Java 客户端保留兼容）

## 项目结构

```
kenaito-conf/
├── server/          # Go 管理端
├── web/             # Vue 前端
├── go-client/       # Go 客户端
├── java-springboot-client/  # Java Spring Boot 客户端
└── sql/             # 数据库脚本
```

## 快速开始

### 1. 初始化数据库

```bash
mysql -u root -p configcenter < sql/init.sql
mysql -u root -p configcenter < sql/testdata.sql  # 测试数据
```

### 2. 启动后端

```bash
cd server
# 修改 config.yaml 中的数据库配置
go run cmd/main.go
# 或编译后运行
go build -o bin/server cmd/main.go
./bin/server -config config.yaml
```

后端默认监听 `:8080`。

### 3. 启动前端

```bash
cd web
npm install
npm run dev
```

前端默认监听 `:3000`，自动代理 API 到后端。

### 4. 访问

打开 http://localhost:3000

默认管理员账号：`admin` / 密码为 BCrypt 哈希对应的原始密码（见 testdata.sql）

## Go 客户端使用

```go
import "kenaito-conf/go-client"

cfg := &client.Config{
    BaseURL:      "http://localhost:8080",
    AppName:      "myapp",
    Env:          "dev",
    Version:      "latest",
    SyncInterval: 30,
    CacheDir:     ".",
}

c := client.New(cfg)

// 注册变更回调
c.OnChange(func(changed map[string]string) {
    for k, v := range changed {
        fmt.Printf("%s = %s\n", k, v)
    }
})

// 启动（拉取全量配置 + 开启定时同步）
c.Start()
defer c.Stop()

// 读取配置
dbHost := c.Get("db.host")
```

## API 概览

### 管理端 API（需 JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /user/login | 登录 |
| GET | /user/login_user_info | 当前用户信息 |
| GET | /admin/menus | 菜单 |
| POST | /admin/config/list | 配置列表 |
| POST | /admin/config/add | 新增配置 |
| POST | /admin/config/update | 更新配置 |
| GET | /admin/config/:id | 配置详情 |
| POST | /admin/config/delete/:id | 删除配置 |
| GET | /admin/config/config_histories/:id | 配置历史 |
| GET | /admin/config/rollback/:id | 回滚配置 |
| POST | /admin/app/list | 应用列表 |
| POST | /admin/app/add | 新增应用 |
| GET | /admin/app/tree | 应用树 |
| GET | /admin/user/list | 用户列表 |
| POST | /admin/user/add | 新增用户 |
| GET | /admin/group/list | 业务组列表 |
| GET | /admin/profile/list | 环境列表 |

### 客户端 API（无需 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/fetch_all_configs | 全量拉取配置 |
| POST | /api/fetch_changed_configs | 增量拉取变更 |
| GET | /api/ping | 健康检查 |

## 与原 Java 版的差异

- **去掉了 ZooKeeper**：配置变更通知改为客户端定时轮询
- **认证方式**：管理端用 JWT，客户端 API 暂开放（可后续加签名验证）
- **前端**：从 layui 静态页面改为 Vue 2 + Element UI 前后端分离
- **加密算法**：保持 AES 兼容，Java 客户端可继续使用
