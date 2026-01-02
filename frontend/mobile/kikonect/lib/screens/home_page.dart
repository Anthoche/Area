import 'dart:convert';
import 'package:flutter/material.dart';
import '../widgets/filter_tag.dart';
import '../widgets/service_card.dart';
import '../widgets/search_bar.dart';
import '../services/api_service.dart';
import 'create_area_page.dart';
import 'profile_page.dart';

/// Home screen showing saved Konects and quick actions.
class Homepage extends StatefulWidget {
  const Homepage({super.key});

  @override
  State<Homepage> createState() => _HomepageState();
}

class _HomepageState extends State<Homepage> {
  final ApiService _apiService = ApiService();
  final TextEditingController _searchController = TextEditingController();
  final Set<String> _activeFilters = {};
  final Set<int> _triggering = {};
  final Set<int> _toggling = {};
  List<dynamic> _workflows = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadWorkflows();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadWorkflows() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final items = await _apiService.getWorkflows();
      if (mounted) {
        setState(() {
          _workflows = items;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
        });
      }
    } finally {
      if (mounted) {
        setState(() {
          _loading = false;
        });
      }
    }
  }

  List<String> get _availableFilters {
    final types = _workflows
        .map((w) => (w['trigger_type'] ?? '').toString())
        .where((t) => t.isNotEmpty)
        .toSet()
        .toList();
    types.sort();
    return types;
  }

  List<dynamic> get _filteredWorkflows {
    final search = _searchController.text.trim().toLowerCase();
    return _workflows.where((w) {
      final name = (w['name'] ?? '').toString().toLowerCase();
      final trigger = (w['trigger_type'] ?? '').toString();
      final matchesSearch = search.isEmpty || name.contains(search);
      final matchesFilter =
          _activeFilters.isEmpty || _activeFilters.contains(trigger);
      return matchesSearch && matchesFilter;
    }).toList();
  }

  Future<void> _triggerManualWorkflow(dynamic item) async {
    final triggerType = (item['trigger_type'] ?? '').toString();
    if (triggerType != 'manual') {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text("Only manual Konects can be triggered here."),
          ),
        );
      }
      return;
    }

    final idValue = item['id'];
    final id = idValue is int ? idValue : int.tryParse(idValue?.toString() ?? '');
    if (id == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text("Invalid Konect id."),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }
    if (_triggering.contains(id)) return;

    setState(() => _triggering.add(id));
    try {
      final payload = _payloadFromWorkflow(item);
      await _apiService.triggerWorkflow(id, payload);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text("Konect triggered!")),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text("Trigger failed: $e"),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _triggering.remove(id));
      }
    }
  }

  void _showWorkflowControls(dynamic item) {
    final triggerType = (item['trigger_type'] ?? '').toString();
    final idValue = item['id'];
    final id = idValue is int ? idValue : int.tryParse(idValue?.toString() ?? '');
    if (id == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text("Invalid Konect id."),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    final initialEnabled = item['enabled'] == true;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        var currentEnabled = initialEnabled;
        var isBusy = false;
        return StatefulBuilder(
          builder: (context, setSheetState) {
            return Container(
              height: MediaQuery.of(context).size.height * 0.45,
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
              ),
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const SizedBox(height: 8),
                  Center(
                    child: Container(
                      width: 50,
                      height: 5,
                      decoration: BoxDecoration(
                        color: Colors.grey[300],
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    (item['name'] ?? 'Konect').toString(),
                    textAlign: TextAlign.center,
                    style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    "Trigger: $triggerType",
                    textAlign: TextAlign.center,
                    style: TextStyle(color: Colors.grey[600]),
                  ),
                  const SizedBox(height: 20),
                  SwitchListTile(
                    title: Text(currentEnabled ? "Enabled" : "Disabled"),
                    value: currentEnabled,
                    onChanged: isBusy
                        ? null
                        : (value) async {
                            final previous = currentEnabled;
                            setSheetState(() {
                              currentEnabled = value;
                              isBusy = true;
                            });
                            setState(() => _toggling.add(id));
                            try {
                              await _apiService.setWorkflowEnabled(id, value);
                              await _loadWorkflows();
                            } catch (e) {
                              setSheetState(() {
                                currentEnabled = previous;
                              });
                              if (mounted) {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  SnackBar(
                                    content: Text("Update failed: $e"),
                                    backgroundColor: Colors.red,
                                  ),
                                );
                              }
                            } finally {
                              if (mounted) {
                                setSheetState(() {
                                  isBusy = false;
                                });
                                setState(() => _toggling.remove(id));
                              }
                            }
                          },
                  ),
                  const SizedBox(height: 6),
                  Text(
                    "This Konect runs automatically when the trigger happens.",
                    textAlign: TextAlign.center,
                    style: TextStyle(color: Colors.grey[600], fontSize: 12),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  Map<String, dynamic> _payloadFromWorkflow(dynamic item) {
    final raw = item['trigger_config'];
    dynamic config = raw;
    if (raw is String && raw.isNotEmpty) {
      try {
        config = jsonDecode(raw);
      } catch (_) {
        config = null;
      }
    }
    if (config is Map<String, dynamic>) {
      final payload = config['payload_template'];
      if (payload is Map<String, dynamic>) {
        return payload;
      }
    }
    return {};
  }

  @override
  Widget build(BuildContext context) {
    final workflows = _filteredWorkflows;

    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        title: const Text(
          'My Konect',
          style: TextStyle(
            color: Colors.black,
            fontWeight: FontWeight.bold,
            fontFamily: 'Serif',
            fontSize: 22,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.person_outline, color: Colors.black),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const ProfilePage()),
              );
            },
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: CustomScrollView(
        slivers: [
          const SliverToBoxAdapter(
            child: Padding(
              padding: EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'My Konects',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w500,
                      color: Colors.black54,
                    ),
                  ),
                  SizedBox(height: 16),
                ],
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Search_bar(
                controller: _searchController,
                onChanged: (_) => setState(() {}),
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
              child: SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: Row(
                  children: [
                    FilterTag(
                      label: 'All',
                      isSelected: _activeFilters.isEmpty,
                      onTap: () {
                        setState(() => _activeFilters.clear());
                      },
                    ),
                    ..._availableFilters.map(
                      (filter) => FilterTag(
                        label: filter,
                        isSelected: _activeFilters.contains(filter),
                        onTap: () {
                          setState(() {
                            if (_activeFilters.contains(filter)) {
                              _activeFilters.remove(filter);
                            } else {
                              _activeFilters.add(filter);
                            }
                          });
                        },
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          if (_loading)
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Center(child: CircularProgressIndicator()),
              ),
            )
          else if (_error != null)
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Center(
                  child: Text(
                    _error!,
                    style: const TextStyle(color: Colors.red),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),
            )
          else if (workflows.isEmpty)
            const SliverToBoxAdapter(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Center(
                  child: Text("No Konect yet. Create the first one!"),
                ),
              ),
            ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            sliver: SliverGrid(
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                mainAxisSpacing: 16,
                crossAxisSpacing: 16,
                childAspectRatio: 1.3,
              ),
              delegate: SliverChildBuilderDelegate(
                (context, index) {
                  final item = workflows[index];
                  final triggerType = (item['trigger_type'] ?? '').toString();
                  final enabled = item['enabled'] == true;
                  final isManual = triggerType == 'manual';
                  return ServiceCard(
                    title: (item['name'] ?? 'Konect #${item['id']}').toString(),
                    color: _getColor(index),
                    subtitle: triggerType.isNotEmpty ? triggerType : null,
                    badgeText: isManual ? "MANUAL" : (enabled ? "ON" : "OFF"),
                    badgeColor: isManual ? Colors.black54 : (enabled ? Colors.green : Colors.red),
                    onTap: () {
                      if (isManual) {
                        _triggerManualWorkflow(item);
                      } else {
                        _showWorkflowControls(item);
                      }
                    },
                  );
                },
                childCount: workflows.length,
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 100)),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        backgroundColor: const Color(0xFF7209B7),
        foregroundColor: Colors.white,
        shape: const CircleBorder(),
        elevation: 4,
        onPressed: () {
          Navigator.push<bool>(
            context,
            MaterialPageRoute(builder: (context) => const CreateAreaPage()),
          ).then((created) {
            if (created == true) {
              _loadWorkflows();
            }
          });
        },
        child: const Icon(Icons.add, size: 30),
      ),
    );
  }

  /// Rotates through a small palette to colorize service cards.
  Color _getColor(int index) {
    final colors = [
      const Color(0xFF00D2FF),
      const Color(0xFFFF4081),
      const Color(0xFFFF4081),
      const Color(0xFF00E676),
      const Color(0xFFD500F9),
    ];
    return colors[index % colors.length];
  }

}
