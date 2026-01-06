import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Manages the app-wide theme mode and persists the choice.
class ThemeController extends ChangeNotifier {
  ThemeController(
    this._storage, {
    ThemeMode initialMode = ThemeMode.system,
  }) : _mode = initialMode;

  static const String _storageKey = 'theme_mode';
  final FlutterSecureStorage _storage;
  ThemeMode _mode;

  ThemeMode get mode => _mode;
  bool get isDark => _mode == ThemeMode.dark;

  Future<void> load() async {
    try {
      final stored = await _storage.read(key: _storageKey);
      final loaded = stored == 'dark'
          ? ThemeMode.dark
          : stored == 'light'
              ? ThemeMode.light
              : stored == 'system'
                  ? ThemeMode.system
                  : _mode;
      if (loaded != _mode) {
        _mode = loaded;
        notifyListeners();
      }
    } catch (_) {}
  }

  Future<void> setMode(ThemeMode mode) async {
    if (_mode == mode) return;
    _mode = mode;
    notifyListeners();
    try {
      await _storage.write(
        key: _storageKey,
        value: mode == ThemeMode.dark
            ? 'dark'
            : mode == ThemeMode.light
                ? 'light'
                : 'system',
      );
    } catch (_) {}
  }
}

/// Exposes the ThemeController to the widget tree.
class ThemeScope extends InheritedNotifier<ThemeController> {
  const ThemeScope({
    super.key,
    required ThemeController controller,
    required Widget child,
  }) : super(notifier: controller, child: child);

  static ThemeController of(BuildContext context) {
    final scope =
        context.dependOnInheritedWidgetOfExactType<ThemeScope>();
    assert(scope != null, 'ThemeScope not found in widget tree.');
    return scope!.notifier!;
  }
}
