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

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        decoration: BoxDecoration(
          color: (service['color'] as Color).withOpacity(0.1),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: (service['color'] as Color).withOpacity(0.3),
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
                  : Icon(Icons.apps, color: service['color']),
            ),
            const SizedBox(height: 12),
            Text(
              service['name'],
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: service['color'],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
