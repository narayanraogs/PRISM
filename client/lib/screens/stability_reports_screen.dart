import 'dart:convert';
import 'dart:math' as math;
import 'dart:typed_data';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:web/web.dart' as web;
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'dart:js_interop';

import '../utils/lttb.dart';

class StabilityReportsScreen extends StatefulWidget {
  const StabilityReportsScreen({super.key});

  @override
  State<StabilityReportsScreen> createState() => _StabilityReportsScreenState();
}

class ParameterData {
  final String name;
  final List<DataPoint> points;
  final List<DataPoint> displayPoints;
  final double min;
  final double max;

  ParameterData({
    required this.name,
    required this.points,
    required this.displayPoints,
    required this.min,
    required this.max,
  });
}

class _StabilityReportsScreenState extends State<StabilityReportsScreen> {
  final GlobalKey _chartKey = GlobalKey();

  List<StabilityReportModel> _sessions = [];
  StabilityReportModel? _selectedSession;
  final List<String> _selectedParams = [];
  Map<String, ParameterData> _loadedData = {};
  bool _isLoadingMetadata = true;
  bool _isLoadingPoints = false;

  // Axis Controls
  final TextEditingController _xMinController = TextEditingController();
  final TextEditingController _xMaxController = TextEditingController();
  final TextEditingController _y1MinController = TextEditingController();
  final TextEditingController _y1MaxController = TextEditingController();
  final TextEditingController _y2MinController = TextEditingController();
  final TextEditingController _y2MaxController = TextEditingController();
  bool _isHelpOpen = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _syncMetadata();
  }

  void _syncMetadata() {
    final serverService = Provider.of<ServerService>(context);
    final metadata = serverService.status.bootstrapData?.stabilityReportsData;

    if (metadata != null && metadata.ok) {
      final newSessions = List.generate(metadata.id.length, (i) {
        return StabilityReportModel(
          id: metadata.id[i],
          date: metadata.date[i],
          time: metadata.time[i],
          parameters: metadata.parameters[i],
        );
      });

      // Update sessions if length changed or they were empty
      if (newSessions.length != _sessions.length || _sessions.isEmpty) {
        debugPrint(
          'StabilityReportsScreen: Syncing ${newSessions.length} sessions',
        );
        setState(() {
          _sessions = newSessions;
          _isLoadingMetadata = false;
          if (_sessions.isNotEmpty &&
              (_selectedSession == null ||
                  !_sessions.any((s) => s.id == _selectedSession!.id))) {
            _selectedSession = _sessions.first;
          }
        });
      } else {
        setState(() => _isLoadingMetadata = false);
      }
    } else {
      setState(() => _isLoadingMetadata = false);
    }
  }

  void _fetchMetadata() {
    // Manually trigger a bootstrap refresh if needed
    setState(() => _isLoadingMetadata = true);
    Provider.of<ServerService>(context, listen: false).fetchBootstrapData();
  }

  Future<void> _fetchParamData(String param) async {
    if (_selectedSession == null) return;

    setState(() => _isLoadingPoints = true);

    final serverService = Provider.of<ServerService>(context, listen: false);
    final response = await serverService.fetchStabilityPoints(
      _selectedSession!.id,
      param,
    );

    if (mounted) {
      if (response != null && response.ok) {
        if (response.points.isEmpty) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('No data points found for this parameter'),
            ),
          );
          setState(() => _isLoadingPoints = false);
          return;
        }

        // Convert TSInt to relative seconds
        // We'll normalize X based on the earliest timestamp in the session's first loaded parameter
        int baseTs = response.points.first.tsInt;
        // If we already have some data loaded, try to find the absolute minimum TS
        for (var data in _loadedData.values) {
          // This is tricky because we don't store raw TSInt in ParameterData.
          // Let's just use the current parameter's first point as base if _loadedData is empty
        }

        final points = response.points.map((p) {
          return DataPoint((p.tsInt - baseTs) / 1000.0, p.value);
        }).toList();

        final minY = points.map((p) => p.y).reduce(math.min);
        final maxY = points.map((p) => p.y).reduce(math.max);

        setState(() {
          _loadedData[param] = ParameterData(
            name: param,
            points: points,
            displayPoints: points.length > 2000
                ? lttb(points, 2000)
                : List.from(points),
            min: minY - (maxY == minY ? 1.0 : (maxY - minY) * 0.1),
            max: maxY + (maxY == minY ? 1.0 : (maxY - minY) * 0.1),
          );
          _isLoadingPoints = false;
          _updateAxisControllers();
        });
      } else {
        setState(() => _isLoadingPoints = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              response?.message ?? 'Failed to load parameter points',
            ),
          ),
        );
      }
    }
  }

  void _toggleParam(String param) async {
    if (_selectedParams.contains(param)) {
      setState(() {
        _selectedParams.remove(param);
        _updateAxisControllers();
      });
    } else {
      if (_selectedParams.length >= 2) {
        _selectedParams.removeAt(0);
      }

      if (!_loadedData.containsKey(param)) {
        await _fetchParamData(param);
      }

      if (_loadedData.containsKey(param)) {
        setState(() {
          _selectedParams.add(param);
          _updateAxisControllers();
        });
      }
    }
  }

  void _updateAxisControllers() {
    if (_selectedParams.isEmpty) return;

    final data1 = _loadedData[_selectedParams[0]]!;
    _xMinController.text = '0';
    _xMaxController.text = data1.points.last.x.toStringAsFixed(0);
    _y1MinController.text = data1.min.toStringAsFixed(2);
    _y1MaxController.text = data1.max.toStringAsFixed(2);

    if (_selectedParams.length > 1) {
      final data2 = _loadedData[_selectedParams[1]]!;
      _y2MinController.text = data2.min.toStringAsFixed(2);
      _y2MaxController.text = data2.max.toStringAsFixed(2);
    }
  }

  double _normalize(
    double val,
    double min2,
    double max2,
    double min1,
    double max1,
  ) {
    if (max2 == min2) return min1;
    return min1 + ((val - min2) / (max2 - min2)) * (max1 - min1);
  }

  double _denormalize(
    double val,
    double min1,
    double max1,
    double min2,
    double max2,
  ) {
    if (max1 == min1) return min2;
    return min2 + ((val - min1) / (max1 - min1)) * (max2 - min2);
  }

  Future<void> _exportImage(String format) async {
    try {
      final boundary =
          _chartKey.currentContext!.findRenderObject() as RenderRepaintBoundary;
      final image = await boundary.toImage(pixelRatio: 3.0);
      final byteData = await image.toByteData(format: ui.ImageByteFormat.png);
      final bytes = byteData!.buffer.asUint8List();

      final blob = web.Blob(
        [bytes.toJS].toJS,
        web.BlobPropertyBag(type: 'image/$format'),
      );
      final url = web.URL.createObjectURL(blob);
      final anchor = web.document.createElement('a') as web.HTMLAnchorElement;
      anchor.href = url;
      anchor.download =
          'Stability_Report_${_selectedSession?.date}_${_selectedSession?.time}.$format';
      anchor.click();
      web.URL.revokeObjectURL(url);

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Report exported as ${format.toUpperCase()}')),
      );
    } catch (e) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('Error exporting image: $e')));
    }
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
                title: 'Stability Reports',
                subtitle: 'Analyze long-term stability data',
                icon: Icons.analytics_rounded,
                trailing: _buildHelpTrigger(theme),
              ),
              Expanded(
                child: Row(
                  children: [
                    _buildSidebar(theme),
                    Expanded(
                      child: ContentCard(
                        margin: const EdgeInsets.all(16),
                        width: double.infinity,
                        padding: const EdgeInsets.all(32),
                        child: _buildPreviewArea(theme),
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

  Widget _buildSidebar(ThemeData theme) {
    return ContentCard(
      width: 380,
      isSidebar: true,
      margin: const EdgeInsets.fromLTRB(16, 16, 0, 16),
      padding: EdgeInsets.zero,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Text(
              'Report Editor',
              style: GoogleFonts.outfit(
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(24),
              children: [
                _buildSectionTitle('SESSION SELECTION'),
                const SizedBox(height: 12),
                _buildSessionDropdown(theme),
                const SizedBox(height: 32),

                _buildSectionTitle('PARAMETERS (MAX 2)'),
                const SizedBox(height: 12),
                _buildParamSelector(theme),
                const SizedBox(height: 32),

                _buildSectionTitle('AXIS LIMITS'),
                const SizedBox(height: 20),
                _buildAxisControls(theme),
                const SizedBox(height: 40),

                _buildActionButtons(theme),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: GoogleFonts.inter(
        fontSize: 11,
        fontWeight: FontWeight.w900,
        color: Colors.grey.shade500,
        letterSpacing: 1.2,
      ),
    );
  }

  Widget _buildSessionDropdown(ThemeData theme) {
    if (_isLoadingMetadata) {
      return const LinearProgressIndicator();
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<StabilityReportModel>(
          value: _selectedSession,
          isExpanded: true,
          items: _sessions.map((s) {
            return DropdownMenuItem(
              value: s,
              child: Text(
                '${s.date} ${s.time}',
                style: GoogleFonts.inter(fontSize: 14),
              ),
            );
          }).toList(),
          onChanged: (val) {
            if (val != null) {
              setState(() {
                _selectedSession = val;
                _selectedParams.clear();
                _loadedData.clear();
              });
            }
          },
        ),
      ),
    );
  }

  Widget _buildParamSelector(ThemeData theme) {
    if (_selectedSession == null) {
      return Text(
        'Select a session first',
        style: TextStyle(color: Colors.grey.shade400),
      );
    }

    return Column(
      children: _selectedSession!.parameters.map((param) {
        final isSelected = _selectedParams.contains(param);
        final index = _selectedParams.indexOf(param);

        return Padding(
          padding: const EdgeInsets.only(bottom: 8.0),
          child: InkWell(
            onTap: () => _toggleParam(param),
            borderRadius: BorderRadius.circular(12),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: isSelected
                    ? theme.colorScheme.primary.withOpacity(0.05)
                    : Colors.transparent,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: isSelected
                      ? theme.colorScheme.primary.withOpacity(0.3)
                      : Colors.grey.shade200,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    isSelected
                        ? Icons.check_circle_rounded
                        : Icons.radio_button_unchecked_rounded,
                    size: 20,
                    color: isSelected
                        ? theme.colorScheme.primary
                        : Colors.grey.shade400,
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      param,
                      style: GoogleFonts.inter(
                        fontSize: 14,
                        fontWeight: isSelected
                            ? FontWeight.w600
                            : FontWeight.normal,
                        color: isSelected
                            ? Colors.black87
                            : Colors.grey.shade700,
                      ),
                    ),
                  ),
                  if (isSelected)
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: index == 0 ? Colors.indigo : Colors.teal,
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        'AXIS ${index + 1}',
                        style: const TextStyle(
                          fontSize: 9,
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
        );
      }).toList(),
    );
  }

  Widget _buildAxisControls(ThemeData theme) {
    return Column(
      children: [
        _buildAxisGroup('Time (X-Axis)', [
          _buildField('Min (s)', _xMinController),
          _buildField('Max (s)', _xMaxController),
        ]),
        const SizedBox(height: 24),
        _buildAxisGroup('Left Axis (Y1)', [
          _buildField('Min', _y1MinController),
          _buildField('Max', _y1MaxController),
        ], color: Colors.indigo),
        const SizedBox(height: 24),
        Opacity(
          opacity: _selectedParams.length > 1 ? 1.0 : 0.4,
          child: _buildAxisGroup('Right Axis (Y2)', [
            _buildField('Min', _y2MinController),
            _buildField('Max', _y2MaxController),
          ], color: Colors.teal),
        ),
      ],
    );
  }

  Widget _buildAxisGroup(String label, List<Widget> fields, {Color? color}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            if (color != null)
              Container(
                width: 3,
                height: 12,
                margin: const EdgeInsets.only(right: 8),
                decoration: BoxDecoration(
                  color: color,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            Text(
              label,
              style: GoogleFonts.inter(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: Colors.black87,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          children:
              fields
                  .expand(
                    (f) => [Expanded(child: f), const SizedBox(width: 12)],
                  )
                  .toList()
                ..removeLast(),
        ),
      ],
    );
  }

  Widget _buildField(String hint, TextEditingController controller) {
    return TextField(
      controller: controller,
      keyboardType: TextInputType.number,
      style: GoogleFonts.robotoMono(fontSize: 13),
      decoration: InputDecoration(
        labelText: hint,
        labelStyle: TextStyle(fontSize: 12, color: Colors.grey.shade500),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        filled: true,
        fillColor: Colors.grey.shade50,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: Colors.grey.shade200),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: Colors.grey.shade200),
        ),
      ),
      onChanged: (val) => setState(() {}),
    );
  }

  Widget _buildActionButtons(ThemeData theme) {
    return Column(
      children: [
        ElevatedButton(
          onPressed: () => setState(() {}),
          style: ElevatedButton.styleFrom(
            backgroundColor: theme.colorScheme.primary,
            foregroundColor: Colors.white,
            minimumSize: const Size(double.infinity, 52),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
          ),
          child: const Text('UPDATE PLOT'),
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: () => _exportImage('png'),
                icon: const Icon(Icons.image_outlined, size: 18),
                label: const Text('PNG'),
                style: OutlinedButton.styleFrom(
                  minimumSize: const Size(0, 48),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: OutlinedButton.icon(
                onPressed: () => _exportImage('jpeg'),
                icon: const Icon(Icons.picture_as_pdf_outlined, size: 18),
                label: const Text('JPG'),
                style: OutlinedButton.styleFrom(
                  minimumSize: const Size(0, 48),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildPreviewArea(ThemeData theme) {
    if (_isLoadingMetadata) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_selectedParams.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.add_chart_rounded,
              size: 64,
              color: Colors.grey.shade300,
            ),
            const SizedBox(height: 16),
            Text(
              _isLoadingPoints
                  ? 'Loading points...'
                  : 'Select parameters to begin plotting',
              style: GoogleFonts.inter(color: Colors.grey.shade500),
            ),
            if (_isLoadingPoints) const SizedBox(height: 20),
            if (_isLoadingPoints)
              const CircularProgressIndicator(strokeWidth: 2),
          ],
        ),
      );
    }

    return Column(
      children: [
        Expanded(
          child: RepaintBoundary(
            key: _chartKey,
            child: Container(
              padding: const EdgeInsets.all(16),
              // Removed inner decoration as the parent ContentCard provides container style
              child: Stack(
                children: [
                  _buildFinalChart(theme),
                  if (_isLoadingPoints)
                    const Center(child: CircularProgressIndicator()),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildFinalChart(ThemeData theme) {
    final xMin = double.tryParse(_xMinController.text) ?? 0;
    final xMax = double.tryParse(_xMaxController.text) ?? 3600;
    final y1Min = double.tryParse(_y1MinController.text) ?? 0;
    final y1Max = double.tryParse(_y1MaxController.text) ?? 100;
    final y2Min = double.tryParse(_y2MinController.text) ?? 0;
    final y2Max = double.tryParse(_y2MaxController.text) ?? 100;

    final List<LineChartBarData> barData = [];

    // Param 1
    final data1 = _loadedData[_selectedParams[0]]!;
    barData.add(
      LineChartBarData(
        spots: data1.displayPoints.map((p) => FlSpot(p.x, p.y)).toList(),
        isCurved: true,
        curveSmoothness: 0.1,
        color: Colors.indigo,
        barWidth: 2,
        dotData: const FlDotData(show: false),
        belowBarData: BarAreaData(
          show: true,
          color: Colors.indigo.withOpacity(0.05),
        ),
      ),
    );

    // Param 2 (Normalized)
    if (_selectedParams.length > 1) {
      final data2 = _loadedData[_selectedParams[1]]!;
      barData.add(
        LineChartBarData(
          spots: data2.displayPoints
              .map(
                (p) => FlSpot(p.x, _normalize(p.y, y2Min, y2Max, y1Min, y1Max)),
              )
              .toList(),
          isCurved: true,
          curveSmoothness: 0.1,
          color: Colors.teal,
          barWidth: 2,
          dotData: const FlDotData(show: false),
          belowBarData: BarAreaData(
            show: true,
            color: Colors.teal.withOpacity(0.05),
          ),
        ),
      );
    }

    return Column(
      children: [
        Row(
          children: [
            _buildChartLegend(_selectedParams[0], Colors.indigo),
            if (_selectedParams.length > 1) ...[
              const SizedBox(width: 24),
              _buildChartLegend(_selectedParams[1], Colors.teal),
            ],
            const Spacer(),
            Text(
              '${_selectedSession?.date} ${_selectedSession?.time}',
              style: GoogleFonts.robotoMono(
                fontSize: 12,
                color: Colors.grey.shade400,
              ),
            ),
          ],
        ),
        const SizedBox(height: 48),
        Expanded(
          child: LineChart(
            LineChartData(
              minX: xMin,
              maxX: xMax,
              minY: y1Min,
              maxY: y1Max,
              lineBarsData: barData,
              gridData: FlGridData(
                show: true,
                drawVerticalLine: true,
                getDrawingHorizontalLine: (val) =>
                    FlLine(color: Colors.grey.shade100, strokeWidth: 1),
                getDrawingVerticalLine: (val) =>
                    FlLine(color: Colors.grey.shade100, strokeWidth: 1),
              ),
              titlesData: FlTitlesData(
                leftTitles: AxisTitles(
                  axisNameWidget: Text(
                    'Axis 1: ${_selectedParams[0]}',
                    style: const TextStyle(
                      fontSize: 10,
                      color: Colors.indigo,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  axisNameSize: 22,
                  sideTitles: SideTitles(
                    showTitles: true,
                    reservedSize: 50,
                    getTitlesWidget: (val, meta) => Text(
                      val.toStringAsFixed(1),
                      style: const TextStyle(fontSize: 10, color: Colors.grey),
                    ),
                  ),
                ),
                rightTitles: AxisTitles(
                  axisNameWidget: _selectedParams.length > 1
                      ? Text(
                          'Axis 2: ${_selectedParams[1]}',
                          style: const TextStyle(
                            fontSize: 10,
                            color: Colors.teal,
                            fontWeight: FontWeight.bold,
                          ),
                        )
                      : null,
                  axisNameSize: 22,
                  sideTitles: SideTitles(
                    showTitles: _selectedParams.length > 1,
                    reservedSize: 50,
                    getTitlesWidget: (val, meta) {
                      final realVal = _denormalize(
                        val,
                        y1Min,
                        y1Max,
                        y2Min,
                        y2Max,
                      );
                      return Text(
                        realVal.toStringAsFixed(1),
                        style: const TextStyle(
                          fontSize: 10,
                          color: Colors.grey,
                        ),
                      );
                    },
                  ),
                ),
                topTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                bottomTitles: AxisTitles(
                  axisNameWidget: const Text(
                    'Time (seconds)',
                    style: TextStyle(fontSize: 10, color: Colors.grey),
                  ),
                  sideTitles: SideTitles(
                    showTitles: true,
                    getTitlesWidget: (val, meta) => Text(
                      '${val.toInt()}s',
                      style: const TextStyle(fontSize: 10, color: Colors.grey),
                    ),
                  ),
                ),
              ),
              borderData: FlBorderData(
                show: true,
                border: Border.all(color: Colors.grey.shade100),
              ),
              lineTouchData: LineTouchData(
                touchTooltipData: LineTouchTooltipData(
                  getTooltipColor: (spot) => Colors.white,
                  getTooltipItems: (touchedSpots) {
                    return touchedSpots.map((spot) {
                      final isSecond = spot.barIndex == 1;
                      final val = isSecond
                          ? _denormalize(spot.y, y1Min, y1Max, y2Min, y2Max)
                          : spot.y;
                      return LineTooltipItem(
                        '${_selectedParams[spot.barIndex]}\n${val.toStringAsFixed(4)}',
                        TextStyle(
                          color: isSecond ? Colors.teal : Colors.indigo,
                          fontWeight: FontWeight.bold,
                          fontSize: 11,
                        ),
                      );
                    }).toList();
                  },
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildChartLegend(String label, Color color) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(width: 8),
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: Colors.grey.shade800,
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
                  'Stability Report Help',
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
                  'Dual-Axis Plotting',
                  'Select up to 2 parameters to compare variations over time. The second parameter is '
                      'automatically normalized to the first axis to allow meaningful visual overlay.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Axis Limits',
                  'PRISM auto-scales axes when parameters are first loaded. You can manually override these '
                      'limits to zoom into specific power fluctuations or frequency drifts.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'LTTB Downsampling',
                  'For sessions with millions of points, the chart uses the Largest Triangle Three Buckets (LTTB) '
                      'algorithm to preserve visual spikes/extreme values while ensuring the UI remains responsive.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Report Export',
                  'Export the current view as High-Resolution PNG or JPG. The filename automatically includes '
                      'the session date and time for easy archival.',
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

class StabilityReportModel {
  final int id;
  final String date;
  final String time;
  final List<String> parameters;

  StabilityReportModel({
    required this.id,
    required this.date,
    required this.time,
    required this.parameters,
  });
}
