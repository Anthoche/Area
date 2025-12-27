import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:url_launcher/url_launcher.dart';

class OAuthService {
  static String get _backendBaseUrl =>
      dotenv.env['API_URL'] ?? 'http://10.0.2.2:8080';

  static String get _redirectUri => dotenv.env['REDIRECT_URI'] ?? 'test';

  Future<Map<String, dynamic>> exchangeCodeForToken(String code,
      {String? state}) async {
    try {
      final url = Uri.parse("$_backendBaseUrl/oauth/google/exchange");
      final response = await http.post(
        url,
        headers: {'Content-Type': 'application/json'},
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

  Future<void> signInWith(String provider) async {
    if (provider != 'google' && provider != 'github') {
      throw Exception("Provider non support': $provider");
    }

    try {
      if (provider == 'google') {
        final url = Uri.parse("$_backendBaseUrl/oauth/google/start")
            .replace(queryParameters: {'redirect_uri': _redirectUri});

        final response = await http.get(url);
        if (response.statusCode != 200) {
          throw Exception("Backend error ${response.statusCode}: ${response.body}");
        }

        final data = jsonDecode(response.body);
        final authUrl = data['auth_url'] as String;
        final uri = Uri.parse(authUrl);
        if (!await launchUrl(uri, mode: LaunchMode.externalApplication)) {
          throw Exception("Impossible d'ouvrir le navigateur");
        }
        return;
      }

      final url = Uri.parse("$_backendBaseUrl/oauth/github/mobile/login")
          .replace(queryParameters: {'ui_redirect': _redirectUri});
      if (!await launchUrl(url, mode: LaunchMode.externalApplication)) {
        throw Exception("Impossible d'ouvrir le navigateur");
      }
    } catch (e) {
      print("Erreur signInWith: $e");
      rethrow;
    }
  }
}
