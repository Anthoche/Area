import 'package:flutter/material.dart';
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

  //register user is the function that will send data to the backend
  Future<void> registerUser() async {
    final url = Uri.parse("http://localhost:8080/register");
    final body = {
      "first_name": firstNameController.text,
      "last_name": lastNameController.text,
      "email": emailController.text,
      "password": passwordController.text,
    };
    final response = await http.post(
      url,
      headers: {"Content-Type": "application/json"},
      body: jsonEncode(body),
    );
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

  // showErrorPopup is a function that will show an error popup
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
                PrimaryButton(text: "Register",
                  onPressed: () async {
                    if (isFieldsEmpty()) {
                      showErrorPopup("Please fill in all fields.");
                      return;
                    }
                    if (!isValidEmail(emailController.text)) {
                      showErrorPopup("Please enter a valid email address.");
                      return; // stop ici
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
