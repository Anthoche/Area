import 'dart:convert';

import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;
import 'package:url_launcher/url_launcher.dart';

/// Handles OAuth login flows through the backend.
class OAuthService {
  static String get _backendBaseUrl =>
      dotenv.env['API_URL'] ?? 'http://10.0.2.2:8080';

  static String get _redirectUri => dotenv.env['REDIRECT_URI'] ?? 'test';
  final _storage = const FlutterSecureStorage();

  /// Exchanges an OAuth code for a backend token payload.
  Future<Map<String, dynamic>> exchangeCodeForToken(
    String code, {
    String? state,
  }) async {
    try {
      final url = Uri.parse("$_backendBaseUrl/oauth/google/exchange");
      final userId = await _storage.read(key: 'user_id');
      final headers = <String, String>{'Content-Type': 'application/json'};
      if (userId != null && userId.isNotEmpty) {
        headers['X-User-ID'] = userId;
      }
      final response = await http.post(
        url,
        headers: headers,
        body: jsonEncode({
          'code': code,
          'state': state ?? '',
          'redirect_uri': _redirectUri,
        }),
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      } else {
        throw Exception(
            'Backend exchange error ${response.statusCode}: ${response.body}');
      }
    } catch (e) {
      print("Erreur exchangeCodeForToken: $e");
      rethrow;
    }
  }

  /// Starts the OAuth flow for the requested provider.
  Future<void> signInWith(String provider) async {
    if (provider != 'google' && provider != 'github') {
      throw Exception("Provider non support': $provider");
    }

    try {
      final userId = await _storage.read(key: 'user_id');
      if (provider == 'google') {
        final query = <String, String>{'redirect_uri': _redirectUri};
        if (userId != null && userId.isNotEmpty) {
          query['user_id'] = userId;
        }
        final url = Uri.parse("$_backendBaseUrl/oauth/google/start")
            .replace(queryParameters: query);

        final response = await http.get(url);
        if (response.statusCode != 200) {
          throw Exception(
              "Backend error ${response.statusCode}: ${response.body}");
        }

        final data = jsonDecode(response.body);
        final authUrl = data['auth_url'] as String;
        final uri = Uri.parse(authUrl);
        if (!await launchUrl(uri, mode: LaunchMode.externalApplication)) {
          throw Exception("Impossible d'ouvrir le navigateur");
        }
        return;
      }

      final query = <String, String>{'ui_redirect': _redirectUri};
      if (userId != null && userId.isNotEmpty) {
        query['user_id'] = userId;
      }
      final url = Uri.parse("$_backendBaseUrl/oauth/github/mobile/login")
          .replace(queryParameters: query);
      if (!await launchUrl(url, mode: LaunchMode.externalApplication)) {
        throw Exception("Impossible d'ouvrir le navigateur");
      }
    } catch (e) {
      print("Erreur signInWith: $e");
      rethrow;
    }
  }
}
