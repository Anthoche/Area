import 'package:flutter/material.dart';

/// Displays a capsule-styled button used to filter service lists.
class FilterTag extends StatelessWidget {
  final String label;
  final bool isSelected;
  final VoidCallback? onTap;

  const FilterTag({
    super.key,
    required this.label,
    this.isSelected = false,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final selectedColor = colorScheme.secondary;
    final backgroundColor =
        isSelected ? selectedColor : colorScheme.surfaceVariant;
    final foregroundColor =
        isSelected ? colorScheme.onSecondary : colorScheme.onSurface;
    final borderColor =
        isSelected ? selectedColor : colorScheme.outlineVariant;
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: OutlinedButton(
        onPressed: onTap,
        style: OutlinedButton.styleFrom(
          backgroundColor: backgroundColor,
          foregroundColor: foregroundColor,
          side: BorderSide(color: borderColor),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          minimumSize: Size.zero,
          tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          elevation: 0,
        ),
        child: Text(
          label,
          style: const TextStyle(
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }
}
