import 'package:flutter/material.dart';
import 'package:prism_client/screens/stability_monitoring_screen.dart';

import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'dart:convert';
import 'dart:js_interop';
import 'package:web/web.dart' as web;

class StabilityScreen extends StatefulWidget {
  const StabilityScreen({super.key});

  @override
  State<StabilityScreen> createState() => _StabilityScreenState();
}

class _StabilityScreenState extends State<StabilityScreen> {
  StabilityMetadata? _metadata;
  bool _isLoading = true;
  bool _isHelpOpen = false;
  String? _errorMessage;

  String? _selectedType;
  String? _selectedInstrument;
  String? _selectedParameter;
  final TextEditingController _descriptionController = TextEditingController();
  final TextEditingController _profileNameController = TextEditingController(
    text: 'Standard Stability Profile',
  );

  // SA Specific
  List<SpectrumProfile> _spectrumProfiles = [];
  String? _selectedProfile;
  final TextEditingController _cfController = TextEditingController();
  final TextEditingController _spanController = TextEditingController();
  final TextEditingController _vbwController = TextEditingController();
  final TextEditingController _rbwController = TextEditingController();
  final TextEditingController _refLevelController = TextEditingController();
  bool _autoRefLevel = true;

  // PM Specific
  final TextEditingController _pmFrequencyController = TextEditingController();
  String _pmFrequencyUnit = 'MHz';

  // PPM Specific
  List<String> _ppmPLConfigs = [];
  List<String> _ppmPulseProfiles = [];
  List<String> _ppmChannels = [];
  String? _selectedPpmPLConfig;
  String? _selectedPpmPulseProfile;
  String? _selectedPpmChannel;

  // TM Specific
  final TextEditingController _tmMnemonicController = TextEditingController();

  final List<StabilityParameterSelection> _selectedParameters = [];

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
    final metadata = service.status.bootstrapData?.stabilityData;

