import 'dart:convert';
import 'dart:typed_data';
import 'package:file_picker/file_picker.dart';
import 'package:file_saver/file_saver.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:prism/app_icons.dart';
import 'package:prism/server_communication/classes.dart';
import 'package:prism/server_communication/messages.dart';
import 'package:prism/structures.dart';
import 'package:prism/theme.dart';


class DownlinkLossEditor extends StatefulWidget {
  final Global global;

  const DownlinkLossEditor(this.global, {super.key});

  @override
  State<DownlinkLossEditor> createState() => _DownlinkLossEditorState();
}

class _DownlinkLossEditorState extends State<DownlinkLossEditor> {
  List<String> _testPhases = [];
  List<String> _configs = [];
  final List<String> _lossTypes = ['Common', 'SA', 'PM', 'SC'];

  String _selectedConfigName = '';
  String _selectedTestPhase = '';

  List<LossEntry> _lossEntries = [];
  int _nextId = 1;

  @override
  void initState() {
    super.initState();
    _getTestPhases();
  }

  @override
  void dispose() {
    for (var entry in _lossEntries) {
      entry.dispose();
    }
    super.dispose();
  }

  void _getTestPhases() async {
    GetResponse resp = await getParameters(widget.global, ['TestPhases']);
    if (!resp.ok) {
      return;
    }
    _testPhases = resp.getValue('TestPhases');
    _selectedTestPhase = _testPhases.isNotEmpty ? _testPhases.first : '';
    _getConfigs();
  }

  void _getConfigs() async {
    Acknowledgement ack = await setParameters(
      widget.global,
      ['SelectedTestPhase'],
      [_selectedTestPhase],
    );
    if (!ack.ok) {
      return;
    }
    GetResponse resp = await getParameters(widget.global, [
      'ConfigsForDownlink',
    ]);
    if (!resp.ok) {
      return;
    }
    _configs = resp.getValue('ConfigsForDownlink');
    _selectedConfigName = _configs.isNotEmpty ? _configs.first : '';
    _fetchLossProfile();
  }

  void _clearEntries() {
    for (var entry in _lossEntries) {
      entry.dispose();
    }
    _lossEntries = [];
  }

  void _fetchLossProfile() async {
    Acknowledgement ack = await setParameters(
      widget.global,
      ['SelectedConfiguration'],
      [_selectedConfigName],
    );
    if (!ack.ok) {
      return;
    }
    GetResponse resp = await getParameters(widget.global, [
      'DownlinkLossProfile',
    ]);
    if (!resp.ok) {
      return;
    }
    List<String> profile = resp.getValue('DownlinkLossProfile');
    final lines = profile[0]
        .split('\n')
        .where((l) => l.trim().isNotEmpty)
        .toList();
    _clearEntries(); // Dispose old controllers before fetching new data

    List<LossEntry> entries = [];
    try {
      for (int i = 0; i < lines.length; i++) {
        final line = lines[i].trim();
        if (line.isEmpty) continue;

        final parts = line.split(',');
        if (parts.length != 4) {
          throw Exception(
            'Invalid format on row ${i + 1}. Expected 4 columns.',
          );
        }

        final description = parts[1].trim();
        final loss = double.tryParse(parts[2].trim());
        final type = parts[3].trim();

        if (loss == null) {
          throw Exception('Invalid loss value on row ${i + 1}.');
        }
        if (!_lossTypes.contains(type)) {
          throw Exception('Invalid type "$type" on row ${i + 1}.');
        }

        entries.add(
          LossEntry(
            id: i + 1,
            description: description,
            loss: loss,
            type: type,
          ),
        );
      }
      _lossEntries = entries;
      _nextId = _lossEntries.length + 1;
    } catch (e) {
      widget.global.updateNotification(
        e.toString(),
        NotificationType.failure,
        null,
      );
    }

    setState(() {
      _lossEntries = entries;
      _nextId = _lossEntries.length + 1;
    });
  }

  void _addRow() {
    setState(() {
      _lossEntries.add(LossEntry.empty(_nextId++));
    });
  }

