# Kenaito Config Center Python Client

[English](README_EN.md) | 中文

Python 客户端 SDK，适用于 Python 3.10+ 项目。

## 功能特性

- 从配置中心拉取应用配置
- 定时同步配置变更
- 支持配置变更回调监听
- 本地文件缓存（配置中心不可用时降级）

## 安装

```bash
cd kenaito-conf-python-client
pip3 install -e .
```

## 快速开始

```python
from kenaito_conf import Config, ConfigClient

config = Config(
    base_url="http://localhost:8080",
    app_name="my-app",
    env="daily",
    sync_interval=30,
)

client = ConfigClient(config)

# 监听配置变更
def on_change(changed):
    for k, v in changed.items():
        print(f"config changed: {k} = {v}")

client.on_change(on_change)

# 启动客户端（拉取配置 + 开启定时同步）
client.start()

# 获取配置
value = client.get("database.host")
all_configs = client.get_all()

# 停止
client.stop()
```

## 配置项

| 参数              | 说明                 | 默认值      |
|-----------------|--------------------|----------|
| `base_url`      | 配置中心地址             | 必填       |
| `app_name`      | 应用名称               | 必填       |
| `env`           | 环境（dev/daily/prod） | `dev`    |
| `version`       | 配置版本               | `latest` |
| `token`         | 认证令牌               | 空        |
| `sync_interval` | 同步间隔（秒）            | `60`     |
| `ignore_global` | 是否忽略全局配置           | `False`  |
| `cache_dir`     | 本地缓存目录             | `.`      |

## Demo

```bash
cd kenaito-conf-python-client-demo
pip3 install -r requirements.txt
python3 app.py
```

Demo 默认监听 `8082` 端口，访问 `http://localhost:8082/` 查看配置。

环境变量覆盖：

| 环境变量                    | 说明      | 默认值                       |
|-------------------------|---------|---------------------------|
| `SERVER_PORT`           | 服务端口    | `8082`                    |
| `KENAITO_BASE_URL`      | 配置中心地址  | `http://localhost:8080`   |
| `KENAITO_APP_NAME`      | 应用名称    | `kenaito-springboot-demo` |
| `KENAITO_PROFILE`       | 环境      | `daily`                   |
| `KENAITO_VERSION`       | 配置版本    | `latest`                  |
| `KENAITO_TOKEN`         | 认证令牌    | 空                         |
| `KENAITO_SYNC_INTERVAL` | 同步间隔（秒） | `30`                      |
| `KENAITO_CACHE_DIR`     | 缓存目录    | `.`                       |
