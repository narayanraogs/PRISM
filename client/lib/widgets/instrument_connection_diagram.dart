import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

enum DiagramType {
  pmZero,
  cableMeasurement,
  attnSG,
  attnGTx,
  attnTSM,
  converterComplex,
  converterSimple,
}

class InstrumentConnectionDiagram extends StatefulWidget {
  final DiagramType type;
  final String? tsmInputName;
  final String? tsmOutputName;
  final String? receiverName;
  final String? inputPortName;
  final String? outputPortName;

  const InstrumentConnectionDiagram({
    super.key,
    required this.type,
    this.tsmInputName,
    this.tsmOutputName,
    this.receiverName,
    this.inputPortName,
    this.outputPortName,
  });

  @override
  State<InstrumentConnectionDiagram> createState() =>
      _InstrumentConnectionDiagramState();
}

class _InstrumentConnectionDiagramState
    extends State<InstrumentConnectionDiagram>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 4),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final primaryColor = theme.colorScheme.primary;
    final accentColor = theme.colorScheme.secondary;

    return Container(
      padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
      decoration: BoxDecoration(
        color: primaryColor.withOpacity(0.03),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: primaryColor.withOpacity(0.1)),
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final width = constraints.maxWidth;
          final height = constraints.maxHeight;

          return Stack(
            clipBehavior: Clip.none,
            children: [
              // Edges (drawn with CustomPaint)
              Positioned.fill(
                child: AnimatedBuilder(
                  animation: _controller,
                  builder: (context, child) {
                    return CustomPaint(
                      painter: _ConnectionPainter(
                        type: widget.type,
                        primaryColor: primaryColor,
                        accentColor: accentColor,
                        animationValue: _controller.value,
                      ),
                    );
                  },
                ),
              ),

              // Nodes
              if (widget.type == DiagramType.pmZero) ...[
                _buildNode(
                  context,
                  label: 'Signal\nGenerator',
                  alignment: Alignment.centerLeft,
                ),
                _buildNode(
                  context,
                  label: 'Power\nMeter',
                  alignment: Alignment.centerRight,
                ),
              ] else if (widget.type == DiagramType.cableMeasurement) ...[
                _buildNode(
                  context,
                  label: 'Signal\nGenerator',
                  alignment: Alignment.centerLeft,
                ),
                _buildNode(
                  context,
                  label: 'Power\nMeter',
                  alignment: Alignment.centerRight,
                ),
              ] else if (widget.type == DiagramType.attnSG) ...[
                _buildNode(
                  context,
                  label: 'Signal\nGenerator',
                  alignment: Alignment.centerLeft,
                  isHighlighted: true,
                ),
                _buildNode(
                  context,
                  label: 'Spectrum\nAnalyzer',
                  alignment: Alignment.centerRight,
                ),
              ] else if (widget.type == DiagramType.attnGTx) ...[
                _buildNode(
                  context,
                  label: 'Ground\nTransmitter',
                  alignment: Alignment.centerLeft,
                  isHighlighted: true,
                ),
                _buildNode(
                  context,
                  label: 'Spectrum\nAnalyzer',
                  alignment: Alignment.centerRight,
                ),
              ] else if (widget.type == DiagramType.attnTSM) ...[
                _buildNode(
                  context,
                  label: 'Signal\nGenerator',
                  alignment: Alignment.centerLeft,
                ),
                Positioned(
                  left: width * 0.5 - 50,
                  top: height * 0.5 - 30,
                  child: _buildNode(
                    context,
                    label: 'Test Selection\nMatrix',
                    alignment: Alignment.center,
                    isHighlighted: true,
                  ),
                ),
                _buildNode(
                  context,
                  label: widget.receiverName ?? 'Spectrum\nAnalyzer',
                  alignment: Alignment.centerRight,
                ),
              ] else if (widget.type == DiagramType.converterComplex) ...[
                _buildNode(
                  context,
                  label: 'Signal\nGenerator 1',
                  alignment: const Alignment(-1.0, -0.6),
                ),
                _buildNode(
                  context,
                  label: 'Signal\nGenerator 2',
                  alignment: const Alignment(-1.0, 0.6),
                ),
                Positioned(
                  left: width * 0.5 - 50,
                  top: height * 0.5 - 30,
                  child: _buildNode(
                    context,
                    label: 'Frequency\nConverter',
                    alignment: Alignment.center,
                    isHighlighted: true,
                  ),
                ),
                _buildNode(
                  context,
                  label: widget.receiverName ?? 'Spectrum\nAnalyzer',
                  alignment: Alignment.centerRight,
                ),
              ] else if (widget.type == DiagramType.converterSimple) ...[
                _buildNode(
                  context,
                  label: 'Frequency\nConverter',
                  alignment: Alignment.centerLeft,
                  isHighlighted: true,
                ),
                _buildNode(
                  context,
                  label: widget.receiverName ?? 'Spectrum\nAnalyzer',
                  alignment: Alignment.centerRight,
                ),
              ],

              // Connection labels
              if (widget.type == DiagramType.pmZero)
                Center(
                  child: _buildEdgeLabel(context, 'Power Sensor', primaryColor),
                )
              else if (widget.type == DiagramType.cableMeasurement) ...[
                Positioned(
                  left: width * 0.25,
                  top: height * 0.5 - 20,
                  width: 80,
                  height: 30, // Fixed size for positioning
                  child: Center(
                    child: _buildEdgeLabel(
                      context,
                      'Cable',
                      Colors.green.shade700,
                    ),
                  ),
                ),
                Positioned(
                  right: width * 0.25,
                  top: height * 0.5 - 20,
                  width: 80,
                  height: 30, // Fixed size for positioning
                  child: Center(
                    child: _buildEdgeLabel(context, 'Sensor', primaryColor),
                  ),
                ),
                // Center connection point
                Center(
                  child: Container(
                    width: 12,
                    height: 12,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      shape: BoxShape.circle,
                      border: Border.all(color: primaryColor, width: 2),
                      boxShadow: [
                        BoxShadow(
                          color: primaryColor.withOpacity(0.2),
                          blurRadius: 4,
                        ),
                      ],
                    ),
                  ),
                ),
              ] else if (widget.type == DiagramType.attnSG ||
                  widget.type == DiagramType.attnGTx)
                Center(child: _buildEdgeLabel(context, 'Cable', primaryColor))
              else if (widget.type == DiagramType.attnTSM) ...[
                Positioned(
                  left: width * 0.5 - 95,
                  top: height * 0.5 - 20,
                  width: 40,
                  child: Center(
                    child: _buildEdgeLabel(
                      context,
                      widget.tsmInputName ?? 'UC',
                      primaryColor,
                    ),
                  ),
                ),
                Positioned(
                  left: width * 0.5 + 55,
                  top: height * 0.5 - 20,
                  width: 40,
                  child: Center(
                    child: _buildEdgeLabel(
                      context,
                      widget.tsmOutputName ?? 'SC',
                      primaryColor,
                    ),
                  ),
                ),
              ] else if (widget.type == DiagramType.converterComplex) ...[
                // Cable Labels (at centers of signal flow)
                Positioned(
                  left: width * 0.25 - 45,
                  top: height * 0.3 - 15,
                  width: 90,
                  child: Center(
                    child: _buildEdgeLabel(
                      context,
                      '${widget.inputPortName ?? "Input"} Cable',
                      primaryColor,
                    ),
                  ),
                ),
                Positioned(
                  left: width * 0.25 - 45,
                  top: height * 0.7 - 15,
                  width: 90,
                  child: Center(
                    child: _buildEdgeLabel(context, 'LO Cable', primaryColor),
                  ),
                ),
                Positioned(
                  right: width * 0.25 - 45,
                  top: height * 0.5 - 15,
                  width: 90,
                  child: Center(
                    child: _buildEdgeLabel(
                      context,
                      '${widget.outputPortName ?? "Output"} Cable',
                      primaryColor,
                    ),
                  ),
                ),

                // Port Identifiers (Near Frequency Converter)
                Positioned(
                  left: width * 0.5 - 125,
                  top: height * 0.4 - 15,
                  width: 70,
                  child: Center(
                    child: _buildEdgeLabel(
                      context,
                      widget.inputPortName ?? 'Input Port',
                      Colors.blueGrey,
                    ),
                  ),
                ),
                Positioned(
                  left: width * 0.5 - 125,
                  top: height * 0.6 - 15,
                  width: 70,
                  child: Center(
                    child: _buildEdgeLabel(
                      context,
                      'Ext LO Port',
                      Colors.blueGrey,
                    ),
                  ),
                ),
                Positioned(
                  left: width * 0.5 + 55,
                  top: height * 0.5 - 15,
                  child: _buildEdgeLabel(
                    context,
                    widget.outputPortName ?? widget.tsmOutputName ?? 'Output',
                    primaryColor,
                  ),
                ),
              ] else if (widget.type == DiagramType.converterSimple) ...[
                // Port Label near converter
                Positioned(
                  left: 105,
                  top: height * 0.5 - 15,
                  child: _buildEdgeLabel(
                    context,
                    widget.inputPortName ?? 'Input Port',
                    primaryColor,
                  ),
                ),
                // Cable label in center
                Center(
                  child: _buildEdgeLabel(
                    context,
                    '${widget.outputPortName ?? "Output"} Cable',
                    primaryColor,
                  ),
                ),
              ],
            ],
          );
        },
      ),
    );
  }

  Widget _buildNode(
    BuildContext context, {
    required String label,
    required Alignment alignment,
    bool isHighlighted = false,
  }) {
    final theme = Theme.of(context);
    final primaryColor = theme.colorScheme.primary;

    return Align(
      alignment: alignment,
      child: Container(
        width: 100,
        height: 60,
        decoration: BoxDecoration(
          color: isHighlighted ? Colors.green.shade700 : primaryColor,
          borderRadius: BorderRadius.circular(8),
          boxShadow: [
            BoxShadow(
              color: primaryColor.withOpacity(0.3),
              blurRadius: 8,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        alignment: Alignment.center,
        child: Text(
          label,
          textAlign: TextAlign.center,
          style: GoogleFonts.inter(
            fontSize: 11,
            fontWeight: FontWeight.bold,
            color: Colors.white,
            height: 1.2,
          ),
        ),
      ),
    );
  }

  Widget _buildEdgeLabel(BuildContext context, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.5)),
        boxShadow: [
          BoxShadow(
            color: color.withOpacity(0.1),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Text(
        label,
        style: GoogleFonts.inter(
          fontSize: 10,
          fontWeight: FontWeight.bold,
          color: color,
        ),
      ),
    );
  }
}

class _ConnectionPainter extends CustomPainter {
  final DiagramType type;
  final Color primaryColor;
  final Color accentColor;
  final double animationValue;

  _ConnectionPainter({
    required this.type,
    required this.primaryColor,
    required this.accentColor,
    required this.animationValue,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final nodeWidth = 100.0;
    final nodeCenterY = size.height / 2;

    if (type == DiagramType.pmZero) {
      final start = Offset(nodeWidth, nodeCenterY);
      final end = Offset(size.width - nodeWidth, nodeCenterY);

      final paint = Paint()
        ..color = primaryColor.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 4
        ..strokeCap = StrokeCap.round;

      canvas.drawLine(start, end, paint);
      _drawMovingDashes(canvas, start, end, primaryColor);
    } else if (type == DiagramType.cableMeasurement) {
      final start = Offset(nodeWidth, nodeCenterY);
      final mid = Offset(size.width / 2, nodeCenterY);
      final end = Offset(size.width - nodeWidth, nodeCenterY);

      final cablePaint = Paint()
        ..color = Colors.green.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 4
        ..strokeCap = StrokeCap.round;

      final sensorPaint = Paint()
        ..color = primaryColor.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 4
        ..strokeCap = StrokeCap.round;

      canvas.drawLine(start, mid, cablePaint);
      canvas.drawLine(mid, end, sensorPaint);

      _drawMovingDashes(canvas, start, mid, Colors.green.shade700);
      _drawMovingDashes(canvas, mid, end, primaryColor);
    } else if (type == DiagramType.attnTSM) {
      final start = Offset(nodeWidth, nodeCenterY);
      final mid = Offset(size.width / 2, nodeCenterY);
      final end = Offset(size.width - nodeWidth, nodeCenterY);

      final paint = Paint()
        ..color = primaryColor.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 4
        ..strokeCap = StrokeCap.round;

      canvas.drawLine(start, mid, paint);
      canvas.drawLine(mid, end, paint);

      _drawMovingDashes(canvas, start, mid, primaryColor);
      _drawMovingDashes(canvas, mid, end, primaryColor);
    } else if (type == DiagramType.converterComplex) {
      final sg1 = Offset(nodeWidth, size.height * 0.2);
      final sg2 = Offset(nodeWidth, size.height * 0.8);
      final mid = Offset(size.width / 2, nodeCenterY);
      final end = Offset(size.width - nodeWidth, nodeCenterY);

      final paint = Paint()
        ..color = primaryColor.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 4
        ..strokeCap = StrokeCap.round;

      canvas.drawLine(sg1, mid, paint);
      canvas.drawLine(sg2, mid, paint);
      canvas.drawLine(mid, end, paint);

      _drawMovingDashes(canvas, sg1, mid, primaryColor);
      _drawMovingDashes(canvas, sg2, mid, primaryColor);
      _drawMovingDashes(canvas, mid, end, primaryColor);
    } else if (type == DiagramType.converterSimple) {
      final start = Offset(nodeWidth, nodeCenterY);
      final end = Offset(size.width - nodeWidth, nodeCenterY);

      final paint = Paint()
        ..color = primaryColor.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 4
        ..strokeCap = StrokeCap.round;

      canvas.drawLine(start, end, paint);
      _drawMovingDashes(canvas, start, end, primaryColor);
    } else {
      // SG or GTx (Default)
      final start = Offset(nodeWidth, nodeCenterY);
      final end = Offset(size.width - nodeWidth, nodeCenterY);

      final paint = Paint()
        ..color = primaryColor.withOpacity(0.2)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 4
        ..strokeCap = StrokeCap.round;

      canvas.drawLine(start, end, paint);
      _drawMovingDashes(canvas, start, end, primaryColor);
    }
  }

  void _drawMovingDashes(Canvas canvas, Offset start, Offset end, Color color) {
    final dashPaint = Paint()
      ..color = color.withOpacity(0.6)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2
      ..strokeCap = StrokeCap.round;

    final distance = (end - start).distance;
    final dashValues = [10.0, 10.0]; // dash, space
    final totalDashLength = dashValues[0] + dashValues[1];

    double currentDist = (animationValue * totalDashLength) % totalDashLength;

    while (currentDist < distance) {
      final dashStart = Offset.lerp(start, end, currentDist / distance)!;
      final dashEnd = Offset.lerp(
        start,
        end,
        (currentDist + dashValues[0]) / distance,
      )!;

      if ((currentDist + dashValues[0]) <= distance) {
        canvas.drawLine(dashStart, dashEnd, dashPaint);
      }
      currentDist += totalDashLength;
    }
  }

  @override
  bool shouldRepaint(covariant _ConnectionPainter oldDelegate) =>
      oldDelegate.animationValue != animationValue || oldDelegate.type != type;
}
