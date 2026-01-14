import 'package:flutter/material.dart';

/// Displays a service card for selection in a grid.
class ServiceSelectionCard extends StatelessWidget {
  final Map<String, dynamic> service;
  final VoidCallback onTap;

  const ServiceSelectionCard({
    super.key,
    required this.service,
    required this.onTap,
  });

  Color _displayColor(Color base, Brightness brightness) {
    final isDark = ThemeData.estimateBrightnessForColor(base) == Brightness.dark;
    if (brightness == Brightness.dark && isDark) {
      final hsl = HSLColor.fromColor(base);
      final lightness = hsl.lightness < 0.6 ? 0.6 : hsl.lightness;
      final saturation = hsl.saturation < 0.5 ? 0.5 : hsl.saturation;
      return hsl
          .withLightness(lightness)
          .withSaturation(saturation)
          .toColor();
    }
    return base;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final baseColor = service['color'] as Color? ?? colorScheme.primary;
    final displayColor = _displayColor(baseColor, theme.brightness);
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        decoration: BoxDecoration(
          color: displayColor.withOpacity(0.18),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: displayColor.withOpacity(0.4),
            width: 1,
          ),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Icon or placeholder.
            Container(
              height: 60,
              width: 60,
              decoration: BoxDecoration(
                color: colorScheme.surface,
                shape: BoxShape.circle,
                boxShadow: [
                  BoxShadow(
                    color: theme.shadowColor.withOpacity(0.2),
                    blurRadius: 8,
                  )
                ],
              ),
              padding: const EdgeInsets.all(12),
              child: service['icon'] != null
                  ? Image.asset(
                      service['icon'],
                      errorBuilder: (c, o, s) => const Icon(Icons.apps),
                    )
                  : Icon(Icons.apps, color: displayColor),
            ),
            const SizedBox(height: 12),
            Text(
              service['name'],
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: displayColor,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
