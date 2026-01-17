import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Stores app-level configuration in secure storage.
class AppConfig {
  static const String _baseUrlKey = 'api_base_url';
  static const FlutterSecureStorage _storage = FlutterSecureStorage();
  static String _baseUrl = _defaultBaseUrl;
  static final _BaseUrlNotifier _baseUrlNotifier =
      _BaseUrlNotifier(_defaultBaseUrl);
  static ValueListenable<String> get baseUrlListenable => _baseUrlNotifier;

  static String get _defaultBaseUrl =>
      dotenv.env['API_URL'] ?? 'http://10.0.2.2:8080';

  /// Returns the current backend base URL.
  static String get baseUrl => _baseUrl;

  /// Loads the persisted config and updates the cached base URL.
  static Future<void> load() async {
    try {
      final stored = await _storage.read(key: _baseUrlKey);
      _updateBaseUrl(_normalize(stored) ?? _defaultBaseUrl);
    } catch (_) {
      _updateBaseUrl(_defaultBaseUrl);
    }
  }

  /// Saves the backend base URL. Use null or empty to reset to default.
  static Future<void> setBaseUrl(String? baseUrl) async {
    final cleaned = _normalize(baseUrl);
    if (cleaned == null) {
      await _storage.delete(key: _baseUrlKey);
      _updateBaseUrl(_defaultBaseUrl);
      return;
    }
    await _storage.write(key: _baseUrlKey, value: cleaned);
    _updateBaseUrl(cleaned);
  }

  /// Returns the stored base URL without falling back to defaults.
  static Future<String?> getStoredBaseUrl() async {
    final stored = await _storage.read(key: _baseUrlKey);
    return _normalize(stored);
  }

  static String? _normalize(String? raw) {
    if (raw == null) return null;
    var value = raw.trim();
    if (value.isEmpty) return null;
    while (value.endsWith('/')) {
      value = value.substring(0, value.length - 1);
    }
    return value;
  }

  static void _updateBaseUrl(String value) {
    _baseUrl = value;
    _baseUrlNotifier.setValue(value);
  }
}

class _BaseUrlNotifier extends ValueNotifier<String> {
  _BaseUrlNotifier(super.value);

  void setValue(String newValue) {
    if (newValue == value) {
      notifyListeners();
      return;
    }
    value = newValue;
  }
}
