# Kenaito Config Center Dart Client

[English](README_EN.md) | 中文

Kenaito Config Center Dart client SDK, compatible with Dart 2.12+ (null-safe).

## Installation

### Option 1: Local Path

Add to your project `pubspec.yaml`:

```yaml
dependencies:
  kenaito_conf:
    path: /path/to/kenaito-conf-dart-client
```

### Option 2: Direct Import

```dart
import '/path/to/kenaito-conf-dart-client/lib/kenaito_conf.dart';
```

## Usage Example

```dart
import 'package:kenaito_conf/kenaito_conf.dart';

void main() async {
  final config = Config(
    baseUrl: 'http://localhost:8080',
    appName: 'my-app',
    env: 'daily',
    syncInterval: 10,
  );

  final client = Client(config);

  // Listen for config changes
  client.onChange((Map<String, String> changed) {
    changed.forEach((key, value) {
      print('config changed: $key = $value');
    });
  });

  // Start client (fetch configs)
  client.start();
  await Future.delayed(Duration(seconds: 2)); // Wait for async completion

  // Get config
  final value = client.get('database.host');
  final allConfigs = client.getAll();

  // Sync loop
  while (true) {
    client.syncOnce();
    await Future.delayed(Duration(seconds: config.syncInterval));
  }
}
```

## Configuration Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| baseUrl | String | Config center backend URL | http://localhost:8080 |
| appName | String | Application name | empty |
| env | String | Environment (dev/daily/prod) | dev |
| version | String | Config version | latest |
| token | String | Application token | empty |
| syncInterval | int | Sync interval (seconds) | 10 |
| ignoreGlobal | bool | Ignore global configs | false |
| cacheDir | String | Cache directory | . |

## Features

- ✅ Full config fetch
- ✅ Incremental sync
- ✅ Local file cache for disaster recovery
- ✅ Config change listener callback
- ✅ Deleted config awareness (deletedKeys)
- ✅ Null safety support
