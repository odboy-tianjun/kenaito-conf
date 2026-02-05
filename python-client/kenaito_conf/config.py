class Config:
    def __init__(
        self,
        base_url: str = "",
        app_name: str = "",
        env: str = "dev",
        version: str = "latest",
        token: str = "",
        sync_interval: int = 60,
        ignore_global: bool = False,
        cache_dir: str = ".",
    ):
        self.base_url = base_url
        self.app_name = app_name
        self.env = env
        self.version = version
        self.token = token
        self.sync_interval = sync_interval
        self.ignore_global = ignore_global
        self.cache_dir = cache_dir
