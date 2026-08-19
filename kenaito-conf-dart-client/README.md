# Kenaito Config Center Dart Client

[English](README_EN.md) | 中文

Kenaito 配置中心 Dart 客户端 SDK，适用于 Dart 2.12+ 项目（支持空安全）。

## 安装

### 方式一：本地路径引入

在你的项目 `pubspec.yaml` 中添加：

```yaml
dependencies:
  kenaito_conf:
    path: /path/to/kenaito-conf-dart-client
```

### 方式二：直接引入

```dart
import '/path/to/kenaito-conf-dart-client/lib/kenaito_conf.dart';
```

## 使用示例

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

  // 监听配置变更
  client.onChange((Map<String, String> changed) {
    changed.forEach((key, value) {
      print('config changed: $key = $value');
    });
  });

  // 启动客户端（拉取配置）
  client.start();
  await Future.delayed(Duration(seconds: 2)); // 等待异步完成

  // 获取配置
  final value = client.get('database.host');
  final allConfigs = client.getAll();

  // 同步循环
  while (true) {
    client.syncOnce();
    await Future.delayed(Duration(seconds: config.syncInterval));
  }
}
```

## 配置项

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| baseUrl | String | 配置中心后端地址 | http://localhost:8080 |
| appName | String | 应用名称 | 空 |
| env | String | 环境（dev/daily/prod） | dev |
| version | String | 配置版本 | latest |
| token | String | 应用 Token | 空 |
| syncInterval | int | 同步间隔（秒） | 10 |
| ignoreGlobal | bool | 是否忽略全局配置 | false |
| cacheDir | String | 缓存目录 | . |

## 功能特性

- ✅ 全量拉取配置
- ✅ 增量同步变更
- ✅ 本地文件缓存容灾
- ✅ 配置变更监听回调
- ✅ 删除配置感知（deletedKeys）
- ✅ 空安全支持
