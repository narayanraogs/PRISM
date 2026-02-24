import 'dart:convert';
import 'dart:typed_data';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';

class SpectrumDumpViewScreen extends StatefulWidget {
  final String base64Image;
  final String instrumentName;

  const SpectrumDumpViewScreen({
    super.key,
    required this.base64Image,
    required this.instrumentName,
  });

  @override
  State<SpectrumDumpViewScreen> createState() => _SpectrumDumpViewScreenState();
}

class DrawingPoint {
  final Offset point;
  final Paint paint;

  DrawingPoint({required this.point, required this.paint});
}

class _SpectrumDumpViewScreenState extends State<SpectrumDumpViewScreen> {
  late Uint8List _imageBytes;
  final List<List<DrawingPoint>> _drawnLines = [];
  Color _selectedColor = Colors.red;
  double _strokeWidth = 3.0;
  final GlobalKey _repaintKey = GlobalKey();
  final TextEditingController _remarkController = TextEditingController();
  bool _isSaving = false;

  @override
  void initState() {
    super.initState();
    try {
      String cleanBase64 = widget.base64Image;
      if (cleanBase64.contains(',')) {
        cleanBase64 = cleanBase64.split(',').last;
      }
      cleanBase64 = cleanBase64.replaceAll(RegExp(r'\s+'), '');
      int padding = cleanBase64.length % 4;
      if (padding != 0) {
        cleanBase64 += '=' * (4 - padding);
      }
      _imageBytes = base64Decode(cleanBase64);
    } catch (e) {
      debugPrint('Error decoding base64 image: $e');
      _imageBytes = Uint8List(0);
    }
  }

  @override
  void dispose() {
    _remarkController.dispose();
    super.dispose();
  }

