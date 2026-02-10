import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';

import 'package:prism_client/services/server_service.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';

class ScpiLogEntry {
  final DateTime timestamp;
  final String type; // 'SENT', 'RECV', 'ERROR'
  final String message;

  ScpiLogEntry({
    required this.timestamp,
    required this.type,
    required this.message,
  });
}

class ScpiSequenceCommand {
  final String command;
  final int delayMs;
  final bool isQuery;

  ScpiSequenceCommand({
    required this.command,
    required this.delayMs,
    required this.isQuery,
  });
}

class ScpiCommanderScreen extends StatefulWidget {
  const ScpiCommanderScreen({super.key});

  @override
  State<ScpiCommanderScreen> createState() => _ScpiCommanderScreenState();
}

class _ScpiCommanderScreenState extends State<ScpiCommanderScreen> {
  // State Variables
  String? _selectedDevice;
  final TextEditingController _ipController = TextEditingController();
  final TextEditingController _portController = TextEditingController();

  CommandDetails? _selectedMnemonic;
  final TextEditingController _commandController = TextEditingController();
  final TextEditingController _replacementController = TextEditingController();
  final TextEditingController _argsController = TextEditingController();
  bool _isWriteAndRead = false;

  final List<ScpiLogEntry> _logs = [];
  final ScrollController _logScrollController = ScrollController();

  // CSV/Sequence State
  List<ScpiSequenceCommand> _sequence = [];
  bool _isExecutingSequence = false;
  int _currentSequenceIndex = -1;

  Key _mnemonicFieldKey = UniqueKey();
  bool _isHelpOpen = false;

  @override
  void dispose() {
    _ipController.dispose();
    _portController.dispose();
    _commandController.dispose();
    _replacementController.dispose();
    _argsController.dispose();
    _logScrollController.dispose();
    super.dispose();
  }

  void _onDeviceSelected(String? device, SCPIDetails? scpiData) {
    setState(() {
      _selectedDevice = device;
      if (device != null && scpiData != null) {
        final details = scpiData.instrumentDetails[device];
        if (details != null) {
          _ipController.text = details.ipAddress;
          _portController.text = details.portNo.toString();
        }
        _selectedMnemonic = null;
        _mnemonicFieldKey = UniqueKey();
        _commandController.clear();
      }
    });
  }

  void _onMnemonicSelected(CommandDetails? cmd) {
    setState(() {
      _selectedMnemonic = cmd;
      if (cmd != null) {
        _commandController.text = cmd.command;
        _isWriteAndRead = !cmd.write;
      }
    });
  }

  String _getFinalCommandString() {
    String base = _commandController.text;
    String replacement = _replacementController.text;
    if (replacement.isNotEmpty) {
      base = base.replaceAll('#', replacement);
    }
    String args = _argsController.text.trim();
    if (args.isNotEmpty) {
      base = "$base $args";
    }
    return base;
  }

  void _sendCommand() {
    final finalCmd = _getFinalCommandString();
    if (finalCmd.isEmpty) return;

    final serverService = Provider.of<ServerService>(context, listen: false);
    final scpiData = serverService.status.bootstrapData?.scpiData;
    if (scpiData == null) return;

    final ip = _ipController.text;
    final port = int.tryParse(_portController.text);
    if (ip.isEmpty || port == null) return;

    final request = SCPICommandRequest(
      ipAddress: ip,
      portNo: port,
      commands: [finalCmd],
      delays: [0.0],
      read: [_isWriteAndRead],
    );

    setState(() {
      _logs.insert(
        0,
        ScpiLogEntry(
          timestamp: DateTime.now(),
          type: 'SENT',
          message: finalCmd,
        ),
      );

      _selectedMnemonic = null;
      _mnemonicFieldKey = UniqueKey();
      _commandController.clear();
      _replacementController.clear();
      _argsController.clear();
      _isWriteAndRead = false;
    });

    serverService
        .connectSCPI(request)
        .listen(
          (response) {
            if (mounted) {
              setState(() {
                _logs.insert(
                  0,
                  ScpiLogEntry(
                    timestamp: DateTime.now(),
                    type: response.ok ? 'RECV' : 'ERROR',
                    message: response.ok
                        ? (response.response.isEmpty
                              ? "Success"
                              : response.response)
                        : response.message,
                  ),
                );
              });
            }
          },
          onDone: () {
            serverService.closeSCPI();
          },
          onError: (err) {
            if (mounted) {
              setState(() {
                _logs.insert(
                  0,
                  ScpiLogEntry(
                    timestamp: DateTime.now(),
                    type: 'ERROR',
                    message: "Connection error: $err",
                  ),
                );
              });
            }
            serverService.closeSCPI();
          },
        );
  }

