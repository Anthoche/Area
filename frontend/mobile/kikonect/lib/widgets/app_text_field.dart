import 'package:flutter/material.dart';

/// Displays a reusable text field styled for authentication and onboarding forms.
class AppTextField extends StatelessWidget {
  final String label;
  final bool obscure;
  final TextEditingController? controller;

  const AppTextField({
    super.key,
    required this.label,
    this.controller,
    this.obscure = false,
  });

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      obscureText: obscure,
      style: const TextStyle(color: Colors.black),
      decoration: InputDecoration(
        labelText: label,
        filled: true,
        fillColor: Colors.white.withOpacity(0.9),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide.none,
        ),
        isDense: true,
        contentPadding: const EdgeInsets.symmetric(
            vertical: 15, horizontal: 20),
      ),
    );
  }
}
