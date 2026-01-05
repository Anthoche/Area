import 'package:flutter/material.dart';

import 'screens/login_page.dart';
import 'theme/app_theme.dart';
import 'theme/theme_controller.dart';

/// Builds the root widget that wires global theming and routes into the app.
class MyApp extends StatelessWidget {
  final ThemeController themeController;

  const MyApp({super.key, required this.themeController});

  @override
  Widget build(BuildContext context) {
    return ThemeScope(
      controller: themeController,
      child: AnimatedBuilder(
        animation: themeController,
        builder: (context, _) {
          return MaterialApp(
            debugShowCheckedModeBanner: false,
            theme: AppTheme.light(),
            darkTheme: AppTheme.dark(),
            themeMode: themeController.mode,
            home: const LoginPage(),
          );
        },
      ),
    );
  }
}