  void _clearLogs() {
    setState(() {
      _logs.clear();
    });
  }

  void _handleCsvUpload() {
    // In a real app, use file_picker. For mock, we'll just populate a sequence.
    setState(() {
      _sequence = [
        ScpiSequenceCommand(command: '*RST', delayMs: 1000, isQuery: false),
        ScpiSequenceCommand(
          command: 'FREQ:CENT 1GHZ',
          delayMs: 500,
          isQuery: false,
        ),
        ScpiSequenceCommand(command: '*IDN?', delayMs: 200, isQuery: true),
        ScpiSequenceCommand(command: 'MEAS:POW?', delayMs: 1000, isQuery: true),
      ];
      _logs.insert(
        0,
        ScpiLogEntry(
          timestamp: DateTime.now(),
          type: 'SENT',
          message: "CSV Loaded: 4 commands identified.",
        ),
      );
    });
  }

  Future<void> _executeSequence() async {
    if (_sequence.isEmpty || _isExecutingSequence) return;

    final serverService = Provider.of<ServerService>(context, listen: false);
    final scpiData = serverService.status.bootstrapData?.scpiData;
    if (scpiData == null) return;

    final ip = _ipController.text;
    final port = int.tryParse(_portController.text);
    if (ip.isEmpty || port == null) return;

    final request = SCPICommandRequest(
      ipAddress: ip,
      portNo: port,
      commands: _sequence.map((e) => e.command).toList(),
      delays: _sequence.map((e) => e.delayMs.toDouble()).toList(),
      read: _sequence.map((e) => e.isQuery).toList(),
    );

    setState(() {
      _isExecutingSequence = true;
      _currentSequenceIndex = 0;
      _logs.insert(
        0,
        ScpiLogEntry(
          timestamp: DateTime.now(),
          type: 'SENT',
          message: "Starting sequence of ${_sequence.length} commands...",
        ),
      );
    });

    int receivedCount = 1;
    serverService
        .connectSCPI(request)
        .listen(
          (response) {
            if (mounted) {
              setState(() {
                _currentSequenceIndex = receivedCount - 1;
                _logs.insert(
                  0,
                  ScpiLogEntry(
                    timestamp: DateTime.now(),
                    type: response.ok ? 'RECV' : 'ERROR',
                    message:
                        "[#$receivedCount] CMD: ${response.command} -> ${response.ok ? (response.response.isEmpty ? "Success" : response.response) : response.message}",
                  ),
                );
                receivedCount++;
              });
            }
          },
          onDone: () {
            if (mounted) {
              setState(() {
                _isExecutingSequence = false;
                _currentSequenceIndex = -1;
                _logs.insert(
                  0,
                  ScpiLogEntry(
                    timestamp: DateTime.now(),
                    type: 'SENT',
                    message: "Sequence execution completed.",
                  ),
                );
              });
            }
            serverService.closeSCPI();
          },
          onError: (err) {
            if (mounted) {
              setState(() {
                _isExecutingSequence = false;
                _currentSequenceIndex = -1;
                _logs.insert(
                  0,
                  ScpiLogEntry(
                    timestamp: DateTime.now(),
                    type: 'ERROR',
                    message: "Sequence failed: $err",
                  ),
                );
              });
            }
            serverService.closeSCPI();
          },
        );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final serverService = Provider.of<ServerService>(context);
    final scpiData = serverService.status.bootstrapData?.scpiData;

    if (scpiData == null || !scpiData.ok) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const CircularProgressIndicator(),
            const SizedBox(height: 16),
            Text(
              scpiData?.message ?? 'Loading SCPI Metadata...',
              style: GoogleFonts.inter(color: Colors.grey),
            ),
          ],
        ),
      );
    }

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildHeader(theme),
                const SizedBox(height: 24),
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // 1. Device Selection Panel
                      Expanded(
                        flex: 3,
                        child: _buildDevicePanel(theme, scpiData),
                      ),
                      const SizedBox(width: 20),
                      // 2. Command Builder Panel
                      Expanded(
                        flex: 4,
                        child: _buildBuilderPanel(
                          theme,
                          scpiData,
                          serverService,
                        ),
                      ),
                      const SizedBox(width: 20),
                      // 3. Response Console Panel
                      Expanded(flex: 5, child: _buildConsolePanel(theme)),
                    ],
                  ),
                ),
              ],
            ),
          ),
          if (_isHelpOpen)
            GestureDetector(
              onTap: () => setState(() => _isHelpOpen = false),
              child: Container(color: Colors.black.withOpacity(0.1)),
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

  Widget _buildHeader(ThemeData theme) {
    return ScreenHeader(
      title: 'SCPI Commander',
      subtitle: 'Send direct SCPI commands to networked instruments',
      icon: Icons.terminal_rounded,
      trailing: _buildHelpTrigger(theme),
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
                : theme.colorScheme.primary.withOpacity(0.2),
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
            color: Colors.black.withOpacity(0.1),
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
                  'SCPI Commander Help',
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
                  'Direct Communication',
                  'SCPI Commander allows you to send raw ASCII commands to instruments over TCP/IP. '
                      'This is useful for debugging or executing specialized functions not covered by the main UI.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Mnemonic Look-up',
                  'The search field in the Command Builder helps you find known mnemonics for the selected device. '
                      'Selecting a mnemonic will automatically populate the command string and toggle the type (Write/Read).',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Placeholders (#)',
                  'Some commands use the `#` character as a placeholder for indices (like channel numbers). '
                      'When detected, a dedicated "Index Number" field will appear to let you fill this in.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Orchestration & CSV',
                  'For repetitive tasks, you can upload a CSV file containing a list of commands and delays. '
                      'PRISM will execute these sequentially from the server side for maximum timing precision.',
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

  Widget _buildDevicePanel(ThemeData theme, SCPIDetails scpiData) {
    bool isSequenceActive = _sequence.isNotEmpty;

    return AnimatedOpacity(
      duration: const Duration(milliseconds: 300),
      opacity: isSequenceActive ? 0.5 : 1.0,
      child: IgnorePointer(
        ignoring: isSequenceActive,
        child: ContentCard(
          isSidebar: true, // Radius 24
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildPanelLabel(theme, 'DEVICE SELECTION'),
              const SizedBox(height: 24),
              _buildDropdownEffect<String>(
                theme,
                label: 'Select Device',
                value: _selectedDevice,
                items: scpiData.instruments,
                icon: Icons.devices_other,
                itemLabel: (d) => d,
                onChanged: (val) => _onDeviceSelected(val, scpiData),
              ),
              const SizedBox(height: 24),
              const Divider(),
              const SizedBox(height: 24),
              _buildPanelLabel(theme, 'CONNECTION DETAILS'),
              const SizedBox(height: 16),
              _buildTextField(
                theme,
                label: 'IP Address',
                controller: _ipController,
                icon: Icons.lan_outlined,
                hint: 'e.g. 192.168.1.1',
              ),
              const SizedBox(height: 16),
              _buildTextField(
                theme,
                label: 'Port Number',
                controller: _portController,
                icon: Icons.numbers,
                hint: 'e.g. 5025',
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildBuilderPanel(
    ThemeData theme,
    SCPIDetails scpiData,
    ServerService serverService,
  ) {
    bool hasHash = _commandController.text.contains('#');
    List<CommandDetails> availableCommands = [];
    if (_selectedDevice != null) {
      availableCommands = List<CommandDetails>.from(
        scpiData.commands[_selectedDevice] ?? [],
      );
      availableCommands.sort((a, b) => a.mnemonic.compareTo(b.mnemonic));
    }

    return ContentCard(
      isSidebar: true,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildPanelLabel(theme, 'COMMAND BUILDER'),
          const SizedBox(height: 24),
          _buildSearchableMnemonic(theme, availableCommands),
          const SizedBox(height: 16),
          _buildTextField(
            theme,
            label: 'Command String',
            controller: _commandController,
            icon: Icons.text_fields,
            hint: 'Enter SCPI command',
            onChanged: (val) => setState(() {}),
          ),
          if (hasHash) ...[
            const SizedBox(height: 16),
            _buildTextField(
              theme,
              label: 'Index Number (#)',
              controller: _replacementController,
              icon: Icons.tag,
              hint: 'Replace # with integer',
            ),
          ],
          if (_selectedMnemonic?.argument == true) ...[
            const SizedBox(height: 16),
            _buildTextField(
              theme,
              label: 'Arguments',
              controller: _argsController,
              icon: Icons.playlist_add_outlined,
              hint: 'Required arguments',
            ),
          ],
          const SizedBox(height: 24),
          Row(
            children: [
              Text(
                'Command Type:',
                style: GoogleFonts.inter(
                  fontSize: 13,
                  fontWeight: FontWeight.bold,
                  color: Colors.grey.shade700,
                ),
              ),
              const Spacer(),
              _buildTypeToggle('Write Only', !_isWriteAndRead, () {
                setState(() => _isWriteAndRead = false);
              }),
              const SizedBox(width: 8),
              _buildTypeToggle('Write & Read', _isWriteAndRead, () {
                setState(() => _isWriteAndRead = true);
              }),
            ],
          ),
          const Spacer(),
          if (_sequence.isEmpty) ...[
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                onPressed: _handleCsvUpload,
                icon: const Icon(Icons.upload_file),
                label: const Text('UPLOAD CSV SEQUENCE'),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
              ),
            ),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _sendCommand,
                icon: const Icon(Icons.send_rounded),
                label: const Text('SEND COMMAND'),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 20),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  backgroundColor: const Color(0xFF1A237E),
                  foregroundColor: Colors.white,
                ),
              ),
            ),
          ] else ...[
            _buildSequenceProgress(theme),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: _isExecutingSequence ? null : _executeSequence,
                    icon: _isExecutingSequence
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Icon(Icons.play_arrow_rounded),
                    label: Text(
                      _isExecutingSequence ? 'RUNNING...' : 'RUN SEQUENCE',
                    ),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 20),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      backgroundColor: Colors.green.shade700,
                      foregroundColor: Colors.white,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                IconButton(
                  onPressed: () {
                    if (_isExecutingSequence) {
                      serverService.abortSCPI();
                    } else {
                      setState(() => _sequence = []);
                    }
                  },
                  icon: Icon(
                    _isExecutingSequence
                        ? Icons.stop_circle_outlined
                        : Icons.close_rounded,
                  ),
                  style: IconButton.styleFrom(
                    backgroundColor: Colors.red.shade50,
                    foregroundColor: Colors.red.shade700,
                    padding: const EdgeInsets.all(16),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSequenceProgress(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.blue.shade50.withOpacity(0.5),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.blue.shade100),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                Icons.checklist_rounded,
                size: 20,
                color: Colors.blue.shade700,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'SEQUENCE LOADED',
                      style: GoogleFonts.inter(
                        fontSize: 10,
                        fontWeight: FontWeight.w900,
                        color: Colors.blue.shade700,
                        letterSpacing: 1.2,
                      ),
                    ),
                    Text(
                      '${_sequence.length} commands identified',
                      style: GoogleFonts.inter(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: Colors.blue.shade900,
                      ),
                    ),
                  ],
                ),
              ),
              IconButton(
                onPressed: _showFullSequencePreview,
                icon: const Icon(Icons.visibility_outlined),
                tooltip: 'Preview Full Sequence',
                style: IconButton.styleFrom(
                  backgroundColor: Colors.white,
                  foregroundColor: Colors.blue.shade700,
                ),
              ),
            ],
          ),
          if (_isExecutingSequence) ...[
            const SizedBox(height: 16),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: (_currentSequenceIndex + 1) / _sequence.length,
                minHeight: 6,
                backgroundColor: Colors.blue.shade100,
                valueColor: AlwaysStoppedAnimation<Color>(Colors.blue.shade700),
              ),
            ),
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Step ${_currentSequenceIndex + 1} of ${_sequence.length}',
                  style: GoogleFonts.inter(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: Colors.blue.shade700,
                  ),
                ),
                Text(
                  '${((_currentSequenceIndex + 1) / _sequence.length * 100).toInt()}%',
                  style: GoogleFonts.inter(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: Colors.blue.shade700,
                  ),
                ),
              ],
            ),
          ] else
            Padding(
              padding: const EdgeInsets.only(top: 12.0),
              child: Text(
                'Ready to orchestrate from server side.',
                style: GoogleFonts.inter(
                  fontSize: 12,
                  color: Colors.blue.shade600,
                  fontStyle: FontStyle.italic,
                ),
              ),
            ),
        ],
      ),
    );
  }

  void _showFullSequencePreview() {
    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: Colors.white,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
        child: Container(
          width: 800,
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFF1A237E).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: const Icon(
                      Icons.table_chart_outlined,
                      color: Color(0xFF1A237E),
                    ),
                  ),
                  const SizedBox(width: 20),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Sequence Preview',
                        style: GoogleFonts.outfit(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          color: const Color(0xFF1A237E),
                        ),
                      ),
                      Text(
                        'Review commands before server-side execution',
                        style: GoogleFonts.inter(
                          fontSize: 14,
                          color: Colors.grey.shade600,
                        ),
                      ),
                    ],
                  ),
                  const Spacer(),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close),
                    style: IconButton.styleFrom(
                      backgroundColor: Colors.grey.shade100,
                      padding: const EdgeInsets.all(12),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 32),
              ConstrainedBox(
                constraints: const BoxConstraints(maxHeight: 400),
                child: Container(
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: Colors.grey.shade200),
                  ),
                  child: SingleChildScrollView(
                    child: DataTable(
                      headingRowColor: MaterialStateProperty.all(
                        Colors.grey.shade50,
                      ),
                      columns: [
                        DataColumn(
                          label: Text(
                            'STEP',
                            style: GoogleFonts.inter(
                              fontWeight: FontWeight.w900,
                              fontSize: 11,
                              color: Colors.grey.shade600,
                            ),
                          ),
                        ),
                        DataColumn(
                          label: Text(
                            'COMMAND',
                            style: GoogleFonts.inter(
                              fontWeight: FontWeight.w900,
                              fontSize: 11,
                              color: Colors.grey.shade600,
                            ),
                          ),
                        ),
                        DataColumn(
                          label: Text(
                            'DELAY (ms)',
                            style: GoogleFonts.inter(
                              fontWeight: FontWeight.w900,
                              fontSize: 11,
                              color: Colors.grey.shade600,
                            ),
                          ),
                        ),
                        DataColumn(
                          label: Text(
                            'MODE',
                            style: GoogleFonts.inter(
                              fontWeight: FontWeight.w900,
                              fontSize: 11,
                              color: Colors.grey.shade600,
                            ),
                          ),
                        ),
                      ],
                      rows: _sequence.asMap().entries.map((entry) {
                        final i = entry.key;
                        final cmd = entry.value;
                        return DataRow(
                          cells: [
                            DataCell(
                              Text(
                                '#${i + 1}',
                                style: const TextStyle(
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                            DataCell(
                              Text(
                                cmd.command,
                                style: GoogleFonts.jetBrainsMono(
                                  fontSize: 13,
                                  color: const Color(0xFF1A237E),
                                ),
                              ),
                            ),
                            DataCell(Text('${cmd.delayMs}')),
                            DataCell(
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 10,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: cmd.isQuery
                                      ? Colors.purple.shade50
                                      : Colors.orange.shade50,
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Text(
                                  cmd.isQuery ? 'READ' : 'WRITE',
                                  style: TextStyle(
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                    color: cmd.isQuery
                                        ? Colors.purple.shade700
                                        : Colors.orange.shade700,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 32),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    child: const Text('Close Preview'),
                  ),
                  const SizedBox(width: 16),
                  ElevatedButton(
                    onPressed: () {
                      Navigator.pop(context);
                      _executeSequence();
                    },
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 32,
                        vertical: 16,
                      ),
                      backgroundColor: Colors.green.shade700,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    child: const Text('Confirm & Run Sequence'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildConsolePanel(ThemeData theme) {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFF0F172A), // Deep navy/black
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.2),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(20.0),
            child: Row(
              children: [
                const Icon(Icons.terminal, color: Colors.greenAccent, size: 20),
                const SizedBox(width: 12),
                Text(
                  'RESPONSE CONSOLE',
                  style: GoogleFonts.jetBrainsMono(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: Colors.white.withOpacity(0.7),
                    letterSpacing: 1.5,
                  ),
                ),
                const Spacer(),
                IconButton(
                  onPressed: _clearLogs,
                  icon: const Icon(
                    Icons.delete_sweep_outlined,
                    color: Colors.white54,
                    size: 20,
                  ),
                  tooltip: 'Clear Console',
                ),
              ],
            ),
          ),
          const Divider(color: Colors.white10, height: 1),
          Expanded(
            child: ListView.builder(
              controller: _logScrollController,
              padding: const EdgeInsets.all(20),
              itemCount: _logs.length,
              itemBuilder: (context, index) {
                final entry = _logs[index];
                return _buildLogEntry(entry);
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLogEntry(ScpiLogEntry entry) {
    Color typeColor = Colors.blueAccent;
    if (entry.type == 'RECV') typeColor = Colors.greenAccent;
    if (entry.type == 'ERROR') typeColor = Colors.redAccent;

    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0),
      child: RichText(
        text: TextSpan(
          style: GoogleFonts.jetBrainsMono(fontSize: 13, height: 1.5),
          children: [
            TextSpan(
              text: '[${DateFormat('HH:mm:ss').format(entry.timestamp)}] ',
              style: const TextStyle(color: Colors.white24),
            ),
            TextSpan(
              text: '${entry.type} ',
              style: TextStyle(color: typeColor, fontWeight: FontWeight.bold),
            ),
            TextSpan(
              text: entry.message,
              style: const TextStyle(color: Colors.white),
            ),
          ],
        ),
      ),
    );
  }

  // --- Helper Widgets ---

  Widget _buildPanelLabel(ThemeData theme, String label) {
    return Text(
      label,
      style: GoogleFonts.inter(
        fontSize: 12,
        fontWeight: FontWeight.w900,
        color: theme.colorScheme.primary,
        letterSpacing: 1.5,
      ),
    );
  }

  Widget _buildTextField(
    ThemeData theme, {
    required String label,
    required TextEditingController controller,
    required IconData icon,
    required String hint,
    ValueChanged<String>? onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label.toUpperCase(),
          style: GoogleFonts.inter(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: Colors.grey.shade500,
            letterSpacing: 1.2,
          ),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: controller,
          onChanged: onChanged,
          style: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w600),
          decoration: InputDecoration(
            hintText: hint,
            prefixIcon: Icon(icon, color: theme.colorScheme.primary, size: 20),
            filled: true,
            fillColor: Colors.grey.shade50,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 12,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildDropdownEffect<T>(
    ThemeData theme, {
    required String label,
    required T? value,
    required List<T> items,
    required IconData icon,
    required String Function(T) itemLabel,
    required ValueChanged<T?> onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label.toUpperCase(),
          style: GoogleFonts.inter(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: Colors.grey.shade500,
            letterSpacing: 1.2,
          ),
        ),
        const SizedBox(height: 8),
        DropdownButtonFormField<T>(
          value: value,
          onChanged: onChanged,
          decoration: InputDecoration(
            prefixIcon: Icon(icon, color: theme.colorScheme.primary, size: 20),
            filled: true,
            fillColor: Colors.grey.shade50,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 12,
            ),
          ),
          items: items.map((T item) {
            return DropdownMenuItem<T>(
              value: item,
              child: Text(
                itemLabel(item),
                style: GoogleFonts.inter(fontSize: 14),
              ),
            );
          }).toList(),
        ),
      ],
    );
  }

  Widget _buildTypeToggle(String label, bool isSelected, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFF1A237E) : Colors.grey.shade100,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 12,
            fontWeight: FontWeight.bold,
            color: isSelected ? Colors.white : Colors.grey.shade600,
          ),
        ),
      ),
    );
  }

  Widget _buildSearchableMnemonic(
    ThemeData theme,
    List<CommandDetails> commands,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'MNEMONIC',
          style: GoogleFonts.inter(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: Colors.grey.shade500,
            letterSpacing: 1.2,
          ),
        ),
        const SizedBox(height: 8),
        Autocomplete<CommandDetails>(
          key: _mnemonicFieldKey,
          displayStringForOption: (CommandDetails option) => option.mnemonic,
          optionsBuilder: (TextEditingValue textEditingValue) {
            if (textEditingValue.text == '') {
              return commands;
            }
            return commands.where((CommandDetails option) {
              return option.mnemonic.toLowerCase().contains(
                textEditingValue.text.toLowerCase(),
              );
            });
          },
          onSelected: _onMnemonicSelected,
          fieldViewBuilder: (context, controller, focusNode, onFieldSubmitted) {
            // No manual syncing needed as we reset via Key
            return TextField(
              controller: controller,
              focusNode: focusNode,
              onSubmitted: (value) => onFieldSubmitted(),
              style: GoogleFonts.inter(
                fontSize: 14,
                fontWeight: FontWeight.w600,
              ),
              decoration: InputDecoration(
                hintText: 'Search or select mnemonic',
                prefixIcon: Icon(
                  Icons.search_rounded,
                  color: theme.colorScheme.primary,
                  size: 20,
                ),
                filled: true,
                fillColor: Colors.grey.shade50,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(color: Colors.grey.shade200),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(color: Colors.grey.shade200),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 12,
                ),
              ),
            );
          },
          optionsViewBuilder: (context, onSelected, options) {
            return Align(
              alignment: Alignment.topLeft,
              child: Material(
                elevation: 4,
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  width: 300, // Fixed width for the suggestion box
                  constraints: const BoxConstraints(maxHeight: 300),
                  child: ListView.builder(
                    padding: EdgeInsets.zero,
                    shrinkWrap: true,
                    itemCount: options.length,
                    itemBuilder: (BuildContext context, int index) {
                      final CommandDetails option = options.elementAt(index);
                      return ListTile(
                        title: Text(
                          option.mnemonic,
                          style: GoogleFonts.inter(
                            fontSize: 13,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        subtitle: Text(
                          option.command,
                          style: GoogleFonts.jetBrainsMono(
                            fontSize: 11,
                            color: Colors.grey.shade500,
                          ),
                        ),
                        onTap: () => onSelected(option),
                      );
                    },
                  ),
                ),
              ),
            );
          },
        ),
      ],
    );
  }
}
