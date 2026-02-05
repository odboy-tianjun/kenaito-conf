import logging
import threading
from typing import Callable

ChangeHandler = Callable[[dict[str, str]], None]

logger = logging.getLogger("kenaito_conf")


class Listener:
    def __init__(self):
        self._handlers: list[ChangeHandler] = []
        self._lock = threading.Lock()

    def on_change(self, handler: ChangeHandler) -> None:
        with self._lock:
            self._handlers.append(handler)

    def notify(self, changed: dict[str, str]) -> None:
        with self._lock:
            handlers = list(self._handlers)
        for h in handlers:
            try:
                h(changed)
            except Exception:
                logger.exception("[ConfigClient] change handler error")
