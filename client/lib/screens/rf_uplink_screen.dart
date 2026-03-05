import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/utils/notifications.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/screens/test_progress_screen.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';

class RFUplinkScreen extends StatefulWidget {
  final bool isActive;
  const RFUplinkScreen({super.key, this.isActive = true});

  @override
  State<RFUplinkScreen> createState() => _RFUplinkScreenState();
}

class _RFUplinkScreenState extends State<RFUplinkScreen> {
  @override
  void didUpdateWidget(RFUplinkScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    // No WebSocket to close here yet, but added for consistency with IndexedStack pattern
  }

  int _selectedOperation = 0; // 0: RF Uplink, 1: Remove Link, 2: Route Path
  bool _isHelpOpen = false;
  RFUplinkMetaData? _metaData;
  LinkStatus? _linkStatus;
  String _selectedTSM = "";
  bool _isMetaLoading = true;
  bool _isStatusLoading = false;

  @override
  void initState() {
    super.initState();
    _fetchInitialData();
  }

  @override
  void dispose() {
    super.dispose();
  }

  void _fetchInitialData() {
    setState(() => _isMetaLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);

    // Use bootstrapped data instead of fetching
    final meta = serverService.status.bootstrapData?.rfuData;

    if (meta != null) {
      debugPrint('RFUplinkScreen: Using Bootstrapped Metadata');
      setState(() {
        _metaData = meta;
        if (_selectedTSM == "" && meta.tsms.isNotEmpty) {
          _selectedTSM = meta.tsms[0];
        }
        _isMetaLoading = false;
      });
      _fetchHardwareStatus();
    } else {
      debugPrint('RFUplinkScreen: Bootstrapped Metadata NOT FOUND');
      setState(() => _isMetaLoading = false);
    }
  }

