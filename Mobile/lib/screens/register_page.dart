import 'package:flutter/material.dart';
import 'home_page.dart';

// widgets
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';

class RegisterPage extends StatelessWidget {
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
                const AppTextField(label: "First Name"),
                const SizedBox(height: 30),
                const AppTextField(label: "Last Name"),
                const SizedBox(height: 30),
                const AppTextField(label: "Email"),
                const SizedBox(height: 30),
                const AppTextField(label: "Password", obscure: true),
                const SizedBox(height: 30),
                const AppTextField(label: "Confirm Password", obscure: true),
                const SizedBox(height: 30),
                PrimaryButton(text: "Register", onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (_) => Homepage()),
                  );
                })
              ],
            ),
          ),
        ),
      ),
    );
  }
}
