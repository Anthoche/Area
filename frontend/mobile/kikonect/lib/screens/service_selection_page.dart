import 'package:flutter/material.dart';
import '../widgets/service_selection_card.dart';
import '../services/api_service.dart';

class ServiceSelectionPage extends StatelessWidget {
  final bool isTrigger; // Pour savoir si on cherche un Trigger ou une Action

  const ServiceSelectionPage({super.key, required this.isTrigger});

  // Helper to parse hex color
  Color _parseColor(String? hexColor) {
    if (hexColor == null || hexColor.isEmpty) return Colors.grey;
    try {
      final hexCode = hexColor.replaceAll('#', '');
      return Color(int.parse('FF$hexCode', radix: 16));
    } catch (e) {
      return Colors.grey;
    }
  }

  @override
  Widget build(BuildContext context) {
    final apiService = ApiService();

    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.close, color: Colors.black),
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          isTrigger ? "Select Trigger" : "Select Action",
          style: const TextStyle(color: Colors.black, fontWeight: FontWeight.bold),
        ),
      ),
      body: FutureBuilder<List<dynamic>>(
        future: apiService.getServices(),
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          } else if (snapshot.hasError) {
            return Center(child: Text("Error: ${snapshot.error}"));
          } else if (!snapshot.hasData || snapshot.data!.isEmpty) {
            return const Center(child: Text("No services available"));
          }

          // Filter services that have at least one trigger (if isTrigger) or action (if !isTrigger)
          final services = snapshot.data!.where((s) {
            if (isTrigger) {
              return s['triggers'] != null && (s['triggers'] as List).isNotEmpty;
            } else {
              return s['reactions'] != null && (s['reactions'] as List).isNotEmpty;
            }
          }).toList();

          if (services.isEmpty) {
            return Center(
              child: Text(isTrigger ? "No triggers available" : "No reactions available"),
            );
          }

          return GridView.builder(
            padding: const EdgeInsets.all(16),
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              crossAxisSpacing: 16,
              mainAxisSpacing: 16,
              childAspectRatio: 1.1,
            ),
            itemCount: services.length,
            itemBuilder: (context, index) {
              final serviceData = services[index];
              final serviceUI = {
                "name": serviceData['name'],
                "icon": serviceData['icon'],
                "color": _parseColor(serviceData['color']),
                "raw": serviceData,
              };

              return ServiceSelectionCard(
                service: serviceUI,
                onTap: () => _showActionsModal(context, serviceUI),
              );
            },
          );
        },
      ),
    );
  }

  void _showActionsModal(BuildContext context, Map<String, dynamic> service) {
    final rawData = service['raw'] as Map<String, dynamic>;
    final List items = isTrigger 
        ? (rawData['triggers'] as List? ?? []) 
        : (rawData['reactions'] as List? ?? []);

    showModalBottomSheet(
      context: context,
      isScrollControlled: true, // Permet de prendre plus de hauteur
      backgroundColor: Colors.transparent,
      builder: (context) {
        return Container(
          height: MediaQuery.of(context).size.height * 0.7, // 70% de l'écran
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
          ),
          child: Column(
            children: [
              // Barre de "poignée"
              const SizedBox(height: 12),
              Container(
                width: 50,
                height: 5,
                decoration: BoxDecoration(
                  color: Colors.grey[300],
                  borderRadius: BorderRadius.circular(10),
                ),
              ),
              const SizedBox(height: 20),
              
              Text(
                "${service['name']} ${isTrigger ? 'Triggers' : 'Actions'}",
                style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 20),

              // Liste des actions (Rectangles)
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  itemCount: items.length,
                  itemBuilder: (context, index) {
                    final item = items[index];
                    return InkWell(
                      onTap: () {
                        // On ferme la modal
                        Navigator.pop(context);
                        // On ferme la page de sélection et on renvoie les données
                        Navigator.pop(context, {
                          "service_id": rawData['id'],
                          "service": service['name'],
                          "id": item['id'],
                          "name": item['name'],
                          "color": service['color'],
                          "icon": service['icon'],
                        });
                      },
                      child: Container(
                        margin: const EdgeInsets.only(bottom: 16),
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          color: Colors.grey[100],
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: Colors.grey[300]!),
                        ),
                        child: Row(
                          children: [
                            Icon(Icons.flash_on, color: service['color']),
                            const SizedBox(width: 16),
                            Text(
                              item['name'] ?? "Unknown",
                              style: const TextStyle(
                                fontSize: 16, 
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const Spacer(),
                            const Icon(Icons.arrow_forward_ios, size: 16, color: Colors.grey),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
