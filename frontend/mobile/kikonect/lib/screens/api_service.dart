import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class ApiService {
  static final String _baseUrl = dotenv.env['API_URL'] ?? 'http://10.0.2.2:8080';
  final _storage = const FlutterSecureStorage();

  Future<Map<String, String>> _getHeaders() async {
    final token = await _storage.read(key: 'jwt_token');
    return {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer $token',
    };
  }

  /// Fetches the list of available services with their triggers and actions.
  Future<List<dynamic>> getServices() async {
    final url = Uri.parse('$_baseUrl/areas');
    final headers = await _getHeaders();

    final response = await http.get(url, headers: headers);

    if (response.statusCode == 200) {
      final body = jsonDecode(response.body);
      if (body is Map<String, dynamic>) {
        return (body['services'] as List?) ?? [];
      }
      return [];
    } else {
      throw Exception('Failed to load services: ${response.statusCode} ${response.body}');
    }
  }

  /// Sends the created area configuration to the backend.
  Future<void> createArea(Map<String, dynamic> areaData) async {
    final url = Uri.parse('$_baseUrl/areas');
    final headers = await _getHeaders();

    final response = await http.post(
      url,
      headers: headers,
      body: jsonEncode(areaData),
    );

    if (response.statusCode != 200 && response.statusCode != 201) {
      throw Exception('Failed to create area: ${response.body}');
    }
  }
}
