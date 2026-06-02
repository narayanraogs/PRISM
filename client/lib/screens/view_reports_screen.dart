import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:pdfrx/pdfrx.dart';
import 'dart:js_interop';
import 'package:web/web.dart' as web;
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';
import '../services/server_service.dart';

class ViewReportsScreen extends StatefulWidget {
  final bool isActive;
  const ViewReportsScreen({super.key, this.isActive = true});

  @override
  State<ViewReportsScreen> createState() => _ViewReportsScreenState();
}

// Local wrapper to avoid repeated parsing
class _ReportEntry {
  final ReportMetadata metadata;
  final DateTime dateTime;

  _ReportEntry({required this.metadata, required this.dateTime});
}

class _ViewReportsScreenState extends State<ViewReportsScreen> {
  final List<_ReportEntry> _allEntries = [];
  List<_ReportEntry> _filteredResultEntries = [];
  _ReportEntry? _selectedEntry;

  bool _isLoading = true;
  String? _errorMessage;

  // Filters
  String _searchQuery = '';
  String _selectedPhase = 'All';
  String _selectedType = 'All';
  DateTimeRange? _dateRange;

  Uint8List? _pdfData;
  bool _isPdfLoading = false;
  String? _pdfErrorMessage;

  // Unique filter values derived from data
  List<String> _phases = ['All'];
  List<String> _types = ['All'];

