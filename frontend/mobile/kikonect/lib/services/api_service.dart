import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class ApiService {
  static final String _baseUrl = dotenv.env['API_URL'] ?? 'http://10.0.2.2:8080';
  static String get baseUrl => _baseUrl;
  final _storage = const FlutterSecureStorage();

  Future<Map<String, String>> _getHeaders() async {
    final token = await _storage.read(key: 'jwt_token');
    final userId = await _storage.read(key: 'user_id');
    return {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer $token',
      if (userId != null && userId.isNotEmpty) 'X-User-ID': userId,
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

  /// Fetches workflows for the current user.
  Future<List<dynamic>> getWorkflows() async {
    final userId = await _storage.read(key: 'user_id');
    if (userId == null || userId.isEmpty) {
      throw Exception('Missing user id. Please login again.');
    }

    final url = Uri.parse('$_baseUrl/workflows');
    final headers = await _getHeaders();

    final response = await http.get(url, headers: headers);

    if (response.statusCode == 200) {
      final body = jsonDecode(response.body);
      if (body is List) {
        return body;
      }
      return [];
    } else {
      throw Exception('Failed to load workflows: ${response.statusCode} ${response.body}');
    }
  }

  /// Sends the created area configuration to the backend.
  Future<void> createArea(Map<String, dynamic> areaData) async {
    final url = Uri.parse('$_baseUrl/workflows');
    final headers = await _getHeaders();

    final response = await http.post(
      url,
      headers: headers,
      body: jsonEncode(areaData),
    );

    if (response.statusCode != 200 && response.statusCode != 201) {
      throw Exception('Failed to create area: ${response.statusCode} ${response.body}');
    }
  }

  /// Triggers a manual workflow run.
  Future<void> triggerWorkflow(int workflowId, Map<String, dynamic> payload) async {
    final url = Uri.parse('$_baseUrl/workflows/$workflowId/trigger');
    final headers = await _getHeaders();

    final response = await http.post(
      url,
      headers: headers,
      body: jsonEncode(payload),
    );

    if (response.statusCode != 202 && response.statusCode != 200) {
      throw Exception('Failed to trigger workflow: ${response.statusCode} ${response.body}');
    }
  }
}
