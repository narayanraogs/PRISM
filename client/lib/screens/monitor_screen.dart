import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';

class MonitorScreen extends StatefulWidget {
  final bool isActive;
  const MonitorScreen({super.key, required this.isActive});

  @override
  State<MonitorScreen> createState() => _MonitorScreenState();
}

class _MonitorScreenState extends State<MonitorScreen> {
  MonitorMetadata? _metadata;
  bool _isLoading = true;
  bool _isHelpOpen = false;
  String? _errorMessage;

  String? _selectedType;
  String? _selectedInstrument;

  Stream<MonitorResponse>? _monitorStream;
  bool _isMonitoring = false;
  DateTime? _lastUpdateTime;

  @override
  void initState() {
    super.initState();
    _loadMetadata();
  }

  @override
  void didUpdateWidget(MonitorScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.isActive && !widget.isActive && _isMonitoring) {
      _stopMonitor();
    }
  }

  void _loadMetadata() {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final service = Provider.of<ServerService>(context, listen: false);
    final metadata = service.status.bootstrapData?.monitorData;

    if (metadata != null) {
      debugPrint('MonitorScreen: Using Bootstrapped Metadata');
      setState(() {
        _metadata = metadata;
        _isLoading = false;
      });
    } else {
      debugPrint('MonitorScreen: Bootstrapped Metadata NOT FOUND');
      setState(() {
        _errorMessage = 'Failed to load metadata';
        _isLoading = false;
      });
    }
  }

  void _startMonitor() {
    if (_selectedType == null || _selectedInstrument == null) return;

    final service = Provider.of<ServerService>(context, listen: false);
    setState(() {
      _isMonitoring = true;
      // Using broadcast stream so we can listen locally and use it in StreamBuilder
      final stream = service
          .connectMonitor(_selectedType!, _selectedInstrument!)
          .asBroadcastStream();
      _monitorStream = stream;

      stream.listen(
        (data) {
          if (mounted) {
            setState(() {
              _lastUpdateTime = DateTime.now();
            });
          }
        },
        onError: (err) {
          if (mounted) {
            setState(() {
              _isMonitoring = false;
              _monitorStream = null;
            });
          }
        },
      );
    });

    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );
    notificationService.addNotification(
      title: 'Monitoring Started',
      message: 'Monitoring $_selectedInstrument...',
      type: NotificationType.info,
    );
  }

  void _stopMonitor() {
    final service = Provider.of<ServerService>(context, listen: false);
    service.closeMonitor();
    setState(() {
      _isMonitoring = false;
      _monitorStream = null;
      _lastUpdateTime = null;
    });

    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );
    notificationService.addNotification(
      title: 'Monitoring Stopped',
      message: 'Disconnected from instrument.',
      type: NotificationType.info,
    );
  }

  @override
  void dispose() {
    final service = Provider.of<ServerService>(context, listen: false);
    service.closeMonitor();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              'Error: $_errorMessage',
              style: TextStyle(color: theme.colorScheme.error),
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadMetadata,
              child: const Text('Retry'),
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
              children: [
                _buildHeader(theme),
                const SizedBox(height: 24),
                Expanded(
                  child: Row(
                    children: [
                      Expanded(flex: 3, child: _buildSelectionPanel(theme)),
                      const SizedBox(width: 20),
                      Expanded(flex: 7, child: _buildMonitorView(theme)),
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
      title: 'Continuous Monitor',
      subtitle: 'Real-time instrument monitoring and visualization',
      icon: Icons.monitor_heart_rounded,
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
                  'Monitor Help',
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
                  'Continuous Monitoring',
                  'This screen provides a live, real-time data stream from the selected instrument. '
                      'Unlike snapshots, this maintains an active connection for constant updates.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Visualizations',
                  '• **SA / VSA**: Displays high-frequency screen captures directly from the hardware.\n'
                      '• **PM / PPM**: Shows numerical power levels in a digital meter format with Peak and Average readings for Peak Power Meters.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Live Indicators',
                  'A "LIVE" badge and "Last Update" timestamp confirm that the data you are seeing is current. '
                      'If the stream disconnects, monitoring will stop automatically.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Usage Tip',
                  'Continuous monitoring consumes more system resources (CPU and Network) than standard captures. '
                      'It is recommended to stop monitoring when not actively observing the signal.',
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

  Widget _buildSelectionPanel(ThemeData theme) {
    return ContentCard(
      isSidebar: true, // Radius 24
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'SELECTION',
            style: GoogleFonts.inter(
              fontWeight: FontWeight.bold,
              fontSize: 12,
              color: theme.colorScheme.primary,
            ),
          ),
          const SizedBox(height: 24),
          _buildDropdown(
            label: 'Type',
            value: _selectedType,
            items: _metadata?.instrumentTypes ?? [],
            onChanged: _isMonitoring
                ? null
                : (val) => setState(() {
                    _selectedType = val;
                    _selectedInstrument = null;
                  }),
          ),
          const SizedBox(height: 16),
          _buildDropdown(
            label: 'Instrument',
            value: _selectedInstrument,
            items: _selectedType != null
                ? (_metadata?.instruments[_selectedType] ?? [])
                : [],
            onChanged: _isMonitoring
                ? null
                : (val) => setState(() => _selectedInstrument = val),
          ),
          const Spacer(),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: (_selectedType != null && _selectedInstrument != null)
                  ? (_isMonitoring ? _stopMonitor : _startMonitor)
                  : null,
              style: ElevatedButton.styleFrom(
                backgroundColor: _isMonitoring
                    ? Colors.red
                    : theme.colorScheme.primary,
                padding: const EdgeInsets.symmetric(vertical: 20),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
              child: Text(
                _isMonitoring ? 'STOP MONITOR' : 'START MONITOR',
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMonitorView(ThemeData theme) {
    if (!_isMonitoring) {
      return ContentCard(
        isSidebar: false,
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.monitor_rounded,
                size: 64,
                color: Colors.grey.shade300,
              ),
              const SizedBox(height: 16),
              const Text('Select instrument to start monitor'),
            ],
          ),
        ),
      );
    }

    return StreamBuilder<MonitorResponse>(
      stream: _monitorStream,
      builder: (context, snapshot) {
        if (snapshot.hasError) {
          return Center(child: Text('Error: ${snapshot.error}'));
        }
        if (!snapshot.hasData) {
          return const Center(child: CircularProgressIndicator());
        }

        final data = snapshot.data!;
        if (!data.ok) {
          return Center(
            child: Text(
              'Device Error: ${data.message}',
              style: TextStyle(color: theme.colorScheme.error),
            ),
          );
        }

        if (_selectedType == 'SA' || _selectedType == 'VSA') {
          return _buildImageMonitor(data);
        } else {
          return _buildPowerMeterMonitor(data);
        }
      },
    );
  }

  Widget _buildImageMonitor(MonitorResponse data) {
    if (data.image.isEmpty)
      return const Center(child: Text('Waiting for image...'));
      
    Uint8List imageBytes;
    try {
      String cleanBase64 = data.image;
      if (cleanBase64.contains(',')) {
        cleanBase64 = cleanBase64.split(',').last;
      }
      cleanBase64 = cleanBase64.replaceAll(RegExp(r'\s+'), '');
      int padding = cleanBase64.length % 4;
      if (padding != 0) {
        cleanBase64 += '=' * (4 - padding);
      }
      imageBytes = base64Decode(cleanBase64);
    } catch (e) {
      debugPrint('Error decoding base64 image: $e');
      imageBytes = Uint8List(0);
    }

    return Container(
      decoration: BoxDecoration(
        color: Colors.black,
        borderRadius: BorderRadius.circular(24),
      ),
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          Positioned.fill(
            child: imageBytes.isEmpty
                ? const Center(
                    child: Text(
                      'Failed to decode image data',
                      style: TextStyle(color: Colors.red),
                    ),
                  )
                : Image.memory(
                    imageBytes,
                    fit: BoxFit.contain,
                    gaplessPlayback: true,
                    errorBuilder: (context, error, stackTrace) {
                      return Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.broken_image, size: 48, color: Colors.grey),
                            const SizedBox(height: 8),
                            Text(
                              'Image render error:\n$error',
                              textAlign: TextAlign.center,
                              style: const TextStyle(color: Colors.red),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
          ),
          if (_lastUpdateTime != null)
            Positioned(
              top: 16,
              right: 16,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.black54,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  'Last Update: ${_lastUpdateTime!.hour.toString().padLeft(2, '0')}:${_lastUpdateTime!.minute.toString().padLeft(2, '0')}:${_lastUpdateTime!.second.toString().padLeft(2, '0')}',
                  style: GoogleFonts.shareTechMono(
                    color: Colors.white,
                    fontSize: 10,
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildPowerMeterMonitor(MonitorResponse data) {
    final isPPM = _selectedType == 'PPM';

    return Container(
      padding: const EdgeInsets.all(32),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1A1A), // Dark background for "screen" feel
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: Colors.grey.shade800, width: 4),
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                _selectedInstrument ?? 'Power Meter',
                style: GoogleFonts.shareTechMono(
                  color: Colors.greenAccent,
                  fontSize: 18,
                ),
              ),
              Row(
                children: [
                  if (_lastUpdateTime != null)
                    Padding(
                      padding: const EdgeInsets.only(right: 12),
                      child: Text(
                        'Last Update: ${_lastUpdateTime!.hour.toString().padLeft(2, '0')}:${_lastUpdateTime!.minute.toString().padLeft(2, '0')}:${_lastUpdateTime!.second.toString().padLeft(2, '0')}',
                        style: GoogleFonts.shareTechMono(
                          color: Colors.grey,
                          fontSize: 12,
                        ),
                      ),
                    ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.greenAccent.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(
                      'LIVE',
                      style: GoogleFonts.shareTechMono(
                        color: Colors.greenAccent,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 40),
          Expanded(
            child: Row(
              children: [
                Expanded(
                  child: _buildChannelDisplay(
                    'CHANNEL A',
                    isPPM ? data.ppmChannelAPeakPower : data.pmChannelA,
                    isPPM ? data.ppmChannelAAvgPower : null,
                  ),
                ),
                const VerticalDivider(
                  color: Colors.white10,
                  width: 64,
                  thickness: 1,
                ),
                Expanded(
                  child: _buildChannelDisplay(
                    'CHANNEL B',
                    isPPM ? data.ppmChannelBPeakPower : data.pmChannelB,
                    isPPM ? data.ppmChannelBAvgPower : null,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChannelDisplay(String label, double mainVal, double? subVal) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.shareTechMono(
            color: Colors.blueAccent,
            fontSize: 14,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 16),
        Row(
          crossAxisAlignment: CrossAxisAlignment.baseline,
          textBaseline: TextBaseline.alphabetic,
          children: [
            Text(
              mainVal.toStringAsFixed(3),
              style: GoogleFonts.orbitron(
                fontSize: 64,
                fontWeight: FontWeight.bold,
                color: Colors.greenAccent,
                shadows: [
                  Shadow(
                    color: Colors.greenAccent.withOpacity(0.5),
                    blurRadius: 10,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            Text(
              'dBm',
              style: GoogleFonts.shareTechMono(
                color: Colors.greenAccent,
                fontSize: 24,
              ),
            ),
          ],
        ),
        if (subVal != null) ...[
          const SizedBox(height: 16),
          Text(
            'AVG POWER',
            style: GoogleFonts.shareTechMono(color: Colors.grey, fontSize: 12),
          ),
          Row(
            crossAxisAlignment: CrossAxisAlignment.baseline,
            textBaseline: TextBaseline.alphabetic,
            children: [
              Text(
                subVal.toStringAsFixed(3),
                style: GoogleFonts.orbitron(
                  fontSize: 24,
                  color: Colors.orangeAccent,
                ),
              ),
              const SizedBox(width: 4),
              Text(
                'dBm',
                style: GoogleFonts.shareTechMono(
                  color: Colors.orangeAccent,
                  fontSize: 12,
                ),
              ),
            ],
          ),
        ],
        const Spacer(),
        // Simple "Bar Meter" visual
        Container(
          height: 8,
          width: double.infinity,
          decoration: BoxDecoration(
            color: Colors.grey.shade900,
            borderRadius: BorderRadius.circular(4),
          ),
          child: FractionallySizedBox(
            widthFactor: ((mainVal + 100) / 130).clamp(0.0, 1.0),
            alignment: Alignment.centerLeft,
            child: Container(
              decoration: BoxDecoration(
                color: Colors.greenAccent,
                borderRadius: BorderRadius.circular(4),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildDropdown({
    required String label,
    required String? value,
    required List<String> items,
    required void Function(String?)? onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label.toUpperCase(),
          style: const TextStyle(
            fontSize: 11,
            fontWeight: FontWeight.bold,
            color: Colors.grey,
          ),
        ),
        const SizedBox(height: 8),
        DropdownButtonFormField<String>(
          value: value,
          onChanged: onChanged,
          decoration: InputDecoration(
            filled: true,
            fillColor: Colors.grey.shade50,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide.none,
            ),
          ),
          items: items
              .map(
                (e) => DropdownMenuItem(
                  value: e,
                  child: Text(
                    e,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
              )
              .toList(),
        ),
      ],
    );
  }
}
