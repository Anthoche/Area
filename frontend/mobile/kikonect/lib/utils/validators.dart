/// Returns true if the email matches a basic pattern.
bool isValidEmail(String email) {
  final trimmed = email.trim();
  if (trimmed.isEmpty) return false;
  final regex = RegExp(r'^[\w\.-]+@[\w\.-]+\.\w+$');
  return regex.hasMatch(trimmed);
}
