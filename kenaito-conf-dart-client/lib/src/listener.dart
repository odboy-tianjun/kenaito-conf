import 'dart:io';

/// 变更监听器
class Listener {
  final List<Function> _handlers = [];

  /// 注册变更回调
  void onChange(Function handler) {
    _handlers.add(handler);
  }

  /// 通知所有监听器
  void notify(Map<String, String> changed) {
    for (final handler in _handlers) {
      try {
        handler(changed);
      } catch (e) {
        stderr.writeln('[ConfigClient] change handler error: $e');
      }
    }
  }
}
