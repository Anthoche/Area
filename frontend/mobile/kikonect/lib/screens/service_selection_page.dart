import 'package:flutter/material.dart';
import '../widgets/service_selection_card.dart';

class ServiceSelectionPage extends StatelessWidget {
  final bool isTrigger; // Pour savoir si on cherche un Trigger ou une Action

  const ServiceSelectionPage({super.key, required this.isTrigger});

  final List<Map<String, dynamic>> services = const [
    {"name": "Google", "icon": "lib/assets/G_logo.png", "color": Colors.redAccent},
    {"name": "Github", "icon": "lib/assets/github_logo.png", "color": Colors.black},
    {"name": "Discord", "icon": null, "color": Colors.indigo},
    {"name": "Spotify", "icon": null, "color": Colors.green},
    {"name": "Twitch", "icon": null, "color": Colors.purple},
    {"name": "Twitter", "icon": null, "color": Colors.blue},
  ];

  @override
  Widget build(BuildContext context) {
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
      body: GridView.builder(
        padding: const EdgeInsets.all(16),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 2,
          crossAxisSpacing: 16,
          mainAxisSpacing: 16,
          childAspectRatio: 1.1,
        ),
        itemCount: services.length,
        itemBuilder: (context, index) {
          final service = services[index];
          return ServiceSelectionCard(
            service: service,
            onTap: () => _showActionsModal(context, service),
          );
        },
      ),
    );
  }

  void _showActionsModal(BuildContext context, Map<String, dynamic> service) {
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
                "${service['name']} actions",
                style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 20),

              // Liste des actions (Rectangles)
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  itemCount: 10, // Juste pour l'exemple
                  itemBuilder: (context, index) {
                    return InkWell(
                      onTap: () {
                        // On ferme la modal
                        Navigator.pop(context);
                        // On ferme la page de sélection et on renvoie les données
                        Navigator.pop(context, {
                          "service": service['name'],
                          "action": "Test Action ${index + 1}",
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
                              "Test Action ${index + 1}",
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
