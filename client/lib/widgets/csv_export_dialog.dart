import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:prism_client/screens/path_loss_planner_screen.dart';

class CsvExportDialog extends StatefulWidget {
  final List<PlannerNode> sources;
  final List<PlannerNode> sinks;
  final Function(String sourceId, String saSinkId, String pmSinkId, String? scSinkId) onExport;

  const CsvExportDialog({
    super.key,
    required this.sources,
    required this.sinks,
    required this.onExport,
  });

  @override
  State<CsvExportDialog> createState() => _CsvExportDialogState();
}

class _CsvExportDialogState extends State<CsvExportDialog> {
  String? _selectedSourceId;
  String? _selectedSaSinkId;
  String? _selectedPmSinkId;
  String? _selectedScSinkId;

  final _formKey = GlobalKey<FormState>();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      elevation: 0,
      backgroundColor: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500),
        padding: const EdgeInsets.all(32),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.1),
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
          ],
        ),
        child: Form(
          key: _formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.primary.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(
                      Icons.description_outlined,
                      color: theme.colorScheme.primary,
                      size: 28,
                    ),
                  ),
                  const SizedBox(width: 16),
                  Text(
                    'Export Path Loss to CSV',
                    style: GoogleFonts.outfit(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                      color: Colors.grey.shade900,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                'Select the source and destination nodes to generate the path loss report.',
                style: GoogleFonts.inter(
                  fontSize: 14,
                  color: Colors.grey.shade600,
                ),
              ),
              const SizedBox(height: 32),
              
              _buildDropdown(
                label: 'Source (Mandatory)',
                hint: 'Select Source Node',
                value: _selectedSourceId,
                items: widget.sources,
                onChanged: (val) => setState(() => _selectedSourceId = val),
                validator: (val) => val == null ? 'Please select a source' : null,
              ),
              const SizedBox(height: 20),
              
              _buildDropdown(
                label: 'Sink for SA Path (Mandatory)',
                hint: 'Select SA Path Destination',
                value: _selectedSaSinkId,
                items: widget.sinks,
                onChanged: (val) => setState(() => _selectedSaSinkId = val),
                validator: (val) => val == null ? 'Please select a sink for SA path' : null,
              ),
              const SizedBox(height: 20),
              
              _buildDropdown(
                label: 'Sink for PM Path (Mandatory)',
                hint: 'Select PM Path Destination',
                value: _selectedPmSinkId,
                items: widget.sinks,
                onChanged: (val) => setState(() => _selectedPmSinkId = val),
                validator: (val) => val == null ? 'Please select a sink for PM path' : null,
              ),
              const SizedBox(height: 20),
              
              _buildDropdown(
                label: 'Sink for Spacecraft Path (Optional)',
                hint: 'Select Spacecraft Path Destination',
                value: _selectedScSinkId,
                items: widget.sinks,
                onChanged: (val) => setState(() => _selectedScSinkId = val),
                isMandatory: false,
              ),
              
              const SizedBox(height: 40),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                    ),
                    child: Text(
                      'CANCEL',
                      style: GoogleFonts.inter(
                        fontWeight: FontWeight.w600,
                        color: Colors.grey.shade600,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  ElevatedButton(
                    onPressed: () {
                      if (_formKey.currentState!.validate()) {
                        widget.onExport(
                          _selectedSourceId!,
                          _selectedSaSinkId!,
                          _selectedPmSinkId!,
                          _selectedScSinkId,
                        );
                        Navigator.pop(context);
                      }
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: theme.colorScheme.primary,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    child: Text(
                      'EXPORT CSV',
                      style: GoogleFonts.inter(
                        fontWeight: FontWeight.bold,
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

  Widget _buildDropdown({
    required String label,
    required String hint,
    required String? value,
    required List<PlannerNode> items,
    required Function(String?) onChanged,
    String? Function(String?)? validator,
    bool isMandatory = true,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: Colors.grey.shade700,
          ),
        ),
        const SizedBox(height: 8),
        DropdownButtonFormField<String>(
          initialValue: value,
          hint: Text(hint, style: GoogleFonts.inter(fontSize: 14, color: Colors.grey.shade400)),
          items: items.map((node) {
            return DropdownMenuItem<String>(
              value: node.id,
              child: Text(node.label, style: GoogleFonts.inter(fontSize: 14)),
            );
          }).toList(),
          onChanged: onChanged,
          validator: validator,
          decoration: InputDecoration(
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade300),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade300),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade500, width: 2),
            ),
            filled: true,
            fillColor: Colors.grey.shade50,
          ),
          icon: const Icon(Icons.keyboard_arrow_down_rounded),
        ),
      ],
    );
  }
}
