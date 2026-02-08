import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:prism_client/services/server_service.dart';

class CableLossScreen extends StatefulWidget {
  const CableLossScreen({super.key});

  @override
  State<CableLossScreen> createState() => _CableLossScreenState();
}

class _CableLossScreenState extends State<CableLossScreen> {
  // Form Controllers
  final TextEditingController _cableNameController = TextEditingController();
  final TextEditingController _lengthController = TextEditingController();

  // State variables
  String _selectedDeviceProfile = 'Laboratory PM';
  String _pmChannel = 'A';
  bool _isPmReferenced = false;
  bool _isMeasuring = false;
  bool _isLoading = true;
  String _measuringStatus = '';
  bool _showConnections = false;

  // Mock Data
  List<String> _cableSuggestions = [
    'RF-CABLE-01',
    'TEST-CABLE-A',
    'CAL-KIT-02',
  ];
  List<String> _deviceProfiles = [
    'Laboratory PM',
    'Production PM',
    'Service PM',
  ];
  List<FrequencyDefinition> _availableFrequencies = [];
  Set<String> _selectedFrequencyNames = {};
  List<CableLossRecord> _history = [];
  CableLossRecord? _selectedPlotRecord;

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  void _loadInitialData() {
    setState(() => _isLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    final metadata = serverService.status.bootstrapData?.cableLossData;

    try {
      if (metadata != null) {
        debugPrint('CableLossScreen: Using Bootstrapped Metadata');
        setState(() {
          _cableSuggestions = metadata.existingCables;
          _deviceProfiles = metadata.deviceProfiles;
          if (_deviceProfiles.isNotEmpty &&
              !_deviceProfiles.contains(_selectedDeviceProfile)) {
            _selectedDeviceProfile = _deviceProfiles.first;
          }
          _isPmReferenced = metadata.isPmZeroed;

          // Parse frequencies: "Name;Value"
          _availableFrequencies = metadata.frequencies.map((f) {
            final parts = f.split(';');
            final name = parts[0];
            final value = parts.length > 1
                ? double.tryParse(parts[1]) ?? 0.0
                : 0.0;
            return FrequencyDefinition(name: name, value: value);
          }).toList();

          _selectedFrequencyNames = _availableFrequencies
              .map((f) => f.name)
              .toSet();
        });
      } else {
        debugPrint('CableLossScreen: Bootstrapped Metadata NOT FOUND');
        // Fallback to mock frequencies if metadata fails
        _initMockFrequencies();
      }

      // Load history
      _loadMeasuredHistory();
    } catch (e) {
      debugPrint('Error loading initial data: $e');
      _initMockFrequencies(); // Fallback
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _loadMeasuredHistory() async {
    final serverService = Provider.of<ServerService>(context, listen: false);
    final historyResp = await serverService.fetchCableMeasuredDetails();
    if (historyResp != null && historyResp.ok) {
      setState(() {
        _history = historyResp.history;
        if (_history.isNotEmpty) {
          _selectedPlotRecord = _history.last;
        }
      });
    } else {
      await _loadMockHistory();
    }
  }

  void _initMockFrequencies() {
    setState(() {
      _availableFrequencies = [
        FrequencyDefinition(name: 'L-Band', value: 1.5),
        FrequencyDefinition(name: 'S-Band', value: 2.2),
        FrequencyDefinition(name: 'C-Band', value: 4.0),
        FrequencyDefinition(name: 'X-Band', value: 8.0),
        FrequencyDefinition(name: 'Ku1', value: 12.0),
        FrequencyDefinition(name: 'Ku2', value: 14.0),
        FrequencyDefinition(name: 'Ka1', value: 18.0),
        FrequencyDefinition(name: 'Ka2', value: 26.0),
      ];
      _selectedFrequencyNames = _availableFrequencies
          .map((f) => f.name)
          .toSet();
    });
  }

  Future<void> _loadMockHistory() async {
    // Current mock history data...
    _history = [
      CableLossRecord(
        slNo: 1,
        cableName: 'RF-CABLE-01',
        length: 5,
        date: '2026-02-01',
        time: '10:30:00',
        measurements: [
          MeasurementPoint(frequency: 1.5, loss: -0.2),
          MeasurementPoint(frequency: 4.0, loss: -0.8),
          MeasurementPoint(frequency: 12.0, loss: -1.8),
          MeasurementPoint(frequency: 18.0, loss: -2.5),
        ],
      ),
    ];
  }

  void _showAppNotification({
    required String title,
    required String message,
    required NotificationType type,
  }) {
    if (!mounted) return;

    // 1. Add to Persistence Notification Center
    context.read<NotificationService>().addNotification(
      title: title,
      message: message,
      type: type,
    );

    // 2. Show Toast (SnackBar)
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Icon(
              type == NotificationType.success
                  ? Icons.check_circle
                  : type == NotificationType.error
                  ? Icons.error
                  : Icons.info,
              color: Colors.white,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                  Text(message, style: const TextStyle(fontSize: 12)),
                ],
              ),
            ),
          ],
        ),
        backgroundColor: type == NotificationType.success
            ? Colors.green.shade600
            : type == NotificationType.error
            ? Colors.red.shade600
            : type == NotificationType.warning
            ? Colors.orange.shade600
            : Colors.blue.shade600,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(12),
        duration: const Duration(seconds: 4),
      ),
    );
  }

  void _performAction(String action) async {
    final serverService = Provider.of<ServerService>(context, listen: false);

    if (action == 'pmreference') {
      setState(() {
        _isMeasuring = true;
        _measuringStatus = 'Zeroing Power Meter...';
      });

      final request = {
        'action': 'pmreference',
        'deviceProfile': _selectedDeviceProfile,
        'channel': _pmChannel,
        'cableName': 'PM',
        'cableLength': 0.0,
        'selectedFrequencies': _availableFrequencies
            .map((f) => f.name)
            .toList(),
      };

      serverService
          .streamCableLossAction(request)
          .listen(
            (status) {
              setState(() {
                _measuringStatus = status.message;
              });

              if (status.completed) {
                setState(() {
                  _isMeasuring = false;
                  _isPmReferenced = status.success;
                  _measuringStatus = '';
                });

                if (status.success) {
                  _showAppNotification(
                    title: 'Success',
                    message: 'Power Meter Zeroed Successfully',
                    type: NotificationType.success,
                  );
                } else {
                  _showAppNotification(
                    title: 'Error',
                    message: status.message,
                    type: NotificationType.error,
                  );
                }
              }
            },
            onError: (e) {
              setState(() {
                _isMeasuring = false;
                _measuringStatus = '';
              });
              _showAppNotification(
                title: 'Connection Error',
                message: e.toString(),
                type: NotificationType.error,
              );
            },
          );
      return;
    }

    // Validation for measurement
    if (_cableNameController.text.trim().isEmpty) {
      _showAppNotification(
        title: 'Validation Error',
        message: 'Please enter a Cable Name',
        type: NotificationType.error,
      );
      return;
    }
    if (_lengthController.text.trim().isEmpty) {
      _showAppNotification(
        title: 'Validation Error',
        message: 'Please enter Cable Length',
        type: NotificationType.error,
      );
      return;
    }

    final length = double.tryParse(_lengthController.text) ?? 0.0;
    final selectedNames = _availableFrequencies
        .where((f) => _selectedFrequencyNames.contains(f.name))
        .map((f) => f.name)
        .toList();

    if (selectedNames.isEmpty) {
      _showAppNotification(
        title: 'Validation Error',
        message: 'Please select at least one frequency',
        type: NotificationType.error,
      );
      return;
    }

    setState(() {
      _isMeasuring = true;
      _measuringStatus = 'Initializing Measurement...';
    });

    final request = {
      'action': 'cableloss',
      'deviceProfile': _selectedDeviceProfile,
      'channel': _pmChannel,
      'cableName': _cableNameController.text.trim(),
      'cableLength': length,
      'selectedFrequencies': selectedNames,
    };

    serverService
        .streamCableLossAction(request)
        .listen(
          (status) {
            setState(() {
              _measuringStatus = status.message;
            });

            if (status.completed) {
              setState(() {
                _isMeasuring = false;
                _measuringStatus = '';
              });

              if (status.success) {
                _showAppNotification(
                  title: 'Measurement Complete',
                  message:
                      'Cable Loss record saved for ${_cableNameController.text}',
                  type: NotificationType.success,
                );
                // Refresh history
                _loadInitialData();
              } else {
                _showAppNotification(
                  title: 'Measurement Failed',
                  message: status.message,
                  type: NotificationType.error,
                );
              }
            }
          },
          onError: (e) {
            setState(() {
              _isMeasuring = false;
              _measuringStatus = '';
            });
            _showAppNotification(
              title: 'Connection Error',
              message: e.toString(),
              type: NotificationType.error,
            );
          },
        );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: Stack(
        children: [
          _isLoading
              ? const Center(child: CircularProgressIndicator())
              : Container(
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        theme.colorScheme.background,
                        theme.colorScheme.primary.withOpacity(0.02),
                        theme.colorScheme.background,
                      ],
                    ),
                  ),
                  child: CustomScrollView(
                    slivers: [
                      _buildAppBar(theme),
                      SliverToBoxAdapter(
                        child: Padding(
                          padding: const EdgeInsets.all(24.0),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              if (_showConnections)
                                _buildConnectionsOverlay(theme),
                              _buildTopStatusCards(theme),
                              const SizedBox(height: 24),
                              Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Expanded(
                                    flex: 2,
                                    child: _buildConfigCard(theme),
                                  ),
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
                      ),
                    ],
                  ),
                ),
          if (_isMeasuring) _buildMeasuringOverlay(theme),
        ],
      ),
    );
  }

  Widget _buildAppBar(ThemeData theme) {
    return SliverAppBar(
      expandedHeight: 120.0,
      floating: false,
      pinned: true,
      elevation: 0,
      backgroundColor: Colors.transparent,
      flexibleSpace: FlexibleSpaceBar(
        titlePadding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
        title: Text(
          'Cable Loss Measurement',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.bold,
            color: theme.colorScheme.primary,
            fontSize: 20,
          ),
        ),
        background: Container(
          decoration: BoxDecoration(
            color: theme.colorScheme.surface.withOpacity(0.8),
          ),
        ),
      ),
      actions: [
        IconButton.filledTonal(
          onPressed: () => setState(() => _showConnections = !_showConnections),
          icon: Icon(_showConnections ? Icons.hub : Icons.hub_outlined),
          tooltip: 'Show Connection Diagrams',
        ),
        const SizedBox(width: 16),
      ],
    );
  }

  Widget _buildConnectionsOverlay(ThemeData theme) {
    return Container(
      margin: const EdgeInsets.only(bottom: 24),
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: theme.colorScheme.primaryContainer.withOpacity(0.4),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: theme.colorScheme.primary.withOpacity(0.2)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Reference Connection Diagrams',
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
          const SizedBox(height: 20),
          Row(
            children: [
              _buildDiagramCard(
                theme,
                '1. PM Zero Connection',
                'Connect Power Sensor directly to PM Zero output for reference.',
                Icons.shutter_speed,
              ),
              const SizedBox(width: 20),
              _buildDiagramCard(
                theme,
                '2. Cable Measurement',
                'Insert the cable under test between the source and the sensor.',
                Icons.settings_input_antenna,
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDiagramCard(
    ThemeData theme,
    String title,
    String desc,
    IconData icon,
  ) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 18, color: theme.colorScheme.primary),
                const SizedBox(width: 10),
                Text(
                  title,
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Container(
              height: 140,
              width: double.infinity,
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.grey.shade100),
              ),
              child: const Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.image_outlined, size: 40, color: Colors.grey),
                  SizedBox(height: 8),
                  Text(
                    '[ Placeholder Image ]',
                    style: TextStyle(color: Colors.grey, fontSize: 11),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            Text(
              desc,
              style: TextStyle(
                fontSize: 12,
                color: Colors.grey.shade600,
                height: 1.4,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTopStatusCards(ThemeData theme) {
    return Row(
      children: [
        _buildStatusCard(
          theme,
          title: 'Power Meter Status',
          value: _isPmReferenced ? 'Referenced' : 'Zeroing Required',
          icon: Icons.speed,
          color: _isPmReferenced ? Colors.green : Colors.orange,
          action: ElevatedButton.icon(
            onPressed: _isMeasuring
                ? null
                : () => _performAction('pmreference'),
            icon: const Icon(Icons.refresh, size: 16),
            label: const Text('Zero PM'),
            style: ElevatedButton.styleFrom(
              backgroundColor: theme.colorScheme.primary,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            ),
          ),
        ),
        const SizedBox(width: 16),
        _buildStatusCard(
          theme,
          title: 'Selected Instrument',
          value: _selectedDeviceProfile,
          icon: Icons.settings_input_antenna,
          color: theme.colorScheme.primary,
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
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            BoxShadow(
              color: theme.colorScheme.primary.withOpacity(0.05),
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
          ],
          border: Border.all(color: color.withOpacity(0.1)),
        ),
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
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    value,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                      fontSize: 18,
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
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 20,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Measurement Configuration',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 24),

          _buildFieldLabel('Cable Name'),
          Autocomplete<String>(
            optionsBuilder: (TextEditingValue textEditingValue) {
              if (textEditingValue.text == '') {
                return const Iterable<String>.empty();
              }
              return _cableSuggestions.where((String option) {
                return option.toLowerCase().contains(
                  textEditingValue.text.toLowerCase(),
                );
              });
            },
            onSelected: (String selection) {
              setState(() {
                _cableNameController.text = selection;
              });
            },
            fieldViewBuilder:
                (context, controller, focusNode, onFieldSubmitted) {
                  return TextField(
                    controller: controller,
                    focusNode: focusNode,
                    decoration: _inputDecoration(
                      'e.g. RF-CABLE-01',
                      Icons.cable,
                    ),
                    onChanged: (val) => _cableNameController.text = val,
                  );
                },
          ),

          const SizedBox(height: 20),

          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildFieldLabel('Length (m)'),
                    TextField(
                      controller: _lengthController,
                      keyboardType: TextInputType.number,
                      decoration: _inputDecoration('5', Icons.straighten),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildFieldLabel('Device Profile'),
                    DropdownButtonFormField<String>(
                      value: _selectedDeviceProfile,
                      items: _deviceProfiles
                          .map(
                            (e) => DropdownMenuItem(value: e, child: Text(e)),
                          )
                          .toList(),
                      onChanged: (val) =>
                          setState(() => _selectedDeviceProfile = val!),
                      decoration: _inputDecoration(
                        '',
                        Icons.radio_button_checked,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),

          const SizedBox(height: 20),

          _buildFieldLabel('Power Meter Channel'),
          SizedBox(
            width: double.infinity,
            child: SegmentedButton<String>(
              segments: const [
                ButtonSegment(
                  value: 'A',
                  label: Text('CH-A'),
                  icon: Icon(Icons.settings_input_composite, size: 16),
                ),
                ButtonSegment(
                  value: 'B',
                  label: Text('CH-B'),
                  icon: Icon(Icons.settings_input_composite, size: 16),
                ),
              ],
              selected: {_pmChannel},
              onSelectionChanged: (Set<String> newSelection) {
                setState(() {
                  _pmChannel = newSelection.first;
                });
              },
            ),
          ),

          const SizedBox(height: 20),
          _buildFieldLabel('Frequencies to Measure (All selected by default)'),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _availableFrequencies.map((f) {
              final isSelected = _selectedFrequencyNames.contains(f.name);
              return FilterChip(
                label: Text('${f.name}: ${f.value.toStringAsFixed(1)}M'),
                selected: isSelected,
                onSelected: (selected) {
                  setState(() {
                    if (selected) {
                      _selectedFrequencyNames.add(f.name);
                    } else {
                      _selectedFrequencyNames.remove(f.name);
                    }
                  });
                },
                selectedColor: theme.colorScheme.primary.withOpacity(0.2),
                checkmarkColor: theme.colorScheme.primary,
                labelStyle: TextStyle(
                  color: isSelected
                      ? theme.colorScheme.primary
                      : Colors.black87,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  fontSize: 11,
                ),
              );
            }).toList(),
          ),

          const SizedBox(height: 32),

          Row(
            children: [
              Expanded(
                child: ElevatedButton.icon(
                  onPressed: _isMeasuring || !_isPmReferenced
                      ? null
                      : () => _performAction('measure'),
                  icon: const Icon(Icons.play_arrow),
                  label: const Text('Start Measurement'),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 20),
                    backgroundColor: theme.colorScheme.primary,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                  ),
                ),
              ),
            ],
          ),
          if (!_isPmReferenced)
            Padding(
              padding: const EdgeInsets.only(top: 12.0),
              child: Center(
                child: Text(
                  'Please zero the Power Meter before starting',
                  style: TextStyle(color: Colors.red.shade400, fontSize: 12),
                ),
              ),
            ),
        ],
      ),
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

  Widget _buildLatestResultCard(ThemeData theme) {
    if (_history.isEmpty || _selectedPlotRecord == null) {
      return Container(
        height: 480,
        decoration: BoxDecoration(
          color: theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(24),
        ),
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.query_stats, size: 64, color: Colors.grey.shade200),
              const SizedBox(height: 16),
              Text(
                'No measurements yet',
                style: TextStyle(color: Colors.grey.shade400),
              ),
            ],
          ),
        ),
      );
    }

    final plotRecord = _selectedPlotRecord!;
    final spots = List.generate(plotRecord.measurements.length, (i) {
      return FlSpot(i.toDouble(), plotRecord.measurements[i].loss);
    });

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 20,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Latest Measurement',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    'Cable: ${plotRecord.cableName} (${plotRecord.length}m)',
                    style: TextStyle(color: Colors.grey.shade600, fontSize: 13),
                  ),
                ],
              ),
              IconButton(
                onPressed: () {},
                icon: const Icon(Icons.download),
                tooltip: 'Export CSV',
              ),
            ],
          ),
          const SizedBox(height: 32),
          SizedBox(
            height: 300,
            child: LineChart(
              LineChartData(
                minX: 0,
                maxX: (plotRecord.measurements.length - 1).toDouble(),
                gridData: FlGridData(
                  show: true,
                  drawHorizontalLine: true,
                  horizontalInterval: 0.5,
                  getDrawingHorizontalLine: (value) => FlLine(
                    color: Colors.grey.withOpacity(0.1),
                    strokeWidth: 1,
                  ),
                ),
                titlesData: FlTitlesData(
                  show: true,
                  bottomTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize: 60,
                      interval: 1,
                      getTitlesWidget: (value, meta) {
                        final index = value.toInt();
                        if (index < 0 ||
                            index >= plotRecord.measurements.length) {
                          return const SizedBox();
                        }
                        final freq = plotRecord.measurements[index].frequency;
                        return Padding(
                          padding: const EdgeInsets.only(top: 8.0),
                          child: RotatedBox(
                            quarterTurns: 1,
                            child: Text(
                              '${freq.toStringAsFixed(1)}',
                              style: const TextStyle(
                                fontSize: 10,
                                color: Colors.grey,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                  leftTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize: 40,
                      getTitlesWidget: (value, meta) => Text(
                        '${value.toStringAsFixed(1)}',
                        style: const TextStyle(
                          fontSize: 10,
                          color: Colors.grey,
                        ),
                      ),
                    ),
                  ),
                  topTitles: const AxisTitles(
                    sideTitles: SideTitles(showTitles: false),
                  ),
                  rightTitles: const AxisTitles(
                    sideTitles: SideTitles(showTitles: false),
                  ),
                ),
                borderData: FlBorderData(show: false),
                lineBarsData: [
                  LineChartBarData(
                    spots: spots,
                    isCurved: false,
                    color: theme.colorScheme.primary,
                    barWidth: 3,
                    isStrokeCapRound: true,
                    dotData: const FlDotData(show: true),
                    belowBarData: BarAreaData(
                      show: true,
                      color: theme.colorScheme.primary.withOpacity(0.05),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          _buildLegend(theme),
        ],
      ),
    );
  }

  Widget _buildLegend(ThemeData theme) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: theme.colorScheme.primary,
            shape: BoxShape.circle,
          ),
        ),
        const SizedBox(width: 8),
        const Text(
          'Loss (dB)',
          style: TextStyle(
            fontSize: 12,
            color: Colors.grey,
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
    );
  }

  Widget _buildHistorySection(ThemeData theme) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(24),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Measurement History',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              TextButton.icon(
                onPressed: () {},
                icon: const Icon(Icons.file_download),
                label: const Text('Export All'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          _buildHistoryTable(theme),
        ],
      ),
    );
  }

  Widget _buildHistoryTable(ThemeData theme) {
    if (_history.isEmpty) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 40),
        child: Center(child: Text('No historical data available')),
      );
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        return Theme(
          data: theme.copyWith(
            dataTableTheme: DataTableThemeData(
              headingRowColor: WidgetStateProperty.all(
                theme.colorScheme.primary.withOpacity(0.02),
              ),
              horizontalMargin: 24,
              columnSpacing: 24,
            ),
          ),
          child: Container(
            decoration: BoxDecoration(
              border: Border.all(color: Colors.grey.withOpacity(0.1)),
              borderRadius: BorderRadius.circular(12),
            ),
            clipBehavior: Clip.antiAlias,
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: ConstrainedBox(
                constraints: BoxConstraints(minWidth: constraints.maxWidth),
                child: DataTable(
                  headingRowHeight: 64,
                  dataRowMaxHeight: 56,
                  columns: [
                    DataColumn(
                      label: Text('Sl.\nNo', style: _tableHeaderStyle()),
                    ),
                    DataColumn(
                      label: Text('Cable Name', style: _tableHeaderStyle()),
                    ),
                    DataColumn(
                      label: Text('Len\n(m)', style: _tableHeaderStyle()),
                    ),
                    DataColumn(
                      label: Text('Date/Time', style: _tableHeaderStyle()),
                    ),
                    ..._availableFrequencies.map(
                      (f) => DataColumn(
                        label: Text(
                          '${f.name}\n${f.value.toStringAsFixed(1)}',
                          style: _tableHeaderStyle(),
                          textAlign: TextAlign.center,
                        ),
                        numeric: true,
                      ),
                    ),
                    DataColumn(
                      label: Text('Action', style: _tableHeaderStyle()),
                    ),
                  ],
                  rows: _history.reversed.map((record) {
                    return DataRow(
                      cells: [
                        DataCell(Text('${record.slNo}')),
                        DataCell(
                          Text(
                            record.cableName,
                            style: const TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                            ),
                          ),
                        ),
                        DataCell(Text('${record.length}')),
                        DataCell(
                          Text(
                            '${record.date}\n${record.time}',
                            style: const TextStyle(
                              fontSize: 10,
                              color: Colors.grey,
                            ),
                          ),
                        ),
                        ..._availableFrequencies.map((f) {
                          final measurement = record.measurements.firstWhere(
                            (m) => m.frequency == f.value,
                            orElse: () =>
                                MeasurementPoint(frequency: f.value, loss: 0),
                          );
                          bool hasData = measurement.loss != 0;
                          return DataCell(
                            Text(
                              hasData
                                  ? measurement.loss.toStringAsFixed(2)
                                  : '-',
                              style: TextStyle(
                                fontSize: 12,
                                color: hasData
                                    ? theme.colorScheme.onSurface
                                    : Colors.grey.shade300,
                                fontWeight: hasData
                                    ? FontWeight.w500
                                    : FontWeight.normal,
                              ),
                            ),
                          );
                        }),
                        DataCell(
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              IconButton(
                                onPressed: () {
                                  setState(() {
                                    _selectedPlotRecord = record;
                                  });
                                  _showAppNotification(
                                    title: 'Chart Updated',
                                    message:
                                        'Plotting data for ${record.cableName}',
                                    type: NotificationType.info,
                                  );
                                },
                                icon: Icon(
                                  Icons.show_chart,
                                  size: 18,
                                  color:
                                      _selectedPlotRecord?.slNo == record.slNo
                                      ? theme.colorScheme.primary
                                      : null,
                                ),
                                tooltip: 'Plot measurement',
                              ),
                              IconButton(
                                onPressed: () {},
                                icon: const Icon(Icons.download, size: 18),
                                tooltip: 'Export CSV',
                              ),
                            ],
                          ),
                        ),
                      ],
                    );
                  }).toList(),
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  TextStyle _tableHeaderStyle() {
    return const TextStyle(
      fontWeight: FontWeight.bold,
      color: Colors.grey,
      fontSize: 12,
    );
  }

  Widget _buildMeasuringOverlay(ThemeData theme) {
    return Container(
      color: Colors.black.withOpacity(0.3),
      child: Center(
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 24),
          decoration: BoxDecoration(
            color: theme.colorScheme.surface,
            borderRadius: BorderRadius.circular(24),
            boxShadow: [
              BoxShadow(color: Colors.black.withOpacity(0.1), blurRadius: 20),
            ],
          ),
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
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () {
                  context.read<ServerService>().abortCableLossMeasurement();
                  setState(() {
                    _isMeasuring = false;
                    _measuringStatus = '';
                  });
                  _showAppNotification(
                    title: 'Measurement Aborted',
                    message: 'The measurement process was stopped by user.',
                    type: NotificationType.warning,
                  );
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.red.shade100,
                  foregroundColor: Colors.red,
                  elevation: 0,
                ),
                child: const Text('Abort Measurement'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class FrequencyDefinition {
  final String name;
  final double value;

  FrequencyDefinition({required this.name, required this.value});
}
