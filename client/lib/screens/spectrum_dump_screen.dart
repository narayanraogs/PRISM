import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/screens/spectrum_dump_view_screen.dart';

class SpectrumDumpScreen extends StatefulWidget {
  const SpectrumDumpScreen({super.key});

  @override
  State<SpectrumDumpScreen> createState() => _SpectrumDumpScreenState();
}

class _SpectrumDumpScreenState extends State<SpectrumDumpScreen> {
  SpectrumDumpMetadata? _metadata;
  bool _isLoading = true;
  String? _errorMessage;

  String? _selectedMode;
  String? _selectedInstrument;

  // Spectrum Dump Specific
  List<SpectrumProfile> _spectrumProfiles = [];
  String? _selectedProfile;
  final TextEditingController _cfController = TextEditingController();
  final TextEditingController _spanController = TextEditingController();
  final TextEditingController _vbwController = TextEditingController();
  final TextEditingController _rbwController = TextEditingController();
  final TextEditingController _refLevelController = TextEditingController();
  final TextEditingController _tracePointsController = TextEditingController(
    text: '1001',
  );
  bool _autoRefLevel = true;
  String? _selectedCaptureMode = 'Clear Write';

  // Screenshot Specific
  String? _selectedScreenshotProfile;

  @override
  void initState() {
    super.initState();
    _loadMetadata();
  }

  void _loadMetadata() {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final service = Provider.of<ServerService>(context, listen: false);
    final metadata = service.status.bootstrapData?.spectrumDumpData;

    if (metadata != null) {
      debugPrint('SpectrumDumpScreen: Using Bootstrapped Metadata');
      setState(() {
        _metadata = metadata;
        _spectrumProfiles = metadata.spectrumProfiles;
        _isLoading = false;
      });
    } else {
      debugPrint('SpectrumDumpScreen: Bootstrapped Metadata NOT FOUND');
      setState(() {
        _errorMessage = 'Failed to load metadata';
        _isLoading = false;
      });
    }
  }