    if (metadata != null) {
      debugPrint('StabilityScreen: Using Bootstrapped Metadata');
      setState(() {
        _metadata = metadata;
        _spectrumProfiles = metadata.profiles;
        _ppmPLConfigs = metadata.plConfigs;
        _ppmPulseProfiles = metadata.pulseProfiles;
        _ppmChannels = metadata.ppmChannels;
        _isLoading = false;
      });
    } else {
      debugPrint('StabilityScreen: Bootstrapped Metadata NOT FOUND');
      setState(() {
        _errorMessage = 'Failed to load metadata';
        _isLoading = false;
      });
    }
  }

  void _addParameter() {
    final description = _descriptionController.text.trim();
    if (description.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a unique description')),
      );
      return;
    }

    if (_selectedParameters.any((p) => p.description == description)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Description must be unique')),
      );
      return;
    }

    if (_selectedType == null ||
        _selectedInstrument == null ||
        _selectedParameter == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select all required fields')),
      );
      return;
    }

    Map<String, dynamic>? extraDetails;
    if (_selectedType == 'SA') {
      if (!_autoRefLevel && _refLevelController.text.trim().isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Please enter Reference Level')),
        );
        return;
      }
      extraDetails = {
        'profileName': _selectedProfile,
        'centerFrequency': double.tryParse(_cfController.text),
        'span': double.tryParse(_spanController.text),
        'vbw': double.tryParse(_vbwController.text),
        'rbw': double.tryParse(_rbwController.text),
        'autoRefLevel': _autoRefLevel,
        'refLevel': _autoRefLevel
            ? null
            : double.tryParse(_refLevelController.text),
      };
    } else if (_selectedType == 'PM') {
      final freq = double.tryParse(_pmFrequencyController.text);
      if (freq == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Please enter a valid Frequency')),
        );
        return;
      }

      double multiplier = 1.0;
      switch (_pmFrequencyUnit) {
        case 'GHz':
          multiplier = 1e9;
          break;
        case 'MHz':
          multiplier = 1e6;
          break;
        case 'KHz':
          multiplier = 1e3;
          break;
        case 'Hz':
          multiplier = 1.0;
          break;
      }

      extraDetails = {'frequencyHz': freq * multiplier};
    } else if (_selectedType == 'PPM') {
      if (_selectedPpmPLConfig == null ||
          _selectedPpmPulseProfile == null ||
          _selectedPpmChannel == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Please select Channel, PL Config and Pulse Profile'),
          ),
        );
        return;
      }
      extraDetails = {
        'channel': _selectedPpmChannel,
        'plConfig': _selectedPpmPLConfig,
        'pulseProfile': _selectedPpmPulseProfile,
      };
    } else if (_selectedType == 'TM') {
      final mnemonic = _tmMnemonicController.text.trim();
      if (mnemonic.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Please enter TM Mnemonic')),
        );
        return;
      }
      extraDetails = {'mnemonic': mnemonic};
    }
    setState(() {
      _selectedParameters.add(
        StabilityParameterSelection(
          description: description,
          instrumentType: _selectedType!,
          instrument: _selectedInstrument!,
          parameter: _selectedParameter!,
          details: '',
          extraDetails: extraDetails,
        ),
      );
      // Reset selections for next add
      _descriptionController.clear();
      _selectedProfile = null;
      _cfController.clear();
      _spanController.clear();
      _vbwController.clear();
      _rbwController.clear();
      _refLevelController.clear();
      _autoRefLevel = true;
      _pmFrequencyController.clear();
      _selectedPpmPLConfig = null;
      _selectedPpmPulseProfile = null;
      _selectedPpmChannel = null;
      _tmMnemonicController.clear();
      _selectedParameter = null;
    });
  }

  void _removeParameter(int index) {
    setState(() {
      _selectedParameters.removeAt(index);
    });
  }

  void _saveProfile() {
    if (_selectedParameters.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Monitoring list is empty')));
      return;
    }

    final data = {
      'profileName': _profileNameController.text,
      'parameters': _selectedParameters.map((e) => e.toJson()).toList(),
    };
    final jsonContent = jsonEncode(data);
    final bytes = utf8.encode(jsonContent);
    final blob = web.Blob([bytes.toJS].toJS);
    final url = web.URL.createObjectURL(blob);
    final anchor = web.document.createElement('a') as web.HTMLAnchorElement;
    anchor.href = url;
    anchor.download =
        "${_profileNameController.text.trim().replaceAll(' ', '_')}.json";
    anchor.click();
    web.URL.revokeObjectURL(url);
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(const SnackBar(content: Text('Stability profile exported')));
  }

  void _loadProfile() {
    final uploadInput =
        web.document.createElement('input') as web.HTMLInputElement;
    uploadInput.type = 'file';
    uploadInput.accept = '.json';
    uploadInput.click();

    uploadInput.onChange.listen((e) {
      final files = uploadInput.files;
      if (files == null || files.length == 0) return;

      final reader = web.FileReader();
      reader.readAsText(files.item(0)!);
      reader.onLoadEnd.listen((e) {
        try {
          final result = reader.result as JSString;
          final dynamic decoded = jsonDecode(result.toDart);

          setState(() {
            if (decoded is Map && decoded.containsKey('parameters')) {
              _profileNameController.text =
                  decoded['profileName'] ?? 'Loaded Profile';
              final List params = decoded['parameters'];
              _selectedParameters.clear();
              _selectedParameters.addAll(
                params
                    .map((e) => StabilityParameterSelection.fromJson(e))
                    .toList(),
              );
            }
          });
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Stability profile loaded')),
          );
        } catch (e) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Failed to load profile: Invalid JSON'),
            ),
          );
        }
      });
    });
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
      body: Stack(
        children: [
          Padding(
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
                      // 1. Basic Selection
                      Expanded(
                        flex: 3,
                        child: _buildBasicSelectionPanel(theme),
                      ),
                      const SizedBox(width: 20),
                      // 2. Instrument Specific Details
                      Expanded(
                        flex: 4,
                        child: _buildInstrumentDetailsPanel(theme),
                      ),
                      const SizedBox(width: 20),
                      // 3. Monitoring List
                      Expanded(flex: 5, child: _buildListPanel(theme)),
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

  Widget _buildBasicSelectionPanel(ThemeData theme) {
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
              'BASIC INFO',
              style: GoogleFonts.inter(
                fontSize: 12,
                fontWeight: FontWeight.w900,
                color: theme.colorScheme.primary,
                letterSpacing: 1.5,
              ),
            ),
            const SizedBox(height: 24),
            Text(
              'DESCRIPTION',
              style: GoogleFonts.inter(
                fontSize: 11,
                fontWeight: FontWeight.w900,
                color: Colors.grey.shade500,
                letterSpacing: 1.2,
              ),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _descriptionController,
              decoration: InputDecoration(
                hintText: 'Entry name',
                hintStyle: TextStyle(color: Colors.grey.shade400, fontSize: 13),
                filled: true,
                fillColor: Colors.white,
                prefixIcon: Icon(
                  Icons.label_outline,
                  color: theme.colorScheme.primary,
                  size: 20,
                ),
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
            ),
            const SizedBox(height: 20),
            _buildDropdownEffect(
              theme,
              label: 'Type',
              value: _selectedType,
              items: _metadata?.instrumentTypes ?? [],
              icon: Icons.category_outlined,
              onChanged: (val) {
                setState(() {
                  _selectedType = val;
                  _selectedInstrument = null;
                  _selectedParameter = null;
                });
              },
            ),
            const SizedBox(height: 20),
            _buildDropdownEffect(
              theme,
              label: 'Instrument',
              value: _selectedInstrument,
              items: _selectedType != null
                  ? (_metadata?.instruments[_selectedType] ?? [])
                  : [],
              icon: Icons.terminal_outlined,
              onChanged: (val) {
                setState(() {
                  _selectedInstrument = val;
                });
              },
              enabled: _selectedType != null,
            ),
            const SizedBox(height: 20),
            _buildDropdownEffect(
              theme,
              label: 'Parameter',
              value: _selectedParameter,
              items: _selectedType != null
                  ? (_metadata?.parameters[_selectedType] ?? [])
                  : [],
              icon: Icons.tune_outlined,
              onChanged: (val) {
                setState(() {
                  _selectedParameter = val;
                });
              },
              enabled: _selectedType != null,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInstrumentDetailsPanel(ThemeData theme) {
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
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'INSTRUMENT DETAILS',
                    style: GoogleFonts.inter(
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                      color: theme.colorScheme.primary,
                      letterSpacing: 1.5,
                    ),
                  ),
                  if (_selectedType == 'SA') ..._buildSADetails(theme),
                  if (_selectedType == 'PM') ..._buildPMDetails(theme),
                  if (_selectedType == 'PPM') ..._buildPPMDetails(theme),
                  if (_selectedType == 'TM') ..._buildTMDetails(theme),
                  if (_selectedType == null)
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 40.0),
                      child: Center(
                        child: Text(
                          'Select an instrument type to see details',
                          textAlign: TextAlign.center,
                          style: TextStyle(
                            color: Colors.grey.shade400,
                            fontSize: 13,
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: _addParameter,
              icon: const Icon(Icons.add_rounded),
              label: const Text('ADD TO MONITORING'),
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

  Widget _buildDropdownEffect(
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
            disabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade100),
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
          dropdownColor: Colors.white,
          isExpanded: true,
        ),
      ],
    );
  }

  Widget _buildListPanel(ThemeData theme) {
    return Column(
      children: [
        Expanded(
          child: Container(
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
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Padding(
                  padding: const EdgeInsets.all(24.0),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _profileNameController,
                          decoration: const InputDecoration(
                            isDense: true,
                            contentPadding: EdgeInsets.zero,
                            border: InputBorder.none,
                            hintText: 'Profile Name',
                          ),
                          style: GoogleFonts.inter(
                            fontSize: 12,
                            fontWeight: FontWeight.w900,
                            color: theme.colorScheme.primary,
                            letterSpacing: 1.5,
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: theme.colorScheme.primary.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Text(
                          '${_selectedParameters.length} ITEMS',
                          style: TextStyle(
                            color: theme.colorScheme.primary,
                            fontWeight: FontWeight.bold,
                            fontSize: 10,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      IconButton(
                        onPressed: _loadProfile,
                        icon: const Icon(Icons.file_upload_outlined, size: 20),
                        tooltip: 'Load Profile',
                        color: theme.colorScheme.primary,
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                      const SizedBox(width: 12),
                      IconButton(
                        onPressed: _saveProfile,
                        icon: const Icon(Icons.download_outlined, size: 20),
                        tooltip: 'Save Profile',
                        color: theme.colorScheme.primary,
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                    ],
                  ),
                ),
                const Divider(height: 1),
                Expanded(
                  child: _selectedParameters.isEmpty
                      ? _buildEmptyState(theme)
                      : ListView.separated(
                          padding: const EdgeInsets.all(32),
                          itemCount: _selectedParameters.length,
                          separatorBuilder: (context, index) =>
                              const SizedBox(height: 12),
                          itemBuilder: (context, index) {
                            final item = _selectedParameters[index];
                            return _buildParameterCard(theme, item, index);
                          },
                        ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 24),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton.icon(
            onPressed: _selectedParameters.isEmpty ? null : _startStability,
            icon: const Icon(Icons.play_arrow_rounded),
            label: const Text('START STABILITY MONITORING'),
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 24),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              elevation: 0,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildEmptyState(ThemeData theme) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.analytics_outlined, size: 64, color: Colors.grey.shade200),
          const SizedBox(height: 16),
          Text(
            'No parameters added yet',
            style: GoogleFonts.inter(
              color: Colors.grey.shade400,
              fontSize: 16,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Use the configuration panel to select and add items',
            style: GoogleFonts.inter(color: Colors.grey.shade300, fontSize: 14),
          ),
        ],
      ),
    );
  }

  Widget _buildParameterCard(
    ThemeData theme,
    StabilityParameterSelection item,
    int index,
  ) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: theme.colorScheme.primary.withOpacity(0.02),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.colorScheme.primary.withOpacity(0.1)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: theme.colorScheme.primary.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(
              _getIconForType(item.instrumentType),
              color: theme.colorScheme.primary,
              size: 24,
            ),
          ),
          const SizedBox(width: 20),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      item.description,
                      style: GoogleFonts.inter(
                        fontWeight: FontWeight.w900,
                        fontSize: 18,
                        color: theme.colorScheme.primary,
                      ),
                    ),
                    const Spacer(),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.primary.withOpacity(0.08),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        item.instrumentType,
                        style: TextStyle(
                          fontSize: 10,
                          color: theme.colorScheme.primary,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Icon(
                      _getIconForType(item.instrumentType),
                      size: 14,
                      color: Colors.grey.shade500,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      '${item.instrument} • ${item.parameter}',
                      style: GoogleFonts.inter(
                        color: Colors.grey.shade600,
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
                if (item.extraDetails != null) ...[
                  const SizedBox(height: 12),
                  _buildExtraDetailsGrid(theme, item.extraDetails!),
                ],
              ],
            ),
          ),
          const SizedBox(width: 12),
          IconButton(
            onPressed: () => _removeParameter(index),
            icon: Icon(
              Icons.delete_outline_rounded,
              color: theme.colorScheme.error,
              size: 24,
            ),
            style: IconButton.styleFrom(
              hoverColor: theme.colorScheme.error.withOpacity(0.05),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildExtraDetailsGrid(ThemeData theme, Map<String, dynamic> extra) {
    List<Widget> chips = [];

    extra.forEach((key, value) {
      if (value != null) {
        String label = key
            .replaceAllMapped(
              RegExp(r'([A-Z])'),
              (match) => ' ${match.group(0)}',
            )
            .trim();
        label = label.substring(0, 1).toUpperCase() + label.substring(1);

        chips.add(
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(6),
              border: Border.all(
                color: theme.colorScheme.primary.withOpacity(0.1),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label.toUpperCase(),
                  style: TextStyle(
                    fontSize: 8,
                    fontWeight: FontWeight.w900,
                    color: Colors.grey.shade500,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  value.toString(),
                  style: GoogleFonts.inter(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: Colors.black,
                  ),
                ),
              ],
            ),
          ),
        );
      }
    });

    return Wrap(spacing: 8, runSpacing: 8, children: chips);
  }

  IconData _getIconForType(String type) {
    switch (type) {
      case 'SA':
        return Icons.signal_cellular_alt_rounded;
      case 'PM':
        return Icons.speed_rounded;
      case 'PPM':
        return Icons.bolt_rounded;
      case 'TM':
        return Icons.settings_remote_rounded;
      default:
        return Icons.tune_rounded;
    }
  }

  @override
  void dispose() {
    _descriptionController.dispose();
    _cfController.dispose();
    _spanController.dispose();
    _vbwController.dispose();
    _rbwController.dispose();
    _refLevelController.dispose();
    _pmFrequencyController.dispose();
    _tmMnemonicController.dispose();
    _profileNameController.dispose();
    super.dispose();
  }

  List<Widget> _buildTMDetails(ThemeData theme) {
    return [
      const SizedBox(height: 24),
      Text(
        'TM DETAILS',
        style: GoogleFonts.inter(
          fontSize: 12,
          fontWeight: FontWeight.w900,
          color: theme.colorScheme.primary,
          letterSpacing: 1.5,
        ),
      ),
      const SizedBox(height: 20),
      _buildTextField(
        theme,
        label: 'TM Mnemonic',
        controller: _tmMnemonicController,
        icon: Icons.abc_rounded,
      ),
    ];
  }

  List<Widget> _buildPPMDetails(ThemeData theme) {
    return [
      const SizedBox(height: 24),
      Text(
        'PEAK POWER METER DETAILS',
        style: GoogleFonts.inter(
          fontSize: 12,
          fontWeight: FontWeight.w900,
          color: theme.colorScheme.primary,
          letterSpacing: 1.5,
        ),
      ),
      const SizedBox(height: 20),
      _buildDropdownEffect(
        theme,
        label: 'PL Configuration',
        value: _selectedPpmPLConfig,
        items: _ppmPLConfigs,
        icon: Icons.settings_applications_outlined,
        onChanged: (val) {
          setState(() {
            _selectedPpmPLConfig = val;
          });
        },
      ),
      const SizedBox(height: 16),
      _buildDropdownEffect(
        theme,
        label: 'Pulse Profile Name',
        value: _selectedPpmPulseProfile,
        items: _ppmPulseProfiles,
        icon: Icons.waves_rounded,
        onChanged: (val) {
          setState(() {
            _selectedPpmPulseProfile = val;
          });
        },
      ),
      const SizedBox(height: 16),
      _buildDropdownEffect(
        theme,
        label: 'Channel',
        value: _selectedPpmChannel,
        items: _ppmChannels,
        icon: Icons.settings_input_component_rounded,
        onChanged: (val) {
          setState(() {
            _selectedPpmChannel = val;
          });
        },
      ),
    ];
  }

  List<Widget> _buildPMDetails(ThemeData theme) {
    return [
      const SizedBox(height: 24),
      Text(
        'POWER METER DETAILS',
        style: GoogleFonts.inter(
          fontSize: 12,
          fontWeight: FontWeight.w900,
          color: theme.colorScheme.primary,
          letterSpacing: 1.5,
        ),
      ),
      const SizedBox(height: 20),
      Row(
        children: [
          Expanded(
            flex: 2,
            child: _buildTextField(
              theme,
              label: 'Frequency',
              controller: _pmFrequencyController,
              icon: Icons.settings_input_antenna,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            flex: 1,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'UNIT',
                  style: GoogleFonts.inter(
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                    color: Colors.grey.shade500,
                    letterSpacing: 1.2,
                  ),
                ),
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: Colors.grey.shade200),
                  ),
                  child: DropdownButtonHideUnderline(
                    child: DropdownButton<String>(
                      value: _pmFrequencyUnit,
                      isExpanded: true,
                      items: ['GHz', 'MHz', 'KHz', 'Hz'].map((String unit) {
                        return DropdownMenuItem<String>(
                          value: unit,
                          child: Text(
                            unit,
                            style: GoogleFonts.inter(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        );
                      }).toList(),
                      onChanged: (val) {
                        setState(() {
                          _pmFrequencyUnit = val!;
                        });
                      },
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    ];
  }

  List<Widget> _buildSADetails(ThemeData theme) {
    return [
      const SizedBox(height: 24),
      Text(
        'SPECTRUM ANALYZER DETAILS',
        style: GoogleFonts.inter(
          fontSize: 12,
          fontWeight: FontWeight.w900,
          color: theme.colorScheme.primary,
          letterSpacing: 1.5,
        ),
      ),
      const SizedBox(height: 24),
      _buildDropdownEffect(
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
    ];
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
          keyboardType: TextInputType.numberWithOptions(decimal: true),
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

  void _startStability() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        surfaceTintColor: Colors.white,
        backgroundColor: Colors.white,
        title: Text(
          'Start Stability Monitoring?',
          style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
        ),
        content: Text(
          'This will initiate real-time monitoring for ${_selectedParameters.length} parameters using the profile "${_profileNameController.text}".',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('CANCEL'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context); // Close dialog
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (context) => StabilityMonitoringScreen(
                    parameters: List.from(_selectedParameters),
                    profileName: _profileNameController.text,
                  ),
                ),
              );
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.primary,
              foregroundColor: Colors.white,
            ),
            child: const Text('CONFIRM'),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'STABILITY CONFIGURATION',
                style: GoogleFonts.outfit(
                  fontSize: 32,
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                'Define instrument parameters for long-term stability monitoring',
                style: GoogleFonts.inter(
                  fontSize: 16,
                  color: Colors.grey.shade600,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
          const Spacer(),
          _buildHelpTrigger(theme),
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
                  'Stability Help',
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
                  'Stability Monitoring',
                  'Long-term stability monitoring allows you to track specific RF parameters over time. '
                      'You can combine multiple measurements from different instruments into a single session.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Configuring Entries',
                  '1. **Basic Info**: Provide a unique description and select the equipment type.\n'
                      '2. **Instrument Details**: Configure specific settings. For SA, you can use existing profiles or set custom spans.\n'
                      '3. **Add to Monitoring**: Once configured, add the entry to your active list.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Instrument Types',
                  '• **SA (Spectrum Analyzer)**: Monitors power level, peak search, or trace data at specific frequencies.\n'
                      '• **PM (Power Meter)**: Continuous total power monitoring at a calibrated frequency.\n'
                      '• **PPM (Peak Power Meter)**: Specialized for pulsed signal analysis.\n'
                      '• **TM (Telemetry)**: Monitors specific satellite telemetry mnemonics via the backend.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Profiles',
                  'Profiles allow you to save your entire monitoring configuration setup. '
                      'This is useful for repetitive test campaigns involving many parameters.',
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
}
