import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

import 'register_middle_page.dart';

import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';

class LoginPage extends StatefulWidget {
  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  // Controllers
  final emailController = TextEditingController();
  final passwordController = TextEditingController();
  String? errorMessage;
  late final String apiBase;

  @override
  void initState() {
    super.initState();
    apiBase = dotenv.env['API_URL'] ?? 'http://localhost:8080';
  }

  @override
  void dispose() {
    emailController.dispose();
    passwordController.dispose();
    super.dispose();
  }

  // isValidEmail is a function that will check if the email is valid
  bool isValidEmail(String email) {
    final emailRegex = RegExp(r'^[\w\.-]+@[\w\.-]+\.\w+$');
    return emailRegex.hasMatch(email);
  }

  // Validate empty fields
  bool isFieldsEmpty() {
    return emailController.text.isEmpty || passwordController.text.isEmpty;
  }

  // Login request
  Future<bool> loginUser() async {
    final url = Uri.parse("$apiBase/login");
    final body = {
      "email": emailController.text,
      "password": passwordController.text,
    };
    try {
      final response = await http.post(
        url,
        headers: {"Content-Type": "application/json"},
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        setState(() => errorMessage = null);
        return true;
      }
      if (response.statusCode == 401 || response.statusCode == 403) {
        setState(() => errorMessage = "Invalid email or password.");
        return false;
      }
      final error = _extractError(response.body);
      setState(() => errorMessage = error ?? "Server error. Please try again later.");
      return false;
    } catch (_) {
      setState(() => errorMessage = "Network error. Please check your connection or backend.");
      return false;
    }
  }

  String? _extractError(String body) {
    try {
      final decoded = jsonDecode(body);
      if (decoded is Map && decoded["error"] is String) return decoded["error"];
    } catch (_) {
      return null;
    }
    return null;
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
              // Logo
              Image.asset(
                'lib/assets/Kikonect_logo.png',
                height: 250,
                width: 250,
                alignment: Alignment.center,
              ),
              // EMAIL FIELD
              AppTextField(label: "Email", controller: emailController),
              const SizedBox(height: 25),
              // PASSWORD FIELD
              AppTextField(label: "Password", obscure: true, controller: passwordController),
              const SizedBox(height: 30),
              // LOGIN BUTTON
              PrimaryButton(
                text: "Login",
                onPressed: () async {
                  setState(() => errorMessage = null);
                  if (isFieldsEmpty()) {
                    setState(() => errorMessage = "Please fill in all fields.");
                    return;
                  }
                  if (!isValidEmail(emailController.text)) {
                    setState(() => errorMessage = "Please enter a valid email address.");
                    return;
                  }
                  await loginUser();
                },
              ),
              if (errorMessage != null) ...[
                const SizedBox(height: 12),
                Text(
                  errorMessage!,
                  style: const TextStyle(color: Colors.red),
                  textAlign: TextAlign.center,
                ),
              ],
              const SizedBox(height: 10),
              // Separation text
              const Text("──────────  or  ──────────"),
              const SizedBox(height: 10),
              // LOGIN google BUTTON
              PrimaryButton(
                text: "Login with Google",
                onPressed: () {},
                icon: Image.asset(
                  'lib/assets/G_logo.png',
                  height: 20,
                ),
              ),
              const SizedBox(height: 10),
              // LOGIN github BUTTON
              PrimaryButton(
                text: "Login with Github",
                onPressed: () {},
                icon: Image.asset(
                  'lib/assets/github_logo.png',
                  height: 20,
                ),
              ),
              const SizedBox(height: 20),
              // Register link
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text("Don't have an account? "),
                  TextButton(
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(builder: (_) => RegisterMiddlePage()),
                      );
                    },
                    child: const Text("Sign up", style: TextStyle(decoration: TextDecoration.underline),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
