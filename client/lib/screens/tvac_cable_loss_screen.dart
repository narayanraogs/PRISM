import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/instrument_connection_diagram.dart';

class TVACCableLossScreen extends StatefulWidget {
  final bool isActive;
  const TVACCableLossScreen({super.key, this.isActive = true});

  @override
  State<TVACCableLossScreen> createState() => _TVACCableLossScreenState();
}

class _TVACCableLossScreenState extends State<TVACCableLossScreen> {
  @override
  void didUpdateWidget(TVACCableLossScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.isActive && !widget.isActive) {
      final serverService = Provider.of<ServerService>(context, listen: false);
      serverService.closeTVACCableLoss();
    }
  }

  // Form Controllers
  final TextEditingController _cableNameController = TextEditingController();
  final TextEditingController _cycleNameController = TextEditingController();

  // State variables
  String _selectedPhase = 'Ambient';
  String _selectedDeviceProfile = '';
  String _pmChannel = 'A';
  bool _isPmReferenced = false;
  bool _isMeasuring = false;
  bool _isLoading = true;
  String _measuringStatus = '';
  bool _showGraph = false;
  bool _showConnections = false;
  bool _isHelpOpen = false;

  List<String> _cableSuggestions = [];
  List<String> _deviceProfiles = [];
  List<double> _frequencies = [];
  List<TVACCableLossRecord> _history = [];

  @override
  void initState() {
    super.initState();
    _loadMetadata();
  }

  void _loadMetadata() {
    setState(() => _isLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    final metadata = serverService.status.bootstrapData?.tvacCableLossData;

    if (metadata != null) {
      debugPrint('TVACCableLossScreen: Using Bootstrapped Metadata');
      setState(() {
        _frequencies = metadata.frequencies;
        _deviceProfiles = metadata.deviceProfiles;
        _cableSuggestions = metadata.existingCables;
        _isPmReferenced = metadata.isPmZeroed;
        if (_deviceProfiles.isNotEmpty) {
          _selectedDeviceProfile = _deviceProfiles.first;
        }
      });
      _loadHistory();
      setState(() => _isLoading = false);
    } else {
      debugPrint('TVACCableLossScreen: Bootstrapped Metadata NOT FOUND');
      setState(() => _isLoading = false);
      if (mounted) {
        Provider.of<NotificationService>(
          context,
          listen: false,
        ).addNotification(
          title: 'Error',
          message: 'Failed to load metadata',
          type: NotificationType.error,
        );
      }
    }
  }

  Future<void> _loadHistory() async {
    final serverService = Provider.of<ServerService>(context, listen: false);
    final response = await serverService.fetchTVACCableMeasuredDetails();
    if (response != null && response.ok) {
      setState(() {
        _history = response.history;
        _isPmReferenced = response.isPmZeroed;
        _cableSuggestions = response.history
            .map((e) => e.cableName)
            .toSet()
            .toList();
      });
    }
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

  void _performAction(String action) {
    if (_cableNameController.text.trim().isEmpty && action != 'pmreference') {
      _showAppNotification(
        title: 'Validation Error',
        message: 'Please enter a Cable Name',
        type: NotificationType.error,
      );
      return;
    }

    setState(() {
      _isMeasuring = true;
      _measuringStatus = 'Initializing measurement...';
    });
    final serverService = Provider.of<ServerService>(context, listen: false);

    final request = {
      'action': action,
      'deviceProfile': _selectedDeviceProfile,
      'channel': _pmChannel,
      'cableName': _cableNameController.text.trim(),
      'cycleName': _cycleNameController.text.trim(),
      'phase': _selectedPhase,
    };

    serverService
        .streamTVACCableLossAction(request)
        .listen(
          (status) {
            if (status.error) {
              if (mounted) {
                context.read<NotificationService>().addNotification(
                  title: 'Measurement Error',
                  message: status.message,
                  type: NotificationType.error,
                );
                setState(() {
                  _isMeasuring = false;
                  _measuringStatus = '';
                });
              }
            } else if (status.message == 'Measurement Completed') {
              if (mounted) {
                context.read<NotificationService>().addNotification(
                  title: 'Success',
                  message: 'Action completed successfully',
                  type: NotificationType.success,
                );
                setState(() {
                  _isMeasuring = false;
                  _measuringStatus = '';
                });
                _loadHistory();
              }
            } else {
              // Progress update
              if (mounted) {
                setState(() => _measuringStatus = status.message);
              }
            }
          },
          onError: (e) {
            if (mounted) {
              setState(() {
                _isMeasuring = false;
                _measuringStatus = '';
              });
              _showAppNotification(
                title: 'Connection Error',
                message: e.toString(),
                type: NotificationType.error,
              );
            }
          },
          onDone: () {
            if (mounted) {
              setState(() {
                _isMeasuring = false;
                _measuringStatus = '';
              });
            }
          },
        );
  }

  @override
  void dispose() {
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.closeTVACCableLoss();
    _cableNameController.dispose();
    _cycleNameController.dispose();
    super.dispose();
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
                title: 'TVAC Cable Loss Measurement',
                subtitle: 'Measure and record cable loss in TVAC environment',
                icon: Icons.thermostat,
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    IconButton.filledTonal(
                      onPressed: () =>
                          setState(() => _showConnections = !_showConnections),
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
              Expanded(
                child: _isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : Stack(
                        children: [
                          SingleChildScrollView(
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
                          if (_isMeasuring)
                            Container(
                              color: Colors.black.withOpacity(0.3),
                              child: Center(
                                child: ContentCard(
                                  width: 400,
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 32,
                                    vertical: 24,
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
                                      const SizedBox(height: 8),
                                      const Text(
                                        'This may take a few minutes...',
                                        style: TextStyle(
                                          fontSize: 12,
                                          color: Colors.grey,
                                        ),
                                      ),
                                      const SizedBox(height: 20),
                                      ElevatedButton.icon(
                                        onPressed: () {
                                          context
                                              .read<ServerService>()
                                              .abortTVACCableLossMeasurement();
                                          setState(() {
                                            _isMeasuring = false;
                                            _measuringStatus = '';
                                          });
                                        },
                                        icon: const Icon(Icons.stop, size: 18),
                                        label: const Text('Abort Measurement'),
                                        style: ElevatedButton.styleFrom(
                                          backgroundColor: Colors.red.shade50,
                                          foregroundColor: Colors.red,
                                          elevation: 0,
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 20,
                                            vertical: 12,
                                          ),
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                            side: BorderSide(
                                              color: Colors.red.shade100,
                                            ),
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
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
                  'TVAC Measurement Help',
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
                  'Thermal Vacuum Testing',
                  'TVAC cable loss measurements compensate for attenuation variations caused by extreme temperature cycles. '
                      'Data is tracked across Ambient, Hot, and Cold phases.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Reference vs Loss',
                  '• **Reference**: The first measurement for a new cable. It sets the baseline (0 dB loss).\n'
                      '• **Cable Loss**: Subsequent measurements are compared against the reference to determine actual attenuation.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Power Meter Zeroing',
                  'Always perform a "Zero PM" action before starting any measurement campaign. '
                      'This removes internal noise offsets and ensures sub-0.1 dB accuracy.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Connection Diagrams',
                  'Click the "Hub" icon in the header or check the "Show Connection Diagrams" button '
                      'to see exactly how to wire the equipment for each stage.',
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
    return ContentCard(
      color: theme.colorScheme.primaryContainer.withOpacity(0.4),
      padding: const EdgeInsets.all(24),
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
    String type,
    String desc,
    IconData icon,
  ) {
    return Expanded(
      child: ContentCard(
        padding: const EdgeInsets.all(16),
        borderRadius: 16,
        color: theme.colorScheme.surface,
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
            icon: _isMeasuring
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.white,
                    ),
                  )
                : const Icon(Icons.refresh, size: 16),
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
        isSidebar: true, // Use simpler style
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
    bool isNewCable =
        !_cableSuggestions.contains(_cableNameController.text.trim()) &&
        _cableNameController.text.trim().isNotEmpty;

    return ContentCard(
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

          // Cable Name Autocomplete
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
                  if (_cableNameController.text != controller.text &&
                      _cableNameController.text.isNotEmpty) {
                    // Keep sync if needed
                  }
                  return TextField(
                    controller: controller,
                    focusNode: focusNode,
                    decoration: _inputDecoration(
                      'e.g. RF-CABLE-01',
                      Icons.cable,
                    ),
                    onChanged: (val) => setState(() {
                      _cableNameController.text = val;
                    }),
                  );
                },
          ),

          if (isNewCable)
            Padding(
              padding: const EdgeInsets.only(top: 8.0),
              child: Row(
                children: [
                  Icon(
                    Icons.new_releases,
                    size: 14,
                    color: theme.colorScheme.primary,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    'New Cable Detected: First measurement will be Reference',
                    style: TextStyle(
                      fontSize: 11,
                      color: theme.colorScheme.primary,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),

          const SizedBox(height: 20),

          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildFieldLabel('Cycle Name'),
                    TextField(
                      controller: _cycleNameController,
                      decoration: _inputDecoration('e.g. Cycle 1', Icons.loop),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildFieldLabel('PM Profile'),
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
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildFieldLabel('PM Channel'),
                    DropdownButtonFormField<String>(
                      value: _pmChannel,
                      items: ['A', 'B']
                          .map(
                            (e) => DropdownMenuItem(
                              value: e,
                              child: Text('Channel $e'),
                            ),
                          )
                          .toList(),
                      onChanged: (val) => setState(() => _pmChannel = val!),
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
          _buildFieldLabel('Test Phase'),
          SegmentedButton<String>(
            segments: const [
              ButtonSegment(
                value: 'Ambient',
                label: Text('Ambient'),
                icon: Icon(Icons.thermostat, size: 16),
              ),
              ButtonSegment(
                value: 'Hot',
                label: Text('Hot'),
                icon: Icon(Icons.wb_sunny, size: 16),
              ),
              ButtonSegment(
                value: 'Cold',
                label: Text('Cold'),
                icon: Icon(Icons.ac_unit, size: 16),
              ),
            ],
            selected: {_selectedPhase},
            onSelectionChanged: (Set<String> newSelection) {
              setState(() {
                _selectedPhase = newSelection.first;
              });
            },
            style: ButtonStyle(visualDensity: VisualDensity.comfortable),
          ),

          const SizedBox(height: 20),
          _buildFieldLabel('Sweep Plan'),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _frequencies.map((f) {
              return Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 5,
                ),
                decoration: BoxDecoration(
                  color: theme.colorScheme.primary.withOpacity(0.05),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: theme.colorScheme.primary.withOpacity(0.1),
                  ),
                ),
                child: Text(
                  '${f.toStringAsFixed(1)} GHz',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: theme.colorScheme.primary,
                  ),
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
                      : () {
                          final isNewCable = !_cableSuggestions.contains(
                            _cableNameController.text.trim(),
                          );
                          _performAction(
                            isNewCable ? 'cablereference' : 'cableloss',
                          );
                        },
                  icon: _isMeasuring
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Colors.white,
                          ),
                        )
                      : const Icon(Icons.play_arrow),
                  label: Text(
                    _isMeasuring ? 'Measuring...' : 'Start Measurement',
                  ),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 20),
                    backgroundColor: theme.colorScheme.primary,
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
                  style: TextStyle(
                    color: Colors.red.shade400,
                    fontSize: 12,
                    fontWeight: FontWeight.normal,
                  ),
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
    if (_history.isEmpty) {
      return Container(
        height: 300,
        decoration: BoxDecoration(
          color: theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(24),
        ),
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.query_stats, size: 48, color: Colors.grey.shade300),
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
    final lastMeas = _history.last;

    return Container(
      padding: const EdgeInsets.all(20),
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
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Latest Measurement Result',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    'Cable: ${lastMeas.cableName} | ${lastMeas.phase}',
                    style: TextStyle(color: Colors.grey.shade600, fontSize: 13),
                  ),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: lastMeas.phase == 'Hot'
                      ? Colors.orange.withOpacity(0.1)
                      : (lastMeas.phase == 'Cold'
                            ? Colors.blue.withOpacity(0.1)
                            : Colors.green.withOpacity(0.1)),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  lastMeas.phase,
                  style: TextStyle(
                    color: lastMeas.phase == 'Hot'
                        ? Colors.orange
                        : (lastMeas.phase == 'Cold'
                              ? Colors.blue
                              : Colors.green),
                    fontWeight: FontWeight.bold,
                    fontSize: 12,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Result Grid
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 5,
              childAspectRatio: 1.3,
              crossAxisSpacing: 10,
              mainAxisSpacing: 10,
            ),
            itemCount: lastMeas.measurements.length,
            itemBuilder: (context, index) {
              final m = lastMeas.measurements[index];
              final freq = m.frequency;
              final val = m.loss;
              final delta = m.delta;

              return Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.grey.shade50,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.grey.shade100),
                ),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      '${freq.toStringAsFixed(1)} GHz',
                      style: TextStyle(
                        fontSize: 10,
                        color: Colors.grey.shade500,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      crossAxisAlignment: CrossAxisAlignment.baseline,
                      textBaseline: TextBaseline.alphabetic,
                      children: [
                        Text(
                          val.toStringAsFixed(2),
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                            color: theme.colorScheme.primary,
                            fontSize: 15,
                          ),
                        ),
                        const SizedBox(width: 2),
                        const Text(
                          'dB',
                          style: TextStyle(fontSize: 10, color: Colors.grey),
                        ),
                      ],
                    ),
                    if (!lastMeas.isReference) ...[
                      const SizedBox(height: 2),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 4,
                          vertical: 1,
                        ),
                        decoration: BoxDecoration(
                          color: (delta.abs() > 1.0)
                              ? Colors.red.withOpacity(0.1)
                              : Colors.blue.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: Text(
                          '${delta >= 0 ? '+' : ''}${delta.toStringAsFixed(2)} Δ',
                          style: TextStyle(
                            fontSize: 9,
                            fontWeight: FontWeight.bold,
                            color: (delta.abs() > 1.0)
                                ? Colors.red
                                : Colors.blue,
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              );
            },
          ),

          const SizedBox(height: 16),
          const Divider(),
          const SizedBox(height: 8),

          Row(
            children: [
              const Icon(Icons.info_outline, size: 14, color: Colors.grey),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  'Delta is computed against the first Reference.',
                  style: TextStyle(
                    fontSize: 11,
                    color: Colors.grey.shade500,
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ),
              TextButton.icon(
                onPressed: () {
                  setState(() {
                    _showGraph = !_showGraph;
                  });
                },
                icon: Icon(
                  _showGraph ? Icons.grid_view : Icons.show_chart,
                  size: 16,
                ),
                label: Text(
                  _showGraph ? 'Grid' : 'Plot',
                  style: const TextStyle(fontSize: 12),
                ),
              ),
            ],
          ),
          if (_showGraph) ...[
            const SizedBox(height: 16),
            SizedBox(
              height: 250,
              child: _buildMeasurementGraph(lastMeas, theme),
            ),
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                _buildLegendItem(theme.colorScheme.primary, 'Current'),
                const SizedBox(width: 16),
                _buildLegendItem(Colors.blue.withOpacity(0.5), 'Reference'),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildLegendItem(Color color, String label) {
    return Row(
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(3),
          ),
        ),
        const SizedBox(width: 4),
        Text(
          label,
          style: const TextStyle(
            fontSize: 10,
            color: Colors.grey,
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
    );
  }

  Widget _buildMeasurementGraph(TVACCableLossRecord record, ThemeData theme) {
    if (record.measurements.isEmpty) return const SizedBox();

    final List<FlSpot> currentSpots = [];
    final List<FlSpot> refSpots = [];
    final List<BarChartGroupData> deltaBars = [];

    double minLoss = 1000;
    double maxLoss = -1000;
    double maxDelta = 0.5;

    for (int i = 0; i < record.measurements.length; i++) {
      final m = record.measurements[i];
      final current = m.loss;
      final ref = current - m.delta;
      final delta = m.delta;

      // Use index as X for categorical axis
      final double xIdx = i.toDouble();

      currentSpots.add(FlSpot(xIdx, current));
      refSpots.add(FlSpot(xIdx, ref));

      if (current < minLoss) minLoss = current;
      if (ref < minLoss) minLoss = ref;
      if (current > maxLoss) maxLoss = current;
      if (ref > maxLoss) maxLoss = ref;

      if (delta.abs() > maxDelta) maxDelta = delta.abs();

      deltaBars.add(
        BarChartGroupData(
          x: i,
          barRods: [
            BarChartRodData(
              toY: delta,
              color: delta.abs() > 1.0 ? Colors.red : Colors.blue,
              width: 12,
              borderRadius: BorderRadius.circular(2),
            ),
          ],
        ),
      );
    }

    minLoss -= 2;
    maxLoss += 2;

    return Column(
      children: [
        Expanded(
          flex: 3,
          child: LineChart(
            LineChartData(
              gridData: const FlGridData(show: true, drawVerticalLine: true),
              titlesData: FlTitlesData(
                show: true,
                rightTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                topTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                bottomTitles: AxisTitles(
                  sideTitles: SideTitles(
                    showTitles: true,
                    reservedSize: 28,
                    interval: 1,
                    getTitlesWidget: (value, meta) {
                      final int idx = value.toInt();
                      if (idx < 0 || idx >= record.measurements.length) {
                        return const SizedBox();
                      }
                      final freq = record.measurements[idx].frequency;
                      return Padding(
                        padding: const EdgeInsets.only(top: 4.0),
                        child: Text(
                          freq.toStringAsFixed(1),
                          style: const TextStyle(
                            fontSize: 9,
                            color: Colors.grey,
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
                    getTitlesWidget: (value, meta) {
                      return Text(
                        value.toStringAsFixed(1),
                        style: const TextStyle(
                          fontSize: 10,
                          color: Colors.grey,
                        ),
                      );
                    },
                  ),
                ),
              ),
              borderData: FlBorderData(show: false),
              minX: 0,
              maxX: (record.measurements.length - 1).toDouble(),
              minY: minLoss,
              maxY: maxLoss,
              lineBarsData: [
                LineChartBarData(
                  spots: refSpots,
                  isCurved: false, // Changed to false for categorical accuracy
                  color: Colors.blue.withOpacity(0.5),
                  barWidth: 2,
                  isStrokeCapRound: true,
                  dotData: const FlDotData(show: true),
                  belowBarData: BarAreaData(show: false),
                ),
                LineChartBarData(
                  spots: currentSpots,
                  isCurved: false,
                  color: theme.colorScheme.primary,
                  barWidth: 4,
                  isStrokeCapRound: true,
                  dotData: const FlDotData(show: true),
                  belowBarData: BarAreaData(
                    show: true,
                    color: theme.colorScheme.primary.withOpacity(0.05),
                  ),
                ),
              ],
              lineTouchData: LineTouchData(
                touchTooltipData: LineTouchTooltipData(
                  maxContentWidth: 200,
                  getTooltipItems: (touchedSpots) {
                    return touchedSpots.map((spot) {
                      final isRef = spot.barIndex == 0;
                      final freq =
                          record.measurements[spot.x.toInt()].frequency;
                      return LineTooltipItem(
                        '${freq.toStringAsFixed(3)} GHz\n',
                        const TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                        ),
                        children: [
                          TextSpan(
                            text:
                                '${isRef ? "Reference" : "Current"}: ${spot.y.toStringAsFixed(2)} dB',
                            style: TextStyle(
                              color: isRef
                                  ? Colors.blue.shade200
                                  : Colors.white,
                              fontSize: 12,
                            ),
                          ),
                        ],
                      );
                    }).toList();
                  },
                ),
              ),
            ),
          ),
        ),
        const SizedBox(height: 12),
        const Text(
          'Delta (dB)',
          style: TextStyle(
            fontSize: 10,
            fontWeight: FontWeight.bold,
            color: Colors.grey,
          ),
        ),
        const SizedBox(height: 4),
        Expanded(
          flex: 1,
          child: BarChart(
            BarChartData(
              alignment: BarChartAlignment.spaceAround,
              maxY: maxDelta * 1.2,
              minY: -maxDelta * 1.2,
              barGroups: deltaBars,
              gridData: const FlGridData(show: false),
              borderData: FlBorderData(show: false),
              titlesData: FlTitlesData(
                show: true,
                leftTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                rightTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                topTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                bottomTitles: AxisTitles(
                  sideTitles: SideTitles(
                    showTitles: true,
                    getTitlesWidget: (value, meta) {
                      final int idx = value.toInt();
                      if (idx < 0 || idx >= record.measurements.length) {
                        return const SizedBox();
                      }
                      final freq = record.measurements[idx].frequency;
                      return Padding(
                        padding: const EdgeInsets.only(top: 4.0),
                        child: Text(
                          freq.toStringAsFixed(1),
                          style: const TextStyle(
                            fontSize: 9,
                            color: Colors.grey,
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildHistorySection(ThemeData theme) {
    return Container(
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
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Measurement History',
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                ElevatedButton.icon(
                  onPressed: () {},
                  icon: const Icon(Icons.download, size: 18),
                  label: const Text('Export CSV'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue.shade50,
                    foregroundColor: Colors.blue.shade800,
                    elevation: 0,
                  ),
                ),
              ],
            ),
          ),

          // History Table Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
            color: Colors.grey.shade50,
            child: Row(
              children: [
                _tableHeaderCell('Sl.', 60),
                _tableHeaderCell('Cable Name', 150),
                _tableHeaderCell('Cycle / Phase', 150),
                _tableHeaderCell('Ref', 60),
                _tableHeaderCell('Timestamp', 180),
                Expanded(child: _tableHeaderCell('Measurements (GHz)', 0)),
              ],
            ),
          ),

          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _history.length,
            separatorBuilder: (context, index) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final item = _history[index];
              return InkWell(
                onTap: () {},
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 16,
                  ),
                  child: Row(
                    children: [
                      _tableCell(item.slNo.toString(), 60),
                      _tableCell(item.cableName, 150, isBold: true),
                      SizedBox(
                        width: 150,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              item.cycleName,
                              style: const TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            Text(
                              item.phase,
                              style: TextStyle(
                                fontSize: 11,
                                color: item.phase == 'Hot'
                                    ? Colors.orange
                                    : (item.phase == 'Cold'
                                          ? Colors.blue
                                          : Colors.green),
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                      SizedBox(
                        width: 60,
                        child: item.isReference
                            ? const Icon(
                                Icons.check_circle,
                                color: Colors.blue,
                                size: 18,
                              )
                            : const SizedBox(),
                      ),
                      _tableCell('${item.date}\n${item.time}', 180),
                      Expanded(
                        child: Wrap(
                          spacing: 8,
                          runSpacing: 4,
                          children: item.measurements.map((m) {
                            final delta = m.delta;
                            return Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.grey.shade100,
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Column(
                                children: [
                                  Text(
                                    '${m.frequency}G',
                                    style: const TextStyle(
                                      fontSize: 9,
                                      color: Colors.grey,
                                    ),
                                  ),
                                  Text(
                                    m.loss.toStringAsFixed(2),
                                    style: const TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  if (!item.isReference)
                                    Text(
                                      '${delta >= 0 ? '+' : ''}${delta.toStringAsFixed(1)}',
                                      style: TextStyle(
                                        fontSize: 9,
                                        color: delta.abs() > 0.1
                                            ? Colors.red
                                            : Colors.green,
                                        fontWeight: FontWeight.w900,
                                      ),
                                    ),
                                ],
                              ),
                            );
                          }).toList(),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _tableHeaderCell(String label, double width) {
    return SizedBox(
      width: width == 0 ? null : width,
      child: Text(
        label,
        style: const TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.bold,
          color: Colors.grey,
        ),
      ),
    );
  }

  Widget _tableCell(String value, double width, {bool isBold = false}) {
    return SizedBox(
      width: width,
      child: Text(
        value,
        style: TextStyle(
          fontSize: 13,
          fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
          color: Colors.grey.shade800,
        ),
      ),
    );
  }
}