  Future<void> _fetchHardwareStatus() async {
    setState(() => _isStatusLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    final status = await serverService.fetchLinkStatus(_selectedTSM);
    if (mounted) {
      setState(() {
        _linkStatus = status;
        _isStatusLoading = false;
      });
    }
  }

  final List<Map<String, dynamic>> _operations = [
    {
      'title': 'RF Uplink',
      'icon': Icons.settings_input_antenna_outlined,
      'selectedIcon': Icons.settings_input_antenna,
    },
    {
      'title': 'Remove Existing Link',
      'icon': Icons.link_off_outlined,
      'selectedIcon': Icons.link_off,
    },
    {
      'title': 'Route Path',
      'icon': Icons.alt_route_outlined,
      'selectedIcon': Icons.alt_route,
    },
  ];

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
                // Persistent Header with Compact Visualization
                ScreenHeader(
                  title: 'RF Operations',
                  subtitle: 'Configure and manage RF uplink paths',
                  icon: Icons.settings_input_antenna,
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      _buildCompactStatusPanel(theme),
                      const SizedBox(width: 16),
                      _buildHelpTrigger(theme),
                    ],
                  ),
                ),
                const SizedBox(height: 32),
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Master Pane (Sidebar)
                      ContentCard(
                        width: 280,
                        isSidebar: true,
                        padding: EdgeInsets.zero,
                        child: Column(
                          children: [
                            const SizedBox(height: 8),
                            Expanded(
                              child: ListView.separated(
                                padding: const EdgeInsets.all(12),
                                itemCount: _operations.length,
                                separatorBuilder: (context, index) =>
                                    const SizedBox(height: 4),
                                itemBuilder: (context, index) {
                                  final isSelected =
                                      _selectedOperation == index;
                                  final op = _operations[index];
                                  return InkWell(
                                    onTap: () {
                                      setState(
                                        () => _selectedOperation = index,
                                      );
                                      if (index == 1 || index == 2) {
                                        _fetchHardwareStatus();
                                      }
                                    },
                                    borderRadius: BorderRadius.circular(16),
                                    child: AnimatedContainer(
                                      duration: const Duration(
                                        milliseconds: 200,
                                      ),
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 16,
                                        vertical: 16,
                                      ),
                                      decoration: BoxDecoration(
                                        color: isSelected
                                            ? theme.colorScheme.primary
                                                  .withOpacity(0.08)
                                            : Colors.transparent,
                                        borderRadius: BorderRadius.circular(16),
                                        border: Border.all(
                                          color: isSelected
                                              ? theme.colorScheme.primary
                                                    .withOpacity(0.2)
                                              : Colors.transparent,
                                        ),
                                      ),
                                      child: Row(
                                        children: [
                                          Icon(
                                            isSelected
                                                ? op['selectedIcon'] as IconData
                                                : op['icon'] as IconData,
                                            color: isSelected
                                                ? theme.colorScheme.primary
                                                : Colors.grey.shade400,
                                            size: 20,
                                          ),
                                          const SizedBox(width: 16),
                                          Expanded(
                                            child: Text(
                                              op['title'] as String,
                                              style: GoogleFonts.inter(
                                                fontWeight: isSelected
                                                    ? FontWeight.w900
                                                    : FontWeight.bold,
                                                color: isSelected
                                                    ? theme.colorScheme.primary
                                                    : Colors.black,
                                                fontSize: 16,
                                              ),
                                            ),
                                          ),
                                          if (isSelected)
                                            Icon(
                                              Icons.chevron_right,
                                              color: theme.colorScheme.primary,
                                              size: 14,
                                            ),
                                        ],
                                      ),
                                    ),
                                  );
                                },
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 32),
                      // Detail Pane
                      Expanded(
                        child: ContentCard(
                          isSidebar: false,
                          child: _isMetaLoading
                              ? Center(
                                  child: Column(
                                    mainAxisAlignment: MainAxisAlignment.center,
                                    children: [
                                      const CircularProgressIndicator(),
                                      const SizedBox(height: 24),
                                      Text(
                                        'Loading Metadata...',
                                        style: GoogleFonts.inter(
                                          color: Colors.grey.shade600,
                                          fontSize: 14,
                                          fontWeight: FontWeight.w500,
                                        ),
                                      ),
                                    ],
                                  ),
                                )
                              : _buildDetailContent(),
                        ),
                      ),
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
            right: _isHelpOpen ? 0 : -400,
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
          boxShadow: [
            if (_isHelpOpen)
              BoxShadow(
                color: theme.colorScheme.primary.withOpacity(0.3),
                blurRadius: 10,
                offset: const Offset(0, 4),
              ),
          ],
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
      width: 400,
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
                  'Contextual Help',
                  style: GoogleFonts.outfit(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: theme.colorScheme.primary,
                  ),
                ),
                IconButton(
                  onPressed: () => setState(() => _isHelpOpen = false),
                  icon: const Icon(Icons.close),
                  style: IconButton.styleFrom(
                    backgroundColor: Colors.grey.shade100,
                  ),
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
                  _operations[_selectedOperation]['title'] as String,
                  _getHelpContentByOperation(),
                  isMain: true,
                ),
                const SizedBox(height: 32),
                _buildHelpItem(
                  theme,
                  'Live Monitoring',
                  'The dashboard header always displays the Current Route Status and Attenuation. '
                      'LEDs represent active path components (Green for ON, Red for OFF).',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Keyboard Shortcuts',
                  '• Press [Enter] to submit forms\n• Press [Esc] to close this help panel',
                ),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.all(24),
            color: Colors.grey.shade50,
            child: Row(
              children: [
                const Icon(
                  Icons.lightbulb_outline,
                  color: Colors.orange,
                  size: 20,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    'Did you find this helpful?',
                    style: TextStyle(color: Colors.grey.shade600, fontSize: 13),
                  ),
                ),
                TextButton(onPressed: () {}, child: const Text('Yes')),
                TextButton(onPressed: () {}, child: const Text('No')),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHelpItem(
    ThemeData theme,
    String title,
    String content, {
    bool isMain = false,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: GoogleFonts.inter(
            fontWeight: FontWeight.bold,
            fontSize: isMain ? 18 : 14,
            color: isMain ? Colors.black : Colors.grey.shade800,
          ),
        ),
        const SizedBox(height: 12),
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

  String _getHelpContentByOperation() {
    switch (_selectedOperation) {
      case 0:
        return 'To establish an RF Uplink, select a valid configuration from the dropdown. '
            'Power at Spacecraft Rx is pre-filled from the database but can be adjusted manually. '
            'Loss estimates are informational and based on the current system calibration.';
      case 1:
        return 'The Remove Link operation safely disconnects active RF configurations. '
            'Use the "Remove All" checkbox for an emergency shutdown of all existing connections.';
      case 2:
        return 'Route Path allows you to define specific signal paths through the TSM matrix. '
            'Set signal attenuation (Attn 1 & 2) to control signal strength. Values are applied immediately to the hardware.';
      default:
        return '';
    }
  }

  Widget _buildCompactStatusPanel(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1C1E),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'LIVE ROUTE STATUS',
                style: GoogleFonts.inter(
                  color: Colors.white.withOpacity(0.5),
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                  letterSpacing: 1.5,
                ),
              ),
              const SizedBox(height: 8),
              if (_isStatusLoading && _linkStatus == null)
                Text(
                  'CONNECTING TO TSM...',
                  style: GoogleFonts.outfit(
                    color: Colors.blue.shade300,
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                  ),
                )
              else if (_linkStatus?.tsmConnected ?? false)
                _buildCompactLedGrid()
              else
                Row(
                  children: [
                    const Icon(
                      Icons.warning_amber_rounded,
                      color: Colors.orange,
                      size: 14,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'TSM DISCONNECTED',
                      style: GoogleFonts.outfit(
                        color: Colors.orange.shade300,
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
            ],
          ),
          const SizedBox(width: 32),
          Container(height: 40, width: 1, color: Colors.white10),
          const SizedBox(width: 24),
          _buildCompactAttnValue(
            'ATTN 1',
            _isStatusLoading && _linkStatus == null
                ? '-- dB'
                : '${_linkStatus?.attn1Value ?? 0.0} dB',
          ),
          const SizedBox(width: 24),
          _buildCompactAttnValue(
            'ATTN 2',
            _isStatusLoading && _linkStatus == null
                ? '-- dB'
                : '${_linkStatus?.attn2Value ?? 0.0} dB',
          ),
        ],
      ),
    );
  }

  Widget _buildCompactLedGrid() {
    if (_linkStatus == null || _linkStatus!.switchStatus.isEmpty) {
      return const SizedBox(height: 20);
    }

    return Row(
      children: [
        for (int i = 0; i < _linkStatus!.switchStatus.length; i++)
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildLedRow(_linkStatus!.switchStatus[i]),
                const SizedBox(height: 4),
                Text(
                  _linkStatus!.switchStatus[i],
                  style: GoogleFonts.robotoMono(
                    color: Colors.blue.shade300,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    letterSpacing: 1.0,
                  ),
                ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildLedRow(String status) {
    final leds = _parseSwitchStatus(status);
    return Row(
      children: [
        for (int j = 0; j < 10; j++)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 2),
            child: Container(
              width: 8,
              height: 8,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: leds[j] ? Colors.green : Colors.red,
                boxShadow: [
                  BoxShadow(
                    color: (leds[j] ? Colors.green : Colors.red).withOpacity(
                      0.5,
                    ),
                    blurRadius: 4,
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }

  List<bool> _parseSwitchStatus(String status) {
    List<bool> ledStatus = List.generate(10, (_) => false);
    // Format: D1A123B456
    int aIndex = status.indexOf('A');
    int bIndex = status.indexOf('B');

    String onPart = '';
    // String offPart = ''; // Not strictly needed for logic but used for parsing

    if (aIndex != -1) {
      if (bIndex != -1) {
        onPart = status.substring(aIndex + 1, bIndex);
        // offPart = status.substring(bIndex + 1);
      } else {
        onPart = status.substring(aIndex + 1);
      }
    }

    for (var i = 0; i < onPart.length; i++) {
      String char = onPart[i];
      int? val = int.tryParse(char);
      if (val != null) {
        if (val == 0) val = 10;
        if (val >= 1 && val <= 10) {
          ledStatus[val - 1] = true;
        }
      }
    }
    return ledStatus;
  }

  Widget _buildCompactAttnValue(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(label, style: const TextStyle(color: Colors.white38, fontSize: 9)),
        Text(
          value,
          style: GoogleFonts.robotoMono(
            color: Colors.blue.shade300,
            fontWeight: FontWeight.bold,
            fontSize: 16,
          ),
        ),
      ],
    );
  }

  Widget _buildDetailContent() {
    switch (_selectedOperation) {
      case 0:
        return RFUplinkForm(
          metaData: _metaData,
          onRefresh: _fetchHardwareStatus,
        );
      case 1:
        return RemoveLinkForm(
          metaData: _metaData,
          linkStatus: _linkStatus,
          onRefresh: _fetchHardwareStatus,
        );
      case 2:
        return RoutePathForm(
          metaData: _metaData,
          linkStatus: _linkStatus,
          selectedTSM: _selectedTSM,
          onTSMChanged: (val) {
            setState(() {
              _selectedTSM = val;
            });
            _fetchHardwareStatus();
          },
          onRefresh: _fetchHardwareStatus,
        );
      default:
        return const Center(child: Text('Select an operation'));
    }
  }
}

class RFUplinkForm extends StatefulWidget {
  final RFUplinkMetaData? metaData;
  final VoidCallback onRefresh;
  const RFUplinkForm({super.key, this.metaData, required this.onRefresh});

  @override
  State<RFUplinkForm> createState() => _RFUplinkFormState();
}

class _RFUplinkFormState extends State<RFUplinkForm> {
  String? _selectedConfig;
  bool _expressUplink = false;
  bool _doppler = false;
  final TextEditingController _powerController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _updatePowerValue();
  }

  @override
  void didUpdateWidget(RFUplinkForm oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.metaData != oldWidget.metaData) {
      _updatePowerValue();
    }
  }

  void _updatePowerValue() {
    if (widget.metaData != null && _selectedConfig != null) {
      final info = widget.metaData!.uplinkConfigInformation[_selectedConfig];
      if (info != null) {
        _powerController.text = info.powerAtSC.toStringAsFixed(1);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'RF UPLINK CONFIGURATION',
          style: GoogleFonts.inter(
            fontWeight: FontWeight.w900,
            fontSize: 12,
            color: theme.colorScheme.primary,
            letterSpacing: 1.5,
          ),
        ),
        const SizedBox(height: 8),
        Text(
          'Establish a new RF uplink with current parameters',
          style: TextStyle(color: Colors.grey.shade500),
        ),
        const SizedBox(height: 40),
        DropdownButtonFormField<String>(
          decoration: InputDecoration(
            labelText: 'Select Configuration',
            prefixIcon: const Icon(Icons.settings_outlined),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
          value: _selectedConfig,
          items: (widget.metaData?.uplinkConfigs ?? []).map((c) {
            return DropdownMenuItem(value: c, child: Text(c));
          }).toList(),
          onChanged: (val) {
            setState(() {
              _selectedConfig = val;
              _updatePowerValue();
            });
          },
        ),
        const SizedBox(height: 32),
        Row(
          children: [
            Expanded(
              child: SwitchListTile(
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: BorderSide(color: Colors.grey.shade200),
                ),
                title: const Text('Express Uplink'),
                subtitle: const Text('Fast track connection'),
                value: _expressUplink,
                onChanged: (val) => setState(() => _expressUplink = val),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: SwitchListTile(
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: BorderSide(color: Colors.grey.shade200),
                ),
                title: const Text('Doppler'),
                subtitle: const Text('Apply doppler correction'),
                value: _doppler,
                onChanged: (val) => setState(() => _doppler = val),
              ),
            ),
          ],
        ),
        const SizedBox(height: 32),
        TextFormField(
          controller: _powerController,
          decoration: InputDecoration(
            labelText: 'Power at Spacecraft Rx (dBm)',
            prefixIcon: const Icon(Icons.bolt),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            helperText: 'Default value prepopulated from database',
          ),
          keyboardType: TextInputType.number,
        ),
        const SizedBox(height: 40),
        Row(
          children: [
            Expanded(
              child: _buildInfoTile(
                'Loss Till SA',
                '${widget.metaData?.uplinkConfigInformation[_selectedConfig]?.saLoss.toStringAsFixed(1) ?? "0.0"} dB',
                Icons.info_outline,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _buildInfoTile(
                'Loss Till Spacecraft',
                '${widget.metaData?.uplinkConfigInformation[_selectedConfig]?.scLoss.toStringAsFixed(1) ?? "0.0"} dB',
                Icons.info_outline,
              ),
            ),
          ],
        ),
        const SizedBox(height: 32),
        SizedBox(
          width: double.infinity,
          height: 56,
          child: ElevatedButton.icon(
            onPressed: _selectedConfig == null
                ? null
                : () {
                    String category = _expressUplink ? "Fast" : "Full";
                    if (_doppler) {
                      category += "-Doppler";
                    }

                    final tests = [
                      TestDescription(
                        testName: 'RFUplink',
                        testCategory: category,
                        configuration: _selectedConfig,
                        extraParameters: [
                          'NominalPower;${_powerController.text}',
                          'Express:$_expressUplink',
                          'Doppler:$_doppler',
                        ],
                      ),
                    ];
                    Navigator.of(context)
                        .push(
                          MaterialPageRoute(
                            builder: (context) =>
                                TestProgressScreen(tests: tests),
                          ),
                        )
                        .then((_) => widget.onRefresh());
                  },
            icon: const Icon(Icons.rocket_launch),
            label: const Text(
              'START UPLINK PROCESS',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.2,
              ),
            ),
            style: ElevatedButton.styleFrom(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              elevation: 4,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildInfoTile(String label, String value, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Icon(icon, size: 20, color: Colors.grey.shade600),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(color: Colors.grey.shade600, fontSize: 12),
              ),
              Text(
                value,
                style: GoogleFonts.robotoMono(
                  fontWeight: FontWeight.bold,
                  color: Colors.black,
                  fontSize: 16,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class RemoveLinkForm extends StatefulWidget {
  final RFUplinkMetaData? metaData;
  final LinkStatus? linkStatus;
  final VoidCallback onRefresh;
  const RemoveLinkForm({
    super.key,
    this.metaData,
    this.linkStatus,
    required this.onRefresh,
  });

  @override
  State<RemoveLinkForm> createState() => _RemoveLinkFormState();
}

class _RemoveLinkFormState extends State<RemoveLinkForm> {
  String? _selectedConfig;
  bool _removeAll = false;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
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
                  'REMOVE EXISTING LINK',
                  style: GoogleFonts.inter(
                    fontWeight: FontWeight.w900,
                    fontSize: 12,
                    color: theme.colorScheme.primary,
                    letterSpacing: 1.5,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Disconnect existing RF links securely',
                  style: TextStyle(color: Colors.grey.shade500),
                ),
              ],
            ),
            Container(
              decoration: BoxDecoration(
                color: theme.colorScheme.primary.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: IconButton(
                icon: Icon(Icons.refresh, color: theme.colorScheme.primary),
                onPressed: widget.onRefresh,
                tooltip: 'Refresh Configurations',
              ),
            ),
          ],
        ),
        const SizedBox(height: 48),
        DropdownButtonFormField<String>(
          decoration: InputDecoration(
            labelText: 'Select Configuration to Remove',
            prefixIcon: const Icon(Icons.link_off),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
          value: _selectedConfig,
          items: (widget.linkStatus?.removeConfigs ?? []).map((c) {
            return DropdownMenuItem(value: c, child: Text(c));
          }).toList(),
          onChanged: (val) => setState(() => _selectedConfig = val),
        ),
        const SizedBox(height: 32),
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: Colors.red.shade50,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.red.shade100),
          ),
          child: CheckboxListTile(
            title: const Text(
              'Remove All Active Links',
              style: TextStyle(
                color: Color(0xFFD32F2F),
                fontWeight: FontWeight.bold,
              ),
            ),
            subtitle: const Text(
              'This will disconnect all existing RF connections',
            ),
            value: _removeAll,
            activeColor: Colors.red,
            onChanged: (val) => setState(() => _removeAll = val ?? false),
            controlAffinity: ListTileControlAffinity.leading,
          ),
        ),
        const Spacer(),
        SizedBox(
          width: double.infinity,
          height: 56,
          child: ElevatedButton.icon(
            onPressed: (_selectedConfig == null && !_removeAll)
                ? null
                : () {
                    final List<TestDescription> tests = [];
                    if (_removeAll) {
                      final configs = widget.linkStatus?.removeConfigs ?? [];
                      for (final config in configs) {
                        tests.add(
                          TestDescription(
                            testName: 'RFUplinkRemoval',
                            testCategory: '',
                            configuration: config,
                          ),
                        );
                      }
                    } else {
                      tests.add(
                        TestDescription(
                          testName: 'RFUplinkRemoval',
                          testCategory: '',
                          configuration: _selectedConfig,
                        ),
                      );
                    }

                    if (tests.isNotEmpty) {
                      Navigator.of(context)
                          .push(
                            MaterialPageRoute(
                              builder: (context) =>
                                  TestProgressScreen(tests: tests),
                            ),
                          )
                          .then((_) => widget.onRefresh());
                    }
                  },
            icon: const Icon(Icons.delete_forever),
            label: const Text(
              'REMOVE RF CONNECTION',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.2,
              ),
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red.shade700,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              elevation: 4,
            ),
          ),
        ),
      ],
    );
  }
}

class RoutePathForm extends StatefulWidget {
  final RFUplinkMetaData? metaData;
  final LinkStatus? linkStatus;
  final String selectedTSM;
  final Function(String) onTSMChanged;
  final VoidCallback onRefresh;

  const RoutePathForm({
    super.key,
    this.metaData,
    this.linkStatus,
    required this.selectedTSM,
    required this.onTSMChanged,
    required this.onRefresh,
  });

  @override
  State<RoutePathForm> createState() => _RoutePathFormState();
}

class _RoutePathFormState extends State<RoutePathForm> {
  String? _selectedConfig;
  String? _selectedPath;
  String? _selectedMnemonic;
  bool _isUserDefinedRoute = false;
  final TextEditingController _attn1Controller = TextEditingController();
  final TextEditingController _attn2Controller = TextEditingController();
  final TextEditingController _userMnemonicController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _updateValues();
  }

  @override
  void didUpdateWidget(RoutePathForm oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.linkStatus != oldWidget.linkStatus) {
      _updateValues();
    }
  }

  @override
  void dispose() {
    _attn1Controller.dispose();
    _attn2Controller.dispose();
    _userMnemonicController.dispose();
    super.dispose();
  }

  void _updateValues() {
    if (widget.linkStatus != null) {
      _attn1Controller.text = widget.linkStatus!.attn1Value.toStringAsFixed(1);
      _attn2Controller.text = widget.linkStatus!.attn2Value.toStringAsFixed(1);
    }
  }

  Future<void> _setAttenuation(int attnNo, String value) async {
    final double? val = double.tryParse(value);
    if (val == null) {
      AppNotifications.show(
        context,
        'Invalid attenuation value',
        type: NotificationType.error,
      );
      return;
    }

    final serverService = Provider.of<ServerService>(context, listen: false);
    final ack = await serverService.setTSMAttn(widget.selectedTSM, attnNo, val);

    if (mounted) {
      if (ack != null && ack.ok) {
        AppNotifications.show(
          context,
          'Successfully set attenuation: ${ack.message}',
          type: NotificationType.success,
        );
        widget.onRefresh();
      } else {
        AppNotifications.show(
          context,
          'Failed to set attenuation: ${ack?.message ?? 'Unknown error'}',
          type: NotificationType.error,
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
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
                  'ROUTE PATH & ATTENUATION',
                  style: GoogleFonts.inter(
                    fontWeight: FontWeight.w900,
                    fontSize: 12,
                    color: theme.colorScheme.primary,
                    letterSpacing: 1.5,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Configure TSM routing and signal attenuation',
                  style: TextStyle(color: Colors.grey.shade500),
                ),
              ],
            ),
            Container(
              decoration: BoxDecoration(
                color: theme.colorScheme.primary.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: IconButton(
                icon: Icon(Icons.refresh, color: theme.colorScheme.primary),
                onPressed: widget.onRefresh,
                tooltip: 'Refresh Current Status',
              ),
            ),
          ],
        ),
        const SizedBox(height: 40),
        Row(
          children: [
            Expanded(
              child: DropdownButtonFormField<String>(
                decoration: InputDecoration(
                  labelText: 'Select TSM',
                  prefixIcon: const Icon(Icons.switch_access_shortcut),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                value: widget.selectedTSM,
                items: (widget.metaData?.tsms ?? []).map((t) {
                  return DropdownMenuItem(value: t, child: Text(t));
                }).toList(),
                onChanged: (val) {
                  if (val != null) widget.onTSMChanged(val);
                },
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: DropdownButtonFormField<String>(
                decoration: InputDecoration(
                  labelText: 'Select Configuration',
                  prefixIcon: const Icon(Icons.tune),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                value: _selectedConfig,
                items: (widget.metaData?.allConfigs ?? []).map((c) {
                  return DropdownMenuItem(value: c, child: Text(c));
                }).toList(),
                onChanged: (val) {
                  setState(() {
                    _selectedConfig = val;
                    _selectedPath = null;
                    _selectedMnemonic = null;
                  });
                },
              ),
            ),
          ],
        ),
        const SizedBox(height: 24),
        DropdownButtonFormField<String>(
          decoration: InputDecoration(
            labelText: 'Path to be Routed',
            prefixIcon: const Icon(Icons.shortcut),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
          value: _selectedPath,
          items: [
            if (_selectedConfig != null && widget.metaData != null)
              ...(widget.metaData!.configPathInformation[_selectedConfig] ?? [])
                  .map((p) {
                    return DropdownMenuItem(
                      value: p.path,
                      child: Text('${p.path} (${p.mnemonic})'),
                    );
                  }),
            const DropdownMenuItem(
              value: 'USER_DEFINED',
              child: Text('User Defined Route...'),
            ),
          ],
          onChanged: (val) {
            setState(() {
              _selectedPath = val;
              if (val == 'USER_DEFINED') {
                _isUserDefinedRoute = true;
                _selectedMnemonic = null;
              } else if (val != null &&
                  _selectedConfig != null &&
                  widget.metaData != null) {
                _isUserDefinedRoute = false;
                final paths =
                    widget.metaData!.configPathInformation[_selectedConfig];
                if (paths != null) {
                  final p = paths.firstWhere((element) => element.path == val);
                  _selectedMnemonic = p.mnemonic;
                }
              }
            });
          },
        ),
        if (_isUserDefinedRoute) ...[
          const SizedBox(height: 24),
          TextFormField(
            controller: _userMnemonicController,
            decoration: InputDecoration(
              labelText: 'User Defined Mnemonic',
              hintText: 'Enter mnemonic directly',
              prefixIcon: const Icon(Icons.edit_note),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
          ),
        ],
        const SizedBox(height: 24),
        Row(
          children: [
            Expanded(
              child: TextFormField(
                controller: _attn1Controller,
                decoration: InputDecoration(
                  labelText: 'Attenuation 1 (dB)',
                  prefixIcon: const Icon(Icons.signal_cellular_alt),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  suffixIcon: Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: IconButton(
                      icon: const Icon(Icons.send_rounded),
                      onPressed: () =>
                          _setAttenuation(1, _attn1Controller.text),
                      tooltip: 'Set Attn 1',
                      color: theme.colorScheme.primary,
                    ),
                  ),
                ),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: TextFormField(
                controller: _attn2Controller,
                decoration: InputDecoration(
                  labelText: 'Attenuation 2 (dB)',
                  prefixIcon: const Icon(Icons.signal_cellular_alt_outlined),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  suffixIcon: Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: IconButton(
                      icon: const Icon(Icons.send_rounded),
                      onPressed: () =>
                          _setAttenuation(2, _attn2Controller.text),
                      tooltip: 'Set Attn 2',
                      color: theme.colorScheme.primary,
                    ),
                  ),
                ),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
              ),
            ),
          ],
        ),
        const Spacer(),
        SizedBox(
          width: double.infinity,
          height: 56,
          child: ElevatedButton.icon(
            onPressed:
                (_selectedMnemonic == null &&
                    (!_isUserDefinedRoute ||
                        _userMnemonicController.text.isEmpty))
                ? null
                : () async {
                    final serverService = Provider.of<ServerService>(
                      context,
                      listen: false,
                    );
                    final String mnemonic = _isUserDefinedRoute
                        ? _userMnemonicController.text
                        : _selectedMnemonic!;

                    final ack = await serverService.setTSMRoute(
                      widget.selectedTSM,
                      mnemonic,
                    );

                    if (mounted) {
                      if (ack != null && ack.ok) {
                        AppNotifications.show(
                          context,
                          'Route successfully established: ${ack.message}',
                          type: NotificationType.success,
                        );
                        widget.onRefresh();
                      } else {
                        AppNotifications.show(
                          context,
                          'Routing error: ${ack?.message ?? 'Unknown error'}',
                          type: NotificationType.error,
                        );
                      }
                    }
                  },
            icon: const Icon(Icons.alt_route),
            label: const Text(
              'ROUTE PATH',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                letterSpacing: 1.2,
              ),
            ),
            style: ElevatedButton.styleFrom(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              elevation: 4,
            ),
          ),
        ),
      ],
    );
  }
}
