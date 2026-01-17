import 'dart:async';
import 'dart:convert';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;

import '../services/app_config.dart';
import '../services/oauth_service.dart';
import '../utils/ui_feedback.dart';
import '../utils/validators.dart';
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';
import 'home_page.dart';
import 'register_middle_page.dart';

/// Displays the login screen and handles OAuth callbacks.
class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final emailController = TextEditingController();
  final passwordController = TextEditingController();

  final OAuthService _authService = OAuthService();
  StreamSubscription? _sub;
  final _appLinks = AppLinks();
  final _storage = const FlutterSecureStorage();

  @override
  void initState() {
    super.initState();
    _initDeepLinks();
  }

  @override
  void dispose() {
    _sub?.cancel();
    emailController.dispose();
    passwordController.dispose();
    super.dispose();
  }

  // Listen for OAuth callbacks.
  Future<void> _initDeepLinks() async {
    _sub = _appLinks.uriLinkStream.listen((uri) {
      if (mounted && ModalRoute.of(context)?.isCurrent == true) {
        _processDeepLink(uri.toString());
      }
    });
  }

  void _processDeepLink(String link) async {
    final uri = Uri.parse(link);

    if (uri.scheme == 'http' && uri.host == 'localhost' && uri.port == 8080) {
      final code = uri.queryParameters['code'];
      final state = uri.queryParameters['state'];
      final error = uri.queryParameters['error'];
      final userIdFromQuery = uri.queryParameters['user_id'];
      final tokenIdFromQuery = uri.queryParameters['token_id'];
      final googleEmail = uri.queryParameters['google_email'];
      final githubEmail = uri.queryParameters['github_email'];
      final githubLogin = uri.queryParameters['github_login'];

      if (error != null) {
        showErrorDialog(context, 'OAuth error: $error');
        return;
      }

      if (userIdFromQuery != null && userIdFromQuery.isNotEmpty) {
        await _saveUserId(userIdFromQuery);
      }

      if (tokenIdFromQuery != null && tokenIdFromQuery.isNotEmpty) {
        if (githubLogin != null || githubEmail != null) {
          await _saveGithubTokenId(tokenIdFromQuery);
        } else if (googleEmail != null) {
          await _saveGoogleTokenId(tokenIdFromQuery);
        }
      }

      if (tokenIdFromQuery != null &&
          tokenIdFromQuery.isNotEmpty &&
          (code == null || code.isEmpty)) {
        if (mounted) {
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (_) => const Homepage()),
          );
        }
        return;
      }

      if (code != null && code.isNotEmpty) {
        print('Code OAuth Google reçu: $code');

        try {
          // Exchange the code for a token via the backend.
          final result =
              await _authService.exchangeCodeForToken(code, state: state);
          final token = result['token'];
          final email = result['email'] ?? '';
          final userId = result['user_id'];
          final tokenId = result['token_id'];

          print('Token reçu: $token, Email: $email');

          // Save the token.
          if (token != null) {
            await _saveToken(token.toString());
          }
          if (userId != null) {
            await _saveUserId(userId);
          }
          if (tokenId != null) {
            await _saveGoogleTokenId(tokenId.toString());
          }

          // Navigate to the home page.
          if (mounted) {
            Navigator.pushReplacement(
              context,
              MaterialPageRoute(builder: (_) => const Homepage()),
            );
          }
        } catch (e) {
          showErrorDialog(context, 'Erreur échange token: $e');
        }
      }
      return;
    }
  }

  Future<void> _saveToken(String token) async {
    await _storage.write(key: 'jwt_token', value: token);
    print("Token saved securely");
  }

  Future<void> _saveUserId(dynamic userId) async {
    if (userId == null) return;
    await _storage.write(key: 'user_id', value: userId.toString());
    print("User id saved securely");
  }

  Future<void> _saveGoogleTokenId(String tokenId) async {
    await _storage.write(key: 'google_token_id', value: tokenId);
    print("Google token id saved securely");
  }

  Future<void> _saveGithubTokenId(String tokenId) async {
    await _storage.write(key: 'github_token_id', value: tokenId);
    print("GitHub token id saved securely");
  }

  Future<void> _loginUser() async {
    final url = Uri.parse("${AppConfig.baseUrl}/login");

    try {
      final res = await http.post(url,
          headers: {"Content-Type": "application/json"},
          body: jsonEncode({
            "email": emailController.text,
            "password": passwordController.text,
          }));

      if (res.statusCode == 200) {
        try {
          final body = jsonDecode(res.body);
          if (body is Map) {
            if (body.containsKey('token')) {
              await _saveToken(body['token'].toString());
            }
            if (body.containsKey('id')) {
              await _saveUserId(body['id']);
            }
          }
        } catch (e) {
          print("Could not parse token from login response: $e");
        }

        if (mounted) {
          Navigator.pushReplacement(
              context, MaterialPageRoute(builder: (_) => const Homepage()));
        }
      } else {
        showErrorDialog(context, "Login failed (${res.statusCode})");
      }
    } catch (e) {
      showErrorDialog(context, "Connection error: $e");
    }
  }

  void _loginOAuth(String provider) async {
    try {
      await _authService.signInWith(provider);
    } catch (e) {
      showErrorDialog(context, "OAuth launch error: $e");
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: SizedBox(
          width: 300,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Image.asset('lib/assets/Kikonect_logo.png', height: 250),
                AppTextField(label: "Email", controller: emailController),
                const SizedBox(height: 25),
                AppTextField(
                    label: "Password",
                    controller: passwordController,
                    obscure: true),
                const SizedBox(height: 30),
                PrimaryButton(
                  text: "Login",
                  onPressed: () {
                    if (emailController.text.isEmpty ||
                        passwordController.text.isEmpty) {
                      showErrorDialog(context, "Please fill in all fields.");
                    } else if (!isValidEmail(emailController.text)) {
                      showErrorDialog(context, "Invalid email.");
                    } else {
                      _loginUser();
                    }
                  },
                ),
                const SizedBox(height: 10),
                const Text("──────────  or  ──────────"),
                const SizedBox(height: 10),
                PrimaryButton(
                  text: "Login with Google",
                  onPressed: () => _loginOAuth('google'),
                  icon: Image.asset('lib/assets/G_logo.png', height: 20),
                ),
                const SizedBox(height: 10),
                PrimaryButton(
                  text: "Login with Github",
                  onPressed: () => _loginOAuth('github'),
                  icon: Image.asset(
                    'lib/assets/github_logo.png',
                    height: 20,
                  ),
                ),
                const SizedBox(height: 20),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Text("Don't have an account? "),
                    TextButton(
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                              builder: (_) => const RegisterMiddlePage()),
                        );
                      },
                      child: const Text(
                        "Sign up",
                        style: TextStyle(decoration: TextDecoration.underline),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
