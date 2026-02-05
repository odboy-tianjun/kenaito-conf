import json
import logging
import os

from flask import Flask
from kenaito_conf import Config, ConfigClient

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(name)s] %(message)s")
logger = logging.getLogger("demo")

app = Flask(__name__)

# Config defaults (overridden by remote config center)
config_values = {
    "database.host": os.getenv("DATABASE_HOST", "192.168.1.1"),
    "database.port": os.getenv("DATABASE_PORT", "8888"),
    "database.name": os.getenv("DATABASE_NAME", "xxxx"),
    "database.user": os.getenv("DATABASE_USER", "xxxx"),
    "database.password": os.getenv("DATABASE_PASSWORD", "xxxxxxxx"),
}

client: ConfigClient | None = None


def on_config_changed(changed: dict[str, str]) -> None:
    for k, v in changed.items():
        if k in config_values:
            config_values[k] = v
            logger.info("config updated: %s = %s", k, v)


def create_client() -> ConfigClient:
    cfg = Config(
        base_url=os.getenv("KENAITO_BASE_URL", "http://localhost:8080"),
        app_name=os.getenv("KENAITO_APP_NAME", "kenaito-springboot-demo"),
        env=os.getenv("KENAITO_PROFILE", "daily"),
        version=os.getenv("KENAITO_VERSION", "latest"),
        token=os.getenv("KENAITO_TOKEN", ""),
        sync_interval=int(os.getenv("KENAITO_SYNC_INTERVAL", "30")),
        cache_dir=os.getenv("KENAITO_CACHE_DIR", "."),
    )
    c = ConfigClient(cfg)
    c.on_change(on_config_changed)
    c.start()
    return c


@app.route("/")
def index():
    return json.dumps(config_values, ensure_ascii=False, indent=2), 200, {
        "Content-Type": "application/json; charset=utf-8",
    }


@app.route("/health")
def health():
    return "OK"


if __name__ == "__main__":
    try:
        client = create_client()
        remote = client.get_all()
        for k in config_values:
            if k in remote:
                config_values[k] = remote[k]
        logger.info("loaded %d configs from remote", len(remote))
    except Exception:
        logger.exception("failed to init config client, using local defaults")

    port = int(os.getenv("SERVER_PORT", "8082"))
    app.run(host="0.0.0.0", port=port)
