import 'package:flutter/material.dart';

/// Lightweight search field used in the home screen header.
class Search_bar extends StatelessWidget {
  final TextEditingController? controller;
  final ValueChanged<String>? onChanged;

  const Search_bar({super.key, this.controller, this.onChanged});

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      onChanged: onChanged,
      decoration: InputDecoration(
        hintText: 'Search in my library',
        hintStyle: TextStyle(color: Colors.grey[400]),
        prefixIcon: Icon(Icons.search, color: Colors.grey[400]),
        filled: true,
        fillColor: const Color(0xFFF3F6F8),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        contentPadding: const EdgeInsets.symmetric(vertical: 12),
      ),
    );
  }
}
