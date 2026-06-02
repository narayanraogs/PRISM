import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/screens/stability_screen.dart';
import 'package:prism_client/utils/lttb.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:intl/intl.dart';
import 'dart:js_interop';
import 'package:web/web.dart' as web;
import 'dart:convert';

class StabilityMonitoringScreen extends StatefulWidget {
  final List<StabilityParameterSelection> parameters;
  final String profileName;

  const StabilityMonitoringScreen({
    super.key,
    required this.parameters,
    required this.profileName,
  });

  @override
  State<StabilityMonitoringScreen> createState() =>
      _StabilityMonitoringScreenState();
}

class ParameterBuffer {
  final String description;
  final String instrument;
  final String parameter;
  final List<StabilityDataUpdate> rawData = [];
  final List<DataPoint> points = [];

  // Cache for downsampled points and FlSpots to avoid object churn
  List<DataPoint> displayPoints = [];
  List<FlSpot> displaySpots = [];
  bool _isDirty = false;

  double min = double.infinity;
  double max = double.negativeInfinity;
  double maxJump = 0;
  int count = 0;

  ParameterBuffer(this.description, this.instrument, this.parameter);

  void add(StabilityDataUpdate update, DateTime sessionStartTime) {
    rawData.add(update);
    final x =
        update.timestamp
            .difference(sessionStartTime)
            .inMilliseconds
            .toDouble() /
        1000.0;
    final y = update.value;

    // Ignore invalid data early
    if (x.isNaN || y.isNaN) return;

    if (points.isNotEmpty) {
      final jump = (y - points.last.y).abs();
      if (jump > maxJump) maxJump = jump;
    }

    // Deduplicate x-axis values to prevent fl_chart from hanging
    if (points.isNotEmpty && points.last.x >= x) {
      points.add(DataPoint(points.last.x + 0.001, y));
    } else {
      points.add(DataPoint(x, y));
    }

    _isDirty = true;

    if (y < min || min == double.infinity) min = y;
    if (y > max || max == double.negativeInfinity) max = y;
    count++;
  }

  /// Recalculates display points only if new data has arrived.
  /// This should be called outside the build() method to prevent frame drops.
  void refreshDisplayPoints(int threshold) {
    if (!_isDirty && displaySpots.isNotEmpty) return;

    displayPoints = points.length > threshold
        ? lttb(points, threshold)
        : List.from(points);

    // Convert to FlSpots once and cache to avoid object churn in build()
    displaySpots = displayPoints.map((p) => FlSpot(p.x, p.y)).toList();
    _isDirty = false;
  }

  // Safe bounds for fl_chart to prevent infinite loops in grid calculation
  double get minBound => (min == double.infinity || min.isNaN)
      ? -0.1
      : (min == max ? min - 0.1 : min);
  double get maxBound => (max == double.negativeInfinity || max.isNaN)
      ? 0.1
      : (min == max ? max + 0.1 : max);

  double get delta => (count < 2) ? 0 : (max - min);
  double get latest => points.isEmpty ? 0 : points.last.y;
}

class _StabilityMonitoringScreenState extends State<StabilityMonitoringScreen> {
  late Stream<StabilityResponse> _stabilityStream;
  StreamSubscription? _subscription;
  final Map<String, ParameterBuffer> _buffers = {};
  late DateTime _startTime;
  DateTime? _endTime;
  Timer? _uiTimer;
  bool _isAborting = false;
  bool _isCompleted = false;
  String? _focusedParameter; // Expanded chart label

  @override
  void initState() {
    super.initState();
    _startTime = DateTime.now();
    for (var p in widget.parameters) {
      _buffers[p.description] = ParameterBuffer(
        p.description,
        p.instrument,
        p.parameter,
      );
    }

    final serverService = Provider.of<ServerService>(context, listen: false);
    _stabilityStream = serverService.connectStability(
      widget.parameters,
      widget.profileName,
    );

    _subscription = _stabilityStream.listen((response) {
      if (mounted) {
        for (var update in response.updates) {
          final buffer = _buffers[update.description];
          if (buffer != null) {
            buffer.add(update, _startTime);
          }
        }
      }
    });

    // Refresh UI every 1000ms for smoothness and lower CPU usage
    _uiTimer = Timer.periodic(const Duration(milliseconds: 1000), (timer) {
      if (mounted) {
        // Run heavy math here (outside of build method)
        for (var buffer in _buffers.values) {
          buffer.refreshDisplayPoints(500); // Downsample to 500 points
        }
        setState(() {});
      }
    });
  }

  @override
  void dispose() {
    _uiTimer?.cancel();
    _subscription?.cancel();
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.closeStability();
    super.dispose();
  }

