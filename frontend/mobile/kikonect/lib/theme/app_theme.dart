import 'package:flutter/material.dart';

/// Centralized light/dark themes for the app.
class AppTheme {
  static const Color _accentColor = Color(0xFF48179C);
  static const Color _secondaryColor = Color(0xFF2563EB);
  static const Color _tertiaryColor = Color(0xFF6A5CFF);

  static const Color _lightBackground = Color(0xFFFFFFFF);
  static const Color _lightSurface = Color(0xFFFFFFFF);
  static const Color _lightSurfaceVariant = Color(0xFFF7FAFC);
  static const Color _lightOutline = Color(0xFFEBE6E7);
  static const Color _lightOnSurface = Color(0xFF111827);
  static const Color _lightOnSurfaceVariant = Color(0xFF4A5565);

  static const Color _darkBackground = Color(0xFF101828);
  static const Color _darkSurface = Color(0xFF18233A);
  static const Color _darkSurfaceVariant = Color(0xFF1F2A44);
  static const Color _darkOutline = Color(0xFF575A6E);
  static const Color _darkOnSurface = Color(0xFFFFFFFF);
  static const Color _darkOnSurfaceVariant = Color(0xFFA9B4C5);

  static ThemeData light() {
    final baseScheme = ColorScheme.fromSeed(
      seedColor: _accentColor,
      brightness: Brightness.light,
    );

    final colorScheme = baseScheme.copyWith(
      primary: _accentColor,
      onPrimary: Colors.white,
      primaryContainer: const Color(0xFFD9C7FF),
      onPrimaryContainer: _accentColor,
      secondary: _secondaryColor,
      onSecondary: Colors.white,
      secondaryContainer: const Color(0xFFEFF6FF),
      onSecondaryContainer: _secondaryColor,
      tertiary: _tertiaryColor,
      onTertiary: Colors.white,
      background: _lightBackground,
      onBackground: _lightOnSurface,
      surface: _lightSurface,
      onSurface: _lightOnSurface,
      surfaceVariant: _lightSurfaceVariant,
      onSurfaceVariant: _lightOnSurfaceVariant,
      outline: _lightOutline,
      outlineVariant: _lightOutline,
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: colorScheme.background,
      appBarTheme: AppBarTheme(
        backgroundColor: colorScheme.surface,
        foregroundColor: colorScheme.onSurface,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
      ),
      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: colorScheme.surface,
        modalBackgroundColor: colorScheme.surface,
        surfaceTintColor: Colors.transparent,
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: colorScheme.surface,
        hintStyle: TextStyle(color: colorScheme.onSurfaceVariant),
        labelStyle: TextStyle(color: colorScheme.onSurfaceVariant),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: colorScheme.outlineVariant),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: colorScheme.primary, width: 1.5),
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: colorScheme.primary,
          foregroundColor: colorScheme.onPrimary,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: colorScheme.onSurface,
          side: BorderSide(color: colorScheme.outlineVariant),
        ),
      ),
      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: colorScheme.primary,
        foregroundColor: colorScheme.onPrimary,
      ),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: colorScheme.inverseSurface,
        contentTextStyle: TextStyle(color: colorScheme.onInverseSurface),
      ),
    );
  }

  static ThemeData dark() {
    final baseScheme = ColorScheme.fromSeed(
      seedColor: _accentColor,
      brightness: Brightness.dark,
    );

    final colorScheme = baseScheme.copyWith(
      primary: _accentColor,
      onPrimary: Colors.white,
      primaryContainer: const Color(0xFF3A1280),
      onPrimaryContainer: Colors.white,
      secondary: _secondaryColor,
      onSecondary: Colors.white,
      secondaryContainer: const Color(0xFF1D4ED8),
      onSecondaryContainer: Colors.white,
      tertiary: _tertiaryColor,
      onTertiary: Colors.white,
      background: _darkBackground,
      onBackground: _darkOnSurface,
      surface: _darkSurface,
      onSurface: _darkOnSurface,
      surfaceVariant: _darkSurfaceVariant,
      onSurfaceVariant: _darkOnSurfaceVariant,
      outline: _darkOutline,
      outlineVariant: _darkOutline,
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: colorScheme.background,
      appBarTheme: AppBarTheme(
        backgroundColor: colorScheme.surface,
        foregroundColor: colorScheme.onSurface,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
      ),
      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: colorScheme.surface,
        modalBackgroundColor: colorScheme.surface,
        surfaceTintColor: Colors.transparent,
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: colorScheme.surfaceVariant,
        hintStyle: TextStyle(color: colorScheme.onSurfaceVariant),
        labelStyle: TextStyle(color: colorScheme.onSurfaceVariant),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: colorScheme.outlineVariant),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: colorScheme.primary, width: 1.5),
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: colorScheme.primary,
          foregroundColor: colorScheme.onPrimary,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: colorScheme.onSurface,
          side: BorderSide(color: colorScheme.outlineVariant),
        ),
      ),
      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: colorScheme.primary,
        foregroundColor: colorScheme.onPrimary,
      ),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: colorScheme.inverseSurface,
        contentTextStyle: TextStyle(color: colorScheme.onInverseSurface),
      ),
    );
  }
}
