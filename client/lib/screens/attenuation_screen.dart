import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/utils/notifications.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/instrument_connection_diagram.dart';

class AttenuationScreen extends StatefulWidget {
  const AttenuationScreen({super.key});

  @override
  State<AttenuationScreen> createState() => _AttenuationScreenState();
}

class _AttenuationScreenState extends State<AttenuationScreen> {
  // Selection State
  String _selectedInstrument = 'TSM'; // TSM, GTx, SG

  // Configuration State
  final TextEditingController _minValueController = TextEditingController(
    text: '0',
  );
  final TextEditingController _maxValueController = TextEditingController(
    text: '63.5',
  );
  final TextEditingController _stepSizeController = TextEditingController(
    text: '0.1',
  );

  String _selectedDeviceProfile = 'Profile 1';
  String _selectedReceiver = 'Receiver 1';
  String _selectedSpectrum = 'Spectrum 1';
  String _selectedTSMConfig = 'Path A';
  String _selectedComponent = 'IFM-1';

  // Measurement State
  bool _isMeasuring = false;
  List<MeasurementPoint> _measuredPoints = [];
  List<MeasurementPoint> _correctedPoints = [];
  bool _showCorrected = false;
  bool _showConnectionImage = false;

  // Real Metadata Lists
  List<String> _deviceProfiles = [];
  List<String> _receivers = [];
  List<String> _spectrumProfiles = [];
  List<String> _tsmConfigs = [];
  List<String> _gtxComponents = [];
  Map<String, AttnRange> _serverRanges = {};
  bool _isLoading = true;
  String _lastStatusMessage = '';
  bool _isHelpOpen = false;

  // Subscription
  StreamSubscription? _attnSubscription;

  @override
  void initState() {
    super.initState();
    _resetToDefaults();
    _fetchMetadata();
  }

