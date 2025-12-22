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

  // Actions (initialement une seule, mais extensible)
  List<Map<String, dynamic>?> actionsData = [null];

  bool _isSubmitting = false;

  Future<void> _submitArea() async {
    if (triggerData == null || actionsData.any((a) => a == null)) return;

    setState(() => _isSubmitting = true);

    try {
      final apiService = ApiService();
      
      // Construct payload for backend
      final payload = {
        "name": "New Area", // You might want to add a text field for this
        "trigger": {
          "service_id": triggerData!['service_id'],
          "trigger_id": triggerData!['id'],
          "fields": triggerData!['fields'] ?? {}
        },
        "actions": actionsData.where((e) => e != null).map((action) => {
          "service_id": action!['service_id'],
          "action_id": action!['id'],
          "fields": action!['fields'] ?? {}
        }).toList()
      };

      await apiService.createArea(payload);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text("Area Created!")));
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text("Error: $e"), backgroundColor: Colors.red));
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
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

              const SizedBox(height: 20),

              // --- ADD NEW CONDITION (Bouton +) ---
              Center(
                child: InkWell(
                  onTap: () {
                    setState(() {
                      actionsData.add(null);
                    });
                  },
                  borderRadius: BorderRadius.circular(30),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                    decoration: BoxDecoration(
                      color: Colors.grey[200],
                      borderRadius: BorderRadius.circular(30),
                      border: Border.all(color: Colors.grey[400]!),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.add_circle_outline, color: Colors.black54),
                        SizedBox(width: 8),
                        Text(
                          "Add New Condition",
                          style: TextStyle(
                            color: Colors.black54,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
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
