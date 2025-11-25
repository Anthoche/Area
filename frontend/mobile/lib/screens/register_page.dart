import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

// widgets
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';

class RegisterPage extends StatefulWidget {
  @override
  State<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends State<RegisterPage> {
  // Controller
  final firstNameController = TextEditingController();
  final lastNameController = TextEditingController();
  final emailController = TextEditingController();
  final passwordController = TextEditingController();
  final confirmPasswordController = TextEditingController();
  String? errorMessage;
  late final String apiBase;

  @override
  void initState() {
    super.initState();
    apiBase = dotenv.env['API_URL'] ?? 'http://localhost:8080';
  }

  @override
  void dispose() {
    firstNameController.dispose();
    lastNameController.dispose();
    emailController.dispose();
    passwordController.dispose();
    confirmPasswordController.dispose();
    super.dispose();
  }

  //register user is the function that will send data to the backend
  Future<bool> registerUser() async {
    final url = Uri.parse("$apiBase/register");
    final body = {
      "firstname": firstNameController.text,
      "lastname": lastNameController.text,
      "email": emailController.text,
      "password": passwordController.text,
    };
    try {
      final response = await http.post(
        url,
        headers: {"Content-Type": "application/json"},
        body: jsonEncode(body),
      );

      if (response.statusCode == 201) {
        setState(() => errorMessage = null);
        return true;
      }
      if (response.statusCode == 409) {
        setState(() => errorMessage = "Email already registered.");
        return false;
      }
      final error = _extractError(response.body);
      setState(() => errorMessage = error ?? "Server error. Please try again later.");
      return false;
    } catch (e) {
      setState(() => errorMessage = "Network error. Please check your connection or backend.");
      return false;
    }
  }

  // passwordMatch is a function that will check if the password and confirm password are the same
  bool passwordMatch() {
    return passwordController.text == confirmPasswordController.text;
  }

  // isValidEmail is a function that will check if the email is valid
  bool isValidEmail(String email) {
    final emailRegex = RegExp(r'^[\w\.-]+@[\w\.-]+\.\w+$');
    return emailRegex.hasMatch(email);
  }

  // isFieldsEmpty is a function that will check if the fields are empty
  bool isFieldsEmpty() {
    return firstNameController.text.isEmpty ||
        lastNameController.text.isEmpty ||
        emailController.text.isEmpty ||
        passwordController.text.isEmpty ||
        confirmPasswordController.text.isEmpty;
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
      body: Container(
        child: Center(
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
                // Register text
                const Text(
                  "Create an account",
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.black,
                  ),
                  textAlign: TextAlign.center,
                ),
                // TEXT FIELDS
                const SizedBox(height: 30),
                AppTextField(label: "First Name", controller: firstNameController),
                const SizedBox(height: 30),
                AppTextField(label: "Last Name", controller: lastNameController),
                const SizedBox(height: 30),
                AppTextField(label: "Email", controller: emailController),
                const SizedBox(height: 30),
                AppTextField(label: "Password", obscure: true, controller: passwordController),
                const SizedBox(height: 30),
                AppTextField(label: "Confirm Password", obscure: true, controller: confirmPasswordController),
                const SizedBox(height: 30),
                // REGISTER BUTTON
                PrimaryButton(
                  text: "Register",
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
                    if (!passwordMatch()) {
                      setState(() => errorMessage = "Passwords do not match.");
                      return;
                    }
                    final ok = await registerUser();
                    if (ok && mounted) {
                      Navigator.pop(context);
                    }
                  },
                ),
                if (errorMessage != null) ...[
                  const SizedBox(height: 12),
                  Text(
                    errorMessage!,
                    style: const TextStyle(color: Colors.red),
                    textAlign: TextAlign.center,
                  ),
                ]
              ],
            ),
          ),
        ),
      ),
    );
  }
}
