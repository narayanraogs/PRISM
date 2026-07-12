import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';

enum InsightDataCategory {
  singleValue, // Category 1: Power, Frequency Error
  fixedMultiple, // Category 2: Lock Dynamic Range
  variableResults, // Category 3: Spurious, Harmonics
}

class InsightsScreen extends StatefulWidget {
  const InsightsScreen({super.key});

  @override
  State<InsightsScreen> createState() => _InsightsScreenState();
}

class _InsightsScreenState extends State<InsightsScreen> {
  InsightDataCategory _selectedCategory = InsightDataCategory.singleValue;
  final List<String> _selectedSessions = [];
  String? _referenceSession;
  bool _useMeanAsReference = false;
  bool _isHelpOpen = false;

  // Real Data State
  List<ReportMetadata> _allReports = [];
  List<ReportMetadata> _filteredReports = [];
  List<TestReport> _loadedReports = [];
  bool _isLoading = true;
  bool _isDataLoading = false;

  String? _selectedConfig;
  String? _selectedTestName;
  String? _selectedTestCategory;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchMetadata();
    });
  }

  Future<void> _fetchMetadata() async {
    setState(() => _isLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    final response = await serverService.fetchReportsMetadata();

    if (mounted) {
      if (response != null && response.ok) {
        setState(() {
          _allReports = response.reports.where((r) => r.success).toList();
          _applyFilters();
          _isLoading = false;
        });
        await _fetchReportsData();
      } else {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _fetchReportsData() async {
    if (_selectedSessions.isEmpty) {
      setState(() {
        _loadedReports = [];
      });
      return;
    }
    setState(() => _isDataLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);
    
    List<Map<String, String>> sessionKeysList = [];
    for (var id in _selectedSessions) {
      final parts = id.split('_');
      if (parts.length >= 2) {
        sessionKeysList.add({
          'date': parts[0],
          'time': parts[1],
        });
      }
    }

    final response = await serverService.fetchReportsData(sessionKeysList);
    if (mounted) {
      if (response != null && response.ok) {
        setState(() {
          _loadedReports = response.reports;
          _isDataLoading = false;
        });
      } else {
        setState(() => _isDataLoading = false);
      }
    }
  }

  void _applyFilters() {
    setState(() {
      _filteredReports = _allReports.where((r) {
        final matchConfig = _selectedConfig == null || r.config == _selectedConfig;
        final matchTest = _selectedTestName == null || r.testType == _selectedTestName;
        final matchCat = _selectedTestCategory == null || r.testCategory == _selectedTestCategory;
        return matchConfig && matchTest && matchCat;
      }).toList();

      if (_selectedTestName != null) {
        final test = _selectedTestName!.toLowerCase();
        if (test == "frequency" ||
            test == "pulsefrequency" ||
            test == "power" ||
            test == "bandwidth" ||
            test == "pulsebandwidth" ||
            test == "pulseuplink" ||
            test == "rfuplinkremoval" ||
            test == "trmanalysis") {
          _selectedCategory = InsightDataCategory.singleValue;
        } else if (test == "lockdynamic" ||
            test == "commanddynamic" ||
            test == "loopstress" ||
            test == "carrieracquisition" ||
            test == "rfuplink" ||
            test == "modindex" ||
            test == "frequencydeviationmeasurement" ||
            test == "pulsemeasurement" ||
            test == "pulseanalysis" ||
            test == "highresolutionpulse" ||
            test == "spectrogramanalysis") {
          _selectedCategory = InsightDataCategory.fixedMultiple;
        } else if (test == "spurious" || test == "harmonics") {
          _selectedCategory = InsightDataCategory.variableResults;
        }
      }

      _selectedSessions.clear();
      final count = math.min(5, _filteredReports.length);
      for (int i = 0; i < count; i++) {
        final id = "${_filteredReports[i].date}_${_filteredReports[i].time}";
        _selectedSessions.add(id);
      }
      if (_selectedSessions.isNotEmpty) {
        _referenceSession = _selectedSessions.first;
      }
    });
  }

  Map<String, dynamic> _extractSingleValuePoint(TestReport report) {
    double value = 0.0;
    String unit = "";
    String label = "Measured";
    double? upperSpec;
    double? lowerSpec;

    final testName = report.testType.toLowerCase();
    final resTable = report.results["Results"];

    if (resTable != null && resTable.data.isNotEmpty) {
      final row = resTable.data.first;
      final header = resTable.header;

      if (testName == "frequency" || testName == "pulsefrequency") {
        int measuredIdx = header.indexOf("Measured [MHz]");
        int devIdx = header.indexOf("Deviation [kHz]");
        int allowedIdx = header.indexOf("Allowed Deviation [kHz]");

        if (devIdx != -1 && devIdx < row.length) {
          value = double.tryParse(row[devIdx].value) ?? 0.0;
          unit = "kHz";
          label = "Deviation";
        } else if (measuredIdx != -1 && measuredIdx < row.length) {
          value = double.tryParse(row[measuredIdx].value) ?? 0.0;
          unit = "MHz";
          label = "Measured Frequency";
        }
        if (allowedIdx != -1 && allowedIdx < row.length) {
          double allowed = double.tryParse(row[allowedIdx].value) ?? 0.0;
          upperSpec = allowed;
          lowerSpec = -allowed;
        }
      } else if (testName == "power") {
        int measuredIdx = header.indexOf("Measured [dBm]");
        int specIdx = header.indexOf("Specified [dBm]");
        if (measuredIdx != -1 && measuredIdx < row.length) {
          value = double.tryParse(row[measuredIdx].value) ?? 0.0;
          unit = "dBm";
          label = "Measured Power";
        }
        if (specIdx != -1 && specIdx < row.length) {
          double spec = double.tryParse(row[specIdx].value) ?? 0.0;
          upperSpec = spec + 0.5;
          lowerSpec = spec - 0.5;
        }
      } else if (testName == "bandwidth" || testName == "pulsebandwidth") {
        int measuredIdx = header.indexOf("Measured BW [kHz]");
        int specIdx = header.indexOf("Specification BW [kHz]");
        if (measuredIdx != -1 && measuredIdx < row.length) {
          value = double.tryParse(row[measuredIdx].value) ?? 0.0;
          unit = "kHz";
          label = "Measured BW";
        }
        if (specIdx != -1 && specIdx < row.length) {
          upperSpec = double.tryParse(row[specIdx].value);
        }
      } else if (testName == "pulseuplink") {
        int measuredIdx = header.indexOf("Measured Power [dBm]");
        int setIdx = header.indexOf("Set Power [dBm]");
        if (measuredIdx != -1 && measuredIdx < row.length) {
          value = double.tryParse(row[measuredIdx].value) ?? 0.0;
          unit = "dBm";
          label = "Measured Power";
        }
        if (setIdx != -1 && setIdx < row.length) {
          double setVal = double.tryParse(row[setIdx].value) ?? 0.0;
          upperSpec = setVal + 0.5;
          lowerSpec = setVal - 0.5;
        }
      } else if (testName == "rfuplinkremoval") {
        int measuredIdx = header.indexOf("Rx I/P Power Measured (dBm)");
        int reqIdx = header.indexOf("Rx I/P Power Required (dBm)");
        if (measuredIdx != -1 && measuredIdx < row.length) {
          value = double.tryParse(row[measuredIdx].value) ?? 0.0;
          unit = "dBm";
          label = "Measured Power";
        }
        if (reqIdx != -1 && reqIdx < row.length) {
          double req = double.tryParse(row[reqIdx].value) ?? 0.0;
          upperSpec = req + 1.0;
          lowerSpec = req - 1.0;
        }
      } else if (testName == "trmanalysis") {
        int trmsIdx = header.indexOf("No Of TRMs");
        if (trmsIdx != -1 && trmsIdx < row.length) {
          value = double.tryParse(row[trmsIdx].value) ?? 0.0;
          unit = "";
          label = "No of TRMs";
        }
      }
    }

    return {
      "value": value,
      "unit": unit,
      "label": label,
      "upperSpec": upperSpec,
      "lowerSpec": lowerSpec,
      "date": report.date,
      "time": report.time,
    };
  }

  List<FlSpot> _extractFixedMultipleSpots(TestReport report) {
    List<FlSpot> spots = [];
    final testName = report.testType.toLowerCase();
    final resTable = report.results["Results"];

    if (resTable == null || resTable.data.isEmpty) {
      var key = report.results.keys.firstWhere((k) => k.contains("Pulse") || k.contains("Results"), orElse: () => "");
      if (key.isEmpty && report.results.isNotEmpty) {
        key = report.results.keys.first;
      }
      final table = report.results[key];
      if (table != null && table.data.isNotEmpty) {
        for (int i = 0; i < table.data.length; i++) {
          final row = table.data[i];
          int measuredIdx = table.header.indexOf("Measured");
          if (measuredIdx == -1) measuredIdx = table.header.indexOf("Mean");
          if (measuredIdx == -1) measuredIdx = table.header.indexOf("Measured [rad]");
          if (measuredIdx == -1 && row.length > 2) measuredIdx = 2;

          if (measuredIdx != -1 && measuredIdx < row.length) {
            double? yVal = double.tryParse(row[measuredIdx].value);
            if (yVal != null) {
              spots.add(FlSpot(i.toDouble(), yVal));
            }
          }
        }
      }
      return spots;
    }

    final header = resTable.header;
    final rows = resTable.data;

    if (testName == "lockdynamic" || testName == "commanddynamic" || testName == "rfuplink") {
      int xIdx = header.indexOf("Receiver Power (dBm)");
      if (xIdx == -1) xIdx = header.indexOf("Rx I/P Power Required (dBm)");
      if (xIdx == -1) xIdx = header.indexOf("Actual Power (dBm)");

      int yIdx = header.indexOf("AGC");
      if (yIdx == -1) yIdx = header.indexOf("Lock Status");

      if (xIdx != -1 && yIdx != -1) {
        for (var row in rows) {
          if (xIdx < row.length && yIdx < row.length) {
            double? xVal = double.tryParse(row[xIdx].value);
            double? yVal;
            if (header[yIdx].contains("Lock")) {
              final valLower = row[yIdx].value.toLowerCase();
              yVal = (valLower.contains("lock") || valLower.contains("yes") || valLower.contains("true") || valLower == "1") ? 1.0 : 0.0;
            } else {
              yVal = double.tryParse(row[yIdx].value);
            }
            if (xVal != null && yVal != null) {
              spots.add(FlSpot(xVal, yVal));
            }
          }
        }
      }
    } else if (testName == "loopstress") {
      int xIdx = header.indexOf("Frequency Offset (kHz)");
      int yIdx = header.indexOf("Loop Stress TM Value");
      if (xIdx != -1 && yIdx != -1) {
        for (var row in rows) {
          if (xIdx < row.length && yIdx < row.length) {
            double? xVal = double.tryParse(row[xIdx].value);
            double? yVal = double.tryParse(row[yIdx].value);
            if (xVal != null && yVal != null) {
              spots.add(FlSpot(xVal, yVal));
            }
          }
        }
      }
    } else if (testName == "carrieracquisition") {
      int xIdx = header.indexOf("Offset Frequency (kHz)");
      int yIdx = header.indexOf("AGC");
      if (xIdx != -1 && yIdx != -1) {
        for (var row in rows) {
          if (xIdx < row.length && yIdx < row.length) {
            double? xVal = double.tryParse(row[xIdx].value);
            double? yVal = double.tryParse(row[yIdx].value);
            if (xVal != null && yVal != null) {
              spots.add(FlSpot(xVal, yVal));
            }
          }
        }
      }
    } else if (testName == "modindex") {
      int xIdx = header.indexOf("SubCarrier Frequency [kHz]");
      int yIdx = header.indexOf("Measured Mod Index [rad]");
      if (xIdx != -1 && yIdx != -1) {
        for (var row in rows) {
          if (xIdx < row.length && yIdx < row.length) {
            double? xVal = double.tryParse(row[xIdx].value);
            double? yVal = double.tryParse(row[yIdx].value);
            if (xVal != null && yVal != null) {
              spots.add(FlSpot(xVal, yVal));
            }
          }
        }
      }
    } else if (testName == "frequencydeviationmeasurement") {
      if (rows.isNotEmpty) {
        final row = rows.first;
        int d1Idx = header.indexOf("Deviation for Frequency1 [kHz]");
        int d2Idx = header.indexOf("Deviation for Frequency2 [kHz]");
        if (d1Idx != -1 && d2Idx != -1 && d1Idx < row.length && d2Idx < row.length) {
          double? d1 = double.tryParse(row[d1Idx].value);
          double? d2 = double.tryParse(row[d2Idx].value);
          if (d1 != null) spots.add(FlSpot(1.0, d1));
          if (d2 != null) spots.add(FlSpot(2.0, d2));
        }
      }
    } else {
      for (int i = 0; i < rows.length; i++) {
        final row = rows[i];
        double? yVal;
        for (var cell in row.skip(1)) {
          yVal = double.tryParse(cell.value);
          if (yVal != null) break;
        }
        spots.add(FlSpot(i.toDouble(), yVal ?? 0.0));
      }
    }

    spots.sort((a, b) => a.x.compareTo(b.x));
    return spots;
  }

  List<FlSpot> _extractVariableResultsSpots(TestReport report) {
    List<FlSpot> spots = [];
    final testName = report.testType.toLowerCase();

    List<TestReportTable> tables = [];
    report.results.forEach((key, value) {
      if (key.toLowerCase().contains("result") ||
          key.toLowerCase().contains("harmonic") ||
          key.toLowerCase().contains("spur")) {
        tables.add(value);
      }
    });

    if (tables.isEmpty && report.results.isNotEmpty) {
      tables.add(report.results.values.first);
    }

    for (var table in tables) {
      final header = table.header;
      final rows = table.data;

      if (testName == "spurious") {
        int fIdx = header.indexOf("Frequency [kHz]");
        int lIdx = header.indexOf("Spurious Level [dBc]");
        if (fIdx != -1 && lIdx != -1) {
          for (var row in rows) {
            if (fIdx < row.length && lIdx < row.length) {
              double? freq = double.tryParse(row[fIdx].value);
              double? level = double.tryParse(row[lIdx].value);
              if (freq != null && level != null) {
                spots.add(FlSpot(freq / 1000000.0, level));
              }
            }
          }
        }
      } else if (testName == "harmonics") {
        int fIdx = header.indexOf("Frequency [MHz]");
        if (fIdx == -1) fIdx = header.indexOf("Measured Frequency [MHz]");
        int lIdx = header.indexOf("Level [dBc]");
        if (fIdx != -1 && lIdx != -1) {
          for (var row in rows) {
            if (fIdx < row.length && lIdx < row.length) {
              double? freq = double.tryParse(row[fIdx].value);
              double? level = double.tryParse(row[lIdx].value);
              if (freq != null && level != null) {
                spots.add(FlSpot(freq / 1000.0, level));
              }
            }
          }
        }
      } else {
        for (var row in rows) {
          if (row.length >= 2) {
            double? xVal = double.tryParse(row[0].value);
            double? yVal = double.tryParse(row[1].value);
            if (xVal != null && yVal != null) {
              spots.add(FlSpot(xVal, yVal));
            }
          }
        }
      }
    }

    spots.sort((a, b) => a.x.compareTo(b.x));
    return spots;
  }

  void _validateSelectionsForConfig(String? config) {
    if (config == null) return;
    if (_selectedTestName != null) {
      final hasMatch = _allReports.any((r) =>
          r.config == config &&
          r.testType == _selectedTestName &&
          (_selectedTestCategory == null || r.testCategory == _selectedTestCategory));
      if (!hasMatch) {
        final hasTestMatch = _allReports.any((r) => r.config == config && r.testType == _selectedTestName);
        if (hasTestMatch) {
          _selectedTestCategory = null;
        } else {
          _selectedTestName = null;
          if (_selectedTestCategory != null) {
            final hasCatMatch = _allReports.any((r) => r.config == config && r.testCategory == _selectedTestCategory);
            if (!hasCatMatch) {
              _selectedTestCategory = null;
            }
          }
        }
      }
    } else if (_selectedTestCategory != null) {
      final hasCatMatch = _allReports.any((r) => r.config == config && r.testCategory == _selectedTestCategory);
      if (!hasCatMatch) {
        _selectedTestCategory = null;
      }
    }
  }

  void _validateSelectionsForTestName(String? testName) {
    if (testName == null) return;
    if (_selectedConfig != null) {
      final hasMatch = _allReports.any((r) =>
          r.testType == testName &&
          r.config == _selectedConfig &&
          (_selectedTestCategory == null || r.testCategory == _selectedTestCategory));
      if (!hasMatch) {
        final hasConfigMatch = _allReports.any((r) => r.testType == testName && r.config == _selectedConfig);
        if (hasConfigMatch) {
          _selectedTestCategory = null;
        } else {
          _selectedConfig = null;
          if (_selectedTestCategory != null) {
            final hasCatMatch = _allReports.any((r) => r.testType == testName && r.testCategory == _selectedTestCategory);
            if (!hasCatMatch) {
              _selectedTestCategory = null;
            }
          }
        }
      }
    } else if (_selectedTestCategory != null) {
      final hasCatMatch = _allReports.any((r) => r.testType == testName && r.testCategory == _selectedTestCategory);
      if (!hasCatMatch) {
        _selectedTestCategory = null;
      }
    }
  }

  void _validateSelectionsForTestCategory(String? testCategory) {
    if (testCategory == null) return;
    if (_selectedConfig != null) {
      final hasMatch = _allReports.any((r) =>
          r.testCategory == testCategory &&
          r.config == _selectedConfig &&
          (_selectedTestName == null || r.testType == _selectedTestName));
      if (!hasMatch) {
        final hasConfigMatch = _allReports.any((r) => r.testCategory == testCategory && r.config == _selectedConfig);
        if (hasConfigMatch) {
          _selectedTestName = null;
        } else {
          _selectedConfig = null;
          if (_selectedTestName != null) {
            final hasTestMatch = _allReports.any((r) => r.testCategory == testCategory && r.testType == _selectedTestName);
            if (!hasTestMatch) {
              _selectedTestName = null;
            }
          }
        }
      }
    } else if (_selectedTestName != null) {
      final hasTestMatch = _allReports.any((r) => r.testCategory == testCategory && r.testType == _selectedTestName);
      if (!hasTestMatch) {
        _selectedTestName = null;
      }
    }
  }

  double _getSessionValue(String sessionId, {double base = 10.0, double variance = 0.5}) {
    // Generate a repeatable pseudo-random value based on the session ID string
    int hash = sessionId.hashCode;
    final rand = math.Random(hash);
    return base + (rand.nextDouble() * variance * 2) - variance;
  }

  List<String> _getUniqueConfigs() {
    return _allReports
        .where((r) {
          final matchTest = _selectedTestName == null || r.testType == _selectedTestName;
          final matchCat = _selectedTestCategory == null || r.testCategory == _selectedTestCategory;
          return matchTest && matchCat;
        })
        .map((e) => e.config)
        .toSet()
        .toList()
      ..sort();
  }

  List<String> _getUniqueTestNames() {
    return _allReports
        .where((r) {
          final matchConfig = _selectedConfig == null || r.config == _selectedConfig;
          final matchCat = _selectedTestCategory == null || r.testCategory == _selectedTestCategory;
          return matchConfig && matchCat;
        })
        .map((e) => e.testType)
        .toSet()
        .toList()
      ..sort();
  }

  List<String> _getUniqueTestCategories() {
    return _allReports
        .where((r) {
          final matchConfig = _selectedConfig == null || r.config == _selectedConfig;
          final matchTest = _selectedTestName == null || r.testType == _selectedTestName;
          return matchConfig && matchTest;
        })
        .map((e) => e.testCategory)
        .toSet()
        .toList()
      ..sort();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      body: Stack(
        children: [
          Column(
            children: [
              ScreenHeader(
                title: 'Insights',
                subtitle:
                    'Comparative analysis and historical trend visualization',
                icon: Icons.lightbulb_rounded,
                trailing: _buildHelpTrigger(theme),
              ),
              Expanded(
                child: Row(
                  children: [
                    _buildSidebar(theme),
                    Expanded(
                      child: Column(
                        children: [
                          Expanded(
                            child: ContentCard(
                              margin: const EdgeInsets.all(16),
                              padding: const EdgeInsets.all(24),
                              child: _buildMainVisualization(theme),
                            ),
                          ),
                          _buildInsightBar(theme),
                        ],
                      ),
                    ),
                    _buildStatsRail(theme),
                  ],
                ),
              ),
            ],
          ),
          if (_isHelpOpen)
            GestureDetector(
              onTap: () => setState(() => _isHelpOpen = false),
              child: Container(color: Colors.black.withValues(alpha: 0.1)),
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

  Widget _buildSidebar(ThemeData theme) {
    return ContentCard(
      width: 320,
      isSidebar: true,
      margin: const EdgeInsets.fromLTRB(16, 16, 0, 16),
      padding: EdgeInsets.zero,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(20.0),
            child: Text(
              'Analysis Scope',
              style: GoogleFonts.outfit(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: theme.colorScheme.primary,
              ),
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                _buildSectionTitle('CONFIGURATION'),
                _buildRealDropdown(
                  _selectedConfig ?? "Select Config",
                  _getUniqueConfigs(),
                  (val) async {
                    setState(() {
                      _selectedConfig = val;
                      _validateSelectionsForConfig(val);
                      _applyFilters();
                    });
                    await _fetchReportsData();
                  },
                ),
                const SizedBox(height: 24),
                _buildSectionTitle('TEST NAME'),
                _buildRealDropdown(
                  _selectedTestName ?? "Select Test",
                  _getUniqueTestNames(),
                  (val) async {
                    setState(() {
                      _selectedTestName = val;
                      _validateSelectionsForTestName(val);
                      _applyFilters();
                    });
                    await _fetchReportsData();
                  },
                ),
                const SizedBox(height: 24),
                _buildSectionTitle('TEST CATEGORY'),
                _buildRealDropdown(
                  _selectedTestCategory ?? "Select Category",
                  _getUniqueTestCategories(),
                  (val) async {
                    setState(() {
                      _selectedTestCategory = val;
                      _validateSelectionsForTestCategory(val);
                      _applyFilters();
                    });
                    await _fetchReportsData();
                  },
                ),
                const SizedBox(height: 24),
                _buildSectionTitle('REFERENCE MODE'),
                const SizedBox(height: 8),
                Row(
                  children: [
                    _modeToggle("Golden", !_useMeanAsReference),
                    const SizedBox(width: 8),
                    _modeToggle("Mean", _useMeanAsReference),
                  ],
                ),
                const SizedBox(height: 24),
                _buildSectionTitle('QUICK SELECT'),
                const SizedBox(height: 8),
                Row(
                  children: [
                    _quickSelectButton("Last 5", 5),
                    const SizedBox(width: 8),
                    _quickSelectButton("Last 10", 10),
                  ],
                ),
                const SizedBox(height: 24),
                _buildSectionTitle('SESSIONS TO COMPARE'),
                const SizedBox(height: 8),
                if (_isLoading)
                  const Center(child: CircularProgressIndicator())
                else if (_filteredReports.isEmpty)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 20),
                    child: Center(
                      child: Text(
                        "No sessions found",
                        style: TextStyle(color: Colors.grey.shade500),
                      ),
                    ),
                  )
                else
                  ..._filteredReports.map((report) {
                    final sessionId = "${report.date}_${report.time}";
                    final isSelected = _selectedSessions.contains(sessionId);
                    final isRef = _referenceSession == sessionId;

                    return Container(
                      margin: const EdgeInsets.only(bottom: 4),
                      decoration: BoxDecoration(
                        color: isRef
                            ? theme.colorScheme.primary.withValues(alpha: 0.05)
                            : Colors.transparent,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: CheckboxListTile(
                        title: Text(
                          '${report.testType}${report.testCategory.isNotEmpty ? ' - ${report.testCategory}' : ''} - ${report.config}',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: isRef
                                ? FontWeight.bold
                                : FontWeight.normal,
                          ),
                        ),
                        subtitle: Text(
                          "${report.date} ${report.time} ${isRef ? '(Ref)' : ''}",
                          style: const TextStyle(fontSize: 11),
                        ),
                        value: isSelected,
                        activeColor: theme.colorScheme.primary,
                        dense: true,
                        controlAffinity: ListTileControlAffinity.leading,
                        contentPadding: const EdgeInsets.only(left: 8, right: 0),
                        onChanged: (val) async {
                          setState(() {
                            if (val!) {
                              _selectedSessions.add(sessionId);
                            } else {
                              _selectedSessions.remove(sessionId);
                              if (_referenceSession == sessionId) {
                                _referenceSession = null;
                              }
                            }
                          });
                          await _fetchReportsData();
                        },
                        secondary: IconButton(
                          icon: Icon(
                            isRef ? Icons.push_pin : Icons.push_pin_outlined,
                            size: 18,
                            color: isRef
                                ? theme.colorScheme.primary
                                : Colors.grey.shade400,
                          ),
                          onPressed: isSelected
                              ? () =>
                                  setState(() => _referenceSession = sessionId)
                              : null,
                          tooltip: 'Set as Golden Reference',
                        ),
                      ),
                    );
                  }),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _modeToggle(String label, bool isActive) {
    return Expanded(
      child: InkWell(
        onTap: () => setState(() => _useMeanAsReference = (label == "Mean")),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 8),
          decoration: BoxDecoration(
            color: isActive
                ? Theme.of(context).colorScheme.primary
                : Colors.transparent,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: isActive ? Colors.transparent : Colors.grey.shade300,
            ),
          ),
          alignment: Alignment.center,
          child: Text(
            label,
            style: TextStyle(
              fontSize: 12,
              color: isActive ? Colors.white : Colors.grey.shade600,
              fontWeight: isActive ? FontWeight.bold : FontWeight.normal,
            ),
          ),
        ),
      ),
    );
  }

  Widget _quickSelectButton(String label, int count) {
    return Expanded(
      child: OutlinedButton(
        onPressed: () async {
          setState(() {
            _selectedSessions.clear();
            final countToSelect = math.min(count, _filteredReports.length);
            for (int i = 0; i < countToSelect; i++) {
              final r = _filteredReports[i];
              _selectedSessions.add("${r.date}_${r.time}");
            }
          });
          await _fetchReportsData();
        },
        style: OutlinedButton.styleFrom(
          padding: const EdgeInsets.symmetric(vertical: 12),
          side: BorderSide(
            color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.5),
          ),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
        child: Text(label, style: const TextStyle(fontSize: 12)),
      ),
    );
  }


  Widget _buildMainVisualization(ThemeData theme) {
    if (_isDataLoading) {
      return const Center(
        child: CircularProgressIndicator(),
      );
    }
    switch (_selectedCategory) {
      case InsightDataCategory.singleValue:
        return _buildSingleValueTrend(theme);
      case InsightDataCategory.fixedMultiple:
        return _buildFixedMultipleTrend(theme);
      case InsightDataCategory.variableResults:
        return _buildVariableResultsOverlay(theme);
    }
  }

  Widget _buildSingleValueTrend(ThemeData theme) {
    final useRealData = _loadedReports.isNotEmpty &&
        _loadedReports.any((r) => r.testType == _selectedTestName);

    double? refValue;
    if (_referenceSession == "Session_010") {
      refValue = 10.4;
    } else if (_referenceSession == "Session_009") {
      refValue = 10.5;
    } else if (_referenceSession == "Session_008") {
      refValue = 10.2;
    }

    List<FlSpot> spots = [];
    double? upperLimit;
    double? lowerLimit;
    String unitLabel = "";
    String valueLabel = "Measured";
    List<Map<String, dynamic>> points = [];

    if (useRealData) {
      points = _loadedReports.map((r) => _extractSingleValuePoint(r)).toList();
      points.sort((a, b) {
        int dateCmp = a["date"].compareTo(b["date"]);
        if (dateCmp != 0) return dateCmp;
        return a["time"].compareTo(b["time"]);
      });
      for (int i = 0; i < points.length; i++) {
        final p = points[i];
        spots.add(FlSpot(i.toDouble(), p["value"]));
        if (p["upperSpec"] != null) upperLimit = p["upperSpec"];
        if (p["lowerSpec"] != null) lowerLimit = p["lowerSpec"];
        unitLabel = p["unit"];
        valueLabel = p["label"];
      }
    } else {
      spots = _selectedSessions.asMap().entries.map((entry) {
        final sessionId = entry.value;
        return FlSpot(
          entry.key.toDouble(),
          _getSessionValue(sessionId, base: 10.5, variance: 0.3),
        );
      }).toList();
      unitLabel = "dBm";
      valueLabel = "Power";
      upperLimit = 10.7;
      lowerLimit = 10.3;
    }

    final hasRef = _referenceSession != null;
    double? benchmarkY;
    if (useRealData) {
      if (_referenceSession != null) {
        final refParts = _referenceSession!.split('_');
        if (refParts.length >= 2) {
          final refDate = refParts[0];
          final refTime = refParts[1];
          final refPt = points.firstWhere(
            (p) => p["date"] == refDate && p["time"] == refTime,
            orElse: () => {},
          );
          if (refPt.isNotEmpty) {
            benchmarkY = refPt["value"];
          }
        }
      }
    } else {
      benchmarkY = refValue ?? 10.4;
    }

    double? minY;
    double? maxY;
    if (lowerLimit != null && upperLimit != null) {
      final range = (upperLimit - lowerLimit).abs();
      final pad = range > 0.0001 ? range * 0.10 : upperLimit.abs() * 0.10;
      double calculatedMin = lowerLimit - pad;
      double calculatedMax = upperLimit + pad;

      if (spots.isNotEmpty) {
        final values = spots.map((s) => s.y).toList();
        final minVal = values.reduce(math.min);
        final maxVal = values.reduce(math.max);
        calculatedMin = math.min(calculatedMin, minVal);
        calculatedMax = math.max(calculatedMax, maxVal);
      }

      minY = calculatedMin;
      maxY = calculatedMax;
    }

    final barGroups = List.generate(spots.length, (index) {
      final spot = spots[index];
      final val = spot.y;
      
      String? sessionId;
      if (useRealData) {
        if (index < points.length) {
          sessionId = "${points[index]["date"]}_${points[index]["time"]}";
        }
      } else {
        if (index < _selectedSessions.length) {
          sessionId = _selectedSessions[index];
        }
      }
      final isGolden = _referenceSession == sessionId;
      final isOutOfSpec = (upperLimit != null && val > upperLimit) || (lowerLimit != null && val < lowerLimit);

      Color barColor = theme.colorScheme.primary;
      if (isGolden) {
        barColor = Colors.amber.shade600;
      } else if (isOutOfSpec) {
        barColor = Colors.red.shade600;
      }

      return BarChartGroupData(
        x: index,
        barRods: [
          BarChartRodData(
            fromY: minY ?? 0.0,
            toY: val,
            color: barColor,
            width: 22,
            borderRadius: const BorderRadius.only(
              topLeft: Radius.circular(4),
              topRight: Radius.circular(4),
            ),
          ),
        ],
      );
    });

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildChartHeader(
          "$valueLabel Over Time${unitLabel.isNotEmpty ? ' ($unitLabel)' : ''}",
          useRealData ? "Database-Driven Historical Telemetry" : "Last 10 Measurements (Mock)",
        ),
        const SizedBox(height: 32),
        Expanded(
          child: BarChart(
            BarChartData(
              minY: minY,
              maxY: maxY,
              barTouchData: BarTouchData(
                touchTooltipData: BarTouchTooltipData(
                  getTooltipColor: (group) => theme.colorScheme.surface.withValues(alpha: 0.9),
                  tooltipRoundedRadius: 8,
                  tooltipBorder: BorderSide(
                    color: theme.colorScheme.primary.withValues(alpha: 0.2),
                  ),
                  getTooltipItem: (group, groupIndex, rod, rodIndex) {
                    String sessionId = "";
                    String date = "";
                    String time = "";
                    double yVal = rod.toY;

                    if (useRealData) {
                      int index = group.x.toInt();
                      if (index >= 0 && index < points.length) {
                        date = points[index]["date"];
                        time = points[index]["time"];
                        sessionId = "${date}_$time";
                      }
                    } else {
                      int index = group.x.toInt() % _selectedSessions.length;
                      if (index >= 0 && index < _selectedSessions.length) {
                        sessionId = _selectedSessions[index];
                      }
                      date = "2026-02-0${10 - index}";
                      time = "14:30:${10 + index}";
                    }

                    return BarTooltipItem(
                      '$sessionId\n',
                      GoogleFonts.outfit(
                        color: theme.colorScheme.primary,
                        fontWeight: FontWeight.bold,
                        fontSize: 12,
                      ),
                      children: [
                        TextSpan(
                          text: '$date | $time\n',
                          style: TextStyle(
                            color: Colors.grey.shade600,
                            fontSize: 10,
                            fontWeight: FontWeight.normal,
                          ),
                        ),
                        TextSpan(
                          text: 'Value: ${yVal.toStringAsFixed(3)} $unitLabel',
                          style: GoogleFonts.robotoMono(
                            color: Colors.black87,
                            fontWeight: FontWeight.bold,
                            fontSize: 13,
                          ),
                        ),
                      ],
                    );
                  },
                ),
                handleBuiltInTouches: true,
              ),
              gridData: FlGridData(
                show: true,
                drawVerticalLine: false,
                getDrawingHorizontalLine: (val) =>
                    FlLine(color: Colors.grey.withValues(alpha: 0.1), strokeWidth: 1),
              ),
              titlesData: FlTitlesData(
                show: true,
                rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
                bottomTitles: AxisTitles(
                  axisNameWidget: const Text(
                    "Timestamp",
                    style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
                  ),
                  sideTitles: SideTitles(
                    showTitles: true,
                    reservedSize: 30,
                    getTitlesWidget: (value, meta) {
                      int idx = value.toInt();
                      if (useRealData) {
                        if (idx >= 0 && idx < points.length) {
                          final date = points[idx]["date"];
                          final time = points[idx]["time"];
                          final displayStr = "${date.substring(math.max(0, date.length - 5))} ${time.substring(0, math.min(time.length, 5))}";
                          return SideTitleWidget(
                            meta: meta,
                            child: Text(
                              displayStr,
                              style: const TextStyle(fontSize: 8),
                            ),
                          );
                        }
                      } else {
                        if (idx >= 0 && idx < _selectedSessions.length) {
                          return SideTitleWidget(
                            meta: meta,
                            child: Text(
                              "S$idx",
                              style: const TextStyle(fontSize: 8),
                            ),
                          );
                        }
                      }
                      return const SizedBox.shrink();
                    },
                  ),
                ),
                leftTitles: AxisTitles(
                  axisNameWidget: Text(
                    "$valueLabel${unitLabel.isNotEmpty ? ' ($unitLabel)' : ''}",
                    style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
                  ),
                  sideTitles: const SideTitles(showTitles: true, reservedSize: 45),
                ),
              ),
              borderData: FlBorderData(show: false),
              rangeAnnotations: RangeAnnotations(
                horizontalRangeAnnotations: [
                  if (upperLimit != null && lowerLimit != null)
                    HorizontalRangeAnnotation(
                      y1: lowerLimit,
                      y2: upperLimit,
                      color: Colors.green.withValues(alpha: 0.05),
                    ),
                ],
              ),
              extraLinesData: ExtraLinesData(
                horizontalLines: [
                  if (hasRef && benchmarkY != null)
                    HorizontalLine(
                      y: benchmarkY,
                      color: theme.colorScheme.primary.withValues(alpha: 0.4),
                      strokeWidth: 2,
                      label: HorizontalLineLabel(
                        show: true,
                        alignment: Alignment.topLeft,
                        labelResolver: (line) => "GOLDEN REF",
                        style: GoogleFonts.inter(
                          fontSize: 10,
                          fontWeight: FontWeight.bold,
                          color: theme.colorScheme.primary,
                        ),
                      ),
                    ),
                  if (upperLimit != null)
                    HorizontalLine(
                      y: upperLimit,
                      color: Colors.green.withValues(alpha: 0.2),
                      strokeWidth: 1,
                      dashArray: [5, 5],
                      label: HorizontalLineLabel(
                        show: true,
                        alignment: Alignment.topRight,
                        labelResolver: (line) => "Upper Spec (${upperLimit!.toStringAsFixed(2)})",
                        style: const TextStyle(fontSize: 10, color: Colors.green),
                      ),
                    ),
                  if (lowerLimit != null)
                    HorizontalLine(
                      y: lowerLimit,
                      color: Colors.red.withValues(alpha: 0.2),
                      strokeWidth: 1,
                      dashArray: [5, 5],
                      label: HorizontalLineLabel(
                        show: true,
                        alignment: Alignment.bottomRight,
                        labelResolver: (line) => "Lower Spec (${lowerLimit!.toStringAsFixed(2)})",
                        style: const TextStyle(fontSize: 10, color: Colors.red),
                      ),
                    ),
                ],
              ),
              barGroups: barGroups,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildFixedMultipleTrend(ThemeData theme) {
    final List<LineChartBarData> bars = [];
    final List<Color> palette = [
      Colors.indigo,
      Colors.teal,
      Colors.deepPurple,
      Colors.blueGrey,
      Colors.cyan,
      const Color(0xFF1E88E5),
      const Color(0xFF43A047),
    ];

    final useRealData = _loadedReports.isNotEmpty &&
        _loadedReports.any((r) => r.testType == _selectedTestName);

    String xTitle = "Index";
    String yTitle = "Measured";

    if (_selectedTestName != null) {
      final test = _selectedTestName!.toLowerCase();
      if (test == "lockdynamic" || test == "commanddynamic" || test == "rfuplink") {
        xTitle = "Receiver Power (dBm)";
        yTitle = "AGC";
      } else if (test == "loopstress") {
        xTitle = "Frequency Offset (kHz)";
        yTitle = "Loop Stress TM";
      } else if (test == "carrieracquisition") {
        xTitle = "Offset Frequency (kHz)";
        yTitle = "AGC";
      } else if (test == "modindex") {
        xTitle = "SubCarrier Freq (kHz)";
        yTitle = "Mod Index (rad)";
      } else if (test == "frequencydeviationmeasurement") {
        xTitle = "Frequency Spec Bins";
        yTitle = "Deviation (kHz)";
      } else if (test.contains("pulse") || test.contains("vsa") || test.contains("ppm")) {
        xTitle = "Parameter Index";
        yTitle = "Value";
      }
    }

    final List<Widget> legendItems = [];
    if (useRealData) {
      for (int i = 0; i < _loadedReports.length; i++) {
        final report = _loadedReports[i];
        final sessionId = "${report.date}_${report.time}";
        final isRef = _referenceSession == sessionId;

        final Color sessionColor = isRef
            ? Colors.amber.shade700
            : palette[i % palette.length].withValues(alpha: 0.6);

        legendItems.add(_buildLegendItem(
          isRef ? Colors.amber.shade700 : palette[i % palette.length],
          "${report.date} ${report.time}",
          isRef,
        ));

        final spots = _extractFixedMultipleSpots(report);
        if (spots.isNotEmpty) {
          bars.add(
            LineChartBarData(
              spots: spots,
              color: sessionColor,
              barWidth: isRef ? 5 : 2.5,
              isCurved: spots.length > 2,
              dotData: FlDotData(
                show: true,
                getDotPainter: (spot, percent, barData, index) {
                  return FlDotCirclePainter(
                    radius: isRef ? 5 : 3.5,
                    color: sessionColor,
                    strokeWidth: isRef ? 1.5 : 1,
                    strokeColor: Colors.white,
                  );
                },
              ),
              belowBarData: isRef
                  ? BarAreaData(show: true, color: Colors.amber.withValues(alpha: 0.05))
                  : null,
            ),
          );
        }
      }
    } else {
      for (int i = 0; i < _selectedSessions.length; i++) {
        final sessionId = _selectedSessions[i];
        final isRef = _referenceSession == sessionId;

        final Color sessionColor = isRef
            ? Colors.amber.shade700
            : palette[i % palette.length].withValues(alpha: 0.6);

        final date = "2026-02-0${10 - i}";
        final time = "14:30:${10 + i}";
        legendItems.add(_buildLegendItem(
          isRef ? Colors.amber.shade700 : palette[i % palette.length],
          "$date $time",
          isRef,
        ));

        List<FlSpot> spots = List.generate(10, (idx) {
          return FlSpot(
            idx.toDouble(),
            _getSessionValue(sessionId, base: 15.0, variance: 0.5) + (idx * 0.05),
          );
        });

        bars.add(
          LineChartBarData(
            spots: spots,
            color: sessionColor,
            barWidth: isRef ? 5 : 2.5,
            isCurved: true,
            dotData: FlDotData(
              show: true,
              getDotPainter: (spot, percent, barData, index) {
                return FlDotCirclePainter(
                  radius: isRef ? 5 : 3.5,
                  color: sessionColor,
                  strokeWidth: isRef ? 1.5 : 1,
                  strokeColor: Colors.white,
                );
              },
            ),
            belowBarData: isRef
                ? BarAreaData(show: true, color: Colors.amber.withValues(alpha: 0.05))
                : null,
          ),
        );
      }
      xTitle = "Freq Bin";
      yTitle = "Gain (dB)";
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildChartHeader(
          useRealData ? "Overlay View - $xTitle vs $yTitle" : "Gain Stability Overlay (dB)",
          "Reference Highlights in Gold",
        ),
        const SizedBox(height: 16),
        Wrap(
          spacing: 12,
          runSpacing: 8,
          children: legendItems,
        ),
        const SizedBox(height: 16),
        Expanded(
          child: LineChart(
            LineChartData(
              lineTouchData: LineTouchData(
                touchTooltipData: LineTouchTooltipData(
                  getTooltipColor: (spot) => theme.colorScheme.surface.withValues(alpha: 0.9),
                  tooltipRoundedRadius: 8,
                  tooltipBorder: BorderSide(
                    color: theme.colorScheme.primary.withValues(alpha: 0.2),
                  ),
                  getTooltipItems: (List<LineBarSpot> touchedSpots) {
                    return touchedSpots.map((LineBarSpot touchedSpot) {
                      String sessionId = "";
                      String date = "";
                      String time = "";

                      if (useRealData) {
                        int barIdx = touchedSpot.barIndex;
                        if (barIdx >= 0 && barIdx < _loadedReports.length) {
                          final report = _loadedReports[barIdx];
                          date = report.date;
                          time = report.time;
                          sessionId = "${date}_$time";
                        }
                      } else {
                        int barIdx = touchedSpot.barIndex % _selectedSessions.length;
                        if (barIdx >= 0 && barIdx < _selectedSessions.length) {
                          sessionId = _selectedSessions[barIdx];
                        }
                        date = "2026-02-0${10 - barIdx}";
                        time = "14:30:${10 + touchedSpot.spotIndex}";
                      }

                      return LineTooltipItem(
                        '$sessionId\n',
                        GoogleFonts.outfit(
                          color: theme.colorScheme.primary,
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                        children: [
                          TextSpan(
                            text: '$date | $time\n',
                            style: TextStyle(
                              color: Colors.grey.shade600,
                              fontSize: 10,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                          TextSpan(
                            text: 'X: ${touchedSpot.x.toStringAsFixed(2)}, Y: ${touchedSpot.y.toStringAsFixed(2)}',
                            style: GoogleFonts.robotoMono(
                              color: Colors.black87,
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                            ),
                          ),
                        ],
                      );
                    }).toList();
                  },
                ),
                handleBuiltInTouches: true,
              ),
              gridData: const FlGridData(show: true, drawVerticalLine: false),
              titlesData: _buildTitlesData(
                xTitle: xTitle,
                yTitle: yTitle,
              ),
              borderData: FlBorderData(show: false),
              lineBarsData: bars,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildVariableResultsOverlay(ThemeData theme) {
    final List<LineChartBarData> layers = [];
    final List<Color> palette = [
      Colors.indigo,
      Colors.teal,
      Colors.deepPurple,
      Colors.blueGrey,
      Colors.cyan,
      const Color(0xFF1E88E5),
      const Color(0xFF43A047),
    ];

    final useRealData = _loadedReports.isNotEmpty &&
        _loadedReports.any((r) => r.testType == _selectedTestName);

    final List<Widget> legendItems = [];
    if (useRealData) {
      for (int i = 0; i < _loadedReports.length; i++) {
        final report = _loadedReports[i];
        final sessionId = "${report.date}_${report.time}";
        final isRef = _referenceSession == sessionId;

        final Color sessionColor = isRef
            ? Colors.amber.shade700
            : palette[i % palette.length].withValues(alpha: 0.6);

        legendItems.add(_buildLegendItem(
          isRef ? Colors.amber.shade700 : palette[i % palette.length],
          "${report.date} ${report.time}",
          isRef,
        ));

        final spots = _extractVariableResultsSpots(report);
        if (spots.isNotEmpty) {
          layers.add(
            LineChartBarData(
              spots: spots,
              barWidth: 0,
              dotData: FlDotData(
                show: true,
                getDotPainter: (spot, percent, barData, index) {
                  bool isAlert = spot.y > -60.0;
                  Color dotColor = isRef
                      ? Colors.amber.shade700
                      : (isAlert ? Colors.red : sessionColor);

                  return FlDotCirclePainter(
                    radius: isRef ? 8 : (isAlert ? 7 : 4),
                    color: dotColor,
                    strokeWidth: isRef ? 3 : 0,
                    strokeColor: Colors.white,
                  );
                },
              ),
            ),
          );
        }
      }
    } else {
      for (int i = 0; i < _selectedSessions.length; i++) {
        final sessionId = _selectedSessions[i];
        final isRef = _referenceSession == sessionId;

        final date = "2026-02-0${10 - i}";
        final time = "14:30:${10 + i}";
        legendItems.add(_buildLegendItem(
          isRef ? Colors.amber.shade700 : palette[i % palette.length],
          "$date $time",
          isRef,
        ));

        layers.add(
          LineChartBarData(
            spots: List.generate(5, (idx) {
              double freq = 1.0 + (idx * 0.5);
              return FlSpot(freq, _getSessionValue(sessionId, base: -70, variance: 10));
            }),
            barWidth: 0,
            dotData: FlDotData(
              show: true,
              getDotPainter: (spot, percent, barData, index) {
                bool isAlert = spot.y > -60;
                Color dotColor = isRef
                    ? Colors.amber.shade700
                    : (isAlert ? Colors.red : Colors.indigo.withValues(alpha: 0.4));

                return FlDotCirclePainter(
                  radius: isRef ? 8 : (isAlert ? 7 : 4),
                  color: dotColor,
                  strokeWidth: isRef ? 3 : 0,
                  strokeColor: Colors.white,
                );
              },
            ),
          ),
        );
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildChartHeader(
          "Spurious / Harmonics Emissions Overlay",
          "Spectral variations highlighting Reference",
        ),
        const SizedBox(height: 16),
        Wrap(
          spacing: 12,
          runSpacing: 8,
          children: legendItems,
        ),
        const SizedBox(height: 16),
        Expanded(
          child: LineChart(
            LineChartData(
              lineTouchData: LineTouchData(
                touchTooltipData: LineTouchTooltipData(
                  getTooltipColor: (spot) => theme.colorScheme.surface.withValues(alpha: 0.9),
                  tooltipRoundedRadius: 8,
                  tooltipBorder: BorderSide(
                    color: theme.colorScheme.primary.withValues(alpha: 0.2),
                  ),
                  getTooltipItems: (List<LineBarSpot> touchedSpots) {
                    return touchedSpots.map((LineBarSpot touchedSpot) {
                      String sessionId = "";
                      String date = "";
                      String time = "";

                      if (useRealData) {
                        int barIdx = touchedSpot.barIndex;
                        if (barIdx >= 0 && barIdx < _loadedReports.length) {
                          final report = _loadedReports[barIdx];
                          date = report.date;
                          time = report.time;
                          sessionId = "${date}_$time";
                        }
                      } else {
                        int barIdx = touchedSpot.barIndex % _selectedSessions.length;
                        if (barIdx >= 0 && barIdx < _selectedSessions.length) {
                          sessionId = _selectedSessions[barIdx];
                        }
                        date = "2026-02-0${10 - barIdx}";
                        time = "14:30:${10 + touchedSpot.spotIndex}";
                      }

                      return LineTooltipItem(
                        '$sessionId\n',
                        GoogleFonts.outfit(
                          color: theme.colorScheme.primary,
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                        children: [
                          TextSpan(
                            text: '$date | $time\n',
                            style: TextStyle(
                              color: Colors.grey.shade600,
                              fontSize: 10,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                          TextSpan(
                            text: 'Freq: ${touchedSpot.x.toStringAsFixed(4)} GHz, Level: ${touchedSpot.y.toStringAsFixed(2)} dBc',
                            style: GoogleFonts.robotoMono(
                              color: Colors.black87,
                              fontWeight: FontWeight.bold,
                              fontSize: 13,
                            ),
                          ),
                        ],
                      );
                    }).toList();
                  },
                ),
                handleBuiltInTouches: true,
              ),
              lineBarsData: layers,
              titlesData: _buildTitlesData(
                xTitle: "Freq (GHz)",
                yTitle: "Level (dBc)",
              ),
              borderData: FlBorderData(show: false),
              gridData: const FlGridData(show: true),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildLegendItem(Color color, String label, bool isRef) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.05),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: color.withValues(alpha: 0.2)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(
              color: color,
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 8),
          Text(
            label + (isRef ? " [REF]" : ""),
            style: GoogleFonts.inter(
              fontSize: 11,
              fontWeight: isRef ? FontWeight.bold : FontWeight.normal,
              color: isRef ? Colors.amber.shade900 : Colors.black87,
            ),
          ),
        ],
      ),
    );
  }

  FlTitlesData _buildTitlesData({
    String xTitle = "Session",
    String yTitle = "Value",
  }) {
    return FlTitlesData(
      show: true,
      rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
      topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
      bottomTitles: AxisTitles(
        axisNameWidget: Text(
          xTitle,
          style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
        ),
        sideTitles: const SideTitles(showTitles: true, reservedSize: 30),
      ),
      leftTitles: AxisTitles(
        axisNameWidget: Text(
          yTitle,
          style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold),
        ),
        sideTitles: const SideTitles(showTitles: true, reservedSize: 45),
      ),
    );
  }

  Widget _buildChartHeader(String title, String subtitle) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: GoogleFonts.outfit(
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              subtitle,
              style: TextStyle(color: Colors.grey.shade500, fontSize: 13),
            ),
          ],
        ),
        Row(
          children: [
            OutlinedButton.icon(
              onPressed: () {},
              icon: const Icon(Icons.download, size: 16),
              label: const Text("Export CSV"),
            ),
            const SizedBox(width: 12),
            ElevatedButton.icon(
              onPressed: () {},
              icon: const Icon(Icons.picture_as_pdf, size: 16),
              label: const Text("Report"),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildInsightBar(ThemeData theme) {
    final useRealData = _loadedReports.isNotEmpty &&
        _loadedReports.any((r) => r.testType == _selectedTestName);

    String message = "All parameters are within nominal historical limits.";
    Color color = Colors.green;
    IconData icon = Icons.check_circle_outline;

    final uiSessionCount = _selectedSessions.length;
    final testLabel = _selectedTestName ?? "Parameters";

    if (useRealData) {
      double? upperLimit;
      double? lowerLimit;
      List<double> values = [];

      if (_selectedCategory == InsightDataCategory.singleValue) {
        List<Map<String, dynamic>> points = _loadedReports.map((r) => _extractSingleValuePoint(r)).toList();
        values = points.map((p) => p["value"] as double).toList();
        if (points.isNotEmpty) {
          upperLimit = points.first["upperSpec"];
          lowerLimit = points.first["lowerSpec"];
        }
      }

      bool hasFailure = false;
      int outOfSpecCount = 0;
      if (_selectedCategory == InsightDataCategory.singleValue && upperLimit != null && lowerLimit != null) {
        for (var val in values) {
          if (val > upperLimit || val < lowerLimit) {
            hasFailure = true;
            outOfSpecCount++;
          }
        }
      }

      if (hasFailure) {
        message = "Analysis: $testLabel shows $outOfSpecCount out-of-spec measurement(s) across $uiSessionCount sessions.";
        color = Colors.red;
        icon = Icons.error_outline;
      } else {
        message = "Analysis: $testLabel shows nominal metrics across $uiSessionCount sessions.";
        color = Colors.green;
        icon = Icons.check_circle_outline;
      }
    } else {
      if (_selectedCategory == InsightDataCategory.singleValue) {
        message = "Analysis: $testLabel shows consistent metrics across $uiSessionCount sessions.";
        color = Colors.green;
        icon = Icons.check_circle_outline;
      } else if (_selectedCategory == InsightDataCategory.variableResults) {
        message = "Overlay Analysis: 1 potential spur detected in the most recent session of $testLabel.";
        color = Colors.orange;
        icon = Icons.warning_amber_rounded;
      }
    }

    return ContentCard(
      margin: const EdgeInsets.fromLTRB(16, 0, 16, 16),
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
      child: Row(
        children: [
          Icon(icon, color: color),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  "DETERMINISTIC INSIGHT",
                  style: GoogleFonts.inter(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: color,
                    letterSpacing: 1.1,
                  ),
                ),
                Text(
                  message,
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
          TextButton(
            onPressed: () => _showBreakdownDialog(context),
            child: const Text("View Breakdown"),
          ),
        ],
      ),
    );
  }

  Widget _buildStatsRail(ThemeData theme) {
    final modeLabel = _useMeanAsReference ? "MEAN" : "GOLDEN";
    final refName = _useMeanAsReference
        ? "Average (Historical)"
        : (_referenceSession ?? "None");

    final useRealData = _loadedReports.isNotEmpty &&
        _loadedReports.any((r) => r.testType == _selectedTestName);

    List<double> values = [];
    String unitLabel = "";
    String valueLabel = "Measured";

    if (useRealData) {
      if (_selectedCategory == InsightDataCategory.singleValue) {
        List<Map<String, dynamic>> points = _loadedReports.map((r) => _extractSingleValuePoint(r)).toList();
        points.sort((a, b) {
          int dateCmp = a["date"].compareTo(b["date"]);
          if (dateCmp != 0) return dateCmp;
          return a["time"].compareTo(b["time"]);
        });
        values = points.map((p) => p["value"] as double).toList();
        if (points.isNotEmpty) {
          unitLabel = points.first["unit"];
          valueLabel = points.first["label"];
        }
      } else {
        for (var report in _loadedReports) {
          final spots = _selectedCategory == InsightDataCategory.fixedMultiple
              ? _extractFixedMultipleSpots(report)
              : _extractVariableResultsSpots(report);
          if (spots.isNotEmpty) {
            final avgY = spots.map((s) => s.y).reduce((a, b) => a + b) / spots.length;
            values.add(avgY);
          }
        }
        unitLabel = "dB";
        valueLabel = "Avg Level";
      }
    } else {
      values = _selectedSessions.map((s) => _getSessionValue(s, base: 10.5, variance: 0.3)).toList();
      unitLabel = "dBm";
      valueLabel = "Power";
    }

    final mean = values.isEmpty ? 0.0 : values.reduce((a, b) => a + b) / values.length;
    final currentVal = values.isNotEmpty ? values.last : 0.0;
    
    double refVal = mean;
    if (!_useMeanAsReference && _referenceSession != null) {
      final refParts = _referenceSession!.split('_');
      if (refParts.length >= 2) {
        final refDate = refParts[0];
        final refTime = refParts[1];
        if (useRealData) {
          final refReport = _loadedReports.firstWhere(
            (r) => r.date == refDate && r.time == refTime,
            orElse: () => TestReport(
              spacecraft: '', testType: '', testCategory: '', config: '',
              date: '', time: '', testPhase: '', results: {}, order: [], remarks: ''
            ),
          );
          if (refReport.date.isNotEmpty) {
            if (_selectedCategory == InsightDataCategory.singleValue) {
              refVal = _extractSingleValuePoint(refReport)["value"];
            } else {
              final spots = _selectedCategory == InsightDataCategory.fixedMultiple
                  ? _extractFixedMultipleSpots(refReport)
                  : _extractVariableResultsSpots(refReport);
              if (spots.isNotEmpty) {
                refVal = spots.map((s) => s.y).reduce((a, b) => a + b) / spots.length;
              }
            }
          }
        } else {
          refVal = _getSessionValue(_referenceSession!, base: 10.5, variance: 0.3);
        }
      }
    }
    
    final delta = currentVal - refVal;
    final maxDrift = values.isEmpty ? 0.0 : values.map((v) => (v - refVal).abs()).reduce(math.max);
    
    double sumSq = 0;
    for (var v in values) {
      sumSq += math.pow(v - mean, 2);
    }
    final stdDev = values.isEmpty ? 0.0 : math.sqrt(sumSq / values.length);

    return ContentCard(
      width: 240,
      margin: const EdgeInsets.fromLTRB(0, 16, 16, 16),
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            "STATISTICS",
            style: GoogleFonts.inter(
              fontSize: 11,
              fontWeight: FontWeight.bold,
              color: Colors.grey,
            ),
          ),
          const SizedBox(height: 20),
          _statItem("Reference ($modeLabel)", refName, isRef: true),
          const Divider(height: 32),
          _statItem(valueLabel, "${currentVal.toStringAsFixed(2)} $unitLabel"),
          _statItem(
            "Delta from $modeLabel",
            "${delta > 0 ? '+' : ''}${delta.toStringAsFixed(2)} $unitLabel",
            valueColor: delta.abs() > 0.2 ? Colors.orange : Colors.green,
          ),
          const Spacer(),
          _statItem("Max Drift", "${maxDrift.toStringAsFixed(2)} $unitLabel"),
          _statItem("Std Dev (σ)", "${stdDev.toStringAsFixed(3)} $unitLabel"),
        ],
      ),
    );
  }

  Widget _statItem(
    String label,
    String value, {
    bool isRef = false,
    Color? valueColor,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 12, color: Colors.grey)),
          const SizedBox(height: 4),
          Text(
            value,
            style: GoogleFonts.robotoMono(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color:
                  valueColor ??
                  (isRef
                      ? Theme.of(context).colorScheme.primary
                      : Colors.black87),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0, top: 4.0),
      child: Text(
        title,
        style: GoogleFonts.inter(
          fontSize: 11,
          fontWeight: FontWeight.w900,
          color: Colors.grey.shade500,
          letterSpacing: 1.2,
        ),
      ),
    );
  }

  Widget _buildRealDropdown(
    String label,
    List<String> items,
    Function(String?) onChanged,
  ) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          value: items.contains(label) ? label : null,
          hint: Text(label, style: const TextStyle(fontSize: 13)),
          isExpanded: true,
          icon: const Icon(Icons.keyboard_arrow_down_rounded, size: 20),
          items: [
            const DropdownMenuItem<String>(
              value: null,
              child: Text("All", style: TextStyle(fontSize: 13)),
            ),
            ...items.map((String value) {
              return DropdownMenuItem<String>(
                value: value,
                child: Text(value, style: const TextStyle(fontSize: 13)),
              );
            }),
          ],
          onChanged: onChanged,
        ),
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
                : theme.colorScheme.primary.withValues(alpha: 0.2),
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
            color: Colors.black.withValues(alpha: 0.1),
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
                  'Insights Help',
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
                  'Golden Reference',
                  'Pin any session as a "Golden" baseline. All other sessions will be compared against this '
                      'reference to calculate deltas in power, noise floor, or gain stability.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Trend Categories',
                  '• **Single Value**: Tracks Scalar data over historical sessions.\n'
                      '• **Fixed Multiple**: Overlays gain/loss vectors across multiple sessions.\n'
                      '• **Variable Results**: Spectral overlay of spurs and harmonics.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Deterministic Insights',
                  'The system automatically compares the current session with historical means. It flags "Trend Drifts" '
                      'and "New Artifacts" (like spurs) that deviate significantly from baseline.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Statistical Rail',
                  'Provides real-time σ (Standard Deviation) and Max Drift metrics based on the currently '
                      'selected Analysis Scope and Reference Mode.',
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

  void _showBreakdownDialog(BuildContext context) {
    final theme = Theme.of(context);
    final useRealData = _loadedReports.isNotEmpty &&
        _loadedReports.any((r) => r.testType == _selectedTestName);

    showDialog(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          title: Text(
            "Session Telemetry Breakdown",
            style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
          ),
          content: SizedBox(
            width: 600,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (!useRealData)
                    Text(
                      "Displaying historical mock dataset comparison. Select a real telemetry configuration to display live database values.",
                      style: TextStyle(color: Colors.grey.shade600, fontSize: 13),
                    )
                  else
                    ..._loadedReports.map((report) {
                      final title = "${report.date} ${report.time}";
                      final isRef = "${report.date}_${report.time}" == _referenceSession;
                      
                      String valueText = "N/A";
                      Color statusColor = Colors.green;
                      IconData statusIcon = Icons.check_circle_outline;

                      if (_selectedCategory == InsightDataCategory.singleValue) {
                        final pt = _extractSingleValuePoint(report);
                        final val = pt["value"] as double;
                        final unit = pt["unit"] as String;
                        final upper = pt["upperSpec"] as double?;
                        final lower = pt["lowerSpec"] as double?;
                        valueText = "${val.toStringAsFixed(3)} $unit";
                        if ((upper != null && val > upper) || (lower != null && val < lower)) {
                          statusColor = Colors.red;
                          statusIcon = Icons.error_outline;
                        }
                      } else {
                        final spots = _selectedCategory == InsightDataCategory.fixedMultiple
                            ? _extractFixedMultipleSpots(report)
                            : _extractVariableResultsSpots(report);
                        if (spots.isNotEmpty) {
                          final avgY = spots.map((s) => s.y).reduce((a, b) => a + b) / spots.length;
                          valueText = "Avg: ${avgY.toStringAsFixed(2)} dB";
                        }
                      }

                      return Container(
                        margin: const EdgeInsets.symmetric(vertical: 4),
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: isRef
                              ? theme.colorScheme.primary.withValues(alpha: 0.05)
                              : Colors.grey.shade50,
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: isRef
                                ? theme.colorScheme.primary.withValues(alpha: 0.2)
                                : Colors.grey.shade200,
                          ),
                        ),
                        child: Row(
                          children: [
                            Icon(statusIcon, color: statusColor, size: 20),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    "${report.testType} (${report.config}) ${isRef ? '[GOLDEN REFERENCE]' : ''}",
                                    style: TextStyle(
                                      fontWeight: isRef ? FontWeight.bold : FontWeight.normal,
                                      fontSize: 13,
                                    ),
                                  ),
                                  Text(
                                    title,
                                    style: const TextStyle(fontSize: 11, color: Colors.grey),
                                  ),
                                ],
                              ),
                            ),
                            Text(
                              valueText,
                              style: GoogleFonts.robotoMono(
                                fontWeight: FontWeight.bold,
                                fontSize: 13,
                              ),
                            ),
                          ],
                        ),
                      );
                    }),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(),
              child: const Text("Close"),
            ),
          ],
        );
      },
    );
  }
}
