import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'app.dart';

/// Entry point of the Flutter app. Loads environment variables before bootstrapping
/// the widget tree so API-dependent screens can access configuration safely.
Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await dotenv.load(fileName: ".env");
  } catch (e) {
    debugPrint("failed to load .env : $e");
  }
  runApp(const MyApp());
}
