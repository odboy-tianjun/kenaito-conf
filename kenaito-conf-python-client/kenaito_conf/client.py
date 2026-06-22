import logging
import threading
import time

import requests

from kenaito_conf.cache import Cache
from kenaito_conf.config import Config
from kenaito_conf.listener import Listener

logger = logging.getLogger("kenaito_conf")


class ConfigClient:
    def __init__(self, config: Config):
        self._config = config
        self._cache = Cache(config.cache_dir)
        self._listener = Listener()
        self._configs: dict[str, str] = {}
        self._last_time: int = 0
        self._lock = threading.RLock()
        self._stop_event = threading.Event()
        self._started = False

    def on_change(self, handler) -> None:
        self._listener.on_change(handler)

    def get(self, key: str) -> str | None:
        with self._lock:
            return self._configs.get(key)

    def get_all(self) -> dict[str, str]:
        with self._lock:
            return dict(self._configs)

    def start(self) -> None:
        with self._lock:
            if self._started:
                return
            self._started = True

        try:
            self._fetch_all()
        except Exception as e:
            logger.warning("[ConfigClient] fetch all configs error: %s, trying cache", e)
            cached = self._cache.read()
            if cached is not None:
                self._apply_configs(cached)
                logger.info("[ConfigClient] loaded %d configs from cache", len(self._configs))
            else:
                raise RuntimeError(f"fetch configs failed and no cache: {e}") from e

        t = threading.Thread(target=self._sync_loop, daemon=True)
        t.start()

    def stop(self) -> None:
        self._stop_event.set()

    def _fetch_all(self) -> None:
        url = f"{self._config.base_url}/api/fetchAllConfigs"
        payload = {
            "appName": self._config.app_name,
            "env": self._config.env,
            "version": self._config.version,
            "ignoreGlobal": self._config.ignore_global,
            "token": self._config.token,
        }
        resp = requests.post(url, json=payload, timeout=10)
        resp.raise_for_status()
        data = resp.json()

        if data.get("code", 0) != 0:
            raise RuntimeError(f"server error: {data.get('msg', 'unknown')}")

        ts = data.get("latest.fetch.timestamp")
        if ts:
            self._last_time = int(ts)

        self._cache.write(data)
        self._apply_configs(data)
        logger.info("[ConfigClient] fetched %d configs", len(self._configs))

    def _fetch_changed(self) -> None:
        url = f"{self._config.base_url}/api/fetchChangedConfigs"
        payload = {
            "appName": self._config.app_name,
            "env": self._config.env,
            "version": self._config.version,
            "lastTime": self._last_time,
            "ignoreGlobal": self._config.ignore_global,
            "token": self._config.token,
        }
        resp = requests.post(url, json=payload, timeout=10)
        resp.raise_for_status()
        result = resp.json()

        if result.get("code", -1) != 0:
            raise RuntimeError(f"server error code: {result.get('code')}")

        changed_data = result.get("data", {})

        ts = changed_data.pop("latest.fetch.timestamp", None)
        if ts:
            self._last_time = int(ts)

        # 处理删除的配置
        deleted_keys_str = changed_data.pop("deletedKeys", None)
        if deleted_keys_str:
            deleted_keys = deleted_keys_str.split(",")
            with self._lock:
                for key in deleted_keys:
                    self._configs.pop(key, None)
            logger.info("[ConfigClient] %d configs deleted: %s", len(deleted_keys), ", ".join(deleted_keys))

        if changed_data:
            changed = {}
            for k, v in changed_data.items():
                changed[k] = v

            with self._lock:
                self._configs.update(changed)

            self._listener.notify(changed)
            logger.info("[ConfigClient] %d configs changed: %s", len(changed), ", ".join(changed.keys()))

    def _apply_configs(self, data: dict) -> None:
        with self._lock:
            for k, v in data.items():
                if k.startswith("kenaito.") or k == "latest.fetch.timestamp":
                    continue
                if not isinstance(v, str):
                    continue
                self._configs[k] = v

    def _sync_loop(self) -> None:
        while not self._stop_event.is_set():
            self._stop_event.wait(self._config.sync_interval)
            if self._stop_event.is_set():
                break
            try:
                self._fetch_changed()
            except Exception:
                logger.exception("[ConfigClient] sync error")
