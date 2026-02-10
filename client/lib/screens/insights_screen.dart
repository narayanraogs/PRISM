import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:fl_chart/fl_chart.dart';
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
  bool _showComparison = false;
  final List<String> _selectedSessions = ["Session_001", "Session_002"];
  String? _referenceSession = "Session_001";
  bool _useMeanAsReference = false;
  bool _isHelpOpen = false;

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
                _buildMockDropdown("Uplink_Profile_A"),
                const SizedBox(height: 24),
                _buildSectionTitle('TEST NAME'),
                _buildTestSelector(theme),
                const SizedBox(height: 24),
                _buildSectionTitle('TEST CATEGORY'),
                _buildMockDropdown("Performance"),
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
                ...List.generate(10, (index) {
                  final sessionId = "Session_0${10 - index}";
                  final isSelected = _selectedSessions.contains(sessionId);
                  final isRef = _referenceSession == sessionId;

                  return Container(
                    margin: const EdgeInsets.only(bottom: 4),
                    decoration: BoxDecoration(
                      color: isRef
                          ? theme.colorScheme.primary.withOpacity(0.05)
                          : Colors.transparent,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: CheckboxListTile(
                      title: Text(
                        sessionId,
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: isRef
                              ? FontWeight.bold
                              : FontWeight.normal,
                        ),
                      ),
                      subtitle: Text(
                        "2026-02-0${10 - index} ${isRef ? '(Reference)' : ''}",
                        style: const TextStyle(fontSize: 11),
                      ),
                      value: isSelected,
                      activeColor: theme.colorScheme.primary,
                      dense: true,
                      controlAffinity: ListTileControlAffinity.leading,
                      contentPadding: const EdgeInsets.only(left: 8, right: 0),
                      onChanged: (val) {
                        setState(() {
                          if (val!) {
                            _selectedSessions.add(sessionId);
                          } else {
                            _selectedSessions.remove(sessionId);
                            if (_referenceSession == sessionId)
                              _referenceSession = null;
                          }
                        });
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
        onPressed: () {
          setState(() {
            _selectedSessions.clear();
            for (int i = 0; i < count; i++) {
              _selectedSessions.add("Session_0${10 - i}");
            }
          });
        },
        style: OutlinedButton.styleFrom(
          padding: const EdgeInsets.symmetric(vertical: 12),
          side: BorderSide(
            color: Theme.of(context).colorScheme.primary.withOpacity(0.5),
          ),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
        child: Text(label, style: const TextStyle(fontSize: 12)),
      ),
    );
  }

  Widget _buildTestSelector(ThemeData theme) {
    final tests = [
      {"name": "Transmit Power", "cat": InsightDataCategory.singleValue},
      {"name": "Dynamic Range", "cat": InsightDataCategory.fixedMultiple},
      {
        "name": "Spurious Emissions",
        "cat": InsightDataCategory.variableResults,
      },
    ];

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<Map<String, dynamic>>(
          value: tests.firstWhere((t) => t['cat'] == _selectedCategory),
          isExpanded: true,
          icon: const Icon(Icons.keyboard_arrow_down_rounded, size: 20),
          items: tests.map((t) {
            return DropdownMenuItem(
              value: t,
              child: Text(
                t['name'] as String,
                style: const TextStyle(fontSize: 13),
              ),
            );
          }).toList(),
          onChanged: (val) {
            if (val != null) {
              setState(() {
                _selectedCategory = val['cat'] as InsightDataCategory;
              });
            }
          },
        ),
      ),
    );
  }

  Widget _buildMainVisualization(ThemeData theme) {
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
    // Mock logic: Session_001 is index 0 in our mock data
    double? refValue;
    if (_referenceSession == "Session_010")
      refValue = 10.4;
    else if (_referenceSession == "Session_009")
      refValue = 10.5;
    else if (_referenceSession == "Session_008")
      refValue = 10.2;
    // ... and so on. For now, let's assume 10.4 is our Golden baseline if something is pinned.
    final hasRef = _referenceSession != null;
    final benchmarkY = refValue ?? 10.4;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildChartHeader("Power Over Time (dBm)", "Last 10 Measurements"),
        const SizedBox(height: 32),
        Expanded(
          child: LineChart(
            LineChartData(
              lineTouchData: _buildTouchData(theme),
              gridData: FlGridData(
                show: true,
                drawVerticalLine: false,
                getDrawingHorizontalLine: (val) =>
                    FlLine(color: Colors.grey.withOpacity(0.1), strokeWidth: 1),
              ),
              titlesData: _buildTitlesData(),
              borderData: FlBorderData(show: false),
              rangeAnnotations: RangeAnnotations(
                horizontalRangeAnnotations: [
                  HorizontalRangeAnnotation(
                    y1: 10.3,
                    y2: 10.7,
                    color: Colors.green.withOpacity(0.05),
                  ),
                ],
              ),
              extraLinesData: ExtraLinesData(
                horizontalLines: [
                  if (hasRef && !_useMeanAsReference)
                    HorizontalLine(
                      y: benchmarkY,
                      color: theme.colorScheme.primary.withOpacity(0.4),
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
                  HorizontalLine(
                    y: 10.7,
                    color: Colors.green.withOpacity(0.2),
                    strokeWidth: 1,
                    dashArray: [5, 5],
                    label: HorizontalLineLabel(
                      show: true,
                      alignment: Alignment.topRight,
                      labelResolver: (line) => "Upper Spec (10.7)",
                      style: const TextStyle(fontSize: 10, color: Colors.green),
                    ),
                  ),
                  HorizontalLine(
                    y: 10.3,
                    color: Colors.red.withOpacity(0.2),
                    strokeWidth: 1,
                    dashArray: [5, 5],
                    label: HorizontalLineLabel(
                      show: true,
                      alignment: Alignment.bottomRight,
                      labelResolver: (line) => "Lower Spec (10.3)",
                      style: const TextStyle(fontSize: 10, color: Colors.red),
                    ),
                  ),
                ],
              ),
              lineBarsData: [
                LineChartBarData(
                  spots: [
                    const FlSpot(0, 10.4),
                    const FlSpot(1, 10.5),
                    const FlSpot(2, 10.2), // Out of spec
                    const FlSpot(3, 10.8), // Out of spec
                    const FlSpot(4, 10.6),
                    const FlSpot(5, 10.4),
                  ],
                  isCurved: true,
                  color: theme.colorScheme.primary,
                  barWidth: 4,
                  dotData: FlDotData(
                    show: true,
                    getDotPainter: (spot, percent, barData, index) {
                      bool isOutOfSpec = spot.y > 10.7 || spot.y < 10.3;
                      // Check if this spot is the Golden Ref spot
                      bool isGolden =
                          hasRef &&
                          spot.y == benchmarkY &&
                          index == 0; // index matching pin logic

                      return FlDotCirclePainter(
                        radius: isGolden ? 8 : (isOutOfSpec ? 6 : 4),
                        color: isGolden
                            ? Colors.white
                            : (isOutOfSpec
                                  ? Colors.red
                                  : theme.colorScheme.primary),
                        strokeWidth: isGolden ? 4 : (isOutOfSpec ? 2 : 0),
                        strokeColor: isGolden
                            ? theme.colorScheme.primary
                            : Colors.white,
                      );
                    },
                  ),
                  belowBarData: BarAreaData(
                    show: true,
                    color: theme.colorScheme.primary.withOpacity(0.1),
                  ),
                ),
              ],
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

    for (int i = 0; i < _selectedSessions.length; i++) {
      final sessionId = _selectedSessions[i];
      final isRef = _referenceSession == sessionId;

      // Determine color: Golden for ref, cycled palette for others
      final Color sessionColor = isRef
          ? Colors.amber.shade700
          : palette[i % palette.length].withOpacity(0.6);

      List<FlSpot> spots = [
        FlSpot(0, 15.0 + (i * 0.15)),
        FlSpot(1, 14.8 + (i * 0.1)),
        FlSpot(2, 15.2 + (i * 0.2)),
        FlSpot(3, 15.1 + (i * 0.1)),
        FlSpot(4, 15.3 + (i * 0.05)),
      ];

      bars.add(
        LineChartBarData(
          spots: spots,
          color: sessionColor,
          barWidth: isRef ? 5 : 2.5,
          isCurved: true,
          dotData: FlDotData(
            show: true,
            getDotPainter: (spot, percent, barData, index) =>
                FlDotCirclePainter(
                  radius: isRef ? 5 : 3.5,
                  color: sessionColor,
                  strokeWidth: isRef ? 2 : 1,
                  strokeColor: Colors.white,
                ),
          ),
          belowBarData: isRef
              ? BarAreaData(show: true, color: Colors.amber.withOpacity(0.05))
              : null,
        ),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildChartHeader(
          "Gain Stability Overlay (dB)",
          "Reference Highlights in Gold",
        ),
        const SizedBox(height: 32),
        Expanded(
          child: LineChart(
            LineChartData(
              lineTouchData: _buildTouchData(theme),
              gridData: const FlGridData(show: true, drawVerticalLine: false),
              titlesData: _buildTitlesData(
                xTitle: "Freq Bin",
                yTitle: "Gain (dB)",
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

    for (int i = 0; i < _selectedSessions.length; i++) {
      final sessionId = _selectedSessions[i];
      final isRef = _referenceSession == sessionId;

      // Mock spectral spots
      List<FlSpot> spots = [];
      if (sessionId == "Session_001") {
        spots = [
          const FlSpot(1.2, -65),
          const FlSpot(2.4, -72),
          const FlSpot(3.1, -80),
        ];
      } else {
        // Randomly offset subsequent sessions
        spots = [
          FlSpot(1.2 + (i * 0.05), -65 - (i * 2)),
          FlSpot(2.4 + (i * 0.03), -72 + (i * 1.5)),
        ];
        if (i == 1)
          spots.add(const FlSpot(1.5, -58)); // New spur in Session_002
      }

      layers.add(
        LineChartBarData(
          spots: spots,
          barWidth: 0,
          dotData: FlDotData(
            show: true,
            getDotPainter: (spot, percent, barData, index) {
              bool isAlert = spot.x == 1.5; // Our dummy alert spur
              Color dotColor = isRef
                  ? Colors.amber.shade700
                  : (isAlert ? Colors.red : Colors.indigo.withOpacity(0.4));

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

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildChartHeader(
          "Spurious Emissions Overlay",
          "Spectral variations highlighting Reference",
        ),
        const SizedBox(height: 32),
        Expanded(
          child: LineChart(
            LineChartData(
              lineTouchData: _buildTouchData(theme),
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

  LineTouchData _buildTouchData(ThemeData theme) {
    return LineTouchData(
      touchTooltipData: LineTouchTooltipData(
        getTooltipColor: (spot) => theme.colorScheme.surface.withOpacity(0.9),
        tooltipRoundedRadius: 8,
        tooltipBorder: BorderSide(
          color: theme.colorScheme.primary.withOpacity(0.2),
        ),
        getTooltipItems: (List<LineBarSpot> touchedSpots) {
          return touchedSpots.map((LineBarSpot touchedSpot) {
            // Mock session mapping based on touchedSpot
            // In a real app, we'd look up the session by spot index or bar index
            final sessionId =
                _selectedSessions[touchedSpot.barIndex %
                    _selectedSessions.length];
            final date = "2026-02-0${10 - (touchedSpot.barIndex % 10)}";
            final time = "14:30:${10 + touchedSpot.spotIndex}";

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
                  text: 'Value: ${touchedSpot.y.toStringAsFixed(2)}',
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
    String message = "All parameters are within nominal historical limits.";
    Color color = Colors.green;
    IconData icon = Icons.check_circle_outline;

    if (_selectedCategory == InsightDataCategory.singleValue) {
      // Check for out of spec points in mock data
      message =
          "Warning: 2 out-of-spec points detected (Low: 10.2, High: 10.8).";
      color = Colors.red;
      icon = Icons.error_outline;
    } else if (_selectedCategory == InsightDataCategory.variableResults) {
      message =
          "Alert: 1 new spur detected at 1.5 GHz that was not in previous sessions.";
      color = Colors.orange;
      icon = Icons.warning_amber_rounded;
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
          TextButton(onPressed: () {}, child: const Text("View Breakdown")),
        ],
      ),
    );
  }

  Widget _buildStatsRail(ThemeData theme) {
    final modeLabel = _useMeanAsReference ? "MEAN" : "GOLDEN";
    final refName = _useMeanAsReference
        ? "Average (Historical)"
        : (_referenceSession ?? "None");

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
          _statItem("Current Value", "10.42 dBm"),
          _statItem(
            "Delta from $modeLabel",
            "+0.04 dB",
            valueColor: Colors.orange,
          ),
          const Spacer(),
          _statItem("Max Drift", "+0.42 dB"),
          _statItem("Std Dev (σ)", "0.24"),
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

  Widget _buildMockDropdown(String value) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(value, style: const TextStyle(fontSize: 13)),
          const Icon(Icons.arrow_drop_down, size: 20),
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
}
