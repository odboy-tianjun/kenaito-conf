import json
import os

CACHE_FILE = "config-cache.json"


class Cache:
    def __init__(self, cache_dir: str = "."):
        self.cache_dir = cache_dir

    def write(self, data: dict) -> None:
        path = os.path.join(self.cache_dir, CACHE_FILE)
        with open(path, "w", encoding="utf-8") as f:
            json.dump(data, f, ensure_ascii=False)

    def read(self) -> dict | None:
        path = os.path.join(self.cache_dir, CACHE_FILE)
        try:
            with open(path, "r", encoding="utf-8") as f:
                return json.load(f)
        except (FileNotFoundError, json.JSONDecodeError):
            return None
