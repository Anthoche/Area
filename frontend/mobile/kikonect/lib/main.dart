import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'app.dart';
import 'services/app_config.dart';
import 'theme/theme_controller.dart';

/// Starts the Flutter app after loading environment variables.
Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await dotenv.load(fileName: ".env");
  } catch (e) {
    debugPrint("failed to load .env : $e");
  }
  await AppConfig.load();
  final themeController = ThemeController(const FlutterSecureStorage());
  await themeController.load();
  runApp(MyApp(themeController: themeController));
}
