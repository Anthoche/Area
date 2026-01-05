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
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    return TextField(
      controller: controller,
      obscureText: obscure,
      style: theme.textTheme.bodyMedium?.copyWith(
        color: colorScheme.onSurface,
      ),
      decoration: InputDecoration(
        labelText: label,
        filled: true,
        fillColor:
            theme.inputDecorationTheme.fillColor ?? colorScheme.surfaceVariant,
        labelStyle: TextStyle(color: colorScheme.onSurfaceVariant),
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
