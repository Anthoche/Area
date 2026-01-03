import 'package:flutter/material.dart';

/// Shows a SnackBar with optional error styling.
void showAppSnackBar(
  BuildContext context,
  String message, {
  bool isError = false,
  Color? backgroundColor,
}) {
  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(
      content: Text(message),
      backgroundColor: backgroundColor ?? (isError ? Colors.red : null),
    ),
  );
}

/// Shows a standard error dialog.
Future<void> showErrorDialog(
  BuildContext context,
  String message, {
  String title = 'Error',
  Color? titleColor,
  Color? buttonColor,
  String buttonLabel = 'OK',
}) {
  return showDialog<void>(
    context: context,
    builder: (_) => AlertDialog(
      title: Text(
        title,
        style: TextStyle(
          color: titleColor ?? Colors.red,
          fontWeight: FontWeight.bold,
        ),
      ),
      content: Text(message),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          style: buttonColor != null
              ? TextButton.styleFrom(foregroundColor: buttonColor)
              : null,
          child: Text(buttonLabel),
        ),
      ],
    ),
  );
}
