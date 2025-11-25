import 'package:flutter/material.dart';
import 'register_page.dart';

// widgets
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';

class RegisterMiddlePage extends StatelessWidget {
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
                // EMAIL FIELD
                const AppTextField(label: "Email"),
                const SizedBox(height: 25),
                // CONTINUE BUTTON
                PrimaryButton(
                  text: "Continue",
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(builder: (_) => RegisterPage()),
                    );
                  },
                ),
                const SizedBox(height: 10),
                // Separation text
                const Text(
                  "──────────  or  ──────────",
                  style: TextStyle(color: Colors.black),
                ),
                const SizedBox(height: 10),
                // LOGIN google BUTTON
                PrimaryButton(
                  text: "Sign in with Google",
                  onPressed: () {},
                  icon: Image.asset(
                    'lib/assets/G_logo.png',
                    height: 20,
                  ),
                ),
                const SizedBox(height: 10),
                // LOGIN github BUTTON
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
