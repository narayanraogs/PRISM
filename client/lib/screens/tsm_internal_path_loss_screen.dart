import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';
import '../services/server_service.dart';

class TSMInternalPathLossScreen extends StatefulWidget {
  const TSMInternalPathLossScreen({super.key});

  @override
  State<TSMInternalPathLossScreen> createState() =>
      _TSMInternalPathLossScreenState();
}

class _TSMInternalPathLossScreenState extends State<TSMInternalPathLossScreen> {
  // State Variables
  String _selectedProfile = "";
  String _selectedChannel = "A";
  String _selectedInput = "";
  String? _selectedOutputPort;

  // Dependency State (from Server)
  bool _isPmReferenced = false;
  bool _isCableMeasured = false;
  List<String> _profiles = [];
  List<InternalLossEntry> _allPaths = [];
  List<String> _uniqueInputPorts = [];
  List<InternalLossEntry> _availableOutputPaths = [];

  bool _isMeasuring = false;
  bool _isLoading = true;
  String _measuringStatus = '';
  bool _showConnections = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchMetadata();
    });
  }

  void _fetchMetadata() {
    setState(() => _isLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    final metadata = serverService.status.bootstrapData?.tsmInternalLossData;

    if (metadata != null) {
      debugPrint('TSMInternalPathLossScreen: Using Bootstrapped Metadata');
      setState(() {
        _profiles = metadata.deviceProfiles;
        if (_profiles.isNotEmpty && _selectedProfile.isEmpty) {
          _selectedProfile = _profiles.first;
        }

        _isPmReferenced = metadata.measuredLoss.pm.measured;
        _isCableMeasured = metadata.measuredLoss.cable.measured;

        _allPaths = metadata.measuredLoss.paths;
        _uniqueInputPorts = _allPaths.map((e) => e.inputPort).toSet().toList()
          ..sort();

        if (_uniqueInputPorts.isNotEmpty && _selectedInput.isEmpty) {
          _selectedInput = _uniqueInputPorts.first;
          _updateAvailableOutputPaths();
        } else if (_selectedInput.isNotEmpty) {
          _updateAvailableOutputPaths();
        }

        _isLoading = false;
      });
    } else {
      debugPrint('TSMInternalPathLossScreen: Bootstrapped Metadata NOT FOUND');
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text("Failed to fetch metadata"),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  void _updateAvailableOutputPaths() {
    _availableOutputPaths = _allPaths
        .where((e) => e.inputPort == _selectedInput)
        .toList();
    if (_availableOutputPaths.isNotEmpty) {
      if (_selectedOutputPort == null ||
          !_availableOutputPaths.any(
            (e) => e.outputPort == _selectedOutputPort,
          )) {
        _selectedOutputPort = _availableOutputPaths.first.outputPort;
      }
    } else {
      _selectedOutputPort = null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    if (_isLoading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    return Scaffold(
      body: Column(
        children: [
          ScreenHeader(
            title: 'TSM Internal Path Loss',
            subtitle: 'Measure internal path loss for TSM',
            icon: Icons.settings_ethernet,
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton.filledTonal(
                  onPressed: _fetchMetadata,
                  icon: const Icon(Icons.refresh),
                  tooltip: 'Refresh Data',
                ),
                const SizedBox(width: 8),
                IconButton.filledTonal(
                  onPressed: () =>
                      setState(() => _showConnections = !_showConnections),
                  icon: Icon(_showConnections ? Icons.hub : Icons.hub_outlined),
                  tooltip: 'Show Path Diagrams',
                ),
              ],
            ),
          ),
          Expanded(
            child: Stack(
              children: [
                SingleChildScrollView(
                  padding: const EdgeInsets.all(24.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (_showConnections) _buildConnectionsOverlay(theme),
                      _buildTopStatusCards(theme),
                      const SizedBox(height: 24),
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(flex: 2, child: _buildConfigCard(theme)),
                          const SizedBox(width: 24),
                          Expanded(
                            flex: 3,
                            child: _buildLatestResultCard(theme),
                          ),
                        ],
                      ),
                      const SizedBox(height: 32),
                      _buildHistorySection(theme),
                    ],
                  ),
                ),
                if (_isMeasuring) _buildMeasuringOverlay(theme),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTopStatusCards(ThemeData theme) {
    return Row(
      children: [
        // 1. Power Meter Card
        _buildStatusCard(
          theme,
          title: 'Step 1: PM Offset',
          value: _isPmReferenced ? 'Measured' : 'Zeroing Required',
          icon: Icons.speed,
          color: _isPmReferenced ? Colors.green : Colors.orange,
          action: ElevatedButton.icon(
            onPressed: () => _startMeasurement('PM'),
            icon: Icon(
              _isPmReferenced ? Icons.check_circle : Icons.refresh,
              size: 16,
            ),
            label: Text(_isPmReferenced ? 'Zeroed' : 'Measure PM'),
            style: ElevatedButton.styleFrom(
              backgroundColor: _isPmReferenced
                  ? Colors.green
                  : theme.colorScheme.primary,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            ),
          ),
        ),
        const SizedBox(width: 16),

        // 2. Cable Loss Card
        _buildStatusCard(
          theme,
          title: 'Step 2: Cable Loss',
          value: _isCableMeasured
              ? 'Measured'
              : (_isPmReferenced ? 'Ready to Measure' : 'Waiting for PM'),
          icon: Icons.cable,
          color: _isCableMeasured
              ? Colors.green
              : (_isPmReferenced ? Colors.blue : Colors.grey),
          action: ElevatedButton.icon(
            onPressed: _isPmReferenced
                ? () => _startMeasurement('Cable Loss')
                : null,
            icon: Icon(
              _isCableMeasured ? Icons.check_circle : Icons.play_arrow,
              size: 16,
            ),
            label: Text(_isCableMeasured ? 'Cable OK' : 'Measure Cable'),
            style: ElevatedButton.styleFrom(
              backgroundColor: _isCableMeasured
                  ? Colors.green
                  : theme.colorScheme.primary,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildStatusCard(
    ThemeData theme, {
    required String title,
    required String value,
    required IconData icon,
    required Color color,
    Widget? action,
  }) {
    return Expanded(
      child: ContentCard(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Icon(icon, color: color, size: 24),
            ),
            const SizedBox(width: 20),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: Colors.grey.shade600,
                      fontWeight: FontWeight.w100,
                      letterSpacing: 1.1,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    value,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                ],
              ),
            ),
            if (action != null) action,
          ],
        ),
      ),
    );
  }

  Widget _buildConfigCard(ThemeData theme) {
    bool canMeasureTSM = _isPmReferenced && _isCableMeasured;

    return ContentCard(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Step 3: Path Selection',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 24),

          _buildFieldLabel('Device Profile'),
          DropdownButtonFormField<String>(
            value: _selectedProfile.isNotEmpty ? _selectedProfile : null,
            items: _profiles
                .map((e) => DropdownMenuItem(value: e, child: Text(e)))
                .toList(),
            onChanged: (val) => setState(() => _selectedProfile = val!),
            decoration: _inputDecoration('', Icons.settings),
          ),

          const SizedBox(height: 20),
          _buildFieldLabel('Power Meter Channel'),
          SizedBox(
            width: double.infinity,
            child: SegmentedButton<String>(
              segments: const [
                ButtonSegment(
                  value: 'A',
                  label: Text('Channel A'),
                  icon: Icon(Icons.settings_input_composite, size: 16),
                ),
                ButtonSegment(
                  value: 'B',
                  label: Text('Channel B'),
                  icon: Icon(Icons.settings_input_composite, size: 16),
                ),
              ],
              selected: {_selectedChannel},
              onSelectionChanged: (val) =>
                  setState(() => _selectedChannel = val.first),
              style: SegmentedButton.styleFrom(
                visualDensity: VisualDensity.comfortable,
                padding: const EdgeInsets.symmetric(vertical: 12),
              ),
            ),
          ),

          const SizedBox(height: 24),
          _buildFieldLabel('Step 3: Select Input Port'),
          Wrap(
            spacing: 12,
            runSpacing: 12,
            children: _uniqueInputPorts.map((portName) {
              final isSelected = _selectedInput == portName;
              final isCompleted = _allPaths
                  .where((e) => e.inputPort == portName)
                  .every((e) => e.measured);

              return ChoiceChip(
                label: Text(portName),
                selected: isSelected,
                avatar: isCompleted
                    ? Icon(
                        Icons.check_circle,
                        size: 16,
                        color: isSelected ? Colors.white : Colors.green,
                      )
                    : null,
                onSelected: (selected) {
                  if (selected) {
                    setState(() {
                      _selectedInput = portName;
                      _updateAvailableOutputPaths();
                    });
                  }
                },
                selectedColor: theme.colorScheme.primaryContainer,
                labelStyle: TextStyle(
                  color: isSelected
                      ? theme.colorScheme.onPrimaryContainer
                      : (isCompleted ? Colors.green.shade700 : Colors.black87),
                  fontWeight: isSelected || isCompleted
                      ? FontWeight.bold
                      : FontWeight.normal,
                ),
                backgroundColor: Colors.grey.shade50,
                showCheckmark: false,
                side: BorderSide(
                  color: isSelected
                      ? theme.colorScheme.primary
                      : (isCompleted
                            ? Colors.green.withOpacity(0.5)
                            : Colors.grey.shade300),
                  width: isSelected ? 2 : 1,
                ),
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 8,
                ),
              );
            }).toList(),
          ),

          const SizedBox(height: 32),
          _buildFieldLabel('Step 4: Select Output Port'),
          Wrap(
            spacing: 12,
            runSpacing: 12,
            children: _availableOutputPaths.map((path) {
              final isSelected = _selectedOutputPort == path.outputPort;
              final isMeasured = path.measured;

              return ChoiceChip(
                label: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(path.outputPort),
                    Text(
                      path.pathMnemonic,
                      style: TextStyle(
                        fontSize: 9,
                        color: isSelected
                            ? Colors.white.withOpacity(0.8)
                            : Colors.grey.shade600,
                      ),
                    ),
                  ],
                ),
                selected: isSelected,
                avatar: isMeasured
                    ? Icon(
                        Icons.check_circle,
                        size: 16,
                        color: isSelected ? Colors.white : Colors.green,
                      )
                    : null,
                onSelected: (selected) {
                  if (selected) {
                    setState(() => _selectedOutputPort = path.outputPort);
                  }
                },
                selectedColor: theme.colorScheme.primaryContainer,
                labelStyle: TextStyle(
                  color: isSelected
                      ? theme.colorScheme.onPrimaryContainer
                      : (isMeasured ? Colors.green.shade700 : Colors.black87),
                  fontWeight: isSelected || isMeasured
                      ? FontWeight.bold
                      : FontWeight.normal,
                ),
                backgroundColor: Colors.grey.shade50,
                showCheckmark: false,
                side: BorderSide(
                  color: isSelected
                      ? theme.colorScheme.primary
                      : (isMeasured
                            ? Colors.green.withOpacity(0.5)
                            : Colors.grey.shade300),
                  width: isSelected ? 2 : 1,
                ),
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 8,
                ),
              );
            }).toList(),
          ),

          const SizedBox(height: 40),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: canMeasureTSM ? () => _startMeasurement('Path') : null,
              icon: const Icon(Icons.play_arrow),
              label: const Text('Measure Internal Path'),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 20),
                backgroundColor: canMeasureTSM
                    ? theme.colorScheme.primary
                    : Colors.grey.shade300,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
            ),
          ),
          if (!canMeasureTSM)
            Padding(
              padding: const EdgeInsets.only(top: 12.0),
              child: Center(
                child: Text(
                  !_isPmReferenced
                      ? 'Complete PM Offset first'
                      : 'Complete Cable Loss measurement first',
                  style: TextStyle(
                    color: Colors.red.shade400,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildLatestResultCard(ThemeData theme) {
    final selectedPath = _availableOutputPaths
        .cast<InternalLossEntry?>()
        .firstWhere(
          (e) => e?.outputPort == _selectedOutputPort,
          orElse: () => null,
        );

    return Column(
      children: [
        ContentCard(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Current Path Status',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.blue.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Text(
                      'Calibration Mode',
                      style: TextStyle(
                        color: Colors.blue,
                        fontWeight: FontWeight.bold,
                        fontSize: 11,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                _selectedInput.isNotEmpty
                    ? '$_selectedInput → ${selectedPath?.outputPort ?? "..."}'
                    : "Select a path",
                style: GoogleFonts.robotoMono(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.primary,
                ),
              ),
              const Divider(height: 32),

              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  _buildMetric(
                    'Frequencies',
                    '${selectedPath?.frequencies.length ?? 0} Points',
                  ),
                  const Spacer(),
                ],
              ),

              const SizedBox(height: 24),
              ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: SizedBox(
                  width: double.infinity,
                  child: DataTable(
                    headingRowHeight: 40,
                    columnSpacing: 20,
                    horizontalMargin: 20,
                    headingRowColor: MaterialStateProperty.all(
                      Colors.grey.shade50,
                    ),
                    columns: const [
                      DataColumn(
                        label: Expanded(
                          child: Text(
                            'Freq (GHz)',
                            textAlign: TextAlign.center,
                          ),
                        ),
                      ),
                      DataColumn(
                        label: Expanded(
                          child: Text('Loss (dB)', textAlign: TextAlign.center),
                        ),
                      ),
                    ],
                    rows: (selectedPath?.frequencies ?? []).asMap().entries.map(
                      (entry) {
                        int idx = entry.key;
                        double freq = entry.value;
                        double loss = selectedPath!.losses[idx];
                        return DataRow(
                          cells: [
                            DataCell(
                              Center(
                                child: Text(
                                  freq.toStringAsFixed(2),
                                  style: GoogleFonts.robotoMono(),
                                ),
                              ),
                            ),
                            DataCell(
                              Center(
                                child: Text(
                                  loss.toStringAsFixed(2),
                                  style: GoogleFonts.robotoMono(
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        );
                      },
                    ).toList(),
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        _buildCharacterizationProgress(theme),
      ],
    );
  }

  Widget _buildCharacterizationProgress(ThemeData theme) {
    int total = _allPaths.length;
    int completed = _allPaths.where((e) => e.measured).length;
    double progress = total > 0 ? completed / total : 0;

    return ContentCard(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Global Characterization Progress',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          LinearProgressIndicator(
            value: progress,
            minHeight: 12,
            backgroundColor: Colors.grey.shade100,
            valueColor: AlwaysStoppedAnimation<Color>(
              theme.colorScheme.primary,
            ),
            borderRadius: BorderRadius.circular(6),
          ),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                '$completed of $total paths completed',
                style: TextStyle(color: Colors.grey.shade600, fontSize: 12),
              ),
              Text(
                '${(progress * 100).toStringAsFixed(0)}%',
                style: TextStyle(
                  color: theme.colorScheme.primary,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildHistorySection(ThemeData theme) {
    return ContentCard(
      padding: EdgeInsets.zero,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Full Port Mapping History',
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Row(
                  children: [
                    _buildIconButton(Icons.picture_as_pdf, 'PDF', Colors.red),
                    const SizedBox(width: 8),
                    _buildIconButton(Icons.table_chart, 'CSV', Colors.green),
                  ],
                ),
              ],
            ),
          ),
          SizedBox(
            width: double.infinity,
            child: DataTable(
              headingRowColor: MaterialStateProperty.all(Colors.grey.shade50),
              horizontalMargin: 24,
              columns: const [
                DataColumn(label: Text('Input Port')),
                DataColumn(label: Text('Output Port')),
                DataColumn(label: Text('Path')),
                DataColumn(label: Text('Freq (GHz)')),
                DataColumn(label: Text('Loss (dB)')),
              ],
              rows: _allPaths.expand((path) {
                if (path.frequencies.isEmpty) {
                  return [
                    DataRow(
                      cells: [
                        DataCell(Text(path.inputPort)),
                        DataCell(Text(path.outputPort)),
                        DataCell(Text(path.pathMnemonic)),
                        const DataCell(Text('N/A')),
                        const DataCell(Text('N/A')),
                      ],
                    ),
                  ];
                }
                return path.frequencies.asMap().entries.map((entry) {
                  int idx = entry.key;
                  double freq = entry.value;
                  double loss = path.losses[idx];
                  return DataRow(
                    cells: [
                      DataCell(Text(path.inputPort)),
                      DataCell(Text(path.outputPort)),
                      DataCell(Text(path.pathMnemonic)),
                      DataCell(
                        Text(
                          freq.toStringAsFixed(2),
                          style: GoogleFonts.robotoMono(),
                        ),
                      ),
                      DataCell(
                        Text(
                          loss.toStringAsFixed(2),
                          style: GoogleFonts.robotoMono(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  );
                });
              }).toList(),
            ),
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildMetric(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 11,
            color: Colors.grey,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          value,
          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
        ),
      ],
    );
  }

  Widget _buildFieldLabel(String label) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0, left: 4),
      child: Text(
        label,
        style: const TextStyle(
          fontSize: 13,
          fontWeight: FontWeight.w600,
          color: Colors.grey,
        ),
      ),
    );
  }

  InputDecoration _inputDecoration(String hint, IconData icon) {
    return InputDecoration(
      hintText: hint,
      prefixIcon: Icon(icon, size: 20),
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
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: BorderSide(
          color: Theme.of(context).colorScheme.primary,
          width: 2,
        ),
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
    );
  }

  Widget _buildIconButton(IconData icon, String label, Color color) {
    return ElevatedButton.icon(
      onPressed: () {},
      icon: Icon(icon, size: 16),
      label: Text(label),
      style: ElevatedButton.styleFrom(
        backgroundColor: color.withOpacity(0.05),
        foregroundColor: color,
        elevation: 0,
        padding: const EdgeInsets.symmetric(horizontal: 12),
      ),
    );
  }

  Widget _buildConnectionsOverlay(ThemeData theme) {
    return ContentCard(
      color: theme.colorScheme.primaryContainer.withOpacity(0.4),
      margin: const EdgeInsets.only(bottom: 24),
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'TSM Path Selection Guide',
                style: GoogleFonts.outfit(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.onPrimaryContainer,
                ),
              ),
              IconButton(
                onPressed: () => setState(() => _showConnections = false),
                icon: const Icon(Icons.close),
              ),
            ],
          ),
          const SizedBox(height: 16),
          const Text(
            'Select the appropriate input and output ports based on the test configuration. Completed paths are marked with a green checkmark.',
          ),
        ],
      ),
    );
  }

  Widget _buildMeasuringOverlay(ThemeData theme) {
    return Container(
      color: Colors.black.withOpacity(0.3),
      child: Center(
        child: ContentCard(
          padding: const EdgeInsets.all(32),
          width: 400,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const CircularProgressIndicator(),
              const SizedBox(height: 24),
              Text(
                _measuringStatus,
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 20),
              ElevatedButton(
                onPressed: () {
                  Provider.of<ServerService>(
                    context,
                    listen: false,
                  ).abortTSMInternalLossMeasurement();
                  setState(() => _isMeasuring = false);
                },
                child: const Text('Abort Measurement'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _startMeasurement(String mode) {
    if (_isMeasuring) return;

    final serverService = Provider.of<ServerService>(context, listen: false);

    final request = InternalLossMeasurementRequest(
      deviceProfile: _selectedProfile,
      pmChannel: _selectedChannel,
      mode: mode,
      inputPort: mode == "Path" ? _selectedInput : "",
      outputPort: mode == "Path" ? (_selectedOutputPort ?? "") : "",
    );

    setState(() {
      _isMeasuring = true;
      _measuringStatus = "Initializing $mode...";
    });

    serverService
        .streamTSMInternalLossAction(request)
        .listen(
          (event) {
            if (!mounted) return;
            if (event is MeasurementStatus) {
              setState(() {
                _measuringStatus = event.message;
                if (event.completed) {
                  _isMeasuring = false;
                  if (event.error) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text(event.message),
                        backgroundColor: Colors.red,
                      ),
                    );
                  }
                }
              });
            } else if (event is TSMInternalLossMeasured) {
              setState(() {
                _isPmReferenced = event.pm.measured;
                _isCableMeasured = event.cable.measured;
                _allPaths = event.paths;
                _uniqueInputPorts =
                    _allPaths.map((e) => e.inputPort).toSet().toList()..sort();
                _updateAvailableOutputPaths();
                _isMeasuring = false;
              });
            }
          },
          onError: (error) {
            if (!mounted) return;
            setState(() {
              _isMeasuring = false;
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text("Error: $error"),
                  backgroundColor: Colors.red,
                ),
              );
            });
          },
          onDone: () {
            if (mounted) {
              setState(() => _isMeasuring = false);
            }
          },
        );
  }
}
