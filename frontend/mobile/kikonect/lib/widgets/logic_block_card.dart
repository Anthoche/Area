import 'package:flutter/material.dart';

/// Displays a tappable logic block for trigger or action selection.
class LogicBlockCard extends StatelessWidget {
  final String typeLabel;
  final String placeholder;
  final Map<String, dynamic>? data;
  final VoidCallback onTap;
  final VoidCallback? onDelete;

  const LogicBlockCard({
    super.key,
    required this.typeLabel,
    required this.placeholder,
    this.data,
    required this.onTap,
    this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final bool isFilled = data != null;
    final Color bgColor =
        isFilled ? (data!['color'] as Color) : colorScheme.surfaceVariant;
    final Color contentColor =
        isFilled ? Colors.white : colorScheme.onSurfaceVariant;

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 300),
        height: 160,
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: bgColor,
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            BoxShadow(
              color: theme.shadowColor.withOpacity(0.12),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Stack(
          children: [
            // Centered content (text or service).
            Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  if (!isFilled) ...[
                    // Empty state (gray).
                    Text(
                      typeLabel,
                      style: TextStyle(
                        color: colorScheme.onSurface,
                        fontSize: 28,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                      decoration: BoxDecoration(
                        color: colorScheme.primary,
                        borderRadius: BorderRadius.circular(30),
                      ),
                      child: Text(
                        placeholder,
                        style: TextStyle(
                          color: colorScheme.onPrimary,
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                    ),
                  ] else ...[
                    // Filled state (service color).
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        if (data!['icon'] != null)
                          Image.asset(data!['icon'], height: 40, width: 40)
                        else
                          Icon(Icons.apps, size: 40, color: contentColor),
                        const SizedBox(width: 10),
                        Text(
                          data!['service'] ?? "",
                          style: TextStyle(
                            color: contentColor,
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      (data!['action'] ?? data!['name'] ?? "").toString(),
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        color: contentColor.withOpacity(0.9),
                        fontSize: 18,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      "Tap to change",
                      style: TextStyle(
                        color: contentColor.withOpacity(0.6),
                        fontSize: 12,
                      ),
                    ),
                  ],
                ],
              ),
            ),

            // Delete button (top right).
            if (onDelete != null)
              Positioned(
                top: -12,
                right: -12,
                child: IconButton(
                  icon: Icon(
                    Icons.cancel,
                    color: isFilled
                        ? Colors.white.withOpacity(0.8)
                        : colorScheme.onSurfaceVariant,
                    size: 28,
                  ),
                  onPressed: onDelete,
                  tooltip: "Remove condition",
                ),
              ),
          ],
        ),
      ),
    );
  }
}
