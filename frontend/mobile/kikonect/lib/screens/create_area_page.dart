import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../utils/ui_feedback.dart';
import '../widgets/logic_block_card.dart';
import 'service_selection_page.dart';

/// Displays the workflow creation screen.
class CreateAreaPage extends StatefulWidget {
  const CreateAreaPage({super.key});

  @override
  State<CreateAreaPage> createState() => _CreateAreaPageState();
}

class _CreateAreaPageState extends State<CreateAreaPage> {
  // Trigger configuration.
  Map<String, dynamic>? triggerData;
  Map<String, dynamic> triggerFieldValues = {};

  // Action blocks (initially one, but extensible).
  List<Map<String, dynamic>?> actionsData = [null];

  bool _isSubmitting = false;
  final TextEditingController _nameController = TextEditingController();

  Future<void> _submitArea() async {
    if (triggerData == null || actionsData.any((a) => a == null)) return;

    final name = _nameController.text.trim();
    if (name.isEmpty) {
      if (mounted) {
        showAppSnackBar(
          context,
          "Please enter a name for this Konect.",
          isError: true,
        );
      }
      return;
    }

    setState(() => _isSubmitting = true);

    try {
      final apiService = ApiService();

      final selectedAction =
          actionsData.firstWhere((a) => a != null, orElse: () => null);
      if (selectedAction == null) return;
      final actionUrl = _buildActionUrl(selectedAction);
      if (actionUrl.isEmpty) {
        if (mounted) {
          showAppSnackBar(
            context,
            "Missing action URL. Please fill required fields.",
            isError: true,
          );
        }
        return;
      }

      final triggerType = triggerData!['id']?.toString() ?? '';
      final actionFields =
          (selectedAction['fields'] as Map<String, dynamic>?) ?? {};
      final triggerConfig = Map<String, dynamic>.from(triggerFieldValues);
      if (triggerType == 'interval') {
        triggerConfig['payload'] = actionFields;
      }
      triggerConfig['payload_template'] = actionFields;

      // Backend expects workflow payload (one trigger + one action_url).
      final payload = {
        "name": name,
        "trigger_type": triggerType,
        "action_url": actionUrl,
        "trigger_config": triggerConfig,
      };

      await apiService.createArea(payload);

      if (mounted) {
        showAppSnackBar(context, "Area Created!");
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        showAppSnackBar(
          context,
          "Error: $e",
          isError: true,
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  String _buildActionUrl(Map<String, dynamic> action) {
    final rawUrl = (action['action_url'] ?? '').toString().trim();
    if (rawUrl.isNotEmpty) {
      if (rawUrl.startsWith('http')) return rawUrl;
      if (rawUrl.startsWith('/')) return '${ApiService.baseUrl}$rawUrl';
      return rawUrl;
    }

    final fields = action['fields'];
    if (fields is Map) {
      final webhookUrl = fields['webhook_url'] ?? fields['url'];
      if (webhookUrl != null && webhookUrl.toString().trim().isNotEmpty) {
        return webhookUrl.toString().trim();
      }
    }

    return '';
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: Icon(Icons.close, color: colorScheme.onSurface, size: 30),
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          "Create",
          style: TextStyle(
            color: colorScheme.onSurface,
            fontWeight: FontWeight.bold,
            fontSize: 24,
          ),
        ),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 10),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextField(
                controller: _nameController,
                decoration: InputDecoration(
                  hintText: "Konect name",
                  hintStyle: TextStyle(color: colorScheme.onSurfaceVariant),
                  filled: true,
                  fillColor: theme.inputDecorationTheme.fillColor ??
                      colorScheme.surfaceVariant,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    vertical: 12,
                    horizontal: 12,
                  ),
                ),
              ),
              const SizedBox(height: 16),
              // IF THIS (trigger).
              LogicBlockCard(
                typeLabel: "IF THIS",
                placeholder: "Add New Service",
                data: triggerData,
                onTap: () async {
                  final result = await Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) =>
                          const ServiceSelectionPage(isTrigger: true),
                    ),
                  );
                  if (!mounted || result == null) return;
                  if (result is Map) {
                    final typedResult = Map<String, dynamic>.from(result);
                    final fields = typedResult['fields'];
                    setState(() {
                      triggerData = typedResult;
                      triggerFieldValues = fields is Map
                          ? Map<String, dynamic>.from(fields)
                          : {};
                    });
                  }
                },
              ),

              const SizedBox(height: 10),

              // Down arrow.
              Center(
                child: Icon(
                  Icons.arrow_downward_rounded,
                  size: 32,
                  color: colorScheme.onSurfaceVariant,
                ),
              ),

              const SizedBox(height: 10),

              // THEN THAT (actions).
              ListView.separated(
                physics: const NeverScrollableScrollPhysics(),
                shrinkWrap: true,
                itemCount: actionsData.length,
                separatorBuilder: (context, index) => Column(
                  children: [
                    const SizedBox(height: 10),
                    Icon(Icons.add, color: colorScheme.onSurfaceVariant),
                    const SizedBox(height: 10),
                  ],
                ),
                itemBuilder: (context, index) {
                  return LogicBlockCard(
                    typeLabel: index == 0 ? "THEN THAT" : "AND THAT",
                    placeholder: "Add New Service",
                    data: actionsData[index],
                    // Enable deletion only for secondary blocks.
                    onDelete: index > 0
                        ? () {
                            setState(() {
                              actionsData.removeAt(index);
                            });
                          }
                        : null,
                    onTap: () async {
                      final result = await Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) =>
                              const ServiceSelectionPage(isTrigger: false),
                        ),
                      );
                      if (!mounted || result == null) return;
                      if (result is Map) {
                        setState(() {
                          actionsData[index] =
                              Map<String, dynamic>.from(result);
                        });
                      }
                    },
                  );
                },
              ),

              const SizedBox(height: 40),

              // Connect button.
              if (triggerData != null && actionsData.any((a) => a != null))
                SizedBox(
                  height: 56,
                  child: ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: colorScheme.primary,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(30),
                      ),
                      elevation: 5,
                    ),
                    onPressed: _isSubmitting ? null : _submitArea,
                    child: _isSubmitting
                        ? CircularProgressIndicator(
                            color: colorScheme.onPrimary,
                          )
                        : Text(
                            "Connect",
                            style: TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.bold,
                              color: colorScheme.onPrimary,
                            ),
                          ),
                  ),
                ),
              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }
}
