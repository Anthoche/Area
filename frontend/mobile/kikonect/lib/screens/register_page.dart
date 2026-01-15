import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;

import '../utils/ui_feedback.dart';
import '../utils/validators.dart';
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';
import 'home_page.dart';

/// Collects full account details and posts a registration request.
class RegisterPage extends StatefulWidget {
  final String? email;

  const RegisterPage({super.key, this.email});

  @override
  State<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends State<RegisterPage> {
  // Form controllers used across validation and submission.
  final firstNameController = TextEditingController();
  final lastNameController = TextEditingController();
  final emailController = TextEditingController();
  final passwordController = TextEditingController();
  final confirmPasswordController = TextEditingController();
  final _storage = const FlutterSecureStorage();

  @override
  void initState() {
    super.initState();
    if (widget.email != null) {
      emailController.text = widget.email!;
    }
  }

  /// Sends the collected profile data to the backend registration endpoint.
  Future<void> registerUser() async {
    final apiUrl = dotenv.env['API_URL'];
    final url = Uri.parse("$apiUrl/register");
    final body = {
      "firstname": firstNameController.text,
      "lastname": lastNameController.text,
      "email": emailController.text,
      "password": passwordController.text,
    };

    final response = await http.post(
      url,
      headers: {"Content-Type": "application/json"},
      body: jsonEncode(body),
    );
    if (response.statusCode == 201 && mounted) {
      try {
        final body = jsonDecode(response.body);
        if (body is Map && body.containsKey('id')) {
          await _storage.write(key: 'user_id', value: body['id'].toString());
        }
      } catch (_) {}

      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (context) => const Homepage()),
      );
    } else {
      if (mounted) {
        showErrorDialog(
          context,
          'Something went wrong',
          buttonColor: Theme.of(context).colorScheme.primary,
        );
      }
    }
  }

  /// Returns true when both password fields match.
  bool passwordMatch() {
    return passwordController.text == confirmPasswordController.text;
  }

  /// Returns true if any required field is empty.
  bool isFieldsEmpty() {
    return firstNameController.text.isEmpty ||
        lastNameController.text.isEmpty ||
        emailController.text.isEmpty ||
        passwordController.text.isEmpty ||
        confirmPasswordController.text.isEmpty;
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
          child: Center(
            child: SizedBox(
              width: 300,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Brand logo.
                  Image.asset(
                    'lib/assets/Kikonect_logo.png',
                    height: 250,
                    width: 250,
                    alignment: Alignment.center,
                  ),
                  Text(
                    "Create an account",
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      color: colorScheme.onSurface,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 30),
                  AppTextField(
                      label: "First Name", controller: firstNameController),
                  const SizedBox(height: 30),
                  AppTextField(
                      label: "Last Name", controller: lastNameController),
                  const SizedBox(height: 30),
                  AppTextField(label: "Email", controller: emailController),
                  const SizedBox(height: 30),
                  AppTextField(
                      label: "Password",
                      obscure: true,
                      controller: passwordController),
                  const SizedBox(height: 30),
                  AppTextField(
                      label: "Confirm Password",
                      obscure: true,
                      controller: confirmPasswordController),
                  const SizedBox(height: 30),
                  PrimaryButton(
                    text: "Register",
                    onPressed: () async {
                      if (isFieldsEmpty()) {
                        showErrorDialog(context, "Please fill in all fields.",
                            buttonColor: colorScheme.primary);
                        return;
                      }
                      if (!isValidEmail(emailController.text)) {
                        showErrorDialog(
                          context,
                          "Please enter a valid email address.",
                          buttonColor: colorScheme.primary,
                        );
                        return;
                      }
                      if (!isValidPassword(passwordController.text)) {
                        showErrorDialog(
                          context,
                          "Password must be at least 8 characters, include a number and a special character.",
                          buttonColor: colorScheme.primary,
                        );
                        return;
                      }
                      if (!passwordMatch()) {
                        showErrorDialog(context, "Passwords do not match",
                            buttonColor: colorScheme.primary);
                        return;
                      }
                      await registerUser();
                    },
                  )
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