  void _fetchMetadata() {
    setState(() => _isLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    final metadata = serverService.status.bootstrapData?.attnData;

    if (metadata != null) {
      debugPrint('AttenuationScreen: Using Bootstrapped Metadata');
      setState(() {
        _deviceProfiles = metadata.deviceProfile;
        _receivers = metadata.receiver;
        _spectrumProfiles = metadata.sprectrumProfile;
        _tsmConfigs = metadata.tsmConfig;
        _gtxComponents = metadata.gtxComponents;
        _serverRanges = metadata.attnRanges;

        // Set initial selections if available
        if (_deviceProfiles.isNotEmpty &&
            (_selectedDeviceProfile == 'Profile 1' ||
                !_deviceProfiles.contains(_selectedDeviceProfile))) {
          _selectedDeviceProfile = _deviceProfiles.first;
        }
        if (_receivers.isNotEmpty &&
            (_selectedReceiver == 'Receiver 1' ||
                !_receivers.contains(_selectedReceiver))) {
          _selectedReceiver = _receivers.first;
        }
        if (_spectrumProfiles.isNotEmpty &&
            (_selectedSpectrum == 'Spectrum 1' ||
                !_spectrumProfiles.contains(_selectedSpectrum))) {
          _selectedSpectrum = _spectrumProfiles.first;
        }
        if (_tsmConfigs.isNotEmpty &&
            (_selectedTSMConfig == 'Path A' ||
                !_tsmConfigs.contains(_selectedTSMConfig))) {
          _selectedTSMConfig = _tsmConfigs.first;
        }
        if (_gtxComponents.isNotEmpty &&
            (_selectedComponent == 'IFM-1' ||
                !_gtxComponents.contains(_selectedComponent))) {
          _selectedComponent = _gtxComponents.first;
        }

        // Apply initial ranges based on current instrument
        _applyServerRange(_selectedInstrument);
        _isLoading = false;
      });
    } else {
      debugPrint('AttenuationScreen: Bootstrapped Metadata NOT FOUND');
      setState(() => _isLoading = false);
    }
  }

  void _resetToDefaults() {
    setState(() {
      if (_serverRanges.containsKey(_selectedInstrument)) {
        _applyServerRange(_selectedInstrument);
      } else {
        // Fallback defaults if server metadata is missing or empty
        if (_selectedInstrument == 'TSM') {
          _minValueController.text = '0';
          _maxValueController.text = '63.5';
          _stepSizeController.text = '0.5';
        } else if (_selectedInstrument == 'SG') {
          _minValueController.text = '-80';
          _maxValueController.text = '0';
          _stepSizeController.text = '1.0';
        } else if (_selectedInstrument == 'GTx') {
          _minValueController.text = '-50';
          _maxValueController.text = '0';
          _stepSizeController.text = '0.1';
        }
      }
      _measuredPoints = [];
      _correctedPoints = [];
      _showCorrected = false;
    });
  }

  void _applyServerRange(String instrument) {
    if (_serverRanges.containsKey(instrument)) {
      final range = _serverRanges[instrument]!;
      _minValueController.text = range.min.toString();
      _maxValueController.text = range.max.toString();
      _stepSizeController.text = range.stepSize.toString();
    }
  }

  void _startMeasurement() {
    final serverService = Provider.of<ServerService>(context, listen: false);

    setState(() {
      _isMeasuring = true;
      _measuredPoints = [];
      _correctedPoints = [];
      _showCorrected = false;
      _lastStatusMessage = 'Initiating measurement...';
    });

    final request = {
      'Type': _selectedInstrument,
      'DeviceProfile': _selectedDeviceProfile,
      'Receiver': _selectedReceiver,
      'SpectrumProfile': _selectedSpectrum,
      'TSMConfig': _selectedTSMConfig,
      'Component': _selectedComponent,
      'Min': double.tryParse(_minValueController.text) ?? 0.0,
      'Max': double.tryParse(_maxValueController.text) ?? 0.0,
      'Step': double.tryParse(_stepSizeController.text) ?? 0.1,
    };

    _attnSubscription = serverService
        .streamAttnAction(request)
        .listen(
          (response) {
            if (!response.ok) {
              setState(() => _isMeasuring = false);
              AppNotifications.showError(context, response.message);
              return;
            }

            final status = response.measurementStatus;
            final deviations = response.deviations;

            setState(() {
              if (status != null) {
                _lastStatusMessage = status.message;
                if (status.error) {
                  _isMeasuring = false;
                  AppNotifications.showError(context, status.message);
                  return;
                }

                if (status.hasData) {
                  _measuredPoints.add(
                    MeasurementPoint(
                      index: status.slNo,
                      setVal: status.setAttn,
                      measuredVal: status.measuredAttn,
                      deviation: status.deviation,
                      showInPlot: status.plotDeviation,
                    ),
                  );
                }

                if (status.completed) {
                  _isMeasuring = false;
                }
              }

              if (deviations != null && deviations.isNotEmpty) {
                _showCorrected = true;
                _correctedPoints = deviations.map((d) {
                  return MeasurementPoint(
                    index: 0,
                    setVal: d.setValue,
                    measuredVal: d.setValue - d.correctedDeviation,
                    deviation: d.correctedDeviation,
                    showInPlot: true,
                  );
                }).toList();
              }
            });
          },
          onError: (e) {
            setState(() => _isMeasuring = false);
          },
          onDone: () {
            setState(() {
              _isMeasuring = false;
              _attnSubscription = null;
            });
          },
        );
  }

  // Removed _finalizeMeasurement as data is now sent via WebSocket consolidated response

  void _stopMeasurement() {
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.abortAttnMeasurement();
    // We don't cancel the subscription or set _isMeasuring = false here.
    // We wait for the server to send the final status and close the stream,
    // which will be handled by the onDone callback in _startMeasurement.
  }

  @override
  void dispose() {
    _attnSubscription?.cancel();
    _minValueController.dispose();
    _maxValueController.dispose();
    _stepSizeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Stack(
        children: [
          Column(
            children: [
              ScreenHeader(
                title: 'Attenuation Measurement',
                subtitle:
                    'Configure and measure attenuation across different instruments',
                icon: Icons.graphic_eq, // Audio wave / signal icon
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    IconButton.filledTonal(
                      onPressed: () => setState(
                        () => _showConnectionImage = !_showConnectionImage,
                      ),
                      icon: Icon(
                        _showConnectionImage ? Icons.hub : Icons.hub_outlined,
                      ),
                      tooltip: 'Show Connection Diagrams',
                    ),
                    const SizedBox(width: 12),
                    _buildHelpTrigger(theme),
                  ],
                ),
              ),
              Expanded(
                child: _isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : Row(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          // Left Sidebar: Instrument Selection
                          _buildInstrumentSidebar(theme),

                          // Center: Configuration
                          Expanded(
                            flex: 3,
                            child: Column(
                              children: [
                                if (_showConnectionImage)
                                  _buildConnectionOverlay(theme),
                                Expanded(
                                  child: _buildConfigurationPanel(theme),
                                ),
                              ],
                            ),
                          ),

                          // Right: Results & Analytics
                          Expanded(flex: 5, child: _buildResultsPanel(theme)),
                        ],
                      ),
              ),
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
    );
  }

  Widget _buildInstrumentSidebar(ThemeData theme) {
    return ContentCard(
      isSidebar: true,
      width: 200,
      margin: const EdgeInsets.only(left: 24, top: 12, bottom: 24),
      padding: EdgeInsets.zero, // Sidebar content has own padding
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(20),
            child: Text(
              'Instruments',
              style: GoogleFonts.outfit(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: theme.colorScheme.primary,
              ),
            ),
          ),
          const Divider(),
          _buildInstrumentTile('TSM', Icons.settings_suggest, theme),
          _buildInstrumentTile('GTx', Icons.sensors, theme),
          _buildInstrumentTile('SG', Icons.waves, theme),
          const Spacer(),
          Padding(
            padding: const EdgeInsets.all(20),
            child: Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: theme.colorScheme.primary.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  Icon(
                    Icons.info_outline,
                    color: theme.colorScheme.primary,
                    size: 20,
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Select an instrument to begin configuration',
                    textAlign: TextAlign.center,
                    style: TextStyle(fontSize: 11, color: Colors.grey),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInstrumentTile(String label, IconData icon, ThemeData theme) {
    bool isSelected = _selectedInstrument == label;
    return InkWell(
      onTap: () {
        setState(() => _selectedInstrument = label);
        _resetToDefaults();
      },
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? theme.colorScheme.primary : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(
              icon,
              color: isSelected ? Colors.white : Colors.grey,
              size: 20,
            ),
            const SizedBox(width: 12),
            Text(
              label,
              style: TextStyle(
                color: isSelected ? Colors.white : Colors.black87,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildConfigurationPanel(ThemeData theme) {
    return ContentCard(
      margin: const EdgeInsets.symmetric(
        horizontal: 24,
        vertical: 12,
      ).copyWith(bottom: 24),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Configuration',
                  style: GoogleFonts.outfit(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),

            _buildDropdown(
              'Device Profile',
              _selectedDeviceProfile,
              _deviceProfiles.isNotEmpty
                  ? _deviceProfiles
                  : [_selectedDeviceProfile],
              (v) => setState(() => _selectedDeviceProfile = v!),
            ),
            _buildDropdown(
              'Receiver',
              _selectedReceiver,
              _receivers.isNotEmpty ? _receivers : [_selectedReceiver],
              (v) => setState(() => _selectedReceiver = v!),
            ),
            _buildDropdown(
              'Spectrum Profile',
              _selectedSpectrum,
              _spectrumProfiles.isNotEmpty
                  ? _spectrumProfiles
                  : [_selectedSpectrum],
              (v) => setState(() => _selectedSpectrum = v!),
            ),

            if (_selectedInstrument == 'TSM')
              _buildDropdown(
                'TSM Config',
                _selectedTSMConfig,
                _tsmConfigs.isNotEmpty ? _tsmConfigs : [_selectedTSMConfig],
                (v) => setState(() => _selectedTSMConfig = v!),
              ),

            if (_selectedInstrument == 'GTx')
              _buildDropdown(
                'Component',
                _selectedComponent,
                _gtxComponents.isNotEmpty
                    ? _gtxComponents
                    : [_selectedComponent],
                (v) => setState(() => _selectedComponent = v!),
              ),

            const Divider(height: 40),

            _buildTextField('Min Value', _minValueController, 'dBm'),
            const SizedBox(height: 16),
            _buildTextField('Max Value', _maxValueController, 'dBm'),
            const SizedBox(height: 16),
            _buildTextField('Step Size', _stepSizeController, 'dB'),

            if (_isMeasuring) ...[
              const SizedBox(height: 24),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: theme.colorScheme.primary.withOpacity(0.05),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: theme.colorScheme.primary.withOpacity(0.1),
                  ),
                ),
                child: Column(
                  children: [
                    Row(
                      children: [
                        SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: theme.colorScheme.primary,
                          ),
                        ),
                        const SizedBox(width: 12),
                        const Expanded(
                          child: Text(
                            'Measurement Active',
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Text(
                      _lastStatusMessage,
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey.shade700,
                        fontStyle: FontStyle.italic,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ],
                ),
              ),
            ],

            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _isMeasuring ? _stopMeasurement : _startMeasurement,
                icon: Icon(_isMeasuring ? Icons.stop : Icons.play_arrow),
                label: Text(
                  _isMeasuring ? 'Abort Measurement' : 'Start Measurement',
                ),
                style: ElevatedButton.styleFrom(
                  backgroundColor: _isMeasuring
                      ? Colors.redAccent
                      : theme.colorScheme.primary,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 20),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDropdown(
    String label,
    String value,
    List<String> items,
    ValueChanged<String?> onChanged,
  ) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.bold,
              color: Colors.grey,
            ),
          ),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            value: value,
            items: items
                .map((e) => DropdownMenuItem(value: e, child: Text(e)))
                .toList(),
            onChanged: onChanged,
            decoration: InputDecoration(
              filled: true,
              fillColor: Colors.grey.shade50,
              contentPadding: const EdgeInsets.symmetric(horizontal: 16),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: BorderSide(color: Colors.grey.shade200),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: BorderSide(color: Colors.grey.shade200),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTextField(
    String label,
    TextEditingController controller,
    String suffix,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.bold,
            color: Colors.grey,
          ),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: controller,
          decoration: InputDecoration(
            suffixText: suffix,
            filled: true,
            fillColor: Colors.grey.shade50,
            contentPadding: const EdgeInsets.symmetric(horizontal: 16),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
          ),
          keyboardType: TextInputType.number,
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
                  'Attenuation Help',
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
                  'Instrument Modes',
                  '• **TSM**: Measures attenuation through the TSM system components.\n'
                      '• **GTx**: Specialized testing for GTx internal modules.\n'
                      '• **SG**: Standard Signal Generator attenuation sweeps.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Range Calibration',
                  'Define the sweep range (Min/Max) and the Step size. PRISM will calculate the necessary command '
                      'sequence and execute them with sub-millisecond precision from the server.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Live Visualization',
                  'The chart displays real-time power deviation. A "Corrected View" may appear after measurement '
                      'tasks complete, showing post-calculated offsets for higher calibration accuracy.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Abort Safety',
                  'If you need to stop a sweep, use the "Abort" button. The system will safely ramp down '
                      'the signal generator or reset the TSM routes before closing the connection.',
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

  Widget _buildResultsPanel(ThemeData theme) {
    return Padding(
      padding: const EdgeInsets.only(right: 24, top: 12, bottom: 24),
      child: Column(
        children: [
          // Top: Chart
          Expanded(
            flex: 4,
            child: ContentCard(
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
                            'Measurement Visualization',
                            style: GoogleFonts.outfit(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          Text(
                            'Real-time power deviation analysis',
                            style: TextStyle(color: Colors.grey, fontSize: 13),
                          ),
                        ],
                      ),
                      if (_showCorrected)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.green.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Row(
                            children: [
                              const Icon(
                                Icons.check_circle,
                                color: Colors.green,
                                size: 16,
                              ),
                              const SizedBox(width: 8),
                              const Text(
                                'Corrected View Enabled',
                                style: TextStyle(
                                  color: Colors.green,
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ],
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 24),
                  Expanded(child: _buildChart(theme)),
                ],
              ),
            ),
          ),

          const SizedBox(height: 12),

          // Bottom: Table
          Expanded(
            flex: 4,
            child: ContentCard(
              width: double.infinity,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Detailed Results',
                    style: GoogleFonts.outfit(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),
                  Expanded(child: _buildTable(theme)),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChart(ThemeData theme) {
    if (_measuredPoints.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.show_chart, size: 64, color: Colors.grey.shade200),
            const SizedBox(height: 16),
            const Text(
              'Start a measurement to view live trace',
              style: TextStyle(color: Colors.grey),
            ),
          ],
        ),
      );
    }

    return LineChart(
      LineChartData(
        gridData: FlGridData(
          show: true,
          drawVerticalLine: true,
          getDrawingHorizontalLine: (v) =>
              FlLine(color: Colors.grey.shade100, strokeWidth: 1),
        ),
        titlesData: FlTitlesData(
          show: true,
          rightTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
          topTitles: AxisTitles(sideTitles: SideTitles(showTitles: false)),
          bottomTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 30,
              getTitlesWidget: (value, meta) => Text(
                value.toStringAsFixed(1),
                style: const TextStyle(fontSize: 10, color: Colors.grey),
              ),
            ),
          ),
          leftTitles: AxisTitles(
            axisNameWidget: const Text(
              'Deviation (dB)',
              style: TextStyle(fontSize: 12, color: Colors.grey),
            ),
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 40,
              getTitlesWidget: (value, meta) => Text(
                value.toStringAsFixed(2),
                style: const TextStyle(fontSize: 10, color: Colors.grey),
              ),
            ),
          ),
        ),
        borderData: FlBorderData(
          show: true,
          border: Border.all(color: Colors.grey.shade200),
        ),
        lineBarsData: [
          // Measured Line
          LineChartBarData(
            spots: _measuredPoints
                .where((p) => p.showInPlot)
                .map((p) => FlSpot(p.setVal, p.deviation))
                .toList(),
            isCurved: true,
            color: _showCorrected
                ? theme.colorScheme.primary.withOpacity(0.3)
                : theme.colorScheme.primary,
            barWidth: 3,
            isStrokeCapRound: true,
            dotData: FlDotData(show: false),
            belowBarData: BarAreaData(
              show: true,
              color: theme.colorScheme.primary.withOpacity(0.05),
            ),
          ),
          // Corrected Line
          if (_showCorrected)
            LineChartBarData(
              spots: _correctedPoints
                  .where((p) => p.showInPlot)
                  .map((p) => FlSpot(p.setVal, p.deviation))
                  .toList(),
              isCurved: true,
              color: Colors.green,
              barWidth: 3,
              isStrokeCapRound: true,
              dotData: FlDotData(show: false),
              belowBarData: BarAreaData(show: false),
            ),
        ],
      ),
    );
  }

  Widget _buildTable(ThemeData theme) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return SingleChildScrollView(
          scrollDirection: Axis.vertical,
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: ConstrainedBox(
              constraints: BoxConstraints(minWidth: constraints.maxWidth),
              child: DataTable(
                headingRowColor: MaterialStateProperty.all(
                  theme.colorScheme.primary.withOpacity(0.05),
                ),
                columnSpacing: 24,
                columns: [
                  const DataColumn(label: Text('Sl. No')),
                  DataColumn(
                    label: Text(
                      _selectedInstrument == 'TSM'
                          ? 'Set Attn (dB)'
                          : 'Set Power (dBm)',
                    ),
                  ),
                  DataColumn(
                    label: Text(
                      _selectedInstrument == 'TSM'
                          ? 'Measured (dB)'
                          : 'Measured (dBm)',
                    ),
                  ),
                  const DataColumn(label: Text('Deviation (dB)')),
                ],
                rows: _measuredPoints
                    .map(
                      (p) => DataRow(
                        cells: [
                          DataCell(Text(p.index.toString())),
                          DataCell(Text(p.setVal.toStringAsFixed(3))),
                          DataCell(Text(p.measuredVal.toStringAsFixed(3))),
                          DataCell(
                            Text(
                              p.deviation.toStringAsFixed(3),
                              style: TextStyle(
                                color: p.deviation.abs() > 0.5
                                    ? Colors.red
                                    : Colors.green,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ],
                      ),
                    )
                    .toList(),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildConnectionOverlay(ThemeData theme) {
    DiagramType dType = DiagramType.attnSG;
    String guideText = '';

    if (_selectedInstrument == 'SG') {
      dType = DiagramType.attnSG;
      guideText =
          'Connect Signal Generator output directly to Spectrum Analyzer.';
    } else if (_selectedInstrument == 'GTx') {
      dType = DiagramType.attnGTx;
      guideText =
          'Connect Ground Transmitter output directly to Spectrum Analyzer.';
    } else if (_selectedInstrument == 'TSM') {
      dType = DiagramType.attnTSM;
      guideText =
          'Connect Signal Generator to TSM input port and Receiver Output to Spectrum Analyzer.';
    }

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
      child: ContentCard(
        color: theme.colorScheme.primaryContainer.withOpacity(0.4),
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Connection Guide: $_selectedInstrument',
                  style: GoogleFonts.outfit(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: theme.colorScheme.onPrimaryContainer,
                  ),
                ),
                IconButton(
                  onPressed: () => setState(() => _showConnectionImage = false),
                  icon: const Icon(Icons.close),
                ),
              ],
            ),
            const SizedBox(height: 16),
            AspectRatio(
              aspectRatio: 3 / 1,
              child: InstrumentConnectionDiagram(type: dType),
            ),
            const SizedBox(height: 16),
            Text(
              guideText,
              style: TextStyle(
                fontSize: 13,
                color: theme.colorScheme.onPrimaryContainer.withOpacity(0.7),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class MeasurementPoint {
  final int index;
  final double setVal;
  final double measuredVal;
  final double deviation;
  final bool showInPlot;

  MeasurementPoint({
    required this.index,
    required this.setVal,
    required this.measuredVal,
    required this.deviation,
    this.showInPlot = true,
  });
}
