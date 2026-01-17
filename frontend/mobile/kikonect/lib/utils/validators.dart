/// Returns true if the email matches a basic pattern.
bool isValidEmail(String email) {
  final trimmed = email.trim();
  if (trimmed.isEmpty) return false;
  final regex = RegExp(r'^[\w\.-]+@[\w\.-]+\.\w+$');
  return regex.hasMatch(trimmed);
}

/// Returns true if the password matches the web password policy.
bool isValidPassword(String password) {
  final regex = RegExp(
    r"""^(?=.*[0-9])(?=.*[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]).{8,}$""",
  );
  return regex.hasMatch(password);
}
