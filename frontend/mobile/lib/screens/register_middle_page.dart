import 'package:flutter/material.dart';
import 'register_page.dart';

// widgets
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';

/// Pre-registration screen that collects an email before showing the full form.
class RegisterMiddlePage extends StatefulWidget {
  const RegisterMiddlePage({super.key});

  @override
  State<RegisterMiddlePage> createState() => _RegisterMiddlePageState();
}

class _RegisterMiddlePageState extends State<RegisterMiddlePage> {
  final TextEditingController emailController = TextEditingController();

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
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.black,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 10),
                const Text(
                  "Enter your email to sign up for this app",
                  style: TextStyle(
                    fontSize: 15,
                    color: Colors.black,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 30),
                AppTextField(
                  label: "Email",
                  controller: emailController,
                ),
                const SizedBox(height: 25),
                PrimaryButton(
                  text: "Continue",
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (_) =>
                            RegisterPage(email: emailController.text),
                      ),
                    );
                  },
                ),
                const SizedBox(height: 10),
                const Text(
                  "-----------  or  -----------",
                  style: TextStyle(color: Colors.black),
                ),
                const SizedBox(height: 10),
                PrimaryButton(
                  text: "Sign in with Google",
                  onPressed: () {},
                  icon: Image.asset(
                    'lib/assets/G_logo.png',
                    height: 20,
                  ),
                ),
                const SizedBox(height: 10),
                PrimaryButton(
                  text: "Sign in with Github",
                  onPressed: () {},
                  icon: Image.asset(
                    'lib/assets/github_logo.png',
                    height: 20,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