  void _removeRow(int id) {
    final entry = _lossEntries.firstWhere((e) => e.id == id);
    entry.dispose(); // Dispose controllers before removing
    setState(() {
      _lossEntries.remove(entry);
    });
  }

  void _saveProfile() async {
    List<String> rows = [];
    for (var i = 0; i < _lossEntries.length; i++) {
      final entry = _lossEntries[i];
      final loss = double.tryParse(entry.lossController.text) ?? 0.0;
      String row =
          "${entry.id},${entry.descriptionController.text},$loss,${entry.type}";
     rows.add(row);
    }
    String loss = rows.join("\n");
    ActionRequest action = ActionRequest();
    action.action = "SaveDownlinkLoss";
    action.addParameter("TestPhase", [_selectedTestPhase]);
    action.addParameter("Configuration", [_selectedConfigName]);
    action.addParameter("Loss", [loss]);
    Acknowledgement ack = await initiateServerAction(
      widget.global,
      action,
    );
    if (!ack.ok) {
      return;
    }
    widget.global.updateNotification(
      "Loss Updated in Database",
      NotificationType.success,
      null,
    );
    _fetchLossProfile();
  }

  Future<void> _importFromCSV() async {
    FilePickerResult? result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['csv'],
      withData: true,
    );

    if (result == null) {
      return;
    }