  // Global parameter lists for regeneration
  List<String> _allVsaParams = [];
  List<String> _selectedVsaParams = [];
  List<String> _allPpmParams = [];
  List<String> _selectedPpmParams = [];
  bool _isHelpOpen = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchData();
    });
  }

  @override
  void didUpdateWidget(ViewReportsScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isActive && !oldWidget.isActive) {
      // Screen became active again, fetch new data
      _fetchData();
    }
  }

  DateTime _parseDateTime(String date, String time) {
    try {
      // Input can be: "12-SEP-2025" "15:10:33" OR "12-09-2025" "15:10:33"
      final parts = date.split('-');
      if (parts.length == 3) {
        // Check if it's a numeric month (like 09) or named month (like SEP)
        final monthItem = parts[1];
        if (int.tryParse(monthItem) == null) {
          // It's a named month, normalize it for DateFormat
          final month = monthItem.toLowerCase();
          parts[1] = month[0].toUpperCase() + month.substring(1);
          final normalizedDate = parts.join('-');
          return DateFormat(
            "dd-MMM-yyyy HH:mm:ss",
          ).parse("$normalizedDate $time");
        } else {
          // It's a numeric month
          final normalizedDate = parts.join('-');
          return DateFormat(
            "dd-MM-yyyy HH:mm:ss",
          ).parse("$normalizedDate $time");
        }
      }
      return DateTime.now();
    } catch (e) {
      debugPrint('Error parsing date: $date $time -> $e');
      return DateTime.now();
    }
  }

  Future<void> _fetchData() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final service = Provider.of<ServerService>(context, listen: false);
    final response = await service.fetchReportsMetadata();

    if (response != null && response.ok) {
      debugPrint('ViewReportsScreen: Using Bootstrapped Metadata');
      _allEntries.clear();
      for (var meta in response.reports) {
        _allEntries.add(
          _ReportEntry(
            metadata: meta,
            dateTime: _parseDateTime(meta.date, meta.time),
          ),
        );
      }

      // Derive unique filters
      _phases = ['All'];
      _types = ['All'];
      for (var entry in _allEntries) {
        if (!_phases.contains(entry.metadata.phase)) {
          _phases.add(entry.metadata.phase);
        }
        if (!_types.contains(entry.metadata.testType)) {
          _types.add(entry.metadata.testType);
        }
      }

      _allEntries.sort((a, b) => b.dateTime.compareTo(a.dateTime));

      // Store global parameter lists
      _allVsaParams = response.allVsaParams;
      _selectedVsaParams = List.from(response.selectedVsaParams);
      _allPpmParams = response.allPpmParams;
      _selectedPpmParams = List.from(response.selectedPpmParams);

      _applyFilters();
      setState(() => _isLoading = false);
    } else {
      debugPrint('ViewReportsScreen: Failed to fetch results from server');
      setState(() {
        _errorMessage = 'Failed to fetch results';
        _isLoading = false;
      });
    }
  }

  Future<void> _fetchPDF(_ReportEntry entry) async {
    setState(() {
      _isPdfLoading = true;
      _pdfErrorMessage = null;
      _pdfData = null;
    });

    try {
      final service = Provider.of<ServerService>(context, listen: false);
      final response = await service.fetchReportPDF(
        entry.metadata.date,
        entry.metadata.time,
      );

      if (response != null && response.ok) {
        setState(() {
          _pdfData = base64Decode(response.message);
        });
      } else {
        setState(() {
          _pdfErrorMessage = response?.message ?? 'Failed to load PDF';
        });
      }
    } catch (e) {
      setState(() {
        _pdfErrorMessage = 'Error: $e';
      });
    } finally {
      setState(() => _isPdfLoading = false);
    }
  }

  void _showRegenDialog(_ReportEntry entry) {
    // Note: We use the local _selectedVsaParams and _selectedPpmParams
    // which represent the current global selection state.
    final res = entry.metadata;
    final theme = Theme.of(context);

    showDialog(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            return AlertDialog(
              title: Text(
                'Regenerate Report',
                style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
              ),
              content: SizedBox(
                width: 500,
                child: SingleChildScrollView(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Adjust global parameters for all ${res.vsaUsed ? "VSA" : ""}${res.vsaUsed && res.ppmUsed ? "/" : ""}${res.ppmUsed ? "PPM" : ""} reports.',
                        style: const TextStyle(
                          fontSize: 13,
                          color: Colors.grey,
                        ),
                      ),
                      const SizedBox(height: 24),
                      if (res.vsaUsed) ...[
                        _buildDialogSection(
                          'VSA Parameters',
                          _allVsaParams,
                          _selectedVsaParams,
                          setDialogState,
                          theme,
                        ),
                        const SizedBox(height: 16),
                      ],
                      if (res.ppmUsed) ...[
                        _buildDialogSection(
                          'PPM Parameters',
                          _allPpmParams,
                          _selectedPpmParams,
                          setDialogState,
                          theme,
                        ),
                      ],
                    ],
                  ),
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('Cancel'),
                ),
                ElevatedButton(
                  onPressed: () async {
                    Navigator.pop(context);
                    setState(() {
                      _isPdfLoading = true;
                      _pdfErrorMessage = null;
                      _pdfData = null;
                    });

                    try {
                      final service = Provider.of<ServerService>(
                        context,
                        listen: false,
                      );
                      final response = await service.regenerateReport(
                        date: res.date,
                        time: res.time,
                        ppmParameters: _selectedPpmParams,
                        vsaParameters: _selectedVsaParams,
                      );

                      if (response != null && response.ok) {
                        setState(() {
                          _pdfData = base64Decode(response.message);
                        });
                      } else {
                        setState(() {
                          _pdfErrorMessage =
                              response?.message ??
                              'Failed to regenerate report';
                        });
                      }
                    } catch (e) {
                      setState(() {
                        _pdfErrorMessage = 'Error: $e';
                      });
                    } finally {
                      setState(() => _isPdfLoading = false);
                    }
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: theme.colorScheme.primary,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  child: const Text('Save & Regenerate'),
                ),
              ],
            );
          },
        );
      },
    );
  }

  Widget _buildDialogSection(
    String title,
    List<String> all,
    List<String> selected,
    StateSetter setDialogState,
    ThemeData theme,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              title,
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
            ),
            Row(
              children: [
                TextButton(
                  onPressed: () => setDialogState(() {
                    selected.clear();
                    selected.addAll(all);
                  }),
                  child: const Text('All', style: TextStyle(fontSize: 12)),
                ),
                TextButton(
                  onPressed: () => setDialogState(() {
                    selected.clear();
                  }),
                  child: const Text('None', style: TextStyle(fontSize: 12)),
                ),
              ],
            ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: all.map((param) {
            final isSelected = selected.contains(param);
            return FilterChip(
              label: Text(param, style: const TextStyle(fontSize: 11)),
              selected: isSelected,
              onSelected: (val) {
                setDialogState(() {
                  if (val) {
                    selected.add(param);
                  } else {
                    selected.remove(param);
                  }
                });
              },
              selectedColor: theme.colorScheme.primary.withOpacity(0.2),
              checkmarkColor: theme.colorScheme.primary,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
                side: BorderSide(
                  color: isSelected
                      ? theme.colorScheme.primary
                      : Colors.grey.shade300,
                ),
              ),
            );
          }).toList(),
        ),
      ],
    );
  }

  void _downloadPDF() {
    if (_pdfData == null || _selectedEntry == null) return;
    final blob = web.Blob(
      [_pdfData!.toJS].toJS,
      web.BlobPropertyBag(type: 'application/pdf'),
    );
    final url = web.URL.createObjectURL(blob);
    final anchor = web.document.createElement('a') as web.HTMLAnchorElement;
    anchor.href = url;
    anchor.download =
        'Report_${_selectedEntry!.metadata.config}_${_selectedEntry!.metadata.date}.pdf';
    
    // Append to DOM, click, and remove to ensure browser respects download filename
    web.document.body?.appendChild(anchor);
    anchor.click();
    anchor.remove();

    // Delay revocation to ensure download is successfully initialized
    Future.delayed(const Duration(milliseconds: 200), () {
      web.URL.revokeObjectURL(url);
    });
  }

  void _applyFilters() {
    setState(() {
      _selectedEntry = null;
      _pdfData = null;
      _pdfErrorMessage = null;
      _filteredResultEntries = _allEntries.where((entry) {
        final meta = entry.metadata;
        final matchesSearch =
            meta.remarks.toLowerCase().contains(_searchQuery.toLowerCase()) ||
            meta.config.toLowerCase().contains(_searchQuery.toLowerCase());
        final matchesPhase =
            _selectedPhase == 'All' || meta.phase == _selectedPhase;
        final matchesType =
            _selectedType == 'All' || meta.testType == _selectedType;
        final matchesDate =
            _dateRange == null ||
            (entry.dateTime.isAfter(_dateRange!.start) &&
                entry.dateTime.isBefore(
                  _dateRange!.end.add(const Duration(days: 1)),
                ));

        return matchesSearch && matchesPhase && matchesType && matchesDate;
      }).toList();
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Column(
        children: [
          ScreenHeader(
            title: 'Test Reports',
            subtitle:
                'Showing ${_filteredResultEntries.length} of ${_allEntries.length} results',
            icon: Icons.assignment,
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                _buildRefreshButton(theme),
                const SizedBox(width: 12),
                _buildDateRangePicker(theme),
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
                    // Left Side: Filters and Table
                    Expanded(
                      flex: 6,
                      child: ContentCard(
                        margin: const EdgeInsets.all(16),
                        width: double.infinity,
                        child: Column(
                          children: [
                            _buildControls(theme),
                            const Divider(height: 1),
                            Expanded(
                              child: _isLoading
                                  ? const Center(
                                      child: CircularProgressIndicator(),
                                    )
                                  : _errorMessage != null
                                  ? Center(
                                      child: Column(
                                        mainAxisAlignment:
                                            MainAxisAlignment.center,
                                        children: [
                                          Icon(
                                            Icons.error_outline,
                                            size: 48,
                                            color: theme.colorScheme.error,
                                          ),
                                          const SizedBox(height: 16),
                                          Text(
                                            _errorMessage!,
                                            style: TextStyle(
                                              color: theme.colorScheme.error,
                                            ),
                                          ),
                                          const SizedBox(height: 16),
                                          ElevatedButton(
                                            onPressed: _fetchData,
                                            child: const Text('Retry'),
                                          ),
                                        ],
                                      ),
                                    )
                                  : _filteredResultEntries.isEmpty
                                  ? const Center(
                                      child: Text(
                                        'No reports found matching filters',
                                      ),
                                    )
                                  : _buildDataTable(theme),
                            ),
                          ],
                        ),
                      ),
                    ),

                    // Right Side: Report Preview
                    Expanded(
                      flex: 4,
                      child: ContentCard(
                        margin: const EdgeInsets.fromLTRB(0, 16, 16, 16),
                        width: double.infinity,
                        child: _isLoading
                            ? const Center(child: CircularProgressIndicator())
                            : _buildPreviewArea(theme),
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
          ),
        ],
      ),
    );
  }

  Widget _buildControls(ThemeData theme) {
    return Padding(
      padding: const EdgeInsets.all(24.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          TextField(
            onChanged: (value) {
              _searchQuery = value;
              _applyFilters();
            },
            decoration: InputDecoration(
              hintText: 'Search by configuration or remarks...',
              prefixIcon: const Icon(Icons.search, size: 20),
              filled: true,
              fillColor: Colors.grey.shade50,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide.none,
              ),
              contentPadding: const EdgeInsets.symmetric(vertical: 16),
            ),
          ),
          const SizedBox(height: 16),
          _buildFilterRibbon(theme),
        ],
      ),
    );
  }

  Widget _buildDateRangePicker(ThemeData theme) {
    return InkWell(
      onTap: () async {
        final picked = await showDateRangePicker(
          context: context,
          firstDate: DateTime(2020),
          lastDate: DateTime.now(),
          initialDateRange: _dateRange,
          builder: (context, child) {
            return Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(
                  maxWidth: 600,
                  maxHeight: 700,
                ),
                child: Theme(
                  data: Theme.of(context).copyWith(
                    colorScheme: theme.colorScheme,
                    // Ensure the background is handled correctly
                    useMaterial3: true,
                  ),
                  child: Dialog(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    clipBehavior: Clip.antiAlias,
                    child: child!,
                  ),
                ),
              ),
            );
          },
        );
        if (picked != null) {
          setState(() => _dateRange = picked);
          _applyFilters();
        }
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: theme.colorScheme.primary.withOpacity(0.1),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(
              Icons.calendar_today,
              size: 16,
              color: theme.colorScheme.primary,
            ),
            const SizedBox(width: 12),
            Text(
              _dateRange == null
                  ? 'All Time'
                  : '${DateFormat('MMM d').format(_dateRange!.start)} - ${DateFormat('MMM d').format(_dateRange!.end)}',
              style: TextStyle(
                color: theme.colorScheme.primary,
                fontWeight: FontWeight.bold,
                fontSize: 13,
              ),
            ),
            if (_dateRange != null) ...[
              const SizedBox(width: 8),
              GestureDetector(
                onTap: () {
                  setState(() => _dateRange = null);
                  _applyFilters();
                },
                child: Icon(
                  Icons.close,
                  size: 14,
                  color: theme.colorScheme.primary,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildRefreshButton(ThemeData theme) {
    return InkWell(
      onTap: _fetchData,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: theme.colorScheme.primary.withOpacity(0.1),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(Icons.refresh, size: 16, color: theme.colorScheme.primary),
            const SizedBox(width: 12),
            Text(
              'Refresh',
              style: TextStyle(
                color: theme.colorScheme.primary,
                fontWeight: FontWeight.bold,
                fontSize: 13,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFilterRibbon(ThemeData theme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildChipRow('Phases', _phases, _selectedPhase, (val) {
          setState(() => _selectedPhase = val);
          _applyFilters();
        }, theme),
        const SizedBox(height: 8),
        _buildChipRow('Types', _types, _selectedType, (val) {
          setState(() => _selectedType = val);
          _applyFilters();
        }, theme),
      ],
    );
  }

  Widget _buildChipRow(
    String label,
    List<String> options,
    String selected,
    Function(String) onSelected,
    ThemeData theme,
  ) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(top: 8.0),
          child: SizedBox(
            width: 80,
            child: Text(
              label,
              style: const TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.bold,
                color: Colors.grey,
              ),
            ),
          ),
        ),
        Expanded(
          child: Wrap(
            spacing: 8.0, // Horizontal space between chips
            runSpacing: 4.0, // Vertical space between lines
            children: options.map((opt) {
              final isSelected = selected == opt;
              return ChoiceChip(
                label: Text(opt),
                selected: isSelected,
                onSelected: (s) => onSelected(opt),
                labelStyle: TextStyle(
                  color: isSelected ? Colors.white : Colors.grey.shade700,
                  fontSize: 12,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                ),
                selectedColor: theme.colorScheme.primary,
                backgroundColor: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: BorderSide(
                    color: isSelected
                        ? Colors.transparent
                        : Colors.grey.shade300,
                  ),
                ),
              );
            }).toList(),
          ),
        ),
      ],
    );
  }

  Widget _buildDataTable(ThemeData theme) {
    return SingleChildScrollView(
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(16),
        child: DataTable(
          headingRowHeight: 48,
          dataRowHeight: 72,
          showCheckboxColumn: false,
          columns: const [
            DataColumn(label: Text('DATE & TIME')),
            DataColumn(label: Text('CONFIG')),
            DataColumn(label: Text('TEST / CATEGORY')),
            DataColumn(label: Text('REMARKS')),
          ],
          rows: _filteredResultEntries.map((entry) {
            final res = entry.metadata;
            final isSelected = _selectedEntry == entry;
            return DataRow(
              selected: isSelected,
              onSelectChanged: (val) {
                setState(() => _selectedEntry = entry);
                if (val == true) {
                  _fetchPDF(entry);
                }
              },
              cells: [
                DataCell(
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        DateFormat('MMM dd, yyyy').format(entry.dateTime),
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                      Text(
                        DateFormat('HH:mm').format(entry.dateTime),
                        style: TextStyle(
                          color: Colors.grey.shade500,
                          fontSize: 11,
                        ),
                      ),
                    ],
                  ),
                ),
                DataCell(
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.secondary.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      res.config,
                      style: TextStyle(
                        color: theme.colorScheme.secondary,
                        fontWeight: FontWeight.bold,
                        fontSize: 12,
                      ),
                    ),
                  ),
                ),
                DataCell(
                  Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        res.testType,
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                      Text(
                        res.testCategory,
                        style: TextStyle(
                          color: Colors.grey.shade500,
                          fontSize: 11,
                        ),
                      ),
                    ],
                  ),
                ),
                DataCell(
                  SizedBox(
                    width: 200,
                    child: Text(
                      res.remarks,
                      style: TextStyle(
                        color: Colors.grey.shade600,
                        fontSize: 12,
                      ),
                      overflow: TextOverflow.ellipsis,
                      maxLines: 2,
                    ),
                  ),
                ),
              ],
            );
          }).toList(),
        ),
      ),
    );
  }

  Widget _buildPreviewArea(ThemeData theme) {
    if (_selectedEntry == null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.description_outlined,
              size: 64,
              color: Colors.grey.shade200,
            ),
            const SizedBox(height: 16),
            const Text(
              'Select a report to preview',
              style: TextStyle(color: Colors.grey),
            ),
          ],
        ),
      );
    }

    final res = _selectedEntry!.metadata;
    final dateTime = _selectedEntry!.dateTime;

    return Column(
      children: [
        // Preview Header
        Padding(
          padding: const EdgeInsets.all(20.0),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: theme.colorScheme.primary.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  Icons.picture_as_pdf,
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Report: ${res.config}',
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    Text(
                      '${DateFormat('MMMM d, yyyy').format(dateTime)} at ${DateFormat('HH:mm').format(dateTime)}',
                      style: const TextStyle(fontSize: 12, color: Colors.grey),
                    ),
                  ],
                ),
              ),
              IconButton(
                onPressed: _pdfData != null ? _downloadPDF : null,
                icon: const Icon(Icons.download_rounded),
              ),
              IconButton(
                onPressed: () {}, // TODO: Implement Print if needed
                icon: const Icon(Icons.print_rounded),
              ),
              IconButton(
                onPressed: (res.vsaUsed || res.ppmUsed)
                    ? () => _showRegenDialog(_selectedEntry!)
                    : null,
                icon: const Icon(Icons.refresh_rounded),
                tooltip: (res.vsaUsed || res.ppmUsed)
                    ? 'Regenerate/Reload'
                    : 'Regeneration only for VSA/PPM',
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        // PDF Content
        Expanded(
          child: Container(
            margin: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: Colors.grey.shade50,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: _isPdfLoading
                ? const Center(child: CircularProgressIndicator())
                : _pdfErrorMessage != null
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.error_outline,
                          size: 48,
                          color: theme.colorScheme.error,
                        ),
                        const SizedBox(height: 16),
                        Text(
                          _pdfErrorMessage!,
                          style: TextStyle(color: theme.colorScheme.error),
                        ),
                        const SizedBox(height: 16),
                        ElevatedButton(
                          onPressed: () => _fetchPDF(_selectedEntry!),
                          child: const Text('Retry'),
                        ),
                      ],
                    ),
                  )
                : _pdfData == null
                ? const Center(
                    child: Text('Click a report to load PDF preview'),
                  )
                : ClipRRect(
                    borderRadius: BorderRadius.circular(12),
                    child: PdfViewer.data(
                      _pdfData!,
                      sourceName:
                          '${res.date}_${res.time}_${DateTime.now().millisecondsSinceEpoch}',
                      params: const PdfViewerParams(maxScale: 10.0),
                    ),
                  ),
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
                  'Reports Help',
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
                  'Finding Reports',
                  'Use the search bar to filter by **Configuration Name** or **Remarks**. Use the Ribbon filters '
                      'below it to drill down by Test Phase (Payload, RF, etc.) or Test Type.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Date Filtering',
                  'Click the "All Time" (or specific date range) badge in the top right header to filter results '
                      'within a specific time window.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Report Regeneration',
                  'If you need to change which parameters (VSA/PPM) are included in the PDF, click "Regenerate" '
                      'on a selected report. This will compute a fresh PDF with your updated parameter selection.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Downloading',
                  'Once a PDF is loaded in the preview pane, use the Download icon in the top right of the preview '
                      'card to save the file locally.',
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
