import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/utils/notifications.dart';
import 'package:file_picker/file_picker.dart';
import 'dart:js_interop';
import 'package:web/web.dart' as web;
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';

class PathLossEntry {
  final int id;
  final TextEditingController descriptionController;
  final TextEditingController lossController;
  String type;

  PathLossEntry({
    required this.id,
    required String description,
    required double loss,
    required this.type,
  }) : descriptionController = TextEditingController(text: description),
       lossController = TextEditingController(text: loss.toString());

  void dispose() {
    descriptionController.dispose();
    lossController.dispose();
  }
}

class LinkLossScreen extends StatefulWidget {
  const LinkLossScreen({super.key});

  @override
  State<LinkLossScreen> createState() => _LinkLossScreenState();
}

class _LinkLossScreenState extends State<LinkLossScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  bool _isLoading = true;
  bool _isHelpOpen = false;

  List<String> _testPhases = [];
  String _selectedTestPhase = '';

  List<String> _configs = [];
  String _selectedConfig = '';

  List<PathLossEntry> _entries = [];
  final List<String> _lossTypes = ['Common', 'SA', 'PM', 'SC'];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _tabController.addListener(_onTabChanged);
    _loadMetadata();
  }

  void _onTabChanged() {
    if (_tabController.indexIsChanging) return;
    if (_selectedTestPhase.isNotEmpty) {
      _loadConfigs();
    }
  }

  @override
  void dispose() {
    _tabController.removeListener(_onTabChanged);
    _tabController.dispose();
    for (var entry in _entries) {
      entry.dispose();
    }
    super.dispose();
  }

  bool _hasSyncedMetadata = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _syncMetadata();
  }

  void _syncMetadata([ServerService? service]) {
    final serverService = service ?? Provider.of<ServerService>(context);
    final metadata = serverService.status.bootstrapData?.databaseData;

    if (metadata != null) {
      if (metadata.ok && !_hasSyncedMetadata) {
        debugPrint(
          'LinkLossScreen: Syncing ${metadata.testPhases.length} test phases',
        );
        setState(() {
          _testPhases = metadata.testPhases;
          if (_testPhases.isNotEmpty && _selectedTestPhase.isEmpty) {
            _selectedTestPhase = _testPhases.first;
          }
          _hasSyncedMetadata = true;
          _isLoading = false;
        });
        if (_selectedTestPhase.isNotEmpty) {
          _loadConfigs();
        }
      } else if (!metadata.ok) {
        setState(() => _isLoading = false);
      }
    }
  }

  void _loadMetadata() {
    setState(() => _isLoading = true);
    _hasSyncedMetadata = false;
    final serverService = Provider.of<ServerService>(context, listen: false);
    if (serverService.status.bootstrapData?.databaseData != null) {
      _syncMetadata(serverService);
    } else {
      serverService.fetchBootstrapData();
    }
  }

  Future<void> _loadConfigs() async {
    final service = context.read<ServerService>();
    final isUplink = _tabController.index == 0;
    final resp = await service.fetchConfigsForLoss(
      _selectedTestPhase,
      isUplink,
    );

    if (resp != null && resp.ok) {
      setState(() {
        _configs = resp.configs;
        if (_configs.isNotEmpty) {
          _selectedConfig = _configs.first;
        } else {
          _selectedConfig = '';
          _entries = [];
        }
      });
      if (_selectedConfig.isNotEmpty) {
        await _loadProfile();
      }
    }
  }

  Future<void> _loadProfile() async {
    final service = context.read<ServerService>();
    final isUplink = _tabController.index == 0;
    final resp = await service.fetchLossProfile(
      _selectedTestPhase,
      _selectedConfig,
      isUplink,
    );

    if (resp != null && resp.ok) {
      final lines = resp.profile
          .split('\n')
          .where((l) => l.trim().isNotEmpty)
          .toList();
      List<PathLossEntry> newEntries = [];
      for (int i = 0; i < lines.length; i++) {
        final parts = lines[i].split(',');
        if (parts.length >= 4) {
          newEntries.add(
            PathLossEntry(
              id: i + 1,
              description: parts[1].trim(),
              loss: double.tryParse(parts[2].trim()) ?? 0.0,
              type: parts[3].trim(),
            ),
          );
        }
      }
      setState(() {
        for (var e in _entries) {
          e.dispose();
        }
        _entries = newEntries;
      });
    }
  }

  Future<void> _saveProfile() async {
    final service = context.read<ServerService>();
    final isUplink = _tabController.index == 0;

    List<String> rows = [];
    for (var entry in _entries) {
      final desc = entry.descriptionController.text;
      final loss = entry.lossController.text;
      rows.add("${entry.id},$desc,$loss,${entry.type}");
    }
    final profile = rows.join('\n');

    final ack = await service.saveLossProfile(
      _selectedTestPhase,
      _selectedConfig,
      profile,
      isUplink,
    );

    if (!mounted) return;

    if (ack != null && ack.ok) {
      AppNotifications.show(
        context,
        ack.message,
        type: NotificationType.success,
      );
      _loadProfile();
    } else {
      AppNotifications.show(
        context,
        ack?.message ?? 'Save failed',
        type: NotificationType.error,
      );
    }
  }

  void _addRow() {
    setState(() {
      _entries.add(
        PathLossEntry(
          id: _entries.length + 1,
          description: '',
          loss: 0.0,
          type: 'Common',
        ),
      );
    });
  }

  void _removeRow(int index) {
    setState(() {
      _entries[index].dispose();
      _entries.removeAt(index);
    });
  }

  Future<void> _importCSV() async {
    FilePickerResult? result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['csv'],
      withData: true,
    );

    if (!mounted) return;

    if (result == null || result.files.single.bytes == null) return;

    try {
      final content = utf8.decode(result.files.single.bytes!);
      final lines = content
          .split('\n')
          .where((l) => l.trim().isNotEmpty)
          .toList();

      // Assume first line is header if it contains "Sl. No" or "Description"
      int startIndex = 0;
      if (lines.isNotEmpty &&
          (lines[0].contains('Sl. No') || lines[0].contains('Description'))) {
        startIndex = 1;
      }

      List<PathLossEntry> imported = [];
      for (int i = startIndex; i < lines.length; i++) {
        final parts = lines[i].split(',');
        if (parts.length >= 4) {
          imported.add(
            PathLossEntry(
              id: imported.length + 1,
              description: parts[1].trim(),
              loss: double.tryParse(parts[2].trim()) ?? 0.0,
              type: parts[3].trim(),
            ),
          );
        }
      }

      setState(() {
        for (var e in _entries) {
          e.dispose();
        }
        _entries = imported;
      });
      AppNotifications.show(
        context,
        "CSV Imported",
        type: NotificationType.success,
      );
    } catch (e) {
      AppNotifications.show(
        context,
        "Import failed: $e",
        type: NotificationType.error,
      );
    }
  }

  void _downloadCSV() {
    if (_selectedConfig.isEmpty) return;

    String csv = "Sl. No,Description,Loss (dB),Type\n";
    for (var entry in _entries) {
      csv +=
          "${entry.id},${entry.descriptionController.text},${entry.lossController.text},${entry.type}\n";
    }

    final bytes = utf8.encode(csv);
    final blob = web.Blob(
      [bytes.toJS].toJS,
      web.BlobPropertyBag(type: 'text/csv'),
    );
    final url = web.URL.createObjectURL(blob);
    final anchor = web.document.createElement('a') as web.HTMLAnchorElement;
    anchor.href = url;
    anchor.download = "$_selectedConfig-loss.csv";
    anchor.click();
    web.URL.revokeObjectURL(url);

    AppNotifications.show(
      context,
      "CSV Downloaded",
      type: NotificationType.success,
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: Stack(
        children: [
          Column(
            children: [
              ScreenHeader(
                title: 'Database Management',
                subtitle: 'Manage test phases and path loss profiles',
                icon: Icons.storage,
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    ElevatedButton.icon(
                      onPressed: _showAddPhaseDialog,
                      icon: const Icon(Icons.add_box_outlined),
                      label: const Text('New Test Phase'),
                    ),
                    const SizedBox(width: 12),
                    _buildHelpTrigger(theme),
                  ],
                ),
              ),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.all(24.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildPhaseSelector(theme),
                      const SizedBox(height: 24),
                      Expanded(
                        child: ContentCard(
                          child: Column(
                            children: [
                              TabBar(
                                controller: _tabController,
                                labelColor: theme.colorScheme.primary,
                                unselectedLabelColor: Colors.grey,
                                indicatorColor: theme.colorScheme.primary,
                                tabs: const [
                                  Tab(text: 'Uplink (RX) Path Loss'),
                                  Tab(text: 'Downlink (TX) Path Loss'),
                                ],
                              ),
                              Expanded(
                                child: _isLoading
                                    ? const Center(
                                        child: CircularProgressIndicator(),
                                      )
                                    : _buildEditorView(theme),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
          if (_isHelpOpen)
            GestureDetector(
              onTap: () => setState(() => _isHelpOpen = false),
              child: Container(color: Colors.black.withValues(alpha: 0.1)),
            ),
          AnimatedPositioned(
            duration: const Duration(milliseconds: 300),
            curve: Curves.easeInOut,
            right: _isHelpOpen ? 0 : -450,
            top: 0,
            bottom: 0,
            child: _buildHelpPanel(theme),
          ),
        ],
      ),
    );
  }

  Widget _buildPhaseSelector(ThemeData theme) {
    return ContentCard(
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
      child: Row(
        children: [
          const Icon(Icons.layers, color: Colors.grey),
          const SizedBox(width: 16),
          const Text(
            'Effective Test Phase:',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: DropdownButton<String>(
              value: _selectedTestPhase,
              isExpanded: true,
              underline: const SizedBox(),
              items: _testPhases.map((phase) {
                return DropdownMenuItem(value: phase, child: Text(phase));
              }).toList(),
              onChanged: (val) {
                if (val != null) {
                  setState(() => _selectedTestPhase = val);
                  _loadConfigs();
                }
              },
            ),
          ),
          const SizedBox(width: 16),
          TextButton.icon(
            onPressed: () async {
              final service = context.read<ServerService>();
              final ack = await service.selectTestPhase(_selectedTestPhase);
              if (ack != null && ack.ok) {
                await service.fetchBootstrapData();
                if (!mounted) return;
                AppNotifications.show(
                  context,
                  "System Test Phase Updated",
                  type: NotificationType.success,
                );
              }
            },
            icon: const Icon(Icons.check_circle_outline),
            label: const Text('Set as System Default'),
          ),
        ],
      ),
    );
  }

  Widget _buildEditorView(ThemeData theme) {
    return Padding(
      padding: const EdgeInsets.all(24.0),
      child: Column(
        children: [
          Row(
            children: [
              const Text(
                'Configuration:',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              const SizedBox(width: 16),
              SizedBox(
                width: 300,
                child: DropdownButtonFormField<String>(
                  initialValue: _selectedConfig.isEmpty ? null : _selectedConfig,
                  decoration: const InputDecoration(
                    contentPadding: EdgeInsets.symmetric(horizontal: 12),
                    border: OutlineInputBorder(),
                  ),
                  items: _configs.map((c) {
                    return DropdownMenuItem(value: c, child: Text(c));
                  }).toList(),
                  onChanged: (val) {
                    if (val != null) {
                      setState(() => _selectedConfig = val);
                      _loadProfile();
                    }
                  },
                ),
              ),
              const Spacer(),
              IconButton.outlined(
                onPressed: _importCSV,
                icon: const Icon(Icons.upload_file),
                tooltip: 'Import CSV',
              ),
              const SizedBox(width: 8),
              IconButton.outlined(
                onPressed: _downloadCSV,
                icon: const Icon(Icons.download),
                tooltip: 'Download CSV',
              ),
              const SizedBox(width: 8),
              ElevatedButton.icon(
                onPressed: _addRow,
                icon: const Icon(Icons.add),
                label: const Text('Add Row'),
              ),
              const SizedBox(width: 8),
              ElevatedButton.icon(
                onPressed: _saveProfile,
                style: ElevatedButton.styleFrom(
                  backgroundColor: theme.colorScheme.primary,
                  foregroundColor: Colors.white,
                ),
                icon: const Icon(Icons.save),
                label: const Text('Save Changes'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          Expanded(
            child: _selectedConfig.isEmpty
                ? const Center(
                    child: Text('No configuration selected or available'),
                  )
                : _buildDataTable(theme),
          ),
        ],
      ),
    );
  }

  Widget _buildDataTable(ThemeData theme) {
    return SingleChildScrollView(
      child: Table(
        columnWidths: const {
          0: FixedColumnWidth(60),
          1: FlexColumnWidth(4),
          2: FlexColumnWidth(2),
          3: FlexColumnWidth(2),
          4: FixedColumnWidth(60),
        },
        border: TableBorder.all(
          color: Colors.grey.shade200,
          borderRadius: BorderRadius.circular(8),
        ),
        children: [
          TableRow(
            decoration: BoxDecoration(color: Colors.grey.shade50),
            children: [
              _buildHeaderCell('Sl.'),
              _buildHeaderCell('Description'),
              _buildHeaderCell('Loss (dB)'),
              _buildHeaderCell('Type'),
              _buildHeaderCell(''),
            ],
          ),
          ..._entries.asMap().entries.map((entry) {
            final idx = entry.key;
            final e = entry.value;
            return TableRow(
              children: [
                _buildCell(Text('${idx + 1}', textAlign: TextAlign.center)),
                _buildCell(
                  TextField(
                    controller: e.descriptionController,
                    decoration: const InputDecoration(
                      border: InputBorder.none,
                      isDense: true,
                    ),
                  ),
                ),
                _buildCell(
                  TextField(
                    controller: e.lossController,
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
                    decoration: const InputDecoration(
                      border: InputBorder.none,
                      isDense: true,
                    ),
                  ),
                ),
                _buildCell(
                  DropdownButton<String>(
                    value: e.type,
                    isExpanded: true,
                    underline: const SizedBox(),
                    items: _lossTypes
                        .map((t) => DropdownMenuItem(value: t, child: Text(t)))
                        .toList(),
                    onChanged: (val) {
                      if (val != null) setState(() => e.type = val);
                    },
                  ),
                ),
                _buildCell(
                  IconButton(
                    icon: const Icon(
                      Icons.delete_outline,
                      color: Colors.red,
                      size: 20,
                    ),
                    onPressed: () => _removeRow(idx),
                  ),
                ),
              ],
            );
          }),
        ],
      ),
    );
  }

  Widget _buildHeaderCell(String text) {
    return Padding(
      padding: const EdgeInsets.all(12.0),
      child: Text(
        text,
        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
      ),
    );
  }

  Widget _buildCell(Widget child) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12.0, vertical: 4.0),
      child: Center(heightFactor: 1, child: child),
    );
  }

  void _showAddPhaseDialog() {
    final nameController = TextEditingController();
    String copyFrom = '';

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('Add New Test Phase'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: nameController,
                decoration: const InputDecoration(labelText: 'Phase Name'),
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                initialValue: copyFrom.isEmpty ? null : copyFrom,
                decoration: const InputDecoration(
                  labelText: 'Copy losses from (Optional)',
                ),
                items: [
                  const DropdownMenuItem(
                    value: '',
                    child: Text('None (Empty losses)'),
                  ),
                  ..._testPhases.map(
                    (p) => DropdownMenuItem(value: p, child: Text(p)),
                  ),
                ],
                onChanged: (val) => setDialogState(() => copyFrom = val ?? ''),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            ElevatedButton(
              onPressed: () async {
                final name = nameController.text.trim();
                if (name.isEmpty) return;

                final service = context.read<ServerService>();
                final ack = await service.addNewTestPhase(name, copyFrom);
                if (!context.mounted) return;
                if (ack != null && ack.ok) {
                  await service.fetchBootstrapData();
                  if (!context.mounted) return;
                  Navigator.pop(context);
                  _loadMetadata();
                  AppNotifications.show(
                    context,
                    "Phase Created",
                    type: NotificationType.success,
                  );
                } else {
                  AppNotifications.show(
                    context,
                    ack?.message ?? "Failed",
                    type: NotificationType.error,
                  );
                }
              },
              child: const Text('Create'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHelpTrigger(ThemeData theme) {
    return InkWell(
      onTap: () => setState(() => _isHelpOpen = !_isHelpOpen),
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: _isHelpOpen
              ? theme.colorScheme.primary
              : theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: _isHelpOpen
                ? theme.colorScheme.primary
                : theme.colorScheme.primary.withValues(alpha: 0.2),
          ),
        ),
        child: Icon(
          Icons.help_outline,
          color: _isHelpOpen ? Colors.white : theme.colorScheme.primary,
          size: 24,
        ),
      ),
    );
  }

  Widget _buildHelpPanel(ThemeData theme) {
    return Container(
      width: 450,
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 30,
            offset: const Offset(-10, 0),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(24, 40, 24, 24),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Database Help',
                  style: GoogleFonts.outfit(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: theme.colorScheme.primary,
                  ),
                ),
                IconButton(
                  onPressed: () => setState(() => _isHelpOpen = false),
                  icon: const Icon(Icons.close),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(24),
              children: [
                _buildHelpItem(
                  theme,
                  'Test Phases',
                  'Phases represent major project milestones (e.g., Payload, Thermal). Switching the "Effective Phase" '
                      'updates the loss reference for the entire PRISM ecosystem in real-time.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Loss Profiles',
                  'Losses are defined per Configuration (VSA/PM). The system sums these values with live '
                      'measurements to provide "Corrected" readings on measurement screens.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Type Categorization',
                  '• **Common**: Shared loss across all paths.\n'
                      '• **SA/PM/SC**: Instrument-specific calibration offsets.\n'
                      'PRISM uses these to narrow down exactly where path loss originates.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Import / Export',
                  'Use CSV files for mass updates. Expected format: `ID, Description, Loss, Type`. '
                      'Always "Save Changes" after an import to commit to the server.',
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHelpItem(ThemeData theme, String title, String content) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: GoogleFonts.inter(fontWeight: FontWeight.bold, fontSize: 16),
        ),
        const SizedBox(height: 8),
        Text(
          content,
          style: GoogleFonts.inter(
            height: 1.6,
            fontSize: 14,
            color: Colors.grey.shade600,
          ),
        ),
      ],
    );
  }
}