    try {
      final bytes = result.files.single.bytes;
      if (bytes == null) {
        _showSnackBar('Could not read file.', isError: true);
        return;
      }
      final content = utf8.decode(bytes);
      final lines = content
          .split('\n')
          .where((l) => l.trim().isNotEmpty)
          .toList();

      if (lines.length < 2) {
        _showSnackBar('CSV file is empty or has no data rows.', isError: true);
        return;
      }

      final List<LossEntry> importedEntries = [];
      int currentId = 1;
      // Start from 1 to skip header
      for (int i = 1; i < lines.length; i++) {
        final line = lines[i].trim();
        if (line.isEmpty) continue;

        final parts = line.split(',');
        if (parts.length != 4) {
          throw Exception(
            'Invalid format on row ${i + 1}. Expected 4 columns.',
          );
        }

        final description = parts[1];
        final loss = double.tryParse(parts[2]);
        final type = parts[3];

        if (loss == null) {
          throw Exception('Invalid loss value on row ${i + 1}.');
        }
        if (!_lossTypes.contains(type)) {
          throw Exception('Invalid type "$type" on row ${i + 1}.');
        }

        importedEntries.add(
          LossEntry(
            id: currentId++,
            description: description,
            loss: loss,
            type: type,
          ),
        );
      }

      setState(() {
        _clearEntries(); // Dispose old controllers
        _lossEntries = importedEntries;
        _nextId = _lossEntries.length + 1;
      });

      _showSnackBar('CSV imported successfully!', isError: false);
    } catch (e) {
      _showSnackBar('Error processing file: $e', isError: true);
    }
  }

  Future<void> _exportToCSV() async {
    if (_lossEntries.isEmpty) {
      _showSnackBar('No data to export.', isError: true);
      return;
    }

    final List<String> rows = ['Sl. No,Description,Loss,Type']; // Header
    for (int i = 0; i < _lossEntries.length; i++) {
      final entry = _lossEntries[i];
      final loss = double.tryParse(entry.lossController.text) ?? 0.0;
      rows.add(
        '${i + 1},${entry.descriptionController.text},$loss,${entry.type}',
      );
    }

    final String csvString = rows.join('\n');
    final Uint8List bytes = utf8.encode(csvString);
    final timeStamp = DateFormat('yyyy-MM-dd_HH-mm').format(DateTime.now());
    final fileName = 'DownlinkLoss_${_selectedConfigName}_$timeStamp';

    try {
      await FileSaver.instance.saveFile(
        name: fileName,
        bytes: bytes,
        fileExtension: 'csv',
        mimeType: MimeType.csv,
      );
      _showSnackBar('Exported successfully!', isError: false);
    } catch (e) {
      _showSnackBar('Error exporting file: $e', isError: true);
    }
  }

  void _showSnackBar(String message, {required bool isError}) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: isError ? Colors.red : Colors.green,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSelectionSection(),
          const SizedBox(height: 16),
          _buildActionButtons(),
          const SizedBox(height: 16),
          Expanded(child: _buildDataTable()),
        ],
      ),
    );
  }

  Widget _buildSelectionSection() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12.0),
        child: Row(
          children: [
            Expanded(
              child: DropdownButtonFormField<String>(
                value: _selectedTestPhase,
                decoration: const InputDecoration(
                  labelText: 'Test Phase',
                  border: OutlineInputBorder(),
                ),
                items: _testPhases
                    .map(
                      (phase) =>
                          DropdownMenuItem(value: phase, child: Text(phase)),
                    )
                    .toList(),
                onChanged: (value) {
                  if (value != null) {
                    setState(() {
                      _selectedTestPhase = value;
                      _getConfigs();
                    });
                  }
                },
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: DropdownButtonFormField<String>(
                value: _selectedConfigName,
                decoration: const InputDecoration(
                  labelText: 'Config Name',
                  border: OutlineInputBorder(),
                ),
                items: _configs
                    .map(
                      (name) =>
                          DropdownMenuItem(value: name, child: Text(name)),
                    )
                    .toList(),
                onChanged: (value) {
                  if (value != null) {
                    setState(() {
                      _selectedConfigName = value;
                      _fetchLossProfile();
                    });
                  }
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildActionButtons() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.end,
      children: [
        OutlinedButton.icon(
          onPressed: _importFromCSV,
          icon: const Icon(AppIcons.upload),
          label: const Text('Import CSV'),
          style: AppTheme.secondaryButtonStyle(context),
        ),
        const SizedBox(width: 8),
        OutlinedButton.icon(
          onPressed: _exportToCSV,
          icon: const Icon(AppIcons.download),
          label: const Text('Export CSV'),
          style: AppTheme.secondaryButtonStyle(context),
        ),
        const SizedBox(width: 8),
        OutlinedButton.icon(
          onPressed: _addRow,
          icon: const Icon(AppIcons.add),
          label: const Text('Add Row'),
          style: AppTheme.secondaryButtonStyle(context),
        ),
        const SizedBox(width: 16),
        ElevatedButton.icon(
          onPressed: _saveProfile,
          icon: Icon(AppIcons.save),
          label: const Text('Save Profile'),
          style: AppTheme.primaryButtonStyle(context),
        ),
      ],
    );
  }

  Widget _buildDataTable() {
    return Card(
      child: SingleChildScrollView(
        scrollDirection: Axis.vertical,
        child: DataTable(
          columns: const [
            DataColumn(label: Text('Sl. No')),
            DataColumn(label: Text('Description')),
            DataColumn(label: Text('Loss (dB)')),
            DataColumn(label: Text('Type')),
            DataColumn(label: Text('Actions')),
          ],
          rows: _lossEntries.map((entry) {
            int displayIndex = _lossEntries.indexOf(entry) + 1;
            return DataRow(
              key: ValueKey(entry.id), // Key is good practice for lists
              cells: [
                DataCell(Text(displayIndex.toString())),
                DataCell(
                  TextFormField(
                    controller: entry.descriptionController,
                    decoration: const InputDecoration(border: InputBorder.none),
                  ),
                ),
                DataCell(
                  TextFormField(
                    controller: entry.lossController,
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
                    decoration: const InputDecoration(border: InputBorder.none),
                  ),
                ),
                DataCell(
                  DropdownButtonFormField<String>(
                    value: entry.type,
                    decoration: const InputDecoration(border: InputBorder.none),
                    items: _lossTypes
                        .map(
                          (type) =>
                              DropdownMenuItem(value: type, child: Text(type)),
                        )
                        .toList(),
                    onChanged: (value) =>
                        setState(() => entry.type = value ?? _lossTypes.first),
                  ),
                ),
                DataCell(
                  IconButton(
                    icon: const Icon(AppIcons.delete, color: Colors.red),
                    onPressed: () => _removeRow(entry.id),
                  ),
                ),
              ],
            );
          }).toList(),
        ),
      ),
    );
  }
}
