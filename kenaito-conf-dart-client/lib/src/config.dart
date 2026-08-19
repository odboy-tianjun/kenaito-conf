/// 配置类
class Config {
  final String baseUrl;
  final String appName;
  final String env;
  final String version;
  final String token;
  final int syncInterval;
  final bool ignoreGlobal;
  final String cacheDir;

  Config({
    this.baseUrl = 'http://localhost:8080',
    this.appName = '',
    this.env = 'dev',
    this.version = 'latest',
    this.token = '',
    this.syncInterval = 10,
    this.ignoreGlobal = false,
    this.cacheDir = '.',
  });
}
