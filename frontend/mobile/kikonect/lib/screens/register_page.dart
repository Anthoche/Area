import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:flutter_dotenv/flutter_dotenv.dart';

// widgets
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
      Navigator.pushReplacement(
        context,
        MaterialPageRoute(builder: (context) => const Homepage()),
      );
    } else {
      if (mounted) {
        showErrorPopup('Something went wrong');
      }
    }
  }

  /// Returns true when both password fields match.
  bool passwordMatch() {
    return passwordController.text == confirmPasswordController.text;
  }

  /// Checks whether the email string matches a simple pattern.
  bool isValidEmail(String email) {
    final emailRegex = RegExp(r'^[\w\.-]+@[\w\.-]+\.\w+$');
    return emailRegex.hasMatch(email);
  }

  /// Returns true if any required field is empty.
  bool isFieldsEmpty() {
    return firstNameController.text.isEmpty ||
        lastNameController.text.isEmpty ||
        emailController.text.isEmpty ||
        passwordController.text.isEmpty ||
        confirmPasswordController.text.isEmpty;
  }

  /// Shows a one-off error dialog with the provided message.
  void showErrorPopup(String message) {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text(
            "Error",
            style: TextStyle(
              fontWeight: FontWeight.bold,
              color: Colors.red,
            ),
          ),
          content: Text(message),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text(
                "OK",
                style: TextStyle(color: Colors.blue),
              ),
            )
          ],
        );
      },
    );
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
                // Brand logo.
                Image.asset(
                  'lib/assets/Kikonect_logo.png',
                  height: 250,
                  width: 250,
                  alignment: Alignment.center,
                ),
                const Text(
                  "Create an account",
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.black,
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
                AppTextField(label: "Password",
                    obscure: true,
                    controller: passwordController),
                const SizedBox(height: 30),
                AppTextField(label: "Confirm Password",
                    obscure: true,
                    controller: confirmPasswordController),
                const SizedBox(height: 30),
                PrimaryButton(
                  text: "Register",
                  onPressed: () async {
                    if (isFieldsEmpty()) {
                      showErrorPopup("Please fill in all fields.");
                      return;
                    }
                    if (!isValidEmail(emailController.text)) {
                      showErrorPopup("Please enter a valid email address.");
                      return;
                    }
                    if (!passwordMatch()) {
                      showErrorPopup("Passwords do not match");
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
    );
  }
}
