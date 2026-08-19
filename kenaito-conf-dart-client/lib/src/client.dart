import 'dart:io';
import 'dart:convert';
import 'config.dart';
import 'cache.dart';
import 'listener.dart';

/// 配置中心客户端
class Client {
  final Config _config;
  final Cache _cache;
  final Listener _listener;
  final Map<String, String> _configs = {};
  int _lastTime = 0;
  bool _started = false;
  bool stop = false;

  Client(Config config)
      : _config = config,
        _cache = Cache(config.cacheDir),
        _listener = Listener();

  /// 注册变更回调
  void onChange(Function handler) {
    _listener.onChange(handler);
  }

  /// 获取配置值
  String? get(String key) {
    return _configs[key];
  }

  /// 获取所有配置
  Map<String, String> getAll() {
    return Map.from(_configs);
  }

  /// 启动客户端
  Future<void> start() async {
    if (_started) {
      return;
    }
    _started = true;

    try {
      await _fetchAll();
    } catch (e) {
      stderr.writeln('[ConfigClient] fetch all configs error: $e, trying cache');
      final cached = _cache.read();
      if (cached != null) {
        _applyConfigs(cached);
        stderr.writeln('[ConfigClient] loaded ${_configs.length} configs from cache');
      } else {
        throw Exception('fetch configs failed and no cache: $e');
      }
    }
  }

  /// 同步一次
  Future<void> syncOnce() async {
    try {
      await _fetchChanged();
    } catch (e) {
      stderr.writeln('[ConfigClient] sync error: $e');
    }
  }

  /// HTTP POST 请求
  Future<Map<String, dynamic>> _httpPost(String url, Map<String, dynamic> payload) async {
    final uri = Uri.parse(url);
    final httpClient = HttpClient();
    
    try {
      final request = await httpClient.postUrl(uri).timeout(Duration(seconds: 10));
      
      request.headers.set('Content-Type', 'application/json');
      request.write(json.encode(payload));
      
      final response = await request.close().timeout(Duration(seconds: 10));
      
      if (response.statusCode != 200) {
        throw Exception('HTTP ${response.statusCode}');
      }
      
      final body = await response.transform(utf8.decoder).join();
      
      try {
        return json.decode(body) as Map<String, dynamic>;
      } catch (e) {
        throw Exception('JSON decode error: $e');
      }
    } finally {
      httpClient.close();
    }
  }

  /// 全量拉取配置
  Future<void> _fetchAll() async {
    final url = '${_config.baseUrl}/api/fetchAllConfigs';
    final payload = {
      'appName': _config.appName,
      'env': _config.env,
      'version': _config.version,
      'ignoreGlobal': _config.ignoreGlobal,
      'token': _config.token,
    };

    final result = await _httpPost(url, payload);

    if (result.containsKey('code') && result['code'] != 0) {
      throw Exception('server error: ${result['msg'] ?? 'unknown'}');
    }

    if (result.containsKey('latest.fetch.timestamp')) {
      final ts = result['latest.fetch.timestamp'];
      if (ts is num) {
        _lastTime = ts.toInt();
      } else if (ts is String) {
        _lastTime = int.tryParse(ts) ?? 0;
      }
    }

    _cache.write(result);
    _applyConfigs(result);

    stderr.writeln('[ConfigClient] fetched ${_configs.length} configs');
  }

  /// 增量拉取变更
  Future<void> _fetchChanged() async {
    final url = '${_config.baseUrl}/api/fetchChangedConfigs';
    final payload = {
      'appName': _config.appName,
      'env': _config.env,
      'version': _config.version,
      'lastTime': _lastTime,
      'ignoreGlobal': _config.ignoreGlobal,
      'token': _config.token,
    };

    final data = await _httpPost(url, payload);

    if (data.containsKey('code') && data['code'] != 0) {
      throw Exception('server error code: ${data['code']}');
    }

    final changedData = (data['data'] ?? {}) as Map<String, dynamic>;

    if (changedData.containsKey('latest.fetch.timestamp')) {
      final ts = changedData['latest.fetch.timestamp'];
      if (ts is num) {
        _lastTime = ts.toInt();
      } else if (ts is String) {
        _lastTime = int.tryParse(ts) ?? 0;
      }
      changedData.remove('latest.fetch.timestamp');
    }

    // 处理删除的配置
    if (changedData.containsKey('deletedKeys') && changedData['deletedKeys'] != null) {
      final deletedKeysStr = changedData['deletedKeys'] as String;
      if (deletedKeysStr.isNotEmpty) {
        final deletedKeys = deletedKeysStr.split(',');
        for (final key in deletedKeys) {
          _configs.remove(key);
        }
        stderr.writeln('[ConfigClient] ${deletedKeys.length} configs deleted: ${deletedKeys.join(', ')}');
      }
      changedData.remove('deletedKeys');
    }

    // 处理变更的配置
    final changed = <String, String>{};
    changedData.forEach((key, value) {
      if (value is String) {
        _configs[key] = value;
        changed[key] = value;
      }
    });

    if (changed.isNotEmpty) {
      _listener.notify(changed);
      stderr.writeln('[ConfigClient] ${changed.length} configs changed: ${changed.keys.join(', ')}');
    }
  }

  /// 应用配置到本地存储
  void _applyConfigs(Map<String, dynamic> data) {
    data.forEach((key, value) {
      if (key.startsWith('kenaito.') || key == 'latest.fetch.timestamp') {
        return;
      }
      if (value is String) {
        _configs[key] = value;
      }
    });
  }
}
