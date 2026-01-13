import 'dart:convert';

import 'package:flutter/material.dart';

import '../services/api_service.dart';
import '../utils/ui_feedback.dart';
import '../widgets/filter_tag.dart';
import '../widgets/search_bar.dart';
import '../widgets/service_card.dart';
import 'create_area_page.dart';
import 'profile_page.dart';

/// Displays the home screen with saved Konects and quick actions.
class Homepage extends StatefulWidget {
  final ApiService? apiService;

  const Homepage({super.key, this.apiService});

  @override
  State<Homepage> createState() => _HomepageState();
}

class _HomepageState extends State<Homepage> {
  late final ApiService _apiService;
  final TextEditingController _searchController = TextEditingController();
  final Set<String> _activeFilters = {};
  final Set<int> _triggering = {};
  List<dynamic> _workflows = [];
  List<dynamic> _services = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _apiService = widget.apiService ?? ApiService();
    _loadWorkflows();
    _loadServices();
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

  Future<void> _loadServices() async {
    try {
      final items = await _apiService.getServices();
      if (mounted) {
        setState(() {
          _services = items;
        });
      }
    } catch (_) {}
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
        showAppSnackBar(
          context,
          "Only manual Konects can be triggered here.",
        );
      }
      return;
    }

    final idValue = item['id'];
    final id =
        idValue is int ? idValue : int.tryParse(idValue?.toString() ?? '');
    if (id == null) {
      if (mounted) {
        showAppSnackBar(
          context,
          "Invalid Konect id.",
          isError: true,
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
        showAppSnackBar(context, "Konect triggered!");
      }
    } catch (e) {
      if (mounted) {
        showAppSnackBar(
          context,
          "Trigger failed: $e",
          isError: true,
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
    final helperText = triggerType == 'manual'
        ? "This Konect runs when you trigger it manually."
        : "This Konect runs automatically when the trigger happens.";
    final idValue = item['id'];
    final id =
        idValue is int ? idValue : int.tryParse(idValue?.toString() ?? '');
    if (id == null) {
      if (mounted) {
        showAppSnackBar(
          context,
          "Invalid Konect id.",
          isError: true,
        );
      }
      return;
    }

    final config = _parseTriggerConfig(item['trigger_config']);
    final actionPayload = _extractActionPayload(config);
    final triggerDetails = _extractTriggerDetails(config);
    final actionUrl = item['action_url']?.toString() ?? '';
    final actionMeta = _findReactionMeta(actionUrl);
    final actionLabel = _actionLabel(actionMeta, actionUrl);
    final triggerMeta = _findTriggerMeta(triggerType);
    final triggerLabel = _triggerLabel(triggerMeta, triggerType);
    final rawFields = actionMeta?['fields'];
    final rawTriggerFields = triggerMeta?['fields'];
    final actionEntries = _formatDetails(
      actionPayload,
      fieldDefs: rawFields is List ? rawFields : const [],
    );
    final triggerEntries = _formatDetails(
      triggerDetails,
      fieldDefs: rawTriggerFields is List ? rawTriggerFields : const [],
    );

    final initialEnabled = item['enabled'] == true;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        final colorScheme = Theme.of(context).colorScheme;
        var currentEnabled = initialEnabled;
        var isBusy = false;
        return StatefulBuilder(
          builder: (context, setSheetState) {
            return Container(
              height: MediaQuery.of(context).size.height * 0.45,
              decoration: BoxDecoration(
                color: colorScheme.surface,
                borderRadius: BorderRadius.vertical(top: Radius.circular(25)),
              ),
              padding: const EdgeInsets.all(20),
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const SizedBox(height: 8),
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
                    const SizedBox(height: 16),
                    Text(
                      (item['name'] ?? 'Konect').toString(),
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                          fontSize: 20, fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      "Trigger: $triggerLabel",
                      textAlign: TextAlign.center,
                      style: TextStyle(color: colorScheme.onSurfaceVariant),
                    ),
                  if (actionLabel.isNotEmpty) ...[
                    const SizedBox(height: 12),
                    Text(
                      "Action: $actionLabel",
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        color: colorScheme.onSurface,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                  const SizedBox(height: 12),
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
                            try {
                              await _apiService.setWorkflowEnabled(id, value);
                              await _loadWorkflows();
                            } catch (e) {
                              setSheetState(() {
                                currentEnabled = previous;
                              });
                              if (mounted) {
                                showAppSnackBar(
                                  context,
                                  "Update failed: $e",
                                  isError: true,
                                );
                              }
                            } finally {
                              if (mounted) {
                                setSheetState(() {
                                  isBusy = false;
                                });
                              }
                            }
                          },
                  ),
                  const SizedBox(height: 6),
                    Text(
                      helperText,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        color: colorScheme.onSurfaceVariant,
                        fontSize: 12,
                    ),
                  ),
                  if (actionEntries.isNotEmpty)
                    _DetailsSection(
                      title: "Action details",
                      entries: actionEntries,
                    ),
                  if (triggerEntries.isNotEmpty)
                    _DetailsSection(
                      title: "Trigger details",
                      entries: triggerEntries,
                    ),
                  if (actionEntries.isEmpty &&
                      triggerEntries.isEmpty &&
                      actionUrl.isNotEmpty)
                    _DetailsSection(
                      title: "Action endpoint",
                      entries: [
                        MapEntry("URL", actionUrl),
                      ],
                    ),
                  const SizedBox(height: 16),
                  OutlinedButton.icon(
                    onPressed: isBusy
                        ? null
                        : () async {
                            final name =
                                (item['name'] ?? 'Konect').toString();
                            final shouldDelete = await showDialog<bool>(
                              context: context,
                              builder: (dialogContext) => AlertDialog(
                                title: const Text("Delete Konect?"),
                                content: Text(
                                  'This will permanently delete "$name".',
                                ),
                                actions: [
                                  TextButton(
                                    onPressed: () => Navigator.pop(
                                      dialogContext,
                                      false,
                                    ),
                                    child: const Text("Cancel"),
                                  ),
                                  TextButton(
                                    onPressed: () =>
                                        Navigator.pop(dialogContext, true),
                                    style: TextButton.styleFrom(
                                      foregroundColor: colorScheme.error,
                                    ),
                                    child: const Text("Delete"),
                                  ),
                                ],
                              ),
                            );
                            if (shouldDelete != true) return;
                            setSheetState(() {
                              isBusy = true;
                            });
                            try {
                              await _apiService.deleteWorkflow(id);
                              if (!mounted) return;
                              Navigator.pop(context);
                              showAppSnackBar(
                                this.context,
                                "Konect deleted.",
                              );
                              await _loadWorkflows();
                            } catch (e) {
                              if (!mounted) return;
                              showAppSnackBar(
                                this.context,
                                "Delete failed: $e",
                                isError: true,
                              );
                              setSheetState(() {
                                isBusy = false;
                              });
                            }
                          },
                    style: OutlinedButton.styleFrom(
                      foregroundColor: colorScheme.error,
                      side: BorderSide(color: colorScheme.error),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    icon: const Icon(Icons.delete_outline),
                    label: const Text("Delete Konect"),
                  ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );
  }

  Map<String, dynamic> _payloadFromWorkflow(dynamic item) {
    final config = _parseTriggerConfig(item['trigger_config']);
    return _extractActionPayload(config);
  }

  Map<String, dynamic> _parseTriggerConfig(dynamic raw) {
    if (raw is Map<String, dynamic>) {
      return Map<String, dynamic>.from(raw);
    }
    if (raw is String && raw.isNotEmpty) {
      try {
        final decoded = jsonDecode(raw);
        if (decoded is Map<String, dynamic>) {
          return Map<String, dynamic>.from(decoded);
        }
      } catch (_) {}
    }
    return {};
  }

  Map<String, dynamic> _extractActionPayload(Map<String, dynamic> config) {
    final payload = config['payload_template'] ?? config['payload'];
    if (payload is Map<String, dynamic>) {
      return Map<String, dynamic>.from(payload);
    }
    return {};
  }

  Map<String, dynamic> _extractTriggerDetails(Map<String, dynamic> config) {
    final details = Map<String, dynamic>.from(config);
    details.remove('payload_template');
    details.remove('payload');
    return details;
  }

  Map<String, dynamic>? _findReactionMeta(String actionUrl) {
    if (_services.isEmpty || actionUrl.isEmpty) return null;
    final path = _actionPath(actionUrl);
    for (final service in _services) {
      if (service is! Map) continue;
      final reactions = service['reactions'];
      if (reactions is! List) continue;
      for (final reaction in reactions) {
        if (reaction is! Map) continue;
        final reactionUrl = reaction['action_url']?.toString() ?? '';
        if (reactionUrl.isEmpty) continue;
        if (_actionUrlMatches(actionUrl, path, reactionUrl)) {
          return {
            'service': service['name'],
            'action': reaction['name'],
            'description': reaction['description'],
            'fields': reaction['fields'],
          };
        }
      }
    }
    return null;
  }

  Map<String, dynamic>? _findTriggerMeta(String triggerType) {
    if (_services.isEmpty || triggerType.isEmpty) return null;
    for (final service in _services) {
      if (service is! Map) continue;
      final triggers = service['triggers'];
      if (triggers is! List) continue;
      for (final trigger in triggers) {
        if (trigger is! Map) continue;
        if (trigger['id']?.toString() == triggerType) {
          return {
            'service': service['name'],
            'trigger': trigger['name'],
            'description': trigger['description'],
            'fields': trigger['fields'],
          };
        }
      }
    }
    return null;
  }

  String _actionPath(String actionUrl) {
    final uri = Uri.tryParse(actionUrl);
    if (uri == null || uri.path.isEmpty) return actionUrl;
    return uri.path;
  }

  bool _actionUrlMatches(String raw, String path, String candidate) {
    if (candidate.isEmpty) return false;
    return raw == candidate ||
        path == candidate ||
        raw.endsWith(candidate) ||
        path.endsWith(candidate);
  }

  String _actionLabel(Map<String, dynamic>? meta, String actionUrl) {
    if (meta != null) {
      final actionName = meta['action']?.toString() ?? '';
      final serviceName = meta['service']?.toString() ?? '';
      if (actionName.isNotEmpty && serviceName.isNotEmpty) {
        return "$actionName ($serviceName)";
      }
      if (actionName.isNotEmpty) return actionName;
    }
    if (actionUrl.isEmpty) return '';
    if (!actionUrl.contains('/actions/')) return "Webhook";
    final path = _actionPath(actionUrl);
    final parts = path.split('/').where((p) => p.isNotEmpty).toList();
    if (parts.isEmpty) return "Action";
    final tail = parts.length >= 2
        ? parts.sublist(parts.length - 2).join(' ')
        : parts.last;
    return _titleize(tail.replaceAll('_', ' '));
  }

  String _triggerLabel(Map<String, dynamic>? meta, String triggerType) {
    if (meta != null) {
      final triggerName = meta['trigger']?.toString() ?? '';
      final serviceName = meta['service']?.toString() ?? '';
      if (triggerName.isNotEmpty && serviceName.isNotEmpty) {
        return "$triggerName ($serviceName)";
      }
      if (triggerName.isNotEmpty) return triggerName;
    }
    return triggerType;
  }

  String _titleize(String raw) {
    final words = raw.split(RegExp(r'\s+'));
    final caps = words.map((word) {
      if (word.isEmpty) return word;
      return "${word[0].toUpperCase()}${word.substring(1)}";
    }).toList();
    return caps.join(' ').trim();
  }

  List<MapEntry<String, String>> _formatDetails(
    Map<String, dynamic> details, {
    List fieldDefs = const [],
  }) {
    final entries = <MapEntry<String, String>>[];
    final usedKeys = <String>{};

    for (final def in fieldDefs) {
      if (def is! Map) continue;
      final key = def['key']?.toString() ?? '';
      if (key.isEmpty || _shouldHideKey(key)) continue;
      if (!details.containsKey(key)) continue;
      final value = _formatValue(details[key]);
      if (value.isEmpty) continue;
      entries.add(MapEntry(_labelForKey(key), value));
      usedKeys.add(key);
    }

    for (final entry in details.entries) {
      if (usedKeys.contains(entry.key) || _shouldHideKey(entry.key)) continue;
      final value = _formatValue(entry.value);
      if (value.isEmpty) continue;
      entries.add(MapEntry(_labelForKey(entry.key), value));
    }

    return entries;
  }

  bool _shouldHideKey(String key) {
    return key.toLowerCase().contains('token');
  }

  String _labelForKey(String key) {
    final cleaned = key.replaceAll('_', ' ').trim();
    if (cleaned.isEmpty) return key;
    final words = cleaned.split(RegExp(r'\s+'));
    final mapped = words.map((word) {
      if (word.isEmpty) return word;
      final lower = word.toLowerCase();
      if (lower == 'id') return 'ID';
      if (lower == 'url') return 'URL';
      if (lower == 'ts') return 'TS';
      return "${word[0].toUpperCase()}${word.substring(1)}";
    }).toList();
    return mapped.join(' ');
  }

  String _formatValue(dynamic value) {
    if (value == null) return '';
    if (value is List) {
      return value.map((item) => item.toString()).join(', ');
    }
    if (value is Map) {
      return jsonEncode(value);
    }
    return value.toString();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final textTheme = theme.textTheme;
    final workflows = _filteredWorkflows;

    return Scaffold(
      appBar: AppBar(
        backgroundColor: colorScheme.surface,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        title: Text(
          'My Konect',
          style: textTheme.titleLarge?.copyWith(
            color: colorScheme.onSurface,
            fontWeight: FontWeight.bold,
            fontFamily: 'Serif',
            fontSize: 22,
          ),
        ),
        actions: [
          IconButton(
            icon: Icon(Icons.person_outline, color: colorScheme.onSurface),
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
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'My Konects',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w500,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 16),
                ],
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: AppSearchBar(
                controller: _searchController,
                onChanged: (_) => setState(() {}),
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 16),
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
                    style: TextStyle(color: colorScheme.error),
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
                    badgeColor: isManual
                        ? colorScheme.onSurfaceVariant
                        : (enabled ? Colors.green : Colors.red),
                    onTap: () {
                      if (isManual) {
                        _triggerManualWorkflow(item);
                      } else {
                        _showWorkflowControls(item);
                      }
                    },
                    onLongPress: () => _showWorkflowControls(item),
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
        backgroundColor: colorScheme.primary,
        foregroundColor: colorScheme.onPrimary,
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
      const Color(0xFF00D0FF),
      const Color(0xFFFF4081),
      const Color(0xFF00E676),
      const Color(0xFFD500F9),
    ];
    return colors[index % colors.length];
  }
}

class _DetailsSection extends StatelessWidget {
  final String title;
  final List<MapEntry<String, String>> entries;

  const _DetailsSection({
    required this.title,
    required this.entries,
  });

  @override
  Widget build(BuildContext context) {
    if (entries.isEmpty) return const SizedBox.shrink();
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      margin: const EdgeInsets.only(top: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: colorScheme.surfaceVariant,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: TextStyle(
              fontWeight: FontWeight.w600,
              color: colorScheme.onSurface,
            ),
          ),
          const SizedBox(height: 8),
          ...entries.map(
            (entry) => Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    entry.key,
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    entry.value,
                    style: TextStyle(
                      fontSize: 13,
                      color: colorScheme.onSurface,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
