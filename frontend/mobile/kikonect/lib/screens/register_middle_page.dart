import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';

import '../services/oauth_service.dart';
import '../utils/ui_feedback.dart';
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';
import 'home_page.dart';
import 'register_page.dart';

/// Displays the entry screen for registration and OAuth sign-in.
class RegisterMiddlePage extends StatefulWidget {
  const RegisterMiddlePage({super.key});

  @override
  State<RegisterMiddlePage> createState() => _RegisterMiddlePageState();
}

class _RegisterMiddlePageState extends State<RegisterMiddlePage> {
  final TextEditingController emailController = TextEditingController();
  final OAuthService _authService = OAuthService();

  StreamSubscription? _sub;
  final _appLinks = AppLinks();

  @override
  void initState() {
    super.initState();
    _initDeepLinkListener();
  }

  @override
  void dispose() {
    _sub?.cancel();
    emailController.dispose();
    super.dispose();
  }

  Future<void> _initDeepLinkListener() async {
    _sub = _appLinks.uriLinkStream.listen((Uri? uri) {
      if (uri != null) _handleDeepLink(uri);
    }, onError: (err) {
      debugPrint("Erreur Deep Link: $err");
    });
  }

  Future<void> _handleDeepLink(Uri uri) async {
    if (uri.scheme != 'http' || uri.host != 'localhost' || uri.port != 8080) {
      return;
    }

    final code = uri.queryParameters['code'];
    final state = uri.queryParameters['state'];
    final error = uri.queryParameters['error'];

    if (error != null) {
      showErrorDialog(context, 'OAuth error: $error', title: 'Erreur');
      return;
    }

    if (code == null || code.isEmpty) {
      return;
    }

    try {
      final result =
          await _authService.exchangeCodeForToken(code, state: state);
      final token = result['token'];
      final email = result['email'] ?? '';

      if (token == null || (token is String && token.isEmpty)) {
        showErrorDialog(context, "Token manquant dans la réponse",
            title: 'Erreur');
        return;
      }

      if (mounted) {
        showDialog(
          context: context,
          barrierDismissible: false,
          builder: (c) => const Center(child: CircularProgressIndicator()),
        );
      }

      await _saveToken(token.toString(), email.toString());

      if (mounted && Navigator.canPop(context)) {
        Navigator.pop(context);
      }

      if (mounted) {
        Navigator.pushAndRemoveUntil(
          context,
          MaterialPageRoute(builder: (_) => const Homepage()),
          (route) => false,
        );
      }
    } catch (e) {
      showErrorDialog(context, 'Erreur exchange token: $e', title: 'Erreur');
    }
  }

  Future<void> _saveToken(String token, String email) async {
    print("Token saved: $token (email: $email)");
  }

  void _onOAuthSignIn(String provider) async {
    try {
      await _authService.signInWith(provider);
    } catch (e) {
      showErrorDialog(context, "Erreur lancement OAuth: $e", title: 'Erreur');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: SizedBox(
          width: 300,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Image.asset(
                'lib/assets/Kikonect_logo.png',
                height: 250,
                width: 250,
                alignment: Alignment.center,
              ),
              const Text(
                "Create an account",
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                  color: Colors.black,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 10),
              const Text(
                "Enter your email to sign up for this app",
                style: TextStyle(fontSize: 15, color: Colors.black),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 30),
              AppTextField(label: "Email", controller: emailController),
              const SizedBox(height: 25),
              PrimaryButton(
                text: "Continue",
                onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => RegisterPage(email: emailController.text),
                    ),
                  );
                },
              ),
              const SizedBox(height: 10),
              const Text(
                "──────────  or  ──────────",
                style: TextStyle(color: Colors.black),
              ),
              const SizedBox(height: 10),
              PrimaryButton(
                text: "Sign in with Google",
                onPressed: () => _onOAuthSignIn('google'),
                icon: Image.asset('lib/assets/G_logo.png', height: 20),
              ),
              const SizedBox(height: 10),
              PrimaryButton(
                text: "Sign in with Github",
                onPressed: () => _onOAuthSignIn('github'),
                icon: Image.asset('lib/assets/github_logo.png', height: 20),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
