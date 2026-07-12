import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/instrument_connection_diagram.dart';
import 'dart:convert';
import 'dart:js_interop';
import 'package:web/web.dart' as web;

class CableLossScreen extends StatefulWidget {
  final bool isActive;
  const CableLossScreen({super.key, this.isActive = true});

  @override
  State<CableLossScreen> createState() => _CableLossScreenState();
}

class _CableLossScreenState extends State<CableLossScreen> {
  @override
  void didUpdateWidget(CableLossScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.isActive && !widget.isActive) {
      final serverService = Provider.of<ServerService>(context, listen: false);
      serverService.closeCableLoss();
    }
  }

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
  List<String> _measuringLogs = [];
  bool _showConnections = false;
  bool _isHelpOpen = false;

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
  final Set<int> _selectedSlNos = {};
  String _searchCableName = '';
  String _searchCableLength = '';
  final List<Color> _plotColors = [
    const Color(0xFF2196F3), // Blue
    const Color(0xFFF44336), // Red
    const Color(0xFF4CAF50), // Green
    const Color(0xFFFF9800), // Orange
    const Color(0xFF9C27B0), // Purple
    const Color(0xFF00BCD4), // Cyan
    const Color(0xFFE91E63), // Pink
    const Color(0xFF795548), // Brown
  ];

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
        if (_history.isNotEmpty && _selectedSlNos.isEmpty) {
          _selectedSlNos.add(_history.last.slNo);
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
        _measuringLogs = [
          '[${DateTime.now().toString().split(' ')[1].split('.')[0]}] Zeroing Power Meter...',
        ];
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
                _measuringLogs.insert(
                  0,
                  "[${DateTime.now().toString().split(' ')[1].split('.')[0]}] ${status.message}",
                );
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
      _measuringLogs = [
        '[${DateTime.now().toString().split(' ')[1].split('.')[0]}] Initializing Measurement...',
      ];
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
              _measuringLogs.insert(
                0,
                "[${DateTime.now().toString().split(' ')[1].split('.')[0]}] ${status.message}",
              );
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
  void dispose() {
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.closeCableLoss();
    _cableNameController.dispose();
    _lengthController.dispose();
    super.dispose();
  }

  void _exportToCSV({CableLossRecord? record}) {
    final historyToExport = record != null ? [record] : _history;
    if (historyToExport.isEmpty) {
      _showAppNotification(
        title: 'Export Failed',
        message: 'No data available to export',
        type: NotificationType.warning,
      );
      return;
    }

    // Header: Sl. No, Cable Name, Length (m), Date, Time, [Frequencies...]
    List<String> header = ['Sl. No', 'Cable Name', 'Length (m)', 'Date', 'Time'];
    for (var f in _availableFrequencies) {
      header.add('${f.name} (${f.value} MHz)');
    }

    String csv = '${header.join(',')}\n';

    for (var rec in historyToExport) {
      List<String> row = [
        rec.slNo.toString(),
        '"${rec.cableName}"',
        rec.length.toString(),
        rec.date,
        rec.time,
      ];

      for (var f in _availableFrequencies) {
        final m = rec.measurements.firstWhere(
          (m) => m.frequency == f.value,
          orElse: () => MeasurementPoint(frequency: f.value, loss: 0),
        );
        row.add(m.loss != 0 ? m.loss.toStringAsFixed(4) : '-');
      }
      csv += '${row.join(',')}\n';
    }

    final bytes = utf8.encode(csv);
    final blob = web.Blob(
      [bytes.toJS].toJS,
      web.BlobPropertyBag(type: 'text/csv'),
    );
    final url = web.URL.createObjectURL(blob);
    final anchor = web.document.createElement('a') as web.HTMLAnchorElement;
    anchor.href = url;
    final filename = record != null
        ? "cable_loss_${record.cableName}_${record.date.replaceAll('-', '')}.csv"
        : "cable_loss_history_${DateTime.now().millisecondsSinceEpoch}.csv";
    anchor.download = filename;
    
    // Append to DOM, click, and remove to ensure browser respects download filename
    web.document.body?.appendChild(anchor);
    anchor.click();
    anchor.remove();

    // Delay revocation to ensure download is successfully initialized
    Future.delayed(const Duration(milliseconds: 200), () {
      web.URL.revokeObjectURL(url);
    });

    _showAppNotification(
      title: 'Export Successful',
      message: 'CSV file has been downloaded',
      type: NotificationType.success,
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              children: [
                ScreenHeader(
                  title: 'Cable Loss Measurement',
                  subtitle:
                      'Measure and record cable loss details across frequency bands',
                  icon: Icons.cable,
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton.filledTonal(
                        onPressed: () => setState(
                          () => _showConnections = !_showConnections,
                        ),
                        icon: Icon(
                          _showConnections ? Icons.hub : Icons.hub_outlined,
                        ),
                        tooltip: 'Show Connection Diagrams',
                      ),
                      const SizedBox(width: 12),
                      _buildHelpTrigger(theme),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                if (_isLoading)
                  const Expanded(
                    child: Center(child: CircularProgressIndicator()),
                  )
                else
                  Expanded(
                    child: SingleChildScrollView(
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
                          const SizedBox(height: 24), // Bottom padding
                        ],
                      ),
                    ),
                  ),
              ],
            ),
          ),
          if (_isMeasuring) _buildMeasuringOverlay(theme),
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
                  'Cable Loss Help',
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
                  'Calibration Principle',
                  'Cable loss measurement determines the attenuation of an RF cable across multiple frequency bands. '
                      'The system compares current power readings against a zeroed reference baseline.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Standard Procedure',
                  '1. **Zero PM**: Connect the sensor directly to the source reference to remove offsets.\n'
                      '2. **Measure**: Insert the cable between the source and sensor to calculate loss.\n'
                      '3. **Save**: Record name and length for traceability in the history logs.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Reference Diagrams',
                  'Toggle the "Hub" icon in the screen header to view visual wiring instructions. '
                      'Ensure all connections are finger-tight and use specific torque wrenches for precision measurements.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Accuracy Tips',
                  '• Always use the same Power Meter profile for reference and loss measurement.\n'
                      '• Ensure the cable length matches the physical tag for accurate loss-per-meter calculation.',
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

  Widget _buildConnectionsOverlay(ThemeData theme) {
    return Container(
      margin: const EdgeInsets.only(bottom: 24),
      child: ContentCard(
        isSidebar: true, // Use simpler style
        color: theme.colorScheme.primaryContainer.withValues(alpha: 0.4),
        borderRadius: 24,
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
      ),
    );
  }

  Widget _buildDiagramCard(
    ThemeData theme,
    String type,
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
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                Icon(icon, size: 18, color: theme.colorScheme.primary),
                const SizedBox(width: 10),
                Text(type, style: const TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),
            const SizedBox(height: 16),
            _buildBlockDiagram(theme, type),
            const SizedBox(height: 16),
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

  Widget _buildBlockDiagram(ThemeData theme, String type) {
    bool isZero = type.contains('Zero');

    return AspectRatio(
      aspectRatio: 3 / 1,
      child: InstrumentConnectionDiagram(
        type: isZero ? DiagramType.pmZero : DiagramType.cableMeasurement,
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
      child: ContentCard(
        isSidebar: true, // Use simpler style for status cards
        borderRadius: 20,
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.1),
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
            ?action,
          ],
        ),
      ),
    );
  }

  Widget _buildConfigCard(ThemeData theme) {
    return ContentCard(
      isSidebar: false,
      padding: const EdgeInsets.all(24),
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
                      initialValue: _selectedDeviceProfile,
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
                selectedColor: theme.colorScheme.primary.withValues(alpha: 0.2),
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
    final selectedRecords = _history
        .where((r) => _selectedSlNos.contains(r.slNo))
        .toList()
      ..sort((a, b) => a.slNo.compareTo(b.slNo));

    if (_history.isEmpty || selectedRecords.isEmpty) {
      return ContentCard(
        isSidebar: false,
        height: 480,
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.query_stats, size: 64, color: Colors.grey.shade200),
              const SizedBox(height: 16),
              const Text(
                'Select records from history to plot',
                style: TextStyle(color: Colors.grey),
              ),
            ],
          ),
        ),
      );
    }

    final isMulti = selectedRecords.length > 1;
    final primaryRecord = selectedRecords.last;

    final sortedFreqValues = _availableFrequencies.map((f) => f.value).toList()..sort();
    final double maxFreq = sortedFreqValues.isNotEmpty ? sortedFreqValues.last : 30.0;
    final double interval = maxFreq > 0 ? (maxFreq / 6).ceilToDouble() : 5.0;

    return ContentCard(
      isSidebar: false,
      padding: const EdgeInsets.all(24),
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
                    isMulti ? 'Comparison Plot' : 'Measurement Plot',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    isMulti
                        ? '${selectedRecords.length} cables selected'
                        : 'Cable: ${primaryRecord.cableName} (${primaryRecord.length}m)',
                    style: TextStyle(color: Colors.grey.shade600, fontSize: 13),
                  ),
                ],
              ),
              Row(
                children: [
                  TextButton.icon(
                    onPressed: () => setState(() => _selectedSlNos.clear()),
                    icon: const Icon(Icons.clear_all, size: 18),
                    label: const Text('Clear All'),
                  ),
                  IconButton(
                    onPressed: () => _exportToCSV(
                      record: isMulti ? null : primaryRecord,
                    ),
                    icon: const Icon(Icons.download),
                    tooltip: 'Export CSV',
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 32),
          SizedBox(
            height: 300,
            child: LineChart(
              LineChartData(
                minX: 0,
                maxX: maxFreq + (interval * 0.5),
                gridData: FlGridData(
                  show: true,
                  drawHorizontalLine: true,
                  drawVerticalLine: true,
                  getDrawingHorizontalLine: (value) => FlLine(
                    color: Colors.grey.withValues(alpha: 0.1),
                    strokeWidth: 1,
                  ),
                  getDrawingVerticalLine: (value) => FlLine(
                    color: Colors.grey.withValues(alpha: 0.1),
                    strokeWidth: 1,
                  ),
                ),
                titlesData: FlTitlesData(
                  show: true,
                  bottomTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize: 40,
                      interval: interval,
                      getTitlesWidget: (value, meta) {
                        if (value < 0 || value > maxFreq + interval) {
                          return const SizedBox();
                        }
                        return Padding(
                          padding: const EdgeInsets.only(top: 8.0),
                          child: Text(
                            value == 0 ? '0' : '${value.toInt()}',
                            style: const TextStyle(
                              fontSize: 10,
                              color: Colors.grey,
                              fontWeight: FontWeight.bold,
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
                        value.toStringAsFixed(1),
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
                lineBarsData: selectedRecords.asMap().entries.map((entry) {
                  final idx = entry.key;
                  final record = entry.value;
                  final color = _plotColors[idx % _plotColors.length];

                  final sortedMeasurements =
                      List<MeasurementPoint>.from(record.measurements)
                        ..sort((a, b) => a.frequency.compareTo(b.frequency));

                  final spots = <FlSpot>[];
                  for (final measurement in sortedMeasurements) {
                    if (measurement.loss != 0) {
                      spots.add(FlSpot(measurement.frequency, measurement.loss));
                    }
                  }

                  return LineChartBarData(
                    spots: spots,
                    isCurved: false,
                    color: color,
                    barWidth: 3,
                    isStrokeCapRound: true,
                    dotData: const FlDotData(show: true),
                    belowBarData: BarAreaData(
                      show: !isMulti,
                      color: color.withValues(alpha: 0.05),
                    ),
                  );
                }).toList(),
              ),
            ),
          ),
          const SizedBox(height: 16),
          _buildLegend(theme, selectedRecords),
        ],
      ),
    );
  }

  Widget _buildLegend(ThemeData theme, List<CableLossRecord> selectedRecords) {
    return Wrap(
      spacing: 16,
      runSpacing: 8,
      alignment: WrapAlignment.center,
      children: selectedRecords.asMap().entries.map((entry) {
        final idx = entry.key;
        final record = entry.value;
        final color = _plotColors[idx % _plotColors.length];
        return Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 10,
              height: 10,
              decoration: BoxDecoration(color: color, shape: BoxShape.circle),
            ),
            const SizedBox(width: 6),
            Text(
              '${record.cableName} (${record.date})',
              style: const TextStyle(
                fontSize: 11,
                color: Colors.grey,
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        );
      }).toList(),
    );
  }

  Widget _buildHistorySection(ThemeData theme) {
    return ContentCard(
      isSidebar: false,
      width: double.infinity,
      padding: const EdgeInsets.all(24),
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
              Row(
                children: [
                  SizedBox(
                    width: 200,
                    child: TextField(
                      decoration: InputDecoration(
                        hintText: 'Filter by Cable Name',
                        prefixIcon: const Icon(Icons.search, size: 18),
                        isDense: true,
                        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      ),
                      onChanged: (val) => setState(() => _searchCableName = val),
                    ),
                  ),
                  const SizedBox(width: 12),
                  SizedBox(
                    width: 150,
                    child: TextField(
                      decoration: InputDecoration(
                        hintText: 'Filter by Length',
                        prefixIcon: const Icon(Icons.straighten, size: 18),
                        isDense: true,
                        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      ),
                      keyboardType: TextInputType.number,
                      onChanged: (val) => setState(() => _searchCableLength = val),
                    ),
                  ),
                  const SizedBox(width: 16),
                  TextButton.icon(
                    onPressed: () => _exportToCSV(),
                    icon: const Icon(Icons.file_download),
                    label: const Text('Export All'),
                  ),
                ],
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
    final filteredHistory = _history.where((record) {
      final matchName = _searchCableName.isEmpty ||
          record.cableName.toLowerCase().contains(_searchCableName.toLowerCase());
      final matchLength = _searchCableLength.isEmpty ||
          record.length.toString().contains(_searchCableLength);
      return matchName && matchLength;
    }).toList();

    if (filteredHistory.isEmpty) {
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
                theme.colorScheme.primary.withValues(alpha: 0.02),
              ),
              horizontalMargin: 24,
              columnSpacing: 24,
            ),
          ),
          child: Container(
            decoration: BoxDecoration(
              border: Border.all(color: Colors.grey.withValues(alpha: 0.1)),
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
                  rows: filteredHistory.reversed.map((record) {
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
                                    if (_selectedSlNos.contains(record.slNo)) {
                                      _selectedSlNos.remove(record.slNo);
                                    } else {
                                      _selectedSlNos.add(record.slNo);
                                    }
                                  });
                                  if (_selectedSlNos.contains(record.slNo)) {
                                    _showAppNotification(
                                      title: 'Chart Updated',
                                      message:
                                          'Added ${record.cableName} to plot',
                                      type: NotificationType.info,
                                    );
                                  }
                                },
                                icon: Icon(
                                  _selectedSlNos.contains(record.slNo)
                                      ? Icons.check_circle
                                      : Icons.show_chart,
                                  size: 18,
                                  color: _selectedSlNos.contains(record.slNo)
                                      ? theme.colorScheme.primary
                                      : null,
                                ),
                                tooltip: 'Plot measurement',
                              ),
                              IconButton(
                                onPressed: () => _exportToCSV(record: record),
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
      color: Colors.black.withValues(alpha: 0.5),
      child: Center(
        child: ContentCard(
          width: 500,
          padding: const EdgeInsets.all(32),
          borderRadius: 24,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.primary.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(
                      Icons.settings_suggest,
                      color: theme.colorScheme.primary,
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          "MEASUREMENT IN PROGRESS",
                          style: GoogleFonts.outfit(
                            fontSize: 12,
                            fontWeight: FontWeight.w900,
                            color: theme.colorScheme.primary,
                            letterSpacing: 1.2,
                          ),
                        ),
                        Text(
                          "Measuring Cable: ${_cableNameController.text}",
                          style: GoogleFonts.inter(
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                            color: Colors.grey.shade700,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 32),
              const LinearProgressIndicator(
                minHeight: 6,
                borderRadius: BorderRadius.all(Radius.circular(3)),
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    _measuringStatus,
                    style: GoogleFonts.inter(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: theme.colorScheme.primary,
                    ),
                  ),
                  const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              Container(
                height: 250,
                width: double.infinity,
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFF1A1C1E),
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.05),
                      blurRadius: 10,
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      "STATUS LOG",
                      style: GoogleFonts.inter(
                        fontSize: 10,
                        fontWeight: FontWeight.w900,
                        color: Colors.grey.shade500,
                        letterSpacing: 1,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Expanded(
                      child: ListView.builder(
                        itemCount: _measuringLogs.length,
                        itemBuilder: (context, index) => Padding(
                          padding: const EdgeInsets.only(bottom: 6),
                          child: Text(
                            _measuringLogs[index],
                            style: GoogleFonts.robotoMono(
                              fontSize: 11,
                              color: Colors.grey.shade400,
                              height: 1.4,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 32),
              Row(
                children: [
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed: () {
                        context
                            .read<ServerService>()
                            .abortCableLossMeasurement();
                        setState(() {
                          _isMeasuring = false;
                          _measuringStatus = '';
                        });
                        _showAppNotification(
                          title: 'Measurement Aborted',
                          message:
                              'The measurement process was stopped by user.',
                          type: NotificationType.warning,
                        );
                      },
                      icon: const Icon(Icons.stop_circle_outlined),
                      label: const Text('ABORT MEASUREMENT'),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.red,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 20),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        elevation: 0,
                      ),
                    ),
                  ),
                ],
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
