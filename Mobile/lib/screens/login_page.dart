import 'package:flutter/material.dart';
import 'home_page.dart';

// widgets
import '../widgets/app_text_field.dart';
import '../widgets/primary_button.dart';
import '../widgets/login_container.dart';

class LoginPage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: LoginContainer(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text(
                "KIKONECT 🔌",
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 50),
              // EMAIL FIELD
              const AppTextField(label: "Email"),
              const SizedBox(height: 30),
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
            ],
          ),
        ),
      ),
    );
  }
}
