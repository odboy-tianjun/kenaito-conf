# Kenaito Config Center Python Client

Python client SDK for the Kenaito Config Center.

## Features

- Fetch application configs from the config center
- Periodic config sync
- Config change callback listener
- Local file cache (fallback when config center is unavailable)

## Installation

```bash
cd kenaito-conf-python-client
pip3 install -e .
```

## Quick Start

```python
from kenaito_conf import Config, ConfigClient

config = Config(
    base_url="http://localhost:8080",
    app_name="my-app",
    env="daily",
    sync_interval=30,
)

client = ConfigClient(config)

# Listen for config changes
def on_change(changed):
    for k, v in changed.items():
        print(f"config changed: {k} = {v}")

client.on_change(on_change)

# Start client (fetch configs + enable periodic sync)
client.start()

# Get config
value = client.get("database.host")
all_configs = client.get_all()

# Stop
client.stop()
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `base_url` | Config center URL | Required |
| `app_name` | Application name | Required |
| `env` | Environment (dev/daily/prod) | `dev` |
| `version` | Config version | `latest` |
| `token` | Auth token | Empty |
| `sync_interval` | Sync interval (seconds) | `60` |
| `ignore_global` | Whether to ignore global configs | `False` |
| `cache_dir` | Local cache directory | `.` |

## Demo

```bash
cd kenaito-conf-python-client-demo
pip3 install -r requirements.txt
python3 app.py
```

The demo listens on port `8082` by default. Visit `http://localhost:8082/` to view configs.

Environment variable overrides:

| Environment Variable | Description | Default |
|----------------------|-------------|---------|
| `SERVER_PORT` | Server port | `8082` |
| `KENAITO_BASE_URL` | Config center URL | `http://localhost:8080` |
| `KENAITO_APP_NAME` | Application name | `kenaito-springboot-demo` |
| `KENAITO_PROFILE` | Environment | `daily` |
| `KENAITO_VERSION` | Config version | `latest` |
| `KENAITO_TOKEN` | Auth token | Empty |
| `KENAITO_SYNC_INTERVAL` | Sync interval (seconds) | `30` |
| `KENAITO_CACHE_DIR` | Cache directory | `.` |
