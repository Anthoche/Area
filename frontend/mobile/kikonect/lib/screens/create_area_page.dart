import 'package:flutter/material.dart';
import 'service_selection_page.dart';
import '../widgets/logic_block_card.dart';
import '../services/api_service.dart';

class CreateAreaPage extends StatefulWidget {
  const CreateAreaPage({super.key});

  @override
  State<CreateAreaPage> createState() => _CreateAreaPageState();
}

class _CreateAreaPageState extends State<CreateAreaPage> {
  // Trigger
  Map<String, dynamic>? triggerData;
  Map<String, dynamic> triggerFieldValues = {};

  // Actions (initialement une seule, mais extensible)
  List<Map<String, dynamic>?> actionsData = [null];

  bool _isSubmitting = false;
  final TextEditingController _nameController = TextEditingController();

  Future<void> _submitArea() async {
    if (triggerData == null || actionsData.any((a) => a == null)) return;

    final name = _nameController.text.trim();
    if (name.isEmpty) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text("Please enter a name for this Konect."),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    setState(() => _isSubmitting = true);

    try {
      final apiService = ApiService();

      final selectedAction = actionsData.firstWhere((a) => a != null, orElse: () => null);
      if (selectedAction == null) return;
      final actionUrl = _buildActionUrl(selectedAction);
      if (actionUrl.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text("Missing action URL. Please fill required fields."),
              backgroundColor: Colors.red,
            ),
          );
        }
        return;
      }

      final triggerType = triggerData!['id']?.toString() ?? '';
      final actionFields = (selectedAction['fields'] as Map<String, dynamic>?) ?? {};
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
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("Area Created!")));
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text("Error: $e"), backgroundColor: Colors.red));
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
    return Scaffold(
      backgroundColor: const Color(0xFFF8F9FA),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.close, color: Colors.black, size: 30),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          "Create",
          style: TextStyle(
            color: Colors.black,
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
              // --- IF THIS (Trigger) ---
              LogicBlockCard(
                typeLabel: "IF THIS",
                placeholder: "Add New Service",
                data: triggerData,
                onTap: () async {
                  final result = await Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ServiceSelectionPage(isTrigger: true),
                    ),
                  );
                  if (result != null) {
                    setState(() {
                      triggerData = result;
                      triggerFieldValues =
                          (result['fields'] as Map<String, dynamic>?) ?? {};
                    });
                  }
                },
              ),

              const SizedBox(height: 10),
              
              // Flèche vers le bas
              const Center(
                child: Icon(Icons.arrow_downward_rounded, size: 32, color: Colors.grey),
              ),
              
              const SizedBox(height: 10),

              TextField(
                controller: _nameController,
                decoration: InputDecoration(
                  hintText: "Konect name",
                  hintStyle: TextStyle(color: Colors.grey[500]),
                  filled: true,
                  fillColor: const Color(0xFFF3F6F8),
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
              
              // --- THEN THAT (Actions) ---
              ListView.separated(
                physics: const NeverScrollableScrollPhysics(),
                shrinkWrap: true,
                itemCount: actionsData.length,
                separatorBuilder: (context, index) => const Column(
                  children: [
                    SizedBox(height: 10),
                    Icon(Icons.add, color: Colors.grey),
                    SizedBox(height: 10),
                  ],
                ),
                itemBuilder: (context, index) {
                  return LogicBlockCard(
                    typeLabel: index == 0 ? "THEN THAT" : "AND THAT",
                    placeholder: "Add New Service",
                    data: actionsData[index],
                    // On active la suppression seulement si ce n'est pas le premier bloc
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
                          builder: (context) => const ServiceSelectionPage(isTrigger: false),
                        ),
                      );
                      if (result != null) {
                        setState(() {
                          actionsData[index] = result;
                        });
                      }
                    },
                  );
                },
              ),

              const SizedBox(height: 40),

              // --- BOUTON CONNECT ---
              if (triggerData != null && actionsData.any((a) => a != null))
                SizedBox(
                  height: 56,
                  child: ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.black,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(30),
                      ),
                      elevation: 5,
                    ),
                    onPressed: _isSubmitting ? null : _submitArea,
                    child: _isSubmitting 
                      ? const CircularProgressIndicator(color: Colors.white)
                      : const Text(
                      "Connect",
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
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
