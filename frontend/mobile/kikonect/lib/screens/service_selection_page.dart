import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../services/api_service.dart';
import '../widgets/service_selection_card.dart';

/// Displays a list of services to select a trigger or action.
class ServiceSelectionPage extends StatelessWidget {
  /// Whether the selection is for a trigger or an action.
  final bool isTrigger;

  const ServiceSelectionPage({super.key, required this.isTrigger});

  static const Map<String, Color> _serviceColors = {
    'core': Color(0xFF5F6368),
    'discord': Color(0xFF5865F2),
    'google': Color(0xFF4285F4),
    'github': Color(0xFF24292E),
    'slack': Color(0xFF4A154B),
    'notion': Color(0xFF111111),
    'weather': Color(0xFF29B6F6),
  };

  static const Map<String, String> _serviceIcons = {
    'core': 'lib/assets/kikonect_icon.png',
    'discord': 'lib/assets/Discord_logo.png',
    'google': 'lib/assets/G_logo.png',
    'github': 'lib/assets/github_logo.png',
    'slack': 'lib/assets/Slack_logo.png',
    'notion': 'lib/assets/Notion_logo.png',
    'weather': 'lib/assets/Weather_logo.png',
    'steam': 'lib/assets/Steam_logo.png',
    'crypto': 'lib/assets/Crypto_logo.png',
    'nasa': 'lib/assets/Nasa_logo.png',
    'air_quality': 'lib/assets/Air-quality_logo.png',
    'trello': 'lib/assets/Trello_logo.png',
    'reddit': 'lib/assets/Reddit_logo.png',
    'youtube': 'lib/assets/Youtube_logo.png',
  };

  /// Parses a hex color string.
  Color? _parseColor(String? hexColor) {
    if (hexColor == null || hexColor.isEmpty) return null;
    try {
      final hexCode = hexColor.replaceAll('#', '');
      return Color(int.parse('FF$hexCode', radix: 16));
    } catch (e) {
      return null;
    }
  }

  String _serviceId(Map<String, dynamic> serviceData) {
    final id = serviceData['id'] ?? serviceData['service_id'] ?? '';
    return id.toString().toLowerCase();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final apiService = ApiService();

    return Scaffold(
      appBar: AppBar(
        backgroundColor: colorScheme.surface,
        elevation: 0,
        leading: IconButton(
          icon: Icon(Icons.close, color: colorScheme.onSurface),
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          isTrigger ? "Select Trigger" : "Select Action",
          style: TextStyle(
            color: colorScheme.onSurface,
            fontWeight: FontWeight.bold,
          ),
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

          // Filter services that expose at least one trigger or action.
          final services = snapshot.data!.where((s) {
            if (s is! Map<String, dynamic>) return false;
            if (!isTrigger && s['enabled'] == false) return false;
            if (isTrigger) {
              return s['triggers'] != null &&
                  (s['triggers'] as List).isNotEmpty;
            } else {
              return s['reactions'] != null &&
                  (s['reactions'] as List).isNotEmpty;
            }
          }).toList();

          if (services.isEmpty) {
            return Center(
              child: Text(isTrigger
                  ? "No triggers available"
                  : "No reactions available"),
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
              final serviceData =
                  services[index] as Map<String, dynamic>;
              final serviceId = _serviceId(serviceData);
              final color = _parseColor(serviceData['color']?.toString()) ??
                  _serviceColors[serviceId] ??
                  Colors.grey;
              final serviceUI = {
                "name": serviceData['name'] ?? serviceId,
                "icon": serviceData['icon'] ?? _serviceIcons[serviceId],
                "color": color,
                "raw": serviceData,
              };

              return ServiceSelectionCard(
                service: serviceUI,
                onTap: () async {
                  final selection = await _showActionsModal(context, serviceUI);
                  if (selection != null) {
                    Navigator.pop(context, selection);
                  }
                },
              );
            },
          );
        },
      ),
    );
  }