  void _handleAbort() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        surfaceTintColor: Colors.white,
        backgroundColor: Colors.white,
        title: Text(
          'Stop Monitoring?',
          style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
        ),
        content: const Text(
          'Are you sure you want to stop stability monitoring? The current session will be ended.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('CONTINUE'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              setState(() {
                _isAborting = true;
              });
              final serverService = Provider.of<ServerService>(
                context,
                listen: false,
              );
              serverService.sendAbortStability();
              // In a real app, we wait for the stream to close or a final response
              Future.delayed(const Duration(seconds: 1), () {
                if (mounted) {
                  setState(() {
                    _endTime = DateTime.now();
                    _isCompleted = true;
                    _uiTimer?.cancel();
                  });
                }
              });
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red.shade600,
              foregroundColor: Colors.white,
            ),
            child: const Text('STOP NOW'),
          ),
        ],
      ),
    );
  }

  String _getElapsedTime() {
    final now = _endTime ?? DateTime.now();
    final diff = now.difference(_startTime);
    final hours = diff.inHours.toString().padLeft(2, '0');
    final minutes = (diff.inMinutes % 60).toString().padLeft(2, '0');
    final seconds = (diff.inSeconds % 60).toString().padLeft(2, '0');
    return '$hours:$minutes:$seconds';
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return PopScope(
      canPop: _isCompleted,
      onPopInvoked: (didPop) {
        if (!didPop && !_isCompleted) {
          _handleAbort();
        }
      },
      child: Scaffold(
        backgroundColor: const Color(0xFFF8FAFC),
        body: Column(
          children: [
            _buildHeader(theme),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.all(24.0),
                child: _buildDynamicGrid(theme),
              ),
            ),
            _buildBottomBar(theme),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 20),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: Colors.grey.shade200)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: theme.colorScheme.primary.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(Icons.speed_rounded, color: theme.colorScheme.primary),
          ),
          const SizedBox(width: 20),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                widget.profileName,
                style: GoogleFonts.outfit(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Colors.black,
                ),
              ),
              Row(
                children: [
                  Container(
                    width: 8,
                    height: 8,
                    decoration: BoxDecoration(
                      color: _isCompleted ? Colors.orange : Colors.green,
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    _isCompleted ? 'Monitoring Stopped' : 'Live Monitoring',
                    style: GoogleFonts.inter(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: Colors.grey.shade600,
                    ),
                  ),
                ],
              ),
            ],
          ),
          const Spacer(),
          _buildMetricHeader('ELAPSED TIME', _getElapsedTime(), theme),
          const SizedBox(width: 40),
          _buildMetricHeader(
            'TOTAL SAMPLES',
            _buffers.values.fold(0, (sum, b) => sum + b.count).toString(),
            theme,
          ),
        ],
      ),
    );
  }

  Widget _buildMetricHeader(String label, String value, ThemeData theme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 10,
            fontWeight: FontWeight.w900,
            color: Colors.grey.shade500,
            letterSpacing: 1.2,
          ),
        ),
        Text(
          value,
          style: GoogleFonts.robotoMono(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: theme.colorScheme.primary,
          ),
        ),
      ],
    );
  }

  Widget _buildDynamicGrid(ThemeData theme) {
    if (_focusedParameter != null) {
      final param = widget.parameters.firstWhere(
        (p) => p.description == _focusedParameter,
      );
      final index = widget.parameters.indexOf(param);
      final buffer = _buffers[param.description]!;

      return Column(
        children: [
          Row(
            children: [
              TextButton.icon(
                onPressed: () => setState(() => _focusedParameter = null),
                icon: const Icon(Icons.arrow_back_rounded),
                label: const Text('BACK TO GRID'),
                style: TextButton.styleFrom(
                  foregroundColor: Colors.grey.shade700,
                  backgroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
              const Spacer(),
              Text(
                'FOCUS MODE',
                style: GoogleFonts.inter(
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                  color: theme.colorScheme.primary,
                  letterSpacing: 2,
                ),
              ),
            ],
          ),
          const SizedBox(height: 20),
          Expanded(
            child: _buildChartCard(
              theme,
              param,
              buffer,
              index,
              isFocused: true,
            ),
          ),
        ],
      );
    }

    int count = widget.parameters.length;
    int crossAxisCount = 1;
    if (count > 1 && count <= 4) crossAxisCount = 2;
    if (count > 4) crossAxisCount = 3;

    return GridView.builder(
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: crossAxisCount,
        crossAxisSpacing: 20,
        mainAxisSpacing: 20,
        childAspectRatio: crossAxisCount == 1 ? 3.5 : 1.8,
      ),
      itemCount: count,
      itemBuilder: (context, index) {
        final param = widget.parameters[index];
        final buffer = _buffers[param.description]!;
        return _buildChartCard(theme, param, buffer, index);
      },
    );
  }

  Widget _buildChartCard(
    ThemeData theme,
    StabilityParameterSelection param,
    ParameterBuffer buffer,
    int index, {
    bool isFocused = false,
  }) {
    // Use distinct colors for each parameter
    final List<Color> chartColors = [
      Colors.blue,
      Colors.green,
      Colors.orange,
      Colors.purple,
      Colors.pink,
      Colors.amber,
      Colors.indigo,
      Colors.teal,
    ];
    final color = chartColors[index % chartColors.length];

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      param.description,
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.bold,
                        fontSize: 16,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                    Text(
                      '${param.instrument} • ${param.parameter}',
                      style: GoogleFonts.inter(
                        fontSize: 12,
                        color: Colors.grey.shade500,
                      ),
                    ),
                  ],
                ),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        buffer.latest.toStringAsFixed(3),
                        style: GoogleFonts.robotoMono(
                          fontSize: isFocused ? 32 : 22,
                          fontWeight: FontWeight.w900,
                          color: color,
                        ),
                      ),
                      const SizedBox(width: 12),
                      IconButton(
                        onPressed: () {
                          setState(() {
                            if (isFocused) {
                              _focusedParameter = null;
                            } else {
                              _focusedParameter = param.description;
                            }
                          });
                        },
                        icon: Icon(
                          isFocused
                              ? Icons.close_fullscreen_rounded
                              : Icons.open_in_full_rounded,
                          size: 20,
                          color: Colors.grey.shade400,
                        ),
                        style: IconButton.styleFrom(
                          hoverColor: theme.colorScheme.primary.withOpacity(
                            0.05,
                          ),
                        ),
                      ),
                    ],
                  ),
                  Text(
                    'LATEST',
                    style: TextStyle(
                      fontSize: 8,
                      fontWeight: FontWeight.w900,
                      color: Colors.grey.shade400,
                    ),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 20),
          Expanded(
            child: LineChart(
              LineChartData(
                gridData: FlGridData(
                  show: true,
                  drawVerticalLine: true,
                  drawHorizontalLine: true,
                  // Disable dynamic grid interval calculation if range is near zero
                  horizontalInterval: (buffer.max - buffer.min).abs() < 0.001
                      ? 1.0
                      : null,
                  getDrawingHorizontalLine: (value) =>
                      FlLine(color: Colors.grey.shade100, strokeWidth: 1),
                  getDrawingVerticalLine: (value) =>
                      FlLine(color: Colors.grey.shade100, strokeWidth: 1),
                ),
                titlesData: FlTitlesData(
                  show: true,
                  rightTitles: const AxisTitles(
                    sideTitles: SideTitles(showTitles: false),
                  ),
                  topTitles: const AxisTitles(
                    sideTitles: SideTitles(showTitles: false),
                  ),
                  bottomTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize: 30,
                      interval: 300, // Labels every 5 mins
                      getTitlesWidget: (value, meta) {
                        if (value == 0) return const Text('0s');
                        final mins = (value / 60).floor();
                        if (mins % 5 == 0) return Text('${mins}m');
                        return const Text('');
                      },
                    ),
                  ),
                  leftTitles: AxisTitles(
                    sideTitles: SideTitles(
                      showTitles: true,
                      reservedSize:
                          50, // Slightly increased to accommodate extra digit
                      getTitlesWidget: (value, meta) {
                        return Text(
                          value.toStringAsFixed(2),
                          style: TextStyle(
                            color: Colors.grey.shade400,
                            fontSize: 10,
                          ),
                        );
                      },
                    ),
                  ),
                ),
                borderData: FlBorderData(show: false),
                minX: buffer.points.isEmpty ? 0 : buffer.points.first.x,
                maxX: buffer.points.isEmpty
                    ? 10
                    : (buffer.points.last.x <= buffer.points.first.x
                          ? buffer.points.first.x + 1
                          : buffer.points.last.x),
                minY: buffer.minBound,
                maxY: buffer.maxBound,
                lineBarsData: [
                  LineChartBarData(
                    spots: buffer.displaySpots,
                    isCurved: false,
                    color: color,
                    barWidth: 2,
                    isStrokeCapRound: true,
                    dotData: const FlDotData(show: false),
                    belowBarData: BarAreaData(
                      show: true,
                      color: color.withOpacity(0.05),
                    ),
                  ),
                ],
                lineTouchData: LineTouchData(
                  touchTooltipData: LineTouchTooltipData(
                    getTooltipColor: (spot) => Colors.white,
                    getTooltipItems: (touchedSpots) {
                      return touchedSpots.map((spot) {
                        return LineTooltipItem(
                          '${spot.y.toStringAsFixed(4)}\n${Duration(seconds: spot.x.toInt()).toString().split('.').first}',
                          GoogleFonts.robotoMono(
                            color: color,
                            fontWeight: FontWeight.bold,
                            fontSize: 12,
                          ),
                        );
                      }).toList();
                    },
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildMiniMetric('MIN', buffer.min.toStringAsFixed(3), theme),
              _buildMiniMetric('MAX', buffer.max.toStringAsFixed(3), theme),
              _buildMiniMetric('ΔMAX', buffer.delta.toStringAsFixed(3), theme),
              _buildMiniMetric(
                'δMAX',
                buffer.maxJump.toStringAsFixed(3),
                theme,
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildMiniMetric(String label, String value, ThemeData theme) {
    return Column(
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 9,
            fontWeight: FontWeight.w900,
            color: Colors.grey.shade400,
            letterSpacing: 1.0,
          ),
        ),
        Text(
          value == 'Infinity' || value == '-Infinity' ? '0.000' : value,
          style: GoogleFonts.robotoMono(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: Colors.black87,
          ),
        ),
      ],
    );
  }

  void _exportCSV() {
    final List<String> csvRows = [];

    // Header
    csvRows.add(
      'ISO Timestamp,Elapsed (s),Instrument,Parameter,Description,Value',
    );

    // Collect all samples from all buffers
    final List<_ExportSample> allSamples = [];
    for (var buffer in _buffers.values) {
      for (var update in buffer.rawData) {
        allSamples.add(
          _ExportSample(
            timestamp: update.timestamp,
            elapsed:
                update.timestamp.difference(_startTime).inMilliseconds / 1000.0,
            instrument: buffer.instrument,
            parameter: buffer.parameter,
            description: buffer.description,
            value: update.value,
          ),
        );
      }
    }

    // Sort by timestamp for proper chronological order across instruments
    allSamples.sort((a, b) => a.timestamp.compareTo(b.timestamp));

    // Convert to CSV strings
    for (var sample in allSamples) {
      csvRows.add(
        '${sample.timestamp.toIso8601String()},${sample.elapsed.toStringAsFixed(3)},'
        '"${sample.instrument}","${sample.parameter}","${sample.description}",${sample.value}',
      );
    }

    final csvString = csvRows.join('\n');
    final bytes = utf8.encode(csvString);
    final blob = web.Blob(
      [bytes.toJS].toJS,
      web.BlobPropertyBag(type: 'text/csv'),
    );
    final url = web.URL.createObjectURL(blob);

    final anchor = web.document.createElement('a') as web.HTMLAnchorElement;
    anchor.href = url;
    anchor.download =
        'Stability_${widget.profileName.replaceAll(' ', '_')}_${DateFormat('yyyyMMdd_HHmmss').format(DateTime.now())}.csv';
    
    // Append to DOM, click, and remove to ensure browser respects download filename
    web.document.body?.appendChild(anchor);
    anchor.click();
    anchor.remove();

    // Delay revocation to ensure download is successfully initialized
    Future.delayed(const Duration(milliseconds: 200), () {
      web.URL.revokeObjectURL(url);
    });

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Stability data exported successfully.')),
    );
  }

  Widget _buildBottomBar(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 20),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(top: BorderSide(color: Colors.grey.shade200)),
      ),
      child: Row(
        children: [
          const Icon(Icons.info_outline, size: 18, color: Colors.blue),
          const SizedBox(width: 12),
          Text(
            _isAborting
                ? 'Stopping monitoring...'
                : 'System running normally. All instruments responding.',
            style: GoogleFonts.inter(
              fontSize: 13,
              fontWeight: FontWeight.w500,
              color: Colors.grey.shade600,
            ),
          ),
          const Spacer(),
          if (_buffers.values.any((b) => b.count > 0))
            Padding(
              padding: const EdgeInsets.only(right: 12.0),
              child: TextButton.icon(
                onPressed: _exportCSV,
                icon: const Icon(Icons.download_rounded),
                label: const Text('EXPORT CSV'),
                style: TextButton.styleFrom(
                  foregroundColor: Colors.blue.shade700,
                ),
              ),
            ),
          if (_isCompleted)
            ElevatedButton.icon(
              onPressed: () => Navigator.pop(context),
              icon: const Icon(Icons.arrow_back),
              label: const Text('RETURN TO CONFIG'),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 16,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            )
          else
            OutlinedButton.icon(
              onPressed: _handleAbort,
              icon: const Icon(Icons.stop_circle_outlined),
              label: const Text('STOP MONITORING'),
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.red,
                side: const BorderSide(color: Colors.red),
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 16,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _ExportSample {
  final DateTime timestamp;
  final double elapsed;
  final String instrument;
  final String parameter;
  final String description;
  final double value;

  _ExportSample({
    required this.timestamp,
    required this.elapsed,
    required this.instrument,
    required this.parameter,
    required this.description,
    required this.value,
  });
}
