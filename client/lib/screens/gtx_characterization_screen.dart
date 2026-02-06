import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class GTxCharacterizationScreen extends StatefulWidget {
  const GTxCharacterizationScreen({super.key});

  @override
  State<GTxCharacterizationScreen> createState() =>
      _GTxCharacterizationScreenState();
}

class _GTxCharacterizationScreenState extends State<GTxCharacterizationScreen> {
  // Mock Data
  final List<String> _profiles = [
    'GTx-Profile-A',
    'GTx-Profile-B',
    'Dual-Source-X',
  ];
  String _selectedProfile = 'GTx-Profile-A';

  final List<String> _components = ['IFM-1', 'IFM-2', 'Direct-Connect'];
  String _selectedComponent = 'IFM-1';

  final List<String> _modulations = ["PM", "FM", "CDMA", "PSK", "FSK"];
  String _selectedModulation = "PM";

  bool _isHelpOpen = false;
  bool _isMeasuring = false;

  // Controllers
  final _cableLossController = TextEditingController(text: "1.2");
  final _ifController = TextEditingController(text: "70,000,000");
  final _subCarFreqController = TextEditingController(text: "8,000");
  final _modIndexController = TextEditingController(text: "1.0");

  // Mock Measurement Results
  final List<Map<String, String>> _mockResults = [
    {"Parameter": "Power Level", "Measured": "-12.4 dBm", "Status": "PASS"},
    {
      "Parameter": "Center Frequency",
      "Measured": "70.0001 MHz",
      "Status": "PASS",
    },
    {
      "Parameter": "In-Band Spurious",
      "Measured": "-65.0 dBc",
      "Status": "PASS",
    },
    {
      "Parameter": "Out-of-Band Spurious",
      "Measured": "-72.1 dBc",
      "Status": "PASS",
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
                _buildHeader(theme),
                const SizedBox(height: 32),
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildProfileSidebar(theme),
                      const SizedBox(width: 32),
                      _buildMainConfiguration(theme),
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
            right: _isHelpOpen ? 0 : -400,
            top: 0,
            bottom: 0,
            child: _buildHelpPanel(theme),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(ThemeData theme) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: theme.colorScheme.primary.withOpacity(0.1),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Icon(Icons.radar, color: theme.colorScheme.primary, size: 32),
        ),
        const SizedBox(width: 20),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'GTx Characterization',
              style: GoogleFonts.outfit(
                fontSize: 28,
                fontWeight: FontWeight.w900,
                color: Colors.black,
              ),
            ),
            Text(
              'Detailed performance analysis of Generator Source',
              style: TextStyle(
                color: Colors.grey.shade600,
                fontSize: 14,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
        const Spacer(),
        _buildInstrumentStatus(theme, "GTx-GEN-01", true),
        const SizedBox(width: 16),
        _buildInstrumentStatus(theme, "SA-SPEC-04", true),
        const SizedBox(width: 24),
        IconButton(
          onPressed: () => setState(() => _isHelpOpen = !_isHelpOpen),
          icon: const Icon(Icons.help_outline),
          style: IconButton.styleFrom(
            backgroundColor: _isHelpOpen
                ? theme.colorScheme.primary
                : Colors.white,
            foregroundColor: _isHelpOpen
                ? Colors.white
                : theme.colorScheme.primary,
          ),
        ),
      ],
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
                final isSelected = _selectedProfile == _profiles[index];
                return Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: InkWell(
                    onTap: () =>
                        setState(() => _selectedProfile = _profiles[index]),
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
                          Text(
                            _profiles[index],
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
                      setState(() => _isMeasuring = !_isMeasuring);
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
                      'LIVE SPECTRUM RESULTS',
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
                ..._mockResults.map((res) => _buildResultRow(res)),
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
                  _buildLogEntry("Initializing GTx Source..."),
                  _buildLogEntry("Applying IF offset 70.0 MHz..."),
                  _buildLogEntry("Calibrating Cable Loss...", isPending: true),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildResultRow(Map<String, String> res) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                res["Parameter"]!,
                style: const TextStyle(color: Colors.white70, fontSize: 12),
              ),
              const SizedBox(height: 4),
              Text(
                res["Measured"]!,
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
              color: Colors.green.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.green.withOpacity(0.3)),
            ),
            child: Text(
              res["Status"]!,
              style: const TextStyle(
                color: Colors.green,
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
                controller: TextEditingController(text: "1,000,000"),
                icon: Icons.unfold_more,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _buildTextField(
                label: "RBW (Hz)",
                controller: TextEditingController(text: "3,000"),
                icon: Icons.grid_view,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _buildTextField(
                label: "VBW (Hz)",
                controller: TextEditingController(text: "1,000"),
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
                    controller: TextEditingController(text: "200,000"),
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
                  "About Characterization",
                  "This screen automates the performance verification of RF Generator sources. It measures Power, Center Frequency, and Spectral Purity (Spurious).",
                ),
                const SizedBox(height: 32),
                _buildHelpItem(
                  "Instrument Setup",
                  "Ensure the GTx is connected to the Spectrum Analyzer via the selected component path. Cable loss must be calibrated for accurate power readings.",
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHelpItem(String title, String content) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: GoogleFonts.inter(fontWeight: FontWeight.bold, fontSize: 16),
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
}