  Future<Map<String, dynamic>?> _showActionsModal(
    BuildContext context,
    Map<String, dynamic> service,
  ) async {
    final rawData = service['raw'] as Map<String, dynamic>;
    final List items = isTrigger
        ? (rawData['triggers'] as List? ?? [])
        : (rawData['reactions'] as List? ?? []);

    return showModalBottomSheet<Map<String, dynamic>>(
      context: context,
      // Allow the sheet to take more height.
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        final colorScheme = Theme.of(context).colorScheme;
        return Container(
          // 70% of the screen height.
          height: MediaQuery.of(context).size.height * 0.7,
          decoration: BoxDecoration(
            color: colorScheme.surface,
            borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
          ),
          child: Column(
            children: [
              // Drag handle.
              const SizedBox(height: 12),
              Container(
                width: 50,
                height: 5,
                decoration: BoxDecoration(
                  color: colorScheme.outlineVariant,
                  borderRadius: BorderRadius.circular(10),
                ),
              ),
              const SizedBox(height: 20),

              Text(
                "${service['name']} ${isTrigger ? 'Triggers' : 'Actions'}",
                style: TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.bold,
                  color: colorScheme.onSurface,
                ),
              ),
              const SizedBox(height: 20),

              // Action list.
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  itemCount: items.length,
                  itemBuilder: (context, index) {
                    final item = items[index];
                    return InkWell(
                      onTap: () async {
                        final fields = (item['fields'] as List? ?? []);
                        if (fields.isNotEmpty) {
                          final selection = await _showFieldsModal(
                            context,
                            service,
                            item,
                            fields,
                          );
                          if (selection != null) {
                            Navigator.pop(context, selection);
                          }
                          return;
                        }
                        Navigator.pop(context, <String, dynamic>{
                          "service_id": rawData['id'],
                          "service": service['name'],
                          "id": item['id'],
                          "name": item['name'],
                          "action": item['name'],
                          "action_url": item['action_url'],
                          "fields": <String, dynamic>{},
                          "color": service['color'],
                          "icon": service['icon'],
                        });
                      },
                      child: Container(
                          margin: const EdgeInsets.only(bottom: 16),
                          padding: const EdgeInsets.all(20),
                          decoration: BoxDecoration(
                            color: colorScheme.surfaceVariant,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: colorScheme.outlineVariant),
                          ),
                          child: Row(
                            children: [
                              Icon(Icons.flash_on, color: service['color']),
                              const SizedBox(width: 16),
                              Expanded(
                                child: Text(
                                  item['name'] ?? "Unknown",
                                  maxLines: 2,
                                  overflow: TextOverflow.ellipsis,
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                    color: colorScheme.onSurface,
                                  ),
                                ),
                              ),
                              const SizedBox(width: 8),
                              Icon(
                                Icons.arrow_forward_ios,
                                size: 16,
                                color: colorScheme.onSurfaceVariant,
                              ),
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

  Future<Map<String, dynamic>?> _showFieldsModal(
    BuildContext context,
    Map<String, dynamic> service,
    Map<String, dynamic> item,
    List fields,
  ) async {
    final rawData = service['raw'] as Map<String, dynamic>;
    final values = <String, dynamic>{};
    final storage = const FlutterSecureStorage();
    var tokenPrefilled = false;
    for (final field in fields) {
      final key = field['key']?.toString() ?? '';
      if (key.isEmpty) continue;
      if (field.containsKey('example')) {
        values[key] = field['example'];
      } else if ((field['type'] ?? '') == 'number') {
        values[key] = 0;
      } else {
        values[key] = '';
      }
    }

    return showModalBottomSheet<Map<String, dynamic>>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setSheetState) {
            final theme = Theme.of(context);
            final colorScheme = theme.colorScheme;
            final viewInsets = MediaQuery.of(context).viewInsets;
            final hasMissingRequired =
                _hasMissingRequiredFields(fields, values);
            if (!tokenPrefilled) {
              tokenPrefilled = true;
              final serviceId = rawData['id']?.toString().toLowerCase() ?? '';
              final itemId = item['id']?.toString().toLowerCase() ?? '';
              final hint = "$serviceId $itemId";
              if (values.containsKey('token_id')) {
                final tokenKey = hint.contains('github')
                    ? 'github_token_id'
                    : (hint.contains('google') || hint.contains('gmail'))
                        ? 'google_token_id'
                        : '';
                if (tokenKey.isNotEmpty) {
                  storage.read(key: tokenKey).then((tokenId) {
                    if (tokenId != null && tokenId.isNotEmpty) {
                      setSheetState(() {
                        values['token_id'] = int.tryParse(tokenId) ?? tokenId;
                      });
                    }
                  });
                }
              }
            }
            return AnimatedPadding(
              duration: const Duration(milliseconds: 200),
              curve: Curves.easeOut,
              padding: EdgeInsets.only(bottom: viewInsets.bottom),
              child: Container(
                height: MediaQuery.of(context).size.height * 0.7,
                decoration: BoxDecoration(
                  color: colorScheme.surface,
                  borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      const SizedBox(height: 12),
                      Center(
                        child: Container(
                          width: 50,
                          height: 5,
                          decoration: BoxDecoration(
                            color: colorScheme.outlineVariant,
                            borderRadius: BorderRadius.circular(10),
                          ),
                        ),
                      ),
                      const SizedBox(height: 20),
                      Text(
                        item['name']?.toString() ?? 'Parameters',
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                          color: colorScheme.onSurface,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 16),
                      if (hasMissingRequired)
                        Padding(
                          padding: EdgeInsets.only(bottom: 10),
                          child: Text(
                            "Please fill required fields.",
                            style: TextStyle(color: colorScheme.error),
                            textAlign: TextAlign.center,
                          ),
                        ),
                      Expanded(
                        child: ListView(
                          keyboardDismissBehavior:
                              ScrollViewKeyboardDismissBehavior.manual,
                          padding: EdgeInsets.only(
                            bottom: viewInsets.bottom > 0 ? 12 : 0,
                          ),
                          children: fields.map<Widget>((field) {
                            final key = field['key']?.toString() ?? '';
                            final type =
                                field['type']?.toString() ?? 'string';
                            final requiredField = field['required'] == true;
                            final description =
                                field['description']?.toString() ?? '';
                            final example = field['example']?.toString() ?? '';
                            final currentValue = values[key];

                            if (key == 'token_id') {
                              final tokenOk = _isTokenIdValid(values[key]);
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 12),
                                child: Text(
                                  tokenOk
                                      ? "Token linked."
                                      : "Missing token id. Please login to this service.",
                                  style: TextStyle(
                                    color: tokenOk
                                        ? colorScheme.secondary
                                        : colorScheme.error,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              );
                            }

                            return Padding(
                              padding: const EdgeInsets.only(bottom: 12),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    requiredField ? "$key *" : key,
                                    style: TextStyle(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w600,
                                      color: colorScheme.onSurface,
                                    ),
                                  ),
                                  const SizedBox(height: 6),
                                  TextFormField(
                                    keyboardType: type == 'number'
                                        ? TextInputType.number
                                        : TextInputType.text,
                                    initialValue:
                                        currentValue?.toString() ?? '',
                                    scrollPadding: EdgeInsets.only(
                                      bottom: viewInsets.bottom + 80,
                                    ),
                                    decoration: InputDecoration(
                                      hintText: example.isNotEmpty
                                          ? example
                                          : description,
                                      hintStyle: TextStyle(
                                        color: colorScheme.onSurfaceVariant,
                                      ),
                                      filled: true,
                                      fillColor:
                                          theme.inputDecorationTheme.fillColor ??
                                              colorScheme.surfaceVariant,
                                      border: OutlineInputBorder(
                                        borderRadius: BorderRadius.circular(12),
                                        borderSide: BorderSide.none,
                                      ),
                                      contentPadding:
                                          const EdgeInsets.symmetric(
                                        vertical: 12,
                                        horizontal: 12,
                                      ),
                                    ),
                                    onChanged: (value) {
                                      setSheetState(() {
                                        if (type == 'number') {
                                          values[key] =
                                              num.tryParse(value) ?? 0;
                                        } else if (type.startsWith('array')) {
                                          values[key] = value
                                              .split(',')
                                              .map((v) => v.trim())
                                              .where((v) => v.isNotEmpty)
                                              .toList();
                                        } else {
                                          values[key] = value;
                                        }
                                      });
                                    },
                                  ),
                                ],
                              ),
                            );
                          }).toList(),
                        ),
                      ),
                      const SizedBox(height: 10),
                      SizedBox(
                        height: 48,
                        child: ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: colorScheme.primary,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(30),
                            ),
                          ),
                        onPressed: hasMissingRequired
                            ? null
                            : () {
                                Navigator.pop(context, <String, dynamic>{
                                  "service_id": rawData['id'],
                                  "service": service['name'],
                                  "id": item['id'],
                                  "name": item['name'],
                                  "action": item['name'],
                                  "action_url": item['action_url'],
                                  "fields": values,
                                  "color": service['color'],
                                  "icon": service['icon'],
                                });
                              },
                          child: const Text(
                            "Validate",
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        );
      },
    );
  }

  bool _hasMissingRequiredFields(List fields, Map<String, dynamic> values) {
    for (final field in fields) {
      if (field is! Map) continue;
      if (_isMissingRequiredField(field, values)) {
        return true;
      }
    }
    return false;
  }

  bool _isMissingRequiredField(Map field, Map<String, dynamic> values) {
    if (field['required'] != true) return false;
    final key = field['key']?.toString() ?? '';
    if (key.isEmpty) return true;
    final type = field['type']?.toString() ?? 'string';
    final value = values[key];

    if (key == 'token_id') {
      return !_isTokenIdValid(value);
    }
    if (type == 'number') {
      final numVal =
          value is num ? value : num.tryParse(value?.toString() ?? '');
      return numVal == null || numVal == 0;
    }
    if (type.startsWith('array')) {
      if (value is List) return value.isEmpty;
      if (value is String) return value.trim().isEmpty;
      return true;
    }
    if (type == 'object') {
      if (value is Map) return value.isEmpty;
      if (value is String) return value.trim().isEmpty;
      return true;
    }
    return value == null || value.toString().trim().isEmpty;
  }

  bool _isTokenIdValid(dynamic value) {
    if (value == null) return false;
    if (value is num) return value > 0;
    final parsed = int.tryParse(value.toString());
    return parsed != null && parsed > 0;
  }
}
