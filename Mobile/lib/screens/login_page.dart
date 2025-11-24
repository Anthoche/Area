import 'package:flutter/material.dart';
import 'home_page.dart';
import 'register_page.dart';

// widgets
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';

class LoginPage extends StatelessWidget {
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
                // EMAIL FIELD
                const AppTextField(label: "Email"),
                const SizedBox(height: 25),
                // PASSWORD FIELD
                const AppTextField(label: "Password", obscure: true),
                const SizedBox(height: 30),
                // LOGIN BUTTON
                PrimaryButton(
                  text: "Login",
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(builder: (_) => Homepage()),
                    );
                  },
                ),
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
                // Register link
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text("Don't have an account?"),
                    TextButton(
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(builder: (_) => RegisterPage()),
                        );
                      },
                      child: Text("Sign up", style: TextStyle(decoration: TextDecoration.underline)),
                    ),
                  ],
                )
              ],
            ),
          ),
        ),
      ),
    );
  }
}
