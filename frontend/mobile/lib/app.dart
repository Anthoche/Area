import 'package:flutter/material.dart';
import 'screens/login_page.dart';

class MyApp extends StatelessWidget {
  const MyApp({super.key});
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF7209B7),
          primary: const Color(0xFF7209B7),
          secondary: const Color(0xFF4CC9F0),
          tertiary: const Color(0xFFF72585),
          brightness: Brightness.light,
        ),
        scaffoldBackgroundColor: const Color(0xFFF8F9FA), // Fond gris très clair neutre
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF3A0CA3),
            foregroundColor: Colors.white,
          ),
        ),
        useMaterial3: true,
      ),
      home: const LoginPage(),
    );
  }
}