  @override
  void dispose() {
    _cfController.dispose();
    _spanController.dispose();
    _vbwController.dispose();
    _rbwController.dispose();
    _refLevelController.dispose();
    _tracePointsController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    if (_isLoading) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const CircularProgressIndicator(),
            const SizedBox(height: 16),
            Text(
              'Loading metadata...',
              style: GoogleFonts.inter(color: Colors.grey.shade600),
            ),
          ],
        ),
      );
    }

    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.error_outline, size: 48, color: theme.colorScheme.error),
            const SizedBox(height: 16),
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
      body: Padding(
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
                  // 1. Mode & Instrument Selection
                  Expanded(flex: 3, child: _buildSelectionPanel(theme)),
                  const SizedBox(width: 20),
                  // 2. Details Pane
                  Expanded(flex: 7, child: _buildDetailsPanel(theme)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(ThemeData theme) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: theme.colorScheme.primary.withOpacity(0.1),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Icon(
            Icons.camera_alt_rounded,
            color: theme.colorScheme.primary,
            size: 32,
          ),
        ),
        const SizedBox(width: 20),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Spectrum Dump / Screenshot',
              style: GoogleFonts.outfit(
                fontSize: 28,
                fontWeight: FontWeight.w900,
                color: Colors.black,
              ),
            ),
            Text(
              'Capture spectrum data or instrument screenshots',
              style: TextStyle(
                color: Colors.grey.shade600,
                fontSize: 14,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildSelectionPanel(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.all(24),
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
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'SELECTION',
              style: GoogleFonts.inter(
                fontSize: 12,
                fontWeight: FontWeight.w900,
                color: theme.colorScheme.primary,
                letterSpacing: 1.5,
              ),
            ),
            const SizedBox(height: 24),
            _buildDropdown(
              theme,
              label: 'Mode',
              value: _selectedMode,
              items: _metadata?.spectrumDumpMode ?? [],
              icon: Icons.tune_rounded,
              onChanged: (val) {
                setState(() {
                  _selectedMode = val;
                  _selectedInstrument = null;
                });
              },
            ),
            const SizedBox(height: 20),
            _buildDropdown(
              theme,
              label: 'Instrument',
              value: _selectedInstrument,
              items: _selectedMode == 'Spectrum Dump'
                  ? (_metadata?.instruments['SA'] ?? [])
                  : (_selectedMode == 'Screenshot'
                        ? (_metadata?.instruments['VSA'] ?? [])
                        : []),
              icon: Icons.terminal_outlined,
              onChanged: (val) {
                setState(() {
                  _selectedInstrument = val;
                });
              },
              enabled: _selectedMode != null,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDetailsPanel(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.all(24),
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
          Text(
            'DETAILS',
            style: GoogleFonts.inter(
              fontSize: 12,
              fontWeight: FontWeight.w900,
              color: theme.colorScheme.primary,
              letterSpacing: 1.5,
            ),
          ),
          const SizedBox(height: 24),
          Expanded(
            child: SingleChildScrollView(child: _buildDetailsContent(theme)),
          ),
          const SizedBox(height: 16),
          if (_selectedMode == 'Spectrum Dump') ...[
            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _selectedInstrument != null
                        ? _executeSetSpectrum
                        : null,
                    icon: const Icon(Icons.settings_remote_rounded, size: 18),
                    label: const Text('SET SPECTRUM'),
                    style: OutlinedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _selectedInstrument != null
                        ? _executeReadFromSA
                        : null,
                    icon: const Icon(Icons.download_rounded, size: 18),
                    label: const Text('READ FROM SA'),
                    style: OutlinedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _selectedInstrument != null
                        ? _executeCaptureTrace
                        : null,
                    icon: const Icon(Icons.shutter_speed_rounded, size: 18),
                    label: const Text('CAPTURE TRACE'),
                    style: OutlinedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
          ],
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: (_selectedMode != null && _selectedInstrument != null)
                  ? _executeCapture
                  : null,
              icon: Icon(
                _selectedMode == 'Screenshot'
                    ? Icons.camera_alt_rounded
                    : Icons.analytics_rounded,
              ),
              label: Text(
                _selectedMode == 'Screenshot'
                    ? 'CAPTURE SCREENSHOT'
                    : 'DUMP SPECTRUM',
              ),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 20),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                elevation: 0,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailsContent(ThemeData theme) {
    if (_selectedMode == 'Spectrum Dump') {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildDropdown(
            theme,
            label: 'Profile Name',
            value: _selectedProfile,
            items: _spectrumProfiles.map((p) => p.profileName).toList(),
            icon: Icons.analytics_outlined,
            onChanged: (val) {
              setState(() {
                _selectedProfile = val;
                final profile = _spectrumProfiles.firstWhere(
                  (p) => p.profileName == val,
                );
                _cfController.text = profile.centerFrequency.toString();
                _spanController.text = profile.span.toString();
                _vbwController.text = profile.vbw.toString();
                _rbwController.text = profile.rbw.toString();
              });
            },
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _buildTextField(
                  theme,
                  label: 'Center Freq (MHz)',
                  controller: _cfController,
                  icon: Icons.settings_input_antenna,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _buildTextField(
                  theme,
                  label: 'Span (MHz)',
                  controller: _spanController,
                  icon: Icons.unfold_more_rounded,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _buildTextField(
                  theme,
                  label: 'VBW (Hz)',
                  controller: _vbwController,
                  icon: Icons.waves_rounded,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _buildTextField(
                  theme,
                  label: 'RBW (Hz)',
                  controller: _rbwController,
                  icon: Icons.grid_view_rounded,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: Row(
                  children: [
                    Checkbox(
                      value: _autoRefLevel,
                      onChanged: (val) {
                        setState(() {
                          _autoRefLevel = val ?? true;
                        });
                      },
                    ),
                    Text(
                      'Auto Ref Level',
                      style: GoogleFonts.inter(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: Colors.grey.shade700,
                      ),
                    ),
                  ],
                ),
              ),
              if (!_autoRefLevel)
                Expanded(
                  child: _buildTextField(
                    theme,
                    label: 'Ref Level (dBm)',
                    controller: _refLevelController,
                    icon: Icons.linear_scale_rounded,
                  ),
                ),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: _buildDropdown(
                  theme,
                  label: 'Mode of Capture',
                  value: _selectedCaptureMode,
                  items: [
                    'Clear Write',
                    'Max Hold',
                    'Min Hold',
                    'Average Hold',
                  ],
                  icon: Icons.flash_on_rounded,
                  onChanged: (val) {
                    setState(() {
                      _selectedCaptureMode = val;
                    });
                  },
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _buildTextField(
                  theme,
                  label: 'Trace Points',
                  controller: _tracePointsController,
                  icon: Icons.grain_rounded,
                ),
              ),
            ],
          ),
        ],
      );
    } else if (_selectedMode == 'Screenshot') {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildDropdown(
            theme,
            label: 'Screenshot Profile',
            value: _selectedScreenshotProfile,
            items: _metadata?.screenshotProfiles ?? [],
            icon: Icons.image_search_rounded,
            onChanged: (val) {
              setState(() {
                _selectedScreenshotProfile = val;
              });
            },
          ),
        ],
      );
    } else {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 40.0),
        child: Center(
          child: Text(
            'Select a mode to continue',
            style: TextStyle(color: Colors.grey.shade400),
          ),
        ),
      );
    }
  }

  Widget _buildDropdown(
    ThemeData theme, {
    required String label,
    required String? value,
    required List<String> items,
    required IconData icon,
    required ValueChanged<String?> onChanged,
    bool enabled = true,
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
        DropdownButtonFormField<String>(
          value: value,
          onChanged: enabled ? onChanged : null,
          decoration: InputDecoration(
            prefixIcon: Icon(
              icon,
              color: enabled ? theme.colorScheme.primary : Colors.grey.shade300,
            ),
            filled: true,
            fillColor: enabled ? Colors.white : Colors.grey.shade50,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 16,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
          ),
          hint: Text(
            'Select $label',
            style: TextStyle(color: Colors.grey.shade400, fontSize: 14),
          ),
          items: items.map((String item) {
            return DropdownMenuItem<String>(
              value: item,
              child: Text(
                item,
                style: GoogleFonts.inter(
                  fontWeight: FontWeight.bold,
                  fontSize: 15,
                ),
              ),
            );
          }).toList(),
          icon: const Icon(Icons.keyboard_arrow_down_rounded),
          isExpanded: true,
        ),
      ],
    );
  }

  Widget _buildTextField(
    ThemeData theme, {
    required String label,
    required TextEditingController controller,
    required IconData icon,
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
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          decoration: InputDecoration(
            prefixIcon: Icon(icon, color: theme.colorScheme.primary, size: 20),
            filled: true,
            fillColor: Colors.white,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 12,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade200),
            ),
          ),
          style: GoogleFonts.inter(fontWeight: FontWeight.bold, fontSize: 14),
        ),
      ],
    );
  }

  void _executeCapture() {
    if (_selectedMode == 'Spectrum Dump') {
      _executeDumpSpectrum();
    } else if (_selectedMode == 'Screenshot') {
      _executeCaptureScreenshot();
    }
  }

  Future<void> _executeCaptureScreenshot() async {
    if (_selectedInstrument == null || _selectedScreenshotProfile == null)
      return;

    final service = Provider.of<ServerService>(context, listen: false);
    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );

    // Show loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final result = await service.dumpScreenshot(
        vsa: _selectedInstrument!,
        mode: _selectedScreenshotProfile!,
      );

      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        if (result != null && result.ok) {
          notificationService.addNotification(
            title: 'Screenshot Captured',
            message:
                'Screenshot captured successfully from $_selectedInstrument',
            type: NotificationType.success,
          );

          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => SpectrumDumpViewScreen(
                base64Image: result.message, // Image is in Message field
                instrumentName: _selectedInstrument!,
              ),
            ),
          );
        } else {
          notificationService.addNotification(
            title: 'Screenshot Failed',
            message:
                result?.message ??
                'Failed to capture screenshot from $_selectedInstrument',
            type: NotificationType.error,
          );
          showDialog(
            context: context,
            builder: (context) => AlertDialog(
              title: const Text('Error'),
              content: Text(result?.message ?? 'Failed to capture screenshot'),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        notificationService.addNotification(
          title: 'Screenshot Error',
          message: 'Error capturing screenshot: $e',
          type: NotificationType.error,
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _executeDumpSpectrum() async {
    if (_selectedInstrument == null) return;

    final service = Provider.of<ServerService>(context, listen: false);
    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );

    // Show loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final result = await service.dumpSpectrum(_selectedInstrument!);

      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        if (result != null && result.ok) {
          notificationService.addNotification(
            title: 'Spectrum Dump Success',
            message:
                'Spectrum dump captured successfully from $_selectedInstrument',
            type: NotificationType.success,
          );

          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => SpectrumDumpViewScreen(
                base64Image: result.message, // Image is in Message field
                instrumentName: _selectedInstrument!,
              ),
            ),
          );
        } else {
          notificationService.addNotification(
            title: 'Spectrum Dump Failed',
            message:
                result?.message ??
                'Failed to capture spectrum dump from $_selectedInstrument',
            type: NotificationType.error,
          );
          showDialog(
            context: context,
            builder: (context) => AlertDialog(
              title: const Text('Error'),
              content: Text(
                result?.message ?? 'Failed to capture spectrum dump',
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        notificationService.addNotification(
          title: 'Spectrum Dump Error',
          message: 'Error captured spectrum dump: $e',
          type: NotificationType.error,
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _executeSetSpectrum() async {
    if (_selectedInstrument == null) return;

    final cf = double.tryParse(_cfController.text) ?? 0.0;
    final span = double.tryParse(_spanController.text) ?? 0.0;
    final rbw = double.tryParse(_rbwController.text) ?? 0.0;
    final vbw = double.tryParse(_vbwController.text) ?? 0.0;
    final refLevel = double.tryParse(_refLevelController.text) ?? 0.0;

    final service = Provider.of<ServerService>(context, listen: false);
    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );

    // Show loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final result = await service.setSpectrum(
        sa: _selectedInstrument!,
        centerFrequency: cf,
        span: span,
        rbw: rbw,
        vbw: vbw,
        autoReference: _autoRefLevel,
        referenceLevel: refLevel,
        mode: _selectedCaptureMode ?? 'Clear Write',
      );

      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        if (result != null && result.ok) {
          notificationService.addNotification(
            title: 'Spectrum Set',
            message:
                'Spectrum settings applied successfully to $_selectedInstrument',
            type: NotificationType.success,
          );
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Spectrum settings applied successfully'),
              backgroundColor: Colors.green,
            ),
          );
        } else {
          notificationService.addNotification(
            title: 'Spectrum Set Failed',
            message:
                result?.message ??
                'Failed to apply settings to $_selectedInstrument',
            type: NotificationType.error,
          );
          showDialog(
            context: context,
            builder: (context) => AlertDialog(
              title: const Text('Error'),
              content: Text(result?.message ?? 'Failed to apply settings'),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        notificationService.addNotification(
          title: 'Spectrum Set Error',
          message: 'Error applying settings: $e',
          type: NotificationType.error,
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _executeReadFromSA() async {
    if (_selectedInstrument == null) return;

    final service = Provider.of<ServerService>(context, listen: false);
    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );

    // Show loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final result = await service.readSpectrum(_selectedInstrument!);

      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        if (result != null && result.ok) {
          setState(() {
            _cfController.text = result.centerFrequency.toString();
            _spanController.text = result.span.toString();
            _rbwController.text = result.rbw.toString();
            _vbwController.text = result.vbw.toString();
            _refLevelController.text = result.referenceLevel.toString();
            _autoRefLevel = false;
          });
          notificationService.addNotification(
            title: 'Read From SA',
            message:
                'Spectrum settings read successfully from $_selectedInstrument',
            type: NotificationType.success,
          );
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Spectrum settings read successfully'),
              backgroundColor: Colors.green,
            ),
          );
        } else {
          notificationService.addNotification(
            title: 'Read From SA Failed',
            message:
                result?.message ??
                'Failed to read settings from $_selectedInstrument',
            type: NotificationType.error,
          );
          showDialog(
            context: context,
            builder: (context) => AlertDialog(
              title: const Text('Error'),
              content: Text(result?.message ?? 'Failed to read settings'),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        notificationService.addNotification(
          title: 'Read From SA Error',
          message: 'Error reading settings: $e',
          type: NotificationType.error,
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _executeCaptureTrace() async {
    if (_selectedInstrument == null) return;

    final service = Provider.of<ServerService>(context, listen: false);
    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );

    int tracePoints = int.tryParse(_tracePointsController.text) ?? 1001;

    // Show loading dialog
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    try {
      final result = await service.dumpTrace(
        sa: _selectedInstrument!,
        tracePoints: tracePoints,
      );

      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        if (result != null && result.ok) {
          notificationService.addNotification(
            title: 'Capture Trace Success',
            message: 'Trace captured successfully from $_selectedInstrument',
            type: NotificationType.success,
          );

          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => SpectrumDumpViewScreen(
                base64Image: result.message, // PNG is in Message field
                instrumentName: _selectedInstrument!,
              ),
            ),
          );
        } else {
          notificationService.addNotification(
            title: 'Capture Trace Failed',
            message:
                result?.message ??
                'Failed to capture trace from $_selectedInstrument',
            type: NotificationType.error,
          );
          showDialog(
            context: context,
            builder: (context) => AlertDialog(
              title: const Text('Error'),
              content: Text(result?.message ?? 'Failed to capture trace'),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('OK'),
                ),
              ],
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        Navigator.pop(context); // Close loading dialog
        notificationService.addNotification(
          title: 'Capture Trace Error',
          message: 'Error capturing trace: $e',
          type: NotificationType.error,
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }
}
