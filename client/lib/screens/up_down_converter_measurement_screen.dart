import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/instrument_connection_diagram.dart';

class UpDownConverterScreen extends StatefulWidget {
  final bool isActive;
  const UpDownConverterScreen({super.key, this.isActive = true});

  @override
  State<UpDownConverterScreen> createState() => _UpDownConverterScreenState();
}

class _UpDownConverterScreenState extends State<UpDownConverterScreen> {
  @override
  void didUpdateWidget(UpDownConverterScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.isActive && !widget.isActive) {
      final serverService = Provider.of<ServerService>(context, listen: false);
      serverService.closeUCDC();
    }
  }

  // --- UI State ---
  bool _isConfigMode = true;
  int _selectedPortIndex = 0;
  bool _isMeasuring = false;
  double _progress = 0.0;
  final List<String> _logs = [];
  int _selectedSpectrumTab = 0;
  bool _isPortConnected = false;
  bool _abortRequested = false;
  bool _isHelpOpen = false;
  bool _showConnectionDiagram = false;
  bool _isReportMode = false;
  int? _selectedReportId;
  List<UCDCResultEntry> _history = [];
  bool _isLoadingHistory = false;

  // --- Real Data from Server ---
  UCDCMetadata? _metadata;
  List<String> _converters = ['UC-S-X-01', 'UC-KA-KU-02', 'DC-Q-V-01'];
  List<String> _deviceProfiles = [
    'Standard Profile',
    'High Temp Profile',
    'Low Power Profile',
  ];
  List<String> _externalSGs = [
    'SG-KEYSIGHT-01',
    'SG-ROHDE-02',
    'Internal Reference',
  ];

  // --- Database Driven Info (Read Only) ---
  String _saName = "N9020B-MXA-01";
  String _sgName = "E8257D-PSG-02";
  String _dbInputFreq = "2100.00 MHz";
  String _dbOutputFreq = "7250.00 MHz";
  String _dbLOFreq = "5150.00 MHz";

  // --- Controllers for Signal Configuration ---
  final TextEditingController _inputPowerController = TextEditingController(
    text: "-30.0",
  );
  final TextEditingController _stepSizeController = TextEditingController(
    text: "5.0",
  );

  // --- Controllers for Cable Losses ---
  final TextEditingController _lossInputCableIn = TextEditingController(
    text: "0.45",
  );
  final TextEditingController _lossOutputCableIn = TextEditingController(
    text: "0.12",
  );
  final TextEditingController _lossOutputCableOut = TextEditingController(
    text: "0.85",
  );
  final TextEditingController _lossOutputCableLO = TextEditingController(
    text: "0.55",
  );
  final TextEditingController _lossExtLOCableLO = TextEditingController(
    text: "0.35",
  );

  // --- Controllers for Spectrum Settings (GTx Style) ---
  final List<String> _spectrumTabs = [
    "Power",
    "Frequency",
    "In-Band Spurious",
    "Out-Band Spurious",
  ];

  final TextEditingController _powerSpanController = TextEditingController(
    text: "1000000",
  );
  final TextEditingController _powerRBWController = TextEditingController(
    text: "3000",
  );
  final TextEditingController _powerVBWController = TextEditingController(
    text: "1000",
  );

  final TextEditingController _freqSpanController = TextEditingController(
    text: "10000000",
  );
  final TextEditingController _freqRBWController = TextEditingController(
    text: "10000",
  );
  final TextEditingController _freqVBWController = TextEditingController(
    text: "3000",
  );

  final TextEditingController _inBandSpanController = TextEditingController(
    text: "1000000",
  );
  final TextEditingController _inBandRBWController = TextEditingController(
    text: "3000",
  );
  final TextEditingController _inBandVBWController = TextEditingController(
    text: "1000",
  );

  final TextEditingController _outBandSpanController = TextEditingController(
    text: "100000000",
  );
  final TextEditingController _outBandRBWController = TextEditingController(
    text: "10000",
  );
  final TextEditingController _outBandVBWController = TextEditingController(
    text: "3000",
  );

  String _selectedConverter = 'UC-S-X-01';
  String _selectedDeviceProfile = 'Standard Profile';
  String _selectedExternalSG = 'SG-KEYSIGHT-01';

  // --- Port & Test Definitions ---
  final List<PortConfig> _ports = [];

  final Map<String, dynamic> _mockResults = {};
  StreamSubscription? _ucdcSubscription;
  List<TestDefinition> _currentlyRunningTests = [];
  int _currentBatchTestIndex = 0;

  @override
  void dispose() {
    _inputPowerController.dispose();
    _stepSizeController.dispose();
    _lossInputCableIn.dispose();
    _lossOutputCableIn.dispose();
    _lossOutputCableOut.dispose();
    _lossOutputCableLO.dispose();
    _lossExtLOCableLO.dispose();

    _powerSpanController.dispose();
    _powerRBWController.dispose();
    _powerVBWController.dispose();
    _freqSpanController.dispose();
    _freqRBWController.dispose();
    _freqVBWController.dispose();
    _inBandSpanController.dispose();
    _inBandRBWController.dispose();
    _inBandVBWController.dispose();
    _outBandSpanController.dispose();
    _outBandRBWController.dispose();
    _outBandVBWController.dispose();
    _ucdcSubscription?.cancel();
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.closeUCDC();
    super.dispose();
  }

  void _handleUCDCEvent(dynamic event) {
    if (event is RTStatus) {
      _addLog(event.message);
      if (event.error) {
        _addLog("ERROR: ${event.message}");
        setState(() {
          if (_currentBatchTestIndex < _currentlyRunningTests.length) {
            _currentlyRunningTests[_currentBatchTestIndex].status = "ERROR";
          }
          _isMeasuring = false;
        });
        _ucdcSubscription?.cancel();
      } else if (event.completed) {
        setState(() {
          if (_currentBatchTestIndex < _currentlyRunningTests.length) {
            _currentlyRunningTests[_currentBatchTestIndex].status = "COMPLETE";
            _currentBatchTestIndex++;
            _progress = _currentBatchTestIndex / _currentlyRunningTests.length;
          }
        });
      }
    } else if (event is ConvertorResults) {
      setState(() {
        // Use TestCode as primary key for results mapping, fall back to TestName
        final key = event.testCode.isNotEmpty ? event.testCode : event.testName;
        _mockResults[key] = event;

        if (_currentBatchTestIndex < _currentlyRunningTests.length) {
          _currentlyRunningTests[_currentBatchTestIndex].status = "MEASURING";
        }
      });
    }
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchUCDCMetadata();
    });
  }

  Future<void> _fetchUCDCMetadata() async {
    final serverService = Provider.of<ServerService>(context, listen: false);
    final metadata = serverService.status.bootstrapData?.ucdcData;
    if (metadata != null && metadata.ok) {
      if (!mounted) return;
      setState(() {
        _metadata = metadata;
        _converters = metadata.converters;
        _deviceProfiles = metadata.deviceProfiles;
        _externalSGs = metadata.signalGenerators;

        if (_converters.isNotEmpty) _selectedConverter = _converters.first;
        if (_deviceProfiles.isNotEmpty)
          _selectedDeviceProfile = _deviceProfiles.first;
        if (_externalSGs.isNotEmpty) _selectedExternalSG = _externalSGs.first;

        _updateDBParams();
        _fetchHistory();
      });
    }
  }

  Future<void> _fetchHistory() async {
    if (_selectedConverter.isEmpty) return;
    setState(() => _isLoadingHistory = true);
    try {
      final serverService = Provider.of<ServerService>(context, listen: false);
      final response = await serverService.getUCDCResults(_selectedConverter);
      if (response.ok) {
        setState(() {
          _history = response.history;
        });
      }
    } catch (e) {
      _addLog("Error fetching history: $e");
    } finally {
      setState(() => _isLoadingHistory = false);
    }
  }

  void _addLog(String msg) {
    setState(() {
      _logs.insert(
        0,
        "[${DateTime.now().toString().split(' ')[1].split('.')[0]}] $msg",
      );
    });
  }

  void _updateDBParams() {
    if (_metadata == null) return;

    setState(() {
      // 1. Update Frequencies from Converter Details
      final converterDetails = _metadata!.converterDetails[_selectedConverter];
      if (converterDetails != null) {
        _dbInputFreq =
            "${(converterDetails.inputFrequency / 1e6).toStringAsFixed(2)} MHz";
        _dbOutputFreq =
            "${(converterDetails.outputFrequency / 1e6).toStringAsFixed(2)} MHz";
        _dbLOFreq =
            "${(converterDetails.loFrequency / 1e6).toStringAsFixed(2)} MHz";
      }

      // 2. Update Instruments from Device Mapping
      final deviceMapping = _metadata!.deviceMapping[_selectedDeviceProfile];
      if (deviceMapping != null) {
        _saName = deviceMapping.saName;
        _sgName = deviceMapping.sgName;
      }

      // 3. Update Port nomenclature and tests based on Up/Down Converter logic
      _updatePortConfigs();
    });
  }

  void _updatePortConfigs() {
    if (_metadata == null) return;
    final converterDetails = _metadata!.converterDetails[_selectedConverter];
    if (converterDetails == null) return;

    setState(() {
      _ports.clear();

      final categoryStyles = {
        "Output Port": {
          "icon": Icons.output,
          "instruction": "Connect SA to OUTPUT PORT.",
        },
        "Input Port": {
          "icon": Icons.login,
          "instruction": "Connect SA to INPUT PORT.",
        },
        "Output Monitor": {
          "icon": Icons.monitor,
          "instruction": "Connect SA to OUTPUT MONITOR PORT.",
        },
        "Input Monitor": {
          "icon": Icons.input,
          "instruction": "Connect SA to INPUT MONITOR PORT.",
        },
        "LO Monitor": {
          "icon": Icons.settings_input_antenna,
          "instruction": "Connect SA to LO MONITOR PORT.",
        },
      };

      final bool isUpConverter =
          converterDetails.outputFrequency > converterDetails.inputFrequency;

      Map<String, List<TestDefinition>> categorizedTests = {};
      for (var test in _metadata!.availableTests) {
        // Filter out radiated gain tests for UpConverters as requested
        if (isUpConverter &&
            (test.code == "UCDC_GAIN_INT_RAD" ||
                test.code == "UCDC_GAIN_EXT_RAD")) {
          continue;
        }

        if (!categorizedTests.containsKey(test.category)) {
          categorizedTests[test.category] = [];
        }
        categorizedTests[test.category]!.add(
          TestDefinition(test.displayName, code: test.code, isSelected: true),
        );
      }

      // Order ports predictably
      final order = [
        "Output Port",
        "Output Monitor",
        "LO Monitor",
        "Input Monitor",
        "Input Port",
      ];

      for (var cat in order) {
        if (categorizedTests.containsKey(cat)) {
          final style =
              categoryStyles[cat] ??
              {
                "icon": Icons.help_outline,
                "instruction": "Connect SA to $cat.",
              };
          _ports.add(
            PortConfig(
              name: cat,
              icon: style["icon"] as IconData,
              instruction: style["instruction"] as String,
              tests: categorizedTests[cat]!,
            ),
          );
        }
      }

      // Ensure index is valid
      if (_selectedPortIndex >= _ports.length) {
        _selectedPortIndex = 0;
      }
    });
  }

  void _abortBatch() {
    setState(() => _abortRequested = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.abortUCDC();
    _addLog("Abort request sent to server.");
  }

  Future<void> _runBatch() async {
    final activePort = _ports[_selectedPortIndex];
    final selectedTests = activePort.tests.where((t) => t.isSelected).toList();

    if (selectedTests.isEmpty) return;

    setState(() {
      _isMeasuring = true;
      _abortRequested = false;
      _progress = 0.0;
      _logs.clear();
      _currentlyRunningTests = selectedTests;
      _currentBatchTestIndex = 0;
      for (var t in selectedTests) t.status = "PENDING";
    });

    _addLog("Starting batch measurement for ${activePort.name}...");

    final serverService = Provider.of<ServerService>(context, listen: false);

    final request = UCDCRequest(
      converterName: _selectedConverter,
      deviceProfile: _selectedDeviceProfile,
      externalSGName: _selectedExternalSG,
      inputCableLoss: double.tryParse(_lossInputCableIn.text) ?? 0.0,
      inputPower: double.tryParse(_inputPowerController.text) ?? 0.0,
      loCableLoss: double.tryParse(_lossExtLOCableLO.text) ?? 0.0,
      outputCableLoss: [
        double.tryParse(_lossOutputCableIn.text) ?? 0.0,
        double.tryParse(_lossOutputCableOut.text) ?? 0.0,
        double.tryParse(_lossOutputCableLO.text) ?? 0.0,
      ],
      powerSpectrum: GTxSpectrum(
        span: double.tryParse(_powerSpanController.text) ?? 0.0,
        rbw: double.tryParse(_powerRBWController.text) ?? 0.0,
        vbw: double.tryParse(_powerVBWController.text) ?? 0.0,
      ),
      frequencySpectrum: GTxSpectrum(
        span: double.tryParse(_freqSpanController.text) ?? 0.0,
        rbw: double.tryParse(_freqRBWController.text) ?? 0.0,
        vbw: double.tryParse(_freqVBWController.text) ?? 0.0,
      ),
      inBandSpectrum: GTxSpectrum(
        span: double.tryParse(_inBandSpanController.text) ?? 0.0,
        rbw: double.tryParse(_inBandRBWController.text) ?? 0.0,
        vbw: double.tryParse(_inBandVBWController.text) ?? 0.0,
      ),
      outBandSpectrum: GTxSpectrum(
        span: double.tryParse(_outBandSpanController.text) ?? 0.0,
        rbw: double.tryParse(_outBandRBWController.text) ?? 0.0,
        vbw: double.tryParse(_outBandVBWController.text) ?? 0.0,
      ),
      stepSize: double.tryParse(_stepSizeController.text) ?? 0.0,
      testsSelected: selectedTests.map((t) => t.code).toList(),
    );

    _ucdcSubscription?.cancel();
    _ucdcSubscription = serverService
        .connectUCDC(request)
        .listen(
          _handleUCDCEvent,
          onError: (err) {
            _addLog("Connection Error: $err");
            setState(() => _isMeasuring = false);
          },
          onDone: () {
            _addLog("Measurement session ended.");
            setState(() => _isMeasuring = false);
          },
        );
  }

  List<Map<String, String>> _generateMockGainData() {
    return List.generate(10, (index) {
      double input = -30.0 + (index * 5.0);
      double output = -10.0 + (index * 4.8);
      double gain = output - input;
      return {
        "input": "${input.toStringAsFixed(1)}",
        "output": "${output.toStringAsFixed(2)}",
        "gain": "${gain.toStringAsFixed(2)}",
      };
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: Column(
        children: [
          ScreenHeader(
            title: 'Up/Down Converter Measurement',
            subtitle: 'Sequential port-based characterization',
            icon: Icons.swap_vert,
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                _buildTopStat(
                  _isReportMode ? "VIEWING" : "CONVERTER",
                  _isReportMode ? "REPORTS" : _selectedConverter,
                ),
                const SizedBox(width: 24),
                _buildTopStat(
                  _isReportMode ? "SESSION" : "PORT",
                  _isReportMode
                      ? "Latest Completion"
                      : _ports[_selectedPortIndex].name,
                ),
                const SizedBox(width: 8),
                IconButton.filledTonal(
                  onPressed: () => setState(
                    () => _showConnectionDiagram = !_showConnectionDiagram,
                  ),
                  icon: Icon(
                    _showConnectionDiagram ? Icons.hub : Icons.hub_outlined,
                  ),
                  tooltip: 'Show Path Diagrams',
                ),
                const SizedBox(width: 12),
                _buildHelpTrigger(theme),
              ],
            ),
          ),
          Expanded(
            child: Stack(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Container(
                        color: const Color(0xFFF8FAFC),
                        padding: const EdgeInsets.all(24),
                        child: SingleChildScrollView(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              _buildInstructionBanner(theme),
                              if (_showConnectionDiagram)
                                _buildConnectionOverlay(theme),
                              const SizedBox(height: 16),
                              if (_isReportMode)
                                _buildReportView(theme)
                              else if (_isConfigMode)
                                _buildConfigurationView(theme)
                              else
                                _buildMeasurementView(theme),
                            ],
                          ),
                        ),
                      ),
                    ),
                    _buildSidebar(theme),
                  ],
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
          ),
        ],
      ),
    );
  }

  Widget _buildSidebar(ThemeData theme) {
    return ContentCard(
      width: 280,
      isSidebar: true,
      margin: const EdgeInsets.only(left: 0),
      borderRadius:
          0, // Using 0 to mimic the previous border-left style or just use default sidebar style
      // Actually, to match the requested style, let's just use ContentCard generally but maybe without left border as it was.
      // But the request is to Standardize. Standard sidebar is floating card.
      // Let's make it a floating card on the right.
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primary,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(
                    Icons.settings_suggest,
                    color: Colors.white,
                  ),
                ),
                const SizedBox(width: 12),
                Text(
                  "MEASUREMENT\nPORTS",
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    fontSize: 14,
                    height: 1.1,
                    letterSpacing: 1,
                    color: theme.colorScheme.primary,
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView.builder(
              itemCount: _ports.length,
              itemBuilder: (context, index) {
                final port = _ports[index];
                final isSelected = _selectedPortIndex == index;
                final isDone = port.tests.every((t) => t.status == "COMPLETE");

                return InkWell(
                  onTap: _isMeasuring
                      ? null
                      : () => setState(() {
                          _selectedPortIndex = index;
                          _isPortConnected =
                              false; // Reset connection status when port changes
                        }),
                  child: Opacity(
                    opacity: _isMeasuring && !isSelected ? 0.5 : 1.0,
                    child: AnimatedContainer(
                      duration: const Duration(milliseconds: 200),
                      margin: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 4,
                      ),
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: isSelected
                            ? theme.colorScheme.primary.withOpacity(0.08)
                            : Colors.transparent,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: isSelected
                              ? theme.colorScheme.primary.withOpacity(0.2)
                              : Colors.transparent,
                        ),
                      ),
                      child: Row(
                        children: [
                          Icon(
                            port.icon,
                            size: 20,
                            color: isSelected
                                ? theme.colorScheme.primary
                                : Colors.grey.shade400,
                          ),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Text(
                              port.name,
                              style: GoogleFonts.inter(
                                fontWeight: isSelected
                                    ? FontWeight.bold
                                    : FontWeight.w500,
                                color: isSelected
                                    ? theme.colorScheme.primary
                                    : Colors.black87,
                              ),
                            ),
                          ),
                          if (isDone)
                            const Icon(
                              Icons.check_circle,
                              color: Colors.green,
                              size: 16,
                            ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
          const Divider(height: 1),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: InkWell(
              onTap: _isMeasuring
                  ? null
                  : () {
                      setState(() => _isReportMode = !_isReportMode);
                      if (_isReportMode) _fetchHistory();
                    },
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: _isReportMode
                      ? theme.colorScheme.primary.withOpacity(0.1)
                      : Colors.transparent,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: _isReportMode
                        ? theme.colorScheme.primary.withOpacity(0.3)
                        : Colors.transparent,
                  ),
                ),
                child: Row(
                  children: [
                    Icon(
                      Icons.picture_as_pdf,
                      size: 20,
                      color: _isReportMode
                          ? theme.colorScheme.primary
                          : Colors.grey.shade400,
                    ),
                    const SizedBox(width: 16),
                    Text(
                      "REPORTS",
                      style: GoogleFonts.inter(
                        fontWeight: _isReportMode
                            ? FontWeight.bold
                            : FontWeight.w500,
                        color: _isReportMode
                            ? theme.colorScheme.primary
                            : Colors.black87,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: ElevatedButton.icon(
              onPressed: (_isMeasuring || _isReportMode)
                  ? null
                  : () => setState(() => _isConfigMode = !_isConfigMode),
              icon: Icon(_isConfigMode ? Icons.analytics : Icons.settings),
              label: Text(
                _isMeasuring
                    ? "BATCH IN PROGRESS"
                    : _isReportMode
                    ? "EXIT REPORT MODE"
                    : (_isConfigMode ? "GO TO MEASUREMENT" : "GO TO SETUP"),
              ),
              style: ElevatedButton.styleFrom(
                minimumSize: const Size(double.infinity, 50),
                backgroundColor: _isMeasuring ? Colors.grey.shade200 : null,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildReportView(ThemeData theme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  "REPORT GENERATOR",
                  style: GoogleFonts.outfit(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: const Color(0xFF1E293B),
                  ),
                ),
                Text(
                  "Historical measurements for $_selectedConverter",
                  style: GoogleFonts.inter(
                    color: Colors.grey.shade600,
                    fontSize: 13,
                  ),
                ),
              ],
            ),
            ElevatedButton.icon(
              onPressed: () {
                _addLog("Generating PDF Report...");
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text("PDF Generation Started...")),
                );
              },
              icon: const Icon(Icons.download),
              label: const Text("EXPORT ALL DATA"),
              style: ElevatedButton.styleFrom(
                backgroundColor: theme.colorScheme.primary,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 16,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 32),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Left Column: Sessions
            Expanded(
              flex: 2,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    "COMPLETED SESSIONS",
                    style: GoogleFonts.inter(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: Colors.grey.shade500,
                      letterSpacing: 1.2,
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (_isLoadingHistory)
                    const Center(
                      child: Padding(
                        padding: EdgeInsets.all(32.0),
                        child: CircularProgressIndicator(),
                      ),
                    )
                  else if (_history.isEmpty)
                    const Center(
                      child: Padding(
                        padding: EdgeInsets.all(32.0),
                        child: Text("No history found"),
                      ),
                    )
                  else
                    ..._history.map((s) => _buildSessionCard(s, theme)),
                ],
              ),
            ),
            const SizedBox(width: 32),
            // Right Column: Details & Selection
            Expanded(
              flex: 3,
              child: _selectedReportId != null
                  ? _buildSessionDetails(theme)
                  : ContentCard(
                      child: Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.analytics_outlined,
                              size: 48,
                              color: Colors.grey.shade300,
                            ),
                            const SizedBox(height: 16),
                            Text(
                              "SELECT A SESSION TO VIEW RESULTS",
                              style: GoogleFonts.inter(
                                fontSize: 12,
                                fontWeight: FontWeight.bold,
                                color: Colors.grey,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildSessionDetails(ThemeData theme) {
    final session = _history.firstWhere((s) => s.id == _selectedReportId);
    return Column(
      children: [
        _buildDynamicResult(
          session.results,
          TestDefinition(
            session.testType,
            code: session.results.testCode,
            status: "COMPLETE",
          ),
        ),
      ],
    );
  }

  Widget _buildSessionCard(UCDCResultEntry session, ThemeData theme) {
    bool isSelected = session.id == _selectedReportId;

    return InkWell(
      onTap: () => setState(() => _selectedReportId = session.id),
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: isSelected ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected
                ? theme.colorScheme.primary.withOpacity(0.3)
                : Colors.grey.shade200,
            width: isSelected ? 2 : 1,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.05),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ]
              : null,
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: isSelected
                    ? theme.colorScheme.primary.withOpacity(0.1)
                    : Colors.grey.shade100,
                shape: BoxShape.circle,
              ),
              child: Icon(
                Icons.calendar_month,
                size: 20,
                color: isSelected ? theme.colorScheme.primary : Colors.grey,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    session.testType,
                    style: GoogleFonts.inter(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Text(
                    "${session.date} • ${session.time}",
                    style: GoogleFonts.inter(
                      color: Colors.grey.shade600,
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            if (isSelected)
              const Icon(Icons.arrow_forward_ios, size: 14, color: Colors.grey),
          ],
        ),
      ),
    );
  }

  Widget _buildTopStat(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 9,
            fontWeight: FontWeight.w900,
            color: Colors.grey.shade400,
            letterSpacing: 1,
          ),
        ),
        Text(
          value,
          style: GoogleFonts.outfit(
            fontSize: 15,
            fontWeight: FontWeight.bold,
            color: const Color(0xFF0D47A1),
          ),
        ),
      ],
    );
  }

  Widget _buildInstructionBanner(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
      decoration: BoxDecoration(
        color: _isPortConnected
            ? const Color(0xFFE8F5E9)
            : const Color(0xFFFFF9C4),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: (_isPortConnected ? Colors.green : Colors.orange).withOpacity(
            0.3,
          ),
        ),
      ),
      child: Row(
        children: [
          Icon(
            _isPortConnected ? Icons.check_circle : Icons.info_outline,
            color: _isPortConnected ? Colors.green : Colors.orange,
            size: 20,
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Text(
              _ports[_selectedPortIndex].instruction,
              style: GoogleFonts.inter(
                fontWeight: FontWeight.w600,
                color: _isPortConnected
                    ? Colors.green.shade900
                    : Colors.orange.shade900,
                fontSize: 13,
              ),
            ),
          ),
          TextButton.icon(
            onPressed: () => setState(() => _isPortConnected = true),
            icon: Icon(
              _isPortConnected ? Icons.verified : Icons.check,
              size: 14,
            ),
            label: Text(
              _isPortConnected ? "VERIFIED" : "CONNECTED",
              style: const TextStyle(fontSize: 12),
            ),
            style: TextButton.styleFrom(
              foregroundColor: _isPortConnected
                  ? Colors.green.shade900
                  : Colors.orange.shade900,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildConfigurationView(ThemeData theme) {
    return Column(
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Device Configuration Card
            Expanded(
              flex: 3,
              child: ContentCard(
                child: Column(
                  children: [
                    Row(
                      children: [
                        Icon(
                          Icons.developer_board,
                          color: theme.colorScheme.primary,
                        ),
                        const SizedBox(width: 12),
                        Text(
                          "DEVICE CONFIGURATION",
                          style: GoogleFonts.inter(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: Colors.grey.shade500,
                            letterSpacing: 1.2,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Expanded(
                          child: _buildDropdown(
                            "Converter Model",
                            _selectedConverter,
                            _converters,
                            (v) {
                              setState(() {
                                _selectedConverter = v!;
                                _selectedReportId = null;
                                _history = [];
                              });
                              _updateDBParams();
                              _fetchHistory();
                            },
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: _buildDropdown(
                            "Device Profile",
                            _selectedDeviceProfile,
                            _deviceProfiles,
                            (v) {
                              setState(() => _selectedDeviceProfile = v!);
                              _updateDBParams();
                            },
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Expanded(
                          child: _buildDropdown(
                            "External SG for LO",
                            _selectedExternalSG,
                            _externalSGs,
                            (v) => setState(() => _selectedExternalSG = v!),
                          ),
                        ),
                        const SizedBox(width: 16),
                        const Spacer(),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Expanded(
                          child: _buildTextField(
                            label: "Nominal Input Power (dBm)",
                            controller: _inputPowerController,
                            icon: Icons.speed,
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: _buildTextField(
                            label: "Step Size for Gain (dB)",
                            controller: _stepSizeController,
                            icon: Icons.linear_scale,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(width: 24),
            // Database Driven Parameters (Read Only)
            Expanded(
              flex: 2,
              child: ContentCard(
                child: Column(
                  children: [
                    Row(
                      children: [
                        Icon(Icons.storage, color: theme.colorScheme.primary),
                        const SizedBox(width: 12),
                        Text(
                          "DATABASE PARAMETERS",
                          style: GoogleFonts.inter(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: Colors.grey.shade500,
                            letterSpacing: 1.2,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    _buildInfoRow(Icons.analytics_outlined, "SA Name", _saName),
                    const Divider(height: 24),
                    _buildInfoRow(
                      Icons.settings_input_antenna,
                      "SG Name",
                      _sgName,
                    ),
                    const Divider(height: 24),
                    _buildInfoRow(Icons.login, "Input Freq", _dbInputFreq),
                    const SizedBox(height: 12),
                    _buildInfoRow(Icons.logout, "Output Freq", _dbOutputFreq),
                    const SizedBox(height: 12),
                    _buildInfoRow(Icons.vibration, "LO Frequency", _dbLOFreq),
                  ],
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 24),
        // Loss Table Card
        ContentCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.grid_on, color: theme.colorScheme.primary),
                  const SizedBox(width: 12),
                  Text(
                    "CABLE LOSS CALIBRATION (dB)",
                    style: GoogleFonts.inter(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: Colors.grey.shade500,
                      letterSpacing: 1.2,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Table(
                columnWidths: const {0: FlexColumnWidth(1.5)},
                children: [
                  _buildTableRow([
                    'Cable',
                    'At Input Freq',
                    'At Output Freq',
                    'At LO Freq',
                  ], isHeader: true),
                  _buildEditableLossRow(
                    'Input Cable',
                    _lossInputCableIn,
                    null,
                    null,
                  ),
                  _buildEditableLossRow(
                    'Output Cable',
                    _lossOutputCableIn,
                    _lossOutputCableOut,
                    _lossOutputCableLO,
                  ),
                  _buildEditableLossRow(
                    'Ext LO Cable',
                    null,
                    null,
                    _lossExtLOCableLO,
                  ),
                ],
              ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        const SizedBox(height: 24),
        // Spectrum Settings Card (GTx Style)
        ContentCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.analytics, color: theme.colorScheme.primary),
                  const SizedBox(width: 12),
                  Text(
                    "SPECTRUM SETTINGS",
                    style: GoogleFonts.inter(
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                      color: Colors.grey.shade500,
                      letterSpacing: 1.2,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              _buildSpectrumTabs(theme),
            ],
          ),
        ),
        const SizedBox(height: 48),
      ],
    );
  }

  Widget _buildSpectrumTabs(ThemeData theme) {
    return Column(
      children: [
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: List.generate(_spectrumTabs.length, (index) {
              final isSelected = _selectedSpectrumTab == index;
              return Padding(
                padding: const EdgeInsets.only(right: 8),
                child: ChoiceChip(
                  label: Text(_spectrumTabs[index]),
                  selected: isSelected,
                  onSelected: (val) =>
                      setState(() => _selectedSpectrumTab = index),
                  selectedColor: theme.colorScheme.primary.withOpacity(0.1),
                  labelStyle: TextStyle(
                    color: isSelected
                        ? theme.colorScheme.primary
                        : Colors.grey.shade600,
                    fontWeight: isSelected
                        ? FontWeight.bold
                        : FontWeight.normal,
                    fontSize: 12,
                  ),
                ),
              );
            }),
          ),
        ),
        const SizedBox(height: 24),
        Row(
          children: [
            Expanded(
              child: _buildTextField(
                label: "Span (Hz)",
                controller: _selectedSpectrumTab == 0
                    ? _powerSpanController
                    : _selectedSpectrumTab == 1
                    ? _freqSpanController
                    : _selectedSpectrumTab == 2
                    ? _inBandSpanController
                    : _outBandSpanController,
                icon: Icons.unfold_more,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _buildTextField(
                label: "RBW (Hz)",
                controller: _selectedSpectrumTab == 0
                    ? _powerRBWController
                    : _selectedSpectrumTab == 1
                    ? _freqRBWController
                    : _selectedSpectrumTab == 2
                    ? _inBandRBWController
                    : _outBandRBWController,
                icon: Icons.grid_view,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _buildTextField(
                label: "VBW (Hz)",
                controller: _selectedSpectrumTab == 0
                    ? _powerVBWController
                    : _selectedSpectrumTab == 1
                    ? _freqVBWController
                    : _selectedSpectrumTab == 2
                    ? _inBandVBWController
                    : _outBandVBWController,
                icon: Icons.blur_on,
              ),
            ),
          ],
        ),
      ],
    );
  }

  TableRow _buildEditableLossRow(
    String cable,
    TextEditingController? c1,
    TextEditingController? c2,
    TextEditingController? c3,
  ) {
    return TableRow(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Text(
            cable,
            style: const TextStyle(fontWeight: FontWeight.w500),
          ),
        ),
        _buildLossInput(c1),
        _buildLossInput(c2),
        _buildLossInput(c3),
      ],
    );
  }

  Widget _buildLossInput(TextEditingController? controller) {
    if (controller == null) {
      return const Center(
        child: Text("NA", style: TextStyle(color: Colors.grey, fontSize: 12)),
      );
    }
    return Padding(
      padding: const EdgeInsets.all(4.0),
      child: TextFormField(
        controller: controller,
        textAlign: TextAlign.center,
        style: const TextStyle(fontSize: 13),
        decoration: const InputDecoration(
          isDense: true,
          border: OutlineInputBorder(),
        ),
      ),
    );
  }

  Widget _buildMeasurementView(ThemeData theme) {
    final activePort = _ports[_selectedPortIndex];

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Left Column: Test Selection & Logs
        Expanded(
          flex: 1,
          child: Column(
            children: [
              ContentCard(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.checklist, color: theme.colorScheme.primary),
                        const SizedBox(width: 12),
                        Text(
                          "TEST SELECTION",
                          style: GoogleFonts.inter(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: Colors.grey.shade500,
                            letterSpacing: 1.2,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    ConstrainedBox(
                      constraints: const BoxConstraints(maxHeight: 250),
                      child: ListView(
                        shrinkWrap: true,
                        children: activePort.tests
                            .map(
                              (test) => CheckboxListTile(
                                title: Text(
                                  test.name,
                                  style: GoogleFonts.inter(fontSize: 13),
                                ),
                                value: test.isSelected,
                                onChanged: _isMeasuring
                                    ? null
                                    : (v) =>
                                          setState(() => test.isSelected = v!),
                                dense: true,
                                contentPadding: EdgeInsets.zero,
                                controlAffinity:
                                    ListTileControlAffinity.leading,
                              ),
                            )
                            .toList(),
                      ),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton.icon(
                      onPressed: _isMeasuring ? _abortBatch : _runBatch,
                      icon: Icon(
                        _isMeasuring
                            ? Icons.stop_circle_outlined
                            : Icons.play_arrow,
                      ),
                      label: Text(
                        _isMeasuring ? "ABORT BATCH" : "RUN SELECTED TESTS",
                      ),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: _isMeasuring
                            ? Colors.red
                            : theme.colorScheme.primary,
                        minimumSize: const Size(double.infinity, 52),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
              SizedBox(
                height: 300,
                child: ContentCard(
                  child: Column(
                    children: [
                      Row(
                        children: [
                          Icon(
                            Icons.terminal,
                            color: theme.colorScheme.primary,
                          ),
                          const SizedBox(width: 12),
                          Text(
                            "STATUS LOG",
                            style: GoogleFonts.inter(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: Colors.grey.shade500,
                              letterSpacing: 1.2,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                      Expanded(
                        child: ListView.builder(
                          itemCount: _logs.length,
                          itemBuilder: (context, index) => Padding(
                            padding: const EdgeInsets.only(bottom: 4),
                            child: Text(
                              _logs[index],
                              style: GoogleFonts.robotoMono(
                                fontSize: 10,
                                color: Colors.grey.shade600,
                              ),
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: 24),
        // Right Column: Progress & Results
        Expanded(
          flex: 2,
          child: Column(
            children: [
              if (_isMeasuring)
                Container(
                  padding: const EdgeInsets.all(16),
                  margin: const EdgeInsets.only(bottom: 16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Column(
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text(
                            "Batch Progress",
                            style: TextStyle(fontWeight: FontWeight.bold),
                          ),
                          Text("${(_progress * 100).toInt()}%"),
                        ],
                      ),
                      const SizedBox(height: 8),
                      LinearProgressIndicator(
                        value: _progress,
                        borderRadius: BorderRadius.circular(4),
                      ),
                    ],
                  ),
                ),
              Column(
                children: activePort.tests
                    .where((t) => t.status != "PENDING")
                    .map((test) => _buildResultCard(test, theme))
                    .toList(),
              ),
            ],
          ),
        ),
      ],
    );
  }

  // --- Helper Methods ---

  Widget _buildResultCard(TestDefinition test, ThemeData theme) {
    // Try to find result by code first, then by name
    final result = _mockResults[test.code] ?? _mockResults[test.name];

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1C1E),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
            child: Row(
              children: [
                Icon(
                  Icons.analytics_outlined,
                  color: Colors.blue.shade300,
                  size: 18,
                ),
                const SizedBox(width: 12),
                Text(
                  test.name.toUpperCase(),
                  style: GoogleFonts.inter(
                    color: Colors.white.withOpacity(0.6),
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                    letterSpacing: 1.2,
                  ),
                ),
                const Spacer(),
                if (test.status == "MEASURING")
                  const SizedBox(
                    width: 14,
                    height: 14,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                else
                  const Icon(Icons.check_circle, color: Colors.green, size: 16),
              ],
            ),
          ),
          const Divider(color: Colors.white10, height: 1),
          Padding(
            padding: const EdgeInsets.all(24),
            child: _buildDynamicResult(result, test),
          ),
        ],
      ),
    );
  }

  Widget _buildDynamicResult(dynamic result, TestDefinition test) {
    if (test.status == "MEASURING" && result == null) {
      return Text(
        "Measuring...",
        style: GoogleFonts.robotoMono(
          fontSize: 18,
          color: Colors.blue.shade300,
        ),
      );
    }

    if (result is! ConvertorResults) {
      return Text(
        result?.toString() ?? "Awaiting Data...",
        style: GoogleFonts.robotoMono(
          fontSize: 18,
          fontWeight: FontWeight.bold,
          color: Colors.blue.shade300,
        ),
      );
    }

    final res = result;
    if (res.gainResults && res.gainResultValue != null) {
      return _buildGainResultTable(res.gainResultValue!);
    }
    if (res.frequencyResults && res.frequencyResultValue != null) {
      return _buildFrequencyResult(res.frequencyResultValue!);
    }
    if (res.harmonicsResults && res.harmonicResultValue != null) {
      return _buildHarmonicResult(res.harmonicResultValue!);
    }
    if (res.spuriousResults && res.spuriousResultValue != null) {
      return _buildSpuriousResult(res.spuriousResultValue!);
    }
    if (res.powerOrLeakageResults && res.powerOrLeakageResultValue != null) {
      return _buildPowerOrLeakageResult(res.powerOrLeakageResultValue!);
    }
    if (res.phaseNoiseResults && res.phaseNoiseResultValue != null) {
      return _buildPhaseNoiseResult(res.phaseNoiseResultValue!);
    }

    return Text(
      "Result received: ${res.testName}",
      style: GoogleFonts.robotoMono(fontSize: 16, color: Colors.blue.shade300),
    );
  }

  Widget _buildFrequencyResult(FrequencyResults res) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _resultRow(
          "Expected Freq",
          "${(res.expectedFrequency / 1e6).toStringAsFixed(6)} MHz",
        ),
        _resultRow(
          "Measured Freq",
          "${(res.measuredFrequency / 1e6).toStringAsFixed(6)} MHz",
        ),
        _resultRow(
          "Deviation",
          "${(res.deviation / 1e3).toStringAsFixed(3)} kHz",
          isError: res.deviation > 1000,
        ),
      ],
    );
  }

  Widget _buildPowerOrLeakageResult(PowerOrLeakageResults res) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _resultRow(
          "Frequency",
          "${(res.frequency / 1e6).toStringAsFixed(2)} MHz",
        ),
        _resultRow("Measured Power", "${res.power.toStringAsFixed(2)} dBm"),
      ],
    );
  }

  Widget _resultRow(String label, String value, {bool isError = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: GoogleFonts.inter(color: Colors.white70, fontSize: 13),
          ),
          Text(
            value,
            style: GoogleFonts.robotoMono(
              color: isError ? Colors.red.shade300 : Colors.blue.shade300,
              fontWeight: FontWeight.bold,
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHarmonicResult(HarmonicResults res) {
    return Table(
      children: [
        TableRow(
          children: [
            _tableHeader("Harmonic"),
            _tableHeader("Frequency"),
            _tableHeader("Level (dBc)"),
          ],
        ),
        for (int i = 0; i < res.harmonicNo.length; i++)
          TableRow(
            children: [
              _tableCell("${res.harmonicNo[i]}"),
              _tableCell(
                "${(double.parse(res.harmonicFrequency[i]) / 1e6).toStringAsFixed(2)} MHz",
              ),
              _tableCell(res.carrierLevel[i]),
            ],
          ),
      ],
    );
  }

  Widget _buildSpuriousResult(SpuriousResults res) {
    return Table(
      children: [
        TableRow(
          children: [
            _tableHeader("Frequency"),
            _tableHeader("Level (dBm)"),
            _tableHeader("dBC"),
          ],
        ),
        for (int i = 0; i < res.frequency.length; i++)
          TableRow(
            children: [
              _tableCell(
                res.frequency[i] == "NIL"
                    ? "NIL"
                    : "${(double.parse(res.frequency[i]) / 1e6).toStringAsFixed(2)} MHz",
              ),
              _tableCell(res.measuredPowerdBm[i]),
              _tableCell(res.spuriousLeveldBC[i]),
            ],
          ),
      ],
    );
  }

  Widget _buildPhaseNoiseResult(PhaseNoiseResults res) {
    return Column(
      children: [
        for (int i = 0; i < res.frequency.length; i++)
          _resultRow(
            "Offset ${res.frequency[i]}",
            "${res.phaseNoise[i]} dBc/Hz",
          ),
      ],
    );
  }

  Widget _buildTextResult(dynamic result, String status) {
    if (status == "MEASURING") {
      return Text(
        "Measuring...",
        style: GoogleFonts.robotoMono(
          fontSize: 18,
          color: Colors.blue.shade300,
        ),
      );
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          result?.toString() ?? "Awaiting Data...",
          style: GoogleFonts.robotoMono(
            fontSize: 22,
            fontWeight: FontWeight.bold,
            color: Colors.blue.shade300,
          ),
        ),
        const SizedBox(height: 8),
        const Text(
          "VERDICT: PASSED",
          style: TextStyle(
            color: Colors.green,
            fontWeight: FontWeight.bold,
            fontSize: 11,
          ),
        ),
      ],
    );
  }

  Widget _buildGainResultTable(GainResults res) {
    return Table(
      children: [
        TableRow(
          children: [
            _tableHeader("In Power (dBm)"),
            _tableHeader("Out Power (dBm)"),
            _tableHeader("Gain (dB)"),
          ],
        ),
        for (int i = 0; i < res.setPower.length; i++)
          TableRow(
            children: [
              _tableCell(res.setPower[i].toStringAsFixed(2)),
              _tableCell(res.outputPower[i].toStringAsFixed(2)),
              _tableCell(
                res.gain[i].toStringAsFixed(2),
                isBold: true,
                color: Colors.blue.shade300,
              ),
            ],
          ),
      ],
    );
  }

  Widget _tableHeader(String text) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: Text(
      text,
      style: const TextStyle(
        color: Colors.white54,
        fontSize: 10,
        fontWeight: FontWeight.bold,
      ),
    ),
  );
  Widget _tableCell(
    String text, {
    bool isBold = false,
    Color color = Colors.white70,
  }) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 8),
    child: Text(
      text,
      style: GoogleFonts.robotoMono(
        color: color,
        fontSize: 13,
        fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
      ),
    ),
  );

  Widget _buildTextField({
    required String label,
    required TextEditingController controller,
    required IconData icon,
  }) {
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        labelText: label,
        prefixIcon: Icon(icon, size: 18),
        border: const OutlineInputBorder(),
        floatingLabelBehavior: FloatingLabelBehavior.always,
      ),
    );
  }

  Widget _buildDropdown(
    String label,
    String value,
    List<String> items,
    Function(String?) onChanged,
  ) {
    return DropdownButtonFormField<String>(
      value: value,
      items: items
          .map((e) => DropdownMenuItem(value: e, child: Text(e)))
          .toList(),
      onChanged: onChanged,
      decoration: InputDecoration(
        labelText: label,
        border: const OutlineInputBorder(),
        floatingLabelBehavior: FloatingLabelBehavior.always,
      ),
    );
  }

  TableRow _buildTableRow(List<String> cells, {bool isHeader = false}) {
    return TableRow(
      children: cells
          .map(
            (cell) => Padding(
              padding: const EdgeInsets.all(12.0),
              child: Text(
                cell,
                style: GoogleFonts.inter(
                  fontWeight: isHeader ? FontWeight.bold : FontWeight.normal,
                  color: isHeader ? Colors.black : Colors.grey.shade700,
                  fontSize: 12,
                ),
              ),
            ),
          )
          .toList(),
    );
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      children: [
        Icon(icon, size: 16, color: Colors.blue.shade700.withOpacity(0.6)),
        const SizedBox(width: 12),
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 12,
            color: Colors.grey.shade600,
            fontWeight: FontWeight.w500,
          ),
        ),
        const Spacer(),
        Text(
          value,
          style: GoogleFonts.robotoMono(
            fontSize: 13,
            color: Colors.blue.shade900,
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
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
                  'Converter Help',
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
                  'Sequential Testing',
                  'Testing is performed port-by-port. Each port (Input, Output, LO Monitor, etc.) has a specific set '
                      'of assigned tests like Gain, Spurious, or Phase Noise.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Port Connection',
                  'Before running a batch, follow the instruction in the top banner. Once the hardware is connected, '
                      'click "CONNECTED" to enable the "RUN BATCH" action.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Reference LO',
                  'Select the external Signal Generator to be used for the Local Oscillator (LO) reference.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Spectrum Automation',
                  'PRISM automatically configures Span, RBW, and VBW on the Spectrum Analyzer based on the '
                      'parameters defined in the Spectrum Settings tab.',
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

  Widget _buildConnectionOverlay(ThemeData theme) {
    final activePort = _ports[_selectedPortIndex];
    final isPhysicalInput = _selectedPortIndex == 4;

    return ContentCard(
      color: theme.colorScheme.primaryContainer.withOpacity(0.4),
      margin: const EdgeInsets.only(top: 16),
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Connection Guide: ${activePort.name}',
                style: GoogleFonts.outfit(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.onPrimaryContainer,
                ),
              ),
              IconButton(
                onPressed: () => setState(() => _showConnectionDiagram = false),
                icon: const Icon(Icons.close),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(
            activePort.instruction,
            style: TextStyle(
              fontSize: 13,
              color: theme.colorScheme.onPrimaryContainer.withOpacity(0.8),
            ),
          ),
          const SizedBox(height: 16),
          AspectRatio(
            aspectRatio: 3 / 1,
            child: InstrumentConnectionDiagram(
              type: isPhysicalInput
                  ? DiagramType.converterSimple
                  : DiagramType.converterComplex,
              tsmOutputName: activePort.name,
              inputPortName: _ports[4].name,
              outputPortName: _ports[0].name,
            ),
          ),
        ],
      ),
    );
  }
}

class PortConfig {
  String name;
  final IconData icon;
  String instruction;
  final List<TestDefinition> tests;
  PortConfig({
    required this.name,
    required this.icon,
    required this.instruction,
    required this.tests,
  });
}

class TestDefinition {
  final String name;
  final String code;
  bool isSelected;
  String status;
  TestDefinition(
    this.name, {
    required this.code,
    this.isSelected = false,
    this.status = "PENDING",
  });
}
