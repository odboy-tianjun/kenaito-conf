import 'dart:io';
import 'dart:convert';

/// 文件缓存
class Cache {
  final String _cacheFile;

  Cache(String cacheDir) : _cacheFile = '$cacheDir/config-cache.json';

  /// 读取缓存
  Map<String, dynamic>? read() {
    try {
      final file = File(_cacheFile);
      if (!file.existsSync()) {
        return null;
      }
      final content = file.readAsStringSync();
      return json.decode(content) as Map<String, dynamic>;
    } catch (e) {
      stderr.writeln('[ConfigClient] cache read error: $e');
      return null;
    }
  }

  /// 写入缓存
  void write(Map<String, dynamic> data) {
    try {
      final file = File(_cacheFile);
      file.writeAsStringSync(json.encode(data));
    } catch (e) {
      stderr.writeln('[ConfigClient] cache write error: $e');
    }
  }
}
