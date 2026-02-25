import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/instrument_connection_diagram.dart';

class GTxCharacterizationScreen extends StatefulWidget {
  final bool isActive;
  const GTxCharacterizationScreen({super.key, this.isActive = true});

  @override
  State<GTxCharacterizationScreen> createState() =>
      _GTxCharacterizationScreenState();
}

class _GTxCharacterizationScreenState extends State<GTxCharacterizationScreen> {
  @override
  void didUpdateWidget(GTxCharacterizationScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.isActive && !widget.isActive) {
      final serverService = Provider.of<ServerService>(context, listen: false);
      serverService.closeGTxTne();
    }
  }

  // Data from Server
  GTxMeasurementMetadata? _metadata;
  List<String> _profiles = [];
  String? _selectedProfile;

  final List<String> _components = ['IFM-1', 'IFM-2'];
  String _selectedComponent = 'IFM-1';

  final List<String> _modulations = ["PM", "FM", "CDMA", "PSK", "FSK"];
  String _selectedModulation = "PM";

  bool _isHelpOpen = false;
  bool _isMeasuring = false;
  bool _showConnections = false;

  // Controllers
  final _cableLossController = TextEditingController(text: "1.2");
  final _ifController = TextEditingController(text: "70,000,000");
  final _subCarFreqController = TextEditingController(text: "8,000");
  final _modIndexController = TextEditingController(text: "1.0");
  final _freqDevController = TextEditingController(text: "200,000");

  final _powerSpanController = TextEditingController(text: "1,000,000");
  final _powerRBWController = TextEditingController(text: "3,000");
  final _powerVBWController = TextEditingController(text: "1,000");

  final _freqSpanController = TextEditingController(text: "1,000,000");
  final _freqRBWController = TextEditingController(text: "3,000");
  final _freqVBWController = TextEditingController(text: "1,000");

  final _inBandSpanController = TextEditingController(text: "1,000,000");
  final _inBandRBWController = TextEditingController(text: "3,000");
  final _inBandVBWController = TextEditingController(text: "1,000");

  final _outBandSpanController = TextEditingController(text: "1,000,000");
  final _outBandRBWController = TextEditingController(text: "3,000");
  final _outBandVBWController = TextEditingController(text: "1,000");

  // Live Data
  GTxResult? _latestResult;
  final List<String> _logs = [];
  StreamSubscription? _subscription;

  @override
  void dispose() {
    _subscription?.cancel();
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.closeGTxTne();
    _cableLossController.dispose();
    _ifController.dispose();
    _subCarFreqController.dispose();
    _modIndexController.dispose();
    _freqDevController.dispose();
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
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchMetadata();
    });
  }

  void _fetchMetadata() {
    final serverService = Provider.of<ServerService>(context, listen: false);
    final metadata = serverService.status.bootstrapData?.gtxData;

    if (metadata != null && metadata.ok) {
      debugPrint('GTxCharacterizationScreen: Using Bootstrapped Metadata');
      setState(() {
        _metadata = metadata;
        _profiles = metadata.deviceProfile;
        if (_profiles.isNotEmpty &&
            (_selectedProfile == null ||
                !_profiles.contains(_selectedProfile))) {
          _selectedProfile = _profiles.first;
        }
      });
    } else {
      debugPrint('GTxCharacterizationScreen: Bootstrapped Metadata NOT FOUND');
    }
  }

  void _startCharacterization() {
    if (_selectedProfile == null) return;

    setState(() {
      _isMeasuring = true;
      _logs.clear();
      _latestResult = null;
      _logs.add("Characterization Protocol Initiated...");
    });

    final serverService = Provider.of<ServerService>(context, listen: false);

    final request = GTxTneRequest(
      deviceProfile: _selectedProfile!,
      component: _selectedComponent,
      intermediateFrequency:
          double.tryParse(_ifController.text.replaceAll(',', '')) ?? 0.0,
      cableLoss: double.tryParse(_cableLossController.text) ?? 0.0,
      modulationScheme: _selectedModulation,
      subCarrierFrequency:
          double.tryParse(_subCarFreqController.text.replaceAll(',', '')) ??
          0.0,
      modIndex: double.tryParse(_modIndexController.text) ?? 0.0,
      frequencyDeviation:
          double.tryParse(_freqDevController.text.replaceAll(',', '')) ?? 0.0,
      frequencySpectrum: GTxSpectrum(
        span:
            double.tryParse(_freqSpanController.text.replaceAll(',', '')) ??
            1000000,
        rbw:
            double.tryParse(_freqRBWController.text.replaceAll(',', '')) ??
            3000,
        vbw:
            double.tryParse(_freqVBWController.text.replaceAll(',', '')) ??
            1000,
      ),
      powerSpectrum: GTxSpectrum(
        span:
            double.tryParse(_powerSpanController.text.replaceAll(',', '')) ??
            1000000,
        rbw:
            double.tryParse(_powerRBWController.text.replaceAll(',', '')) ??
            3000,
        vbw:
            double.tryParse(_powerVBWController.text.replaceAll(',', '')) ??
            1000,
      ),
      inBandSpectrum: GTxSpectrum(
        span:
            double.tryParse(_inBandSpanController.text.replaceAll(',', '')) ??
            1000000,
        rbw:
            double.tryParse(_inBandRBWController.text.replaceAll(',', '')) ??
            3000,
        vbw:
            double.tryParse(_inBandVBWController.text.replaceAll(',', '')) ??
            1000,
      ),
      outBandSpectrum: GTxSpectrum(
        span:
            double.tryParse(_outBandSpanController.text.replaceAll(',', '')) ??
            1000000,
        rbw:
            double.tryParse(_outBandRBWController.text.replaceAll(',', '')) ??
            3000,
        vbw:
            double.tryParse(_outBandVBWController.text.replaceAll(',', '')) ??
            1000,
      ),
    );

    _subscription = serverService
        .connectGTxTne(request)
        .listen(
          (data) {
            if (data is RTStatus) {
              setState(() {
                _logs.add(data.message);
                if (data.completed) {
                  _isMeasuring = false;
                }
              });
            } else if (data is GTxResult) {
              setState(() {
                _latestResult = data;
              });
            }
          },
          onError: (e) {
            setState(() {
              _logs.add("Critical Error: $e");
              _isMeasuring = false;
            });
          },
          onDone: () {
            setState(() {
              _isMeasuring = false;
            });
          },
        );
  }

  void _stopCharacterization() {
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.abortGTxTne();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ScreenHeader(
                  title: 'GTx Characterization',
                  subtitle: 'Detailed performance analysis of Generator Source',
                  icon: Icons.radar,
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (_metadata != null &&
                          _selectedProfile != null &&
                          _metadata!.deviceMapping.containsKey(
                            _selectedProfile,
                          )) ...[
                        _buildInstrumentStatus(
                          theme,
                          _metadata!.deviceMapping[_selectedProfile!]!.gtxName,
                          true,
                        ),
                        const SizedBox(width: 16),
                        _buildInstrumentStatus(
                          theme,
                          _metadata!.deviceMapping[_selectedProfile!]!.saName,
                          true,
                        ),
                      ] else ...[
                        _buildInstrumentStatus(theme, "GTx-GEN-01", true),
                        const SizedBox(width: 16),
                        _buildInstrumentStatus(theme, "SA-SPEC-04", true),
                      ],
                      const SizedBox(width: 8),
                      IconButton.filledTonal(
                        onPressed: () => setState(
                          () => _showConnections = !_showConnections,
                        ),
                        icon: Icon(
                          _showConnections ? Icons.hub : Icons.hub_outlined,
                        ),
                        tooltip: 'Show Path Diagrams',
                      ),
                      const SizedBox(width: 24),
                      _buildHelpTrigger(theme),
                    ],
                  ),
                ),
                const SizedBox(height: 32),
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildProfileSidebar(theme),
                      const SizedBox(width: 32),
                      Expanded(
                        flex: 3,
                        child: Column(
                          children: [
                            if (_showConnections)
                              _buildConnectionOverlay(theme),
                            Expanded(child: _buildMainConfiguration(theme)),
                          ],
                        ),
                      ),
                      const SizedBox(width: 32),
                      _buildResultsPanel(theme),
                    ],
                  ),
                ),
              ],
            ),
          ),
          // Help Side-Panel
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

  Widget _buildHelpTrigger(ThemeData theme) {
    return IconButton(
      onPressed: () => setState(() => _isHelpOpen = !_isHelpOpen),
      icon: const Icon(Icons.help_outline),
      style: IconButton.styleFrom(
        backgroundColor: _isHelpOpen ? theme.colorScheme.primary : Colors.white,
        foregroundColor: _isHelpOpen ? Colors.white : theme.colorScheme.primary,
        padding: const EdgeInsets.all(12),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  Widget _buildInstrumentStatus(ThemeData theme, String name, bool online) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1C1E),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: online ? Colors.green : Colors.red,
              boxShadow: [
                BoxShadow(
                  color: (online ? Colors.green : Colors.red).withOpacity(0.5),
                  blurRadius: 4,
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Text(
            name,
            style: GoogleFonts.robotoMono(
              color: Colors.blue.shade300,
              fontSize: 12,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildProfileSidebar(ThemeData theme) {
    return Container(
      width: 260,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: theme.colorScheme.primary.withOpacity(0.05),
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
            child: Text(
              'DEVICE PROFILES',
              style: GoogleFonts.inter(
                fontWeight: FontWeight.w900,
                fontSize: 11,
                color: Colors.grey.shade500,
                letterSpacing: 1.2,
              ),
            ),
          ),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              itemCount: _profiles.length,
              itemBuilder: (context, index) {
                final profile = _profiles[index];
                final isSelected = _selectedProfile == profile;
                return Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: InkWell(
                    onTap: () => setState(() => _selectedProfile = profile),
                    borderRadius: BorderRadius.circular(16),
                    child: AnimatedContainer(
                      duration: const Duration(milliseconds: 200),
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
                            Icons.phonelink_setup,
                            size: 18,
                            color: isSelected
                                ? theme.colorScheme.primary
                                : Colors.grey.shade400,
                          ),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Text(
                              profile,
                              overflow: TextOverflow.ellipsis,
                              style: GoogleFonts.inter(
                                fontSize: 14,
                                fontWeight: isSelected
                                    ? FontWeight.bold
                                    : FontWeight.w500,
                                color: isSelected
                                    ? theme.colorScheme.primary
                                    : Colors.black87,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMainConfiguration(ThemeData theme) {
    return Expanded(
      flex: 3,
      child: Container(
        padding: const EdgeInsets.all(32),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(32),
          boxShadow: [
            BoxShadow(
              color: theme.colorScheme.primary.withOpacity(0.03),
              blurRadius: 30,
              offset: const Offset(0, 15),
            ),
          ],
        ),
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildSectionHeader(
                theme,
                "SIGNAL CONFIGURATION",
                Icons.settings_input_component,
              ),
              const SizedBox(height: 24),
              Row(
                children: [
                  Expanded(
                    child: _buildDropdown(
                      label: "Component",
                      value: _selectedComponent,
                      items: _components,
                      onChanged: (val) =>
                          setState(() => _selectedComponent = val!),
                    ),
                  ),
                  const SizedBox(width: 24),
                  Expanded(
                    child: _buildTextField(
                      label: "Cable Loss (dB)",
                      controller: _cableLossController,
                      icon: Icons.linear_scale,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              _buildTextField(
                label: "Intermediate Frequency (Hz)",
                controller: _ifController,
                icon: Icons.radar,
              ),
              const SizedBox(height: 40),
              _buildSectionHeader(theme, "MODULATION PARAMETERS", Icons.waves),
              const SizedBox(height: 24),
              _buildDropdown(
                label: "Modulation Mode",
                value: _selectedModulation,
                items: _modulations,
                onChanged: (val) => setState(() => _selectedModulation = val!),
              ),
              const SizedBox(height: 24),
              _buildModulationSpecificFields(),
              const SizedBox(height: 40),
              _buildSectionHeader(theme, "SPECTRUM SETTINGS", Icons.analytics),
              const SizedBox(height: 24),
              _buildSpectrumTabs(theme),
              const SizedBox(height: 48),
              Center(
                child: SizedBox(
                  width: 300,
                  height: 56,
                  child: ElevatedButton.icon(
                    onPressed: () {
                      if (_isMeasuring) {
                        _stopCharacterization();
                      } else {
                        _startCharacterization();
                      }
                    },
                    icon: _isMeasuring
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Icon(Icons.play_arrow_rounded, size: 28),
                    label: Text(
                      _isMeasuring ? 'STOPPING...' : 'START CHARACTERIZATION',
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: _isMeasuring
                          ? Colors.red
                          : theme.colorScheme.primary,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      elevation: 0,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildResultsPanel(ThemeData theme) {
    return Expanded(
      flex: 2,
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: const Color(0xFF1E1E2D),
              borderRadius: BorderRadius.circular(24),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'RESULTS',
                      style: GoogleFonts.inter(
                        color: Colors.white.withOpacity(0.6),
                        fontSize: 11,
                        fontWeight: FontWeight.w900,
                        letterSpacing: 1.5,
                      ),
                    ),
                    Icon(
                      Icons.analytics_outlined,
                      color: Colors.blue.shade300,
                      size: 20,
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                if (_latestResult == null)
                  Center(
                    child: Padding(
                      padding: const EdgeInsets.only(top: 40),
                      child: Text(
                        'Awaiting Measurement Data...',
                        style: TextStyle(
                          color: Colors.white.withOpacity(0.3),
                          fontSize: 13,
                        ),
                      ),
                    ),
                  )
                else ...[
                  _buildResultRow(
                    "Power Level",
                    _latestResult!.powerMeasurementCompleted
                        ? "${_latestResult!.powerMeasured.toStringAsFixed(2)} dBm"
                        : "Measuring...",
                    _latestResult!.powerMeasurementCompleted
                        ? "COMPLETE"
                        : "PENDING",
                  ),
                  _buildResultRow(
                    "Center Freq",
                    _latestResult!.freqMeasurementCompleted
                        ? "${_latestResult!.freqMeasuredMHz.toStringAsFixed(6)} MHz"
                        : "Measuring...",
                    _latestResult!.freqMeasurementCompleted
                        ? "COMPLETE"
                        : "PENDING",
                  ),
                  if (_latestResult!.inBandSpuriousMeasurementCompleted)
                    _buildResultRow(
                      "In-Band Spurious",
                      _latestResult!.inBandSpuriousFreqOffsetskHz.isEmpty
                          ? "NIL"
                          : "${_latestResult!.inBandPowerOffsets.first.toStringAsFixed(1)} dBc",
                      "COMPLETE",
                    ),
                  if (_latestResult!.outBandSpuriousMeasurementCompleted)
                    _buildResultRow(
                      "Out-Band Spurious",
                      _latestResult!.outBandSpuriousFreqOffsetskHz.isEmpty
                          ? "NIL"
                          : "${_latestResult!.outBandPowerOffsets.first.toStringAsFixed(1)} dBc",
                      "COMPLETE",
                    ),
                  if (_latestResult!.modIndexApplicable)
                    _buildResultRow(
                      "Mod Index",
                      _latestResult!.modIndexMeasurementCompleted
                          ? _latestResult!.modIndexMeasured.toStringAsFixed(2)
                          : "Measuring...",
                      _latestResult!.modIndexMeasurementCompleted
                          ? "COMPLETE"
                          : "PENDING",
                    ),
                  if (_latestResult!.frequencyDeviationApplicable)
                    _buildResultRow(
                      "Freq Dev",
                      _latestResult!.frequencyDeviationMeasurementCompleted
                          ? "${(_latestResult!.frequencyDeviationMeasured / 1000).toStringAsFixed(1)} kHz"
                          : "Measuring...",
                      _latestResult!.frequencyDeviationMeasurementCompleted
                          ? "COMPLETE"
                          : "PENDING",
                    ),
                  if (_latestResult!.harmonicsMeasurementCompleted)
                    _buildResultRow(
                      "2nd Harmonic",
                      _latestResult!.harmonicsFreqMHz.length > 1
                          ? (_latestResult!.harmonicsPresent[1]
                                ? "${_latestResult!.harmonicsMeasureddBm[1].toStringAsFixed(1)} dBm"
                                : "NIL")
                          : "-",
                      "COMPLETE",
                    ),
                ],
              ],
            ),
          ),
          const SizedBox(height: 24),
          Expanded(
            child: Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
                border: Border.all(color: Colors.grey.shade100),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'MEASUREMENT LOG',
                    style: GoogleFonts.inter(
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                      color: Colors.grey.shade400,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildLogEntry("Measurement Session Ready"),
                  const SizedBox(height: 8),
                  Expanded(
                    child: ListView.builder(
                      itemCount: _logs.length,
                      itemBuilder: (context, index) {
                        return _buildLogEntry(_logs[index]);
                      },
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

  Widget _buildResultRow(String parameter, String value, String status) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                parameter,
                style: const TextStyle(color: Colors.white70, fontSize: 12),
              ),
              const SizedBox(height: 4),
              Text(
                value,
                style: GoogleFonts.robotoMono(
                  color: Colors.blue.shade300,
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: status == "COMPLETE"
                  ? Colors.green.withOpacity(0.1)
                  : Colors.orange.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: status == "COMPLETE"
                    ? Colors.green.withOpacity(0.3)
                    : Colors.orange.withOpacity(0.3),
              ),
            ),
            child: Text(
              status,
              style: TextStyle(
                color: status == "COMPLETE" ? Colors.green : Colors.orange,
                fontSize: 10,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLogEntry(String msg, {bool isPending = false}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Icon(
            isPending ? Icons.pending_outlined : Icons.check_circle_outline,
            size: 14,
            color: isPending ? Colors.orange : Colors.green,
          ),
          const SizedBox(width: 12),
          Text(
            msg,
            style: GoogleFonts.inter(fontSize: 13, color: Colors.grey.shade600),
          ),
        ],
      ),
    );
  }

  int _selectedSpectrumTab = 0;
  final List<String> _spectrumTabs = [
    "Power",
    "Frequency",
    "In-Band",
    "Out-Band",
  ];

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

  Widget _buildSectionHeader(ThemeData theme, String title, IconData icon) {
    return Row(
      children: [
        Icon(icon, size: 18, color: theme.colorScheme.primary),
        const SizedBox(width: 12),
        Text(
          title,
          style: GoogleFonts.inter(
            fontWeight: FontWeight.w900,
            fontSize: 12,
            color: theme.colorScheme.primary,
            letterSpacing: 1.5,
          ),
        ),
      ],
    );
  }

  Widget _buildTextField({
    required String label,
    required TextEditingController controller,
    required IconData icon,
  }) {
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        labelText: label,
        prefixIcon: Icon(icon, size: 20),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
        floatingLabelBehavior: FloatingLabelBehavior.always,
      ),
    );
  }

  Widget _buildDropdown({
    required String label,
    required String value,
    required List<String> items,
    required Function(String?) onChanged,
  }) {
    return DropdownButtonFormField<String>(
      value: value,
      items: items
          .map((e) => DropdownMenuItem(value: e, child: Text(e)))
          .toList(),
      onChanged: onChanged,
      decoration: InputDecoration(
        labelText: label,
        prefixIcon: const Icon(Icons.layers_outlined, size: 20),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
        floatingLabelBehavior: FloatingLabelBehavior.always,
      ),
    );
  }

  Widget _buildModulationSpecificFields() {
    if (_selectedModulation == "PM" || _selectedModulation == "FM") {
      return Row(
        children: [
          Expanded(
            child: _buildTextField(
              label: "Sub Carrier Freq (Hz)",
              controller: _subCarFreqController,
              icon: Icons.settings_input_composite,
            ),
          ),
          const SizedBox(width: 24),
          Expanded(
            child: _selectedModulation == "PM"
                ? _buildTextField(
                    label: "Modulation Index",
                    controller: _modIndexController,
                    icon: Icons.tune,
                  )
                : _buildTextField(
                    label: "Freq Deviation (Hz)",
                    controller: _freqDevController,
                    icon: Icons.compare_arrows,
                  ),
          ),
        ],
      );
    }
    return const SizedBox.shrink();
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
                  'GTx Help',
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
                  'Modulation Profiles',
                  'Characterization depends on the modulation scheme (PM, FM, PSK, etc.). The system calculates '
                      'parameters like Modulation Index or Frequency Deviation based on carrier/sub-carrier power ratios.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Spectrum Automation',
                  'PRISM automatically scales the Spectrum Analyzer span and resolution bandwidth (RBW) for each test '
                      'phase (Power, Spurious, Harmonics) to ensure the noise floor does not mask critical spurs.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Spurious Detection',
                  'In-Band and Out-Band spurious tests sweep the spectrum around the carrier. Any peak detected '
                      'above the defined threshold is recorded with its absolute power and its offset in dBc.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Log Feedback',
                  'The "Measurement Log" at the bottom right provides real-time SCPI command status and hardware '
                      'settling events. Check this if measurements appear stuck.',
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
                'GTx Connection Guide',
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
            'Connect the Ground Transmitter (GTx) output directly to the Spectrum Analyzer. Any offset required for high-power protection must be entered in the "Cable Loss" field.',
          ),
          const SizedBox(height: 16),
          const AspectRatio(
            aspectRatio: 3 / 1,
            child: InstrumentConnectionDiagram(type: DiagramType.attnGTx),
          ),
        ],
      ),
    );
  }
}