  Future<void> _saveAndClose() async {
    setState(() => _isSaving = true);

    try {
      // 1. Capture the annotated image
      final boundary =
          _repaintKey.currentContext?.findRenderObject()
              as RenderRepaintBoundary?;
      if (boundary == null) throw Exception('Could not capture image');

      final image = await boundary.toImage(pixelRatio: 2.0);
      final byteData = await image.toByteData(format: ui.ImageByteFormat.png);
      if (byteData == null) throw Exception('Could not convert image to data');

      final pngBytes = byteData.buffer.asUint8List();
      final base64Image = base64Encode(pngBytes);

      // 2. Call server to save
      if (mounted) {
        final service = Provider.of<ServerService>(context, listen: false);
        final notificationService = Provider.of<NotificationService>(
          context,
          listen: false,
        );

        final result = await service.saveSpectrum(
          spectrumBase64: base64Image,
          remark: _remarkController.text,
        );

        if (mounted) {
          if (result != null && result.ok) {
            notificationService.addNotification(
              title: 'Spectrum Saved',
              message: result.message,
              type: NotificationType.success,
            );
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(result.message),
                backgroundColor: Colors.green,
              ),
            );
            Navigator.pop(context);
          } else {
            notificationService.addNotification(
              title: 'Save Failed',
              message: result?.message ?? 'Unknown error',
              type: NotificationType.error,
            );
            showDialog(
              context: context,
              builder: (context) => AlertDialog(
                title: const Text('Error'),
                content: Text(result?.message ?? 'Failed to save spectrum'),
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
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isSaving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade100,
      appBar: AppBar(
        title: Text(
          'Spectrum Dump: ${widget.instrumentName}',
          style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.undo_rounded),
            onPressed: () {
              if (_drawnLines.isNotEmpty) {
                setState(() => _drawnLines.removeLast());
              }
            },
            tooltip: 'Undo',
          ),
          IconButton(
            icon: const Icon(Icons.delete_sweep_rounded),
            onPressed: () {
              setState(() => _drawnLines.clear());
            },
            tooltip: 'Clear All',
          ),
          const SizedBox(width: 8),
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 8.0, horizontal: 16),
            child: ElevatedButton.icon(
              onPressed: _isSaving ? null : _saveAndClose,
              icon: _isSaving
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : const Icon(Icons.check_rounded),
              label: Text(_isSaving ? 'SAVING...' : 'SAVE & CLOSE'),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          // Toolbar
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
            color: Colors.white,
            child: Row(
              children: [
                _buildToolItem(
                  'COLOR',
                  Row(
                    children: [
                      _colorCircle(Colors.red),
                      _colorCircle(Colors.green),
                      _colorCircle(Colors.blue),
                      _colorCircle(Colors.yellow),
                      _colorCircle(Colors.black),
                    ],
                  ),
                ),
                const VerticalDivider(width: 40),
                _buildToolItem(
                  'SIZE',
                  Row(
                    children: [
                      _sizeButton(2.0, 'Small'),
                      _sizeButton(4.0, 'Medium'),
                      _sizeButton(8.0, 'Large'),
                    ],
                  ),
                ),
                const VerticalDivider(width: 40),
                Expanded(
                  child: _buildToolItem(
                    'REMARK',
                    TextField(
                      controller: _remarkController,
                      decoration: InputDecoration(
                        hintText: 'Add a remark for this capture...',
                        hintStyle: TextStyle(
                          fontSize: 13,
                          color: Colors.grey.shade400,
                        ),
                        isDense: true,
                        contentPadding: const EdgeInsets.symmetric(
                          vertical: 10,
                          horizontal: 12,
                        ),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                          borderSide: BorderSide(color: Colors.grey.shade200),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(8),
                          borderSide: BorderSide(color: Colors.grey.shade200),
                        ),
                      ),
                      style: const TextStyle(fontSize: 13),
                    ),
                  ),
                ),
              ],
            ),
          ),
          // Drawing Area
          Expanded(
            child: Center(
              child: Padding(
                padding: const EdgeInsets.all(32.0),
                child: Container(
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.1),
                        blurRadius: 20,
                        offset: const Offset(0, 10),
                      ),
                    ],
                  ),
                  clipBehavior: Clip.antiAlias,
                  child: AspectRatio(
                    aspectRatio: 16 / 9, // Adjust if needed
                    child: RepaintBoundary(
                      key: _repaintKey,
                      child: Stack(
                        children: [
                          // The Image
                          Positioned.fill(
                            child: _imageBytes.isEmpty
                                ? const Center(
                                    child: Text(
                                      'Failed to decode image data',
                                      style: TextStyle(color: Colors.red),
                                    ),
                                  )
                                : Image.memory(
                                    _imageBytes,
                                    fit: BoxFit.contain,
                                    gaplessPlayback: true,
                                    errorBuilder: (context, error, stackTrace) {
                                      return Center(
                                        child: Column(
                                          mainAxisAlignment: MainAxisAlignment.center,
                                          children: [
                                            const Icon(Icons.broken_image, size: 48, color: Colors.grey),
                                            const SizedBox(height: 8),
                                            Text(
                                              'Image render error:\n$error',
                                              textAlign: TextAlign.center,
                                              style: const TextStyle(color: Colors.red),
                                            ),
                                          ],
                                        ),
                                      );
                                    },
                                  ),
                          ),
                          // The Drawing Layer
                          Positioned.fill(
                            child: GestureDetector(
                              onPanStart: (details) {
                                setState(() {
                                  final point = DrawingPoint(
                                    point: details.localPosition,
                                    paint: Paint()
                                      ..color = _selectedColor
                                      ..strokeCap = StrokeCap.round
                                      ..strokeWidth = _strokeWidth,
                                  );
                                  _drawnLines.add([point]);
                                });
                              },
                              onPanUpdate: (details) {
                                setState(() {
                                  final point = DrawingPoint(
                                    point: details.localPosition,
                                    paint: Paint()
                                      ..color = _selectedColor
                                      ..strokeCap = StrokeCap.round
                                      ..strokeWidth = _strokeWidth,
                                  );
                                  _drawnLines.last.add(point);
                                });
                              },
                              child: CustomPaint(
                                painter: DrawingPainter(lines: _drawnLines),
                                size: Size.infinite,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildToolItem(String label, Widget content) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
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
        const SizedBox(height: 8),
        content,
      ],
    );
  }

  Widget _colorCircle(Color color) {
    final isSelected = _selectedColor == color;
    return GestureDetector(
      onTap: () => setState(() => _selectedColor = color),
      child: Container(
        margin: const EdgeInsets.only(right: 8),
        width: 28,
        height: 28,
        decoration: BoxDecoration(
          color: color,
          shape: BoxShape.circle,
          border: Border.all(
            color: isSelected ? Colors.black : Colors.grey.shade300,
            width: isSelected ? 2 : 1,
          ),
        ),
        child: isSelected
            ? const Icon(Icons.check, color: Colors.white, size: 16)
            : null,
      ),
    );
  }

  Widget _sizeButton(double size, String label) {
    final isSelected = _strokeWidth == size;
    return GestureDetector(
      onTap: () => setState(() => _strokeWidth = size),
      child: Container(
        margin: const EdgeInsets.only(right: 8),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: isSelected ? Colors.blue.withOpacity(0.1) : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: isSelected ? Colors.blue : Colors.grey.shade300,
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? Colors.blue : Colors.black,
            fontSize: 12,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
  }
}

class DrawingPainter extends CustomPainter {
  final List<List<DrawingPoint>> lines;

  DrawingPainter({required this.lines});

  @override
  void paint(Canvas canvas, Size size) {
    for (final line in lines) {
      for (int i = 0; i < line.length - 1; i++) {
        canvas.drawLine(line[i].point, line[i + 1].point, line[i].paint);
      }
    }
  }

  @override
  bool shouldRepaint(covariant DrawingPainter oldDelegate) => true;
}
