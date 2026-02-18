import logging
import threading
import time

import requests

from kenaito_conf.cache import Cache
from kenaito_conf.config import Config
from kenaito_conf.crypto import decrypt_config
from kenaito_conf.listener import Listener

logger = logging.getLogger("kenaito_conf")


class ConfigClient:
    def __init__(self, config: Config):
        self._config = config
        self._cache = Cache(config.cache_dir)
        self._listener = Listener()
        self._configs: dict[str, str] = {}
        self._secret: str = ""
        self._global_secret: str = ""
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
        url = (
            f"{self._config.base_url}/api/fetchAllConfigs"
            f"?appName={self._config.app_name}"
            f"&env={self._config.env}"
            f"&version={self._config.version}"
            f"&ignoreGlobal={str(self._config.ignore_global).lower()}"
            f"&token={self._config.token}"
        )
        resp = requests.get(url, timeout=10)
        resp.raise_for_status()
        data = resp.json()

        if data.get("code", 0) != 0:
            raise RuntimeError(f"server error: {data.get('msg', 'unknown')}")

        secret = data.get("kenaito.config.encrypt-secret")
        if secret:
            self._secret = secret

        global_secret = data.get("kenaito.config.global-encrypt-secret")
        if global_secret:
            self._global_secret = global_secret

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

        if changed_data:
            changed = {}
            for k, v in changed_data.items():
                decrypted = self._try_decrypt(v)
                changed[k] = decrypted

            with self._lock:
                self._configs.update(changed)

            self._listener.notify(changed)
            logger.info("[ConfigClient] %d configs changed", len(changed))

    def _apply_configs(self, data: dict) -> None:
        with self._lock:
            for k, v in data.items():
                if k.startswith("kenaito.") or k == "latest.fetch.timestamp":
                    continue
                if not isinstance(v, str):
                    continue
                self._configs[k] = self._try_decrypt(v)

    def _try_decrypt(self, value: str) -> str:
        decrypted = decrypt_config(self._secret, value)
        if decrypted == value and self._global_secret:
            decrypted = decrypt_config(self._global_secret, value)
        return decrypted

    def _sync_loop(self) -> None:
        while not self._stop_event.is_set():
            self._stop_event.wait(self._config.sync_interval)
            if self._stop_event.is_set():
                break
            try:
                self._fetch_changed()
            except Exception:
                logger.exception("[ConfigClient] sync error")
