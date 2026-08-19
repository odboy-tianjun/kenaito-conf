import 'dart:io';
import '../kenaito-conf-dart-client/lib/kenaito_conf.dart';

String getEnvVar(String key, String defaultValue) {
  final value = Platform.environment[key];
  return (value != null && value.isNotEmpty) ? value : defaultValue;
}

void main() async {
  // 配置值（可被远程配置覆盖）
  final configValues = <String, String>{
    'database.host': getEnvVar('DATABASE_HOST', '192.168.1.1'),
    'database.port': getEnvVar('DATABASE_PORT', '8888'),
    'database.name': getEnvVar('DATABASE_NAME', 'xxxx'),
    'database.user': getEnvVar('DATABASE_USER', 'xxxx'),
    'database.password': getEnvVar('DATABASE_PASSWORD', 'xxxxxxxx'),
  };

  final config = Config(
    baseUrl: getEnvVar('KENAITO_BASE_URL', 'http://localhost:8080'),
    appName: getEnvVar('KENAITO_APP_NAME', 'kenaito-dart-demo'),
    env: getEnvVar('KENAITO_ENV', 'daily'),
    version: getEnvVar('KENAITO_VERSION', 'latest'),
    token: getEnvVar('KENAITO_TOKEN', ''),
    syncInterval: int.parse(getEnvVar('KENAITO_SYNC_INTERVAL', '10')),
    cacheDir: getEnvVar('KENAITO_CACHE_DIR', '.'),
  );

  final client = Client(config);

  // 监听配置变更
  client.onChange((Map<String, String> changed) {
    print('Config changed:');
    changed.forEach((key, value) {
      print('  $key = $value');
      if (configValues.containsKey(key)) {
        configValues[key] = value;
        print('  [updated] $key = $value');
      }
    });
  });

  // 启动客户端
  try {
    await client.start();
    
    final remote = client.getAll();
    configValues.forEach((key, _) {
      if (remote.containsKey(key)) {
        configValues[key] = remote[key]!;
      }
    });
    print('loaded ${remote.length} configs from remote\n');
  } catch (e) {
    print('Failed to init config client, using local defaults: $e\n');
  }

  // 显示当前配置
  print('Current configs:');
  final sortedKeys = configValues.keys.toList()..sort();
  for (final key in sortedKeys) {
    print('  $key = ${configValues[key]}');
  }

  print('\nListening for config changes... (press Ctrl+C to stop)\n');

  // 信号处理
  ProcessSignal.sigint.watch().listen((_) {
    print('\nShutting down...');
    client.stop = true;
    exit(0);
  });

  // 同步循环
  while (!client.stop) {
    await client.syncOnce();
    await Future.delayed(Duration(seconds: config.syncInterval));
  }
}
