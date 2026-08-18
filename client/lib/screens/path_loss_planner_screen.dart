
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/utils/notifications.dart';
import 'package:prism_client/utils/svg_exporter.dart';
import 'package:prism_client/widgets/csv_export_dialog.dart';

import 'package:web/web.dart' as web;

enum NodeType { source, component, branching, instrument, hub, converter }

enum NodeDirection { up, down, center }

class PlannerNode {
  final String id;
  String label;
  double lossDb;
  NodeType type;
  NodeDirection direction;
  String? calibratedCableId;
  double loOffset;
  double sourceFrequency;
  List<PlannerNode> children;

  PlannerNode({
    required this.id,
    required this.label,
    this.lossDb = 0.0,
    required this.type,
    this.direction = NodeDirection.down,
    this.calibratedCableId,
    this.loOffset = 0.0,
    this.sourceFrequency = 1000.0,
    List<PlannerNode>? children,
  }) : children = children ?? [];

  Map<String, dynamic> toJson() => {
    'id': id,
    'label': label,
    'lossDb': lossDb,
    'type': type.index,
    'direction': direction.index,
    'calibratedCableId': calibratedCableId,
    'loOffset': loOffset,
    'sourceFrequency': sourceFrequency,
    'children': children.map((c) => c.toJson()).toList(),
  };

  factory PlannerNode.fromJson(Map<String, dynamic> json) {
    return PlannerNode(
      id: json['id'],
      label: json['label'],
      lossDb: (json['lossDb'] as num).toDouble(),
      type: NodeType.values[json['type']],
      direction: NodeDirection.values[json['direction']],
      calibratedCableId: json['calibratedCableId'],
      loOffset: (json['loOffset'] as num?)?.toDouble() ?? 0.0,
      sourceFrequency: (json['sourceFrequency'] as num?)?.toDouble() ?? 1000.0,
      children: (json['children'] as List?)
          ?.map((c) => PlannerNode.fromJson(c))
          .toList(),
    );
  }
}

class SolveResult {
  final double totalLoss;
  final double finalPower;
  final List<PathStep> steps;
  SolveResult({
    required this.totalLoss,
    required this.finalPower,
    required this.steps,
  });
}

class PathStep {
  final String nodeId;
  final String label;
  final double frequency;
  final double loss;
  final double outputFrequency;
  final double inputPower;
  final double outputPower;

  PathStep({
    required this.nodeId,
    required this.label,
    required this.frequency,
    required this.loss,
    required this.outputFrequency,
    required this.inputPower,
    required this.outputPower,
  });
}

class PathLossPlannerScreen extends StatefulWidget {
  final bool isActive;
  const PathLossPlannerScreen({super.key, this.isActive = false});

  @override
  State<PathLossPlannerScreen> createState() => _PathLossPlannerScreenState();
}

class _PathLossPlannerScreenState extends State<PathLossPlannerScreen> {
  PlannerNode? _hubNode;
  String? _startNodeId;
  String? _endNodeId;
  List<String> _calibratedCables = [];
  TSMInternalLossMetadata? _tsmData;
  final TransformationController _transformationController =
      TransformationController();
  List<CableLossRecord> _allCalibratedRecords = [];
  List<String> _plannerList = [];
  String? _selectedPlanner = 'plannerState';
  final ScrollController _summaryVerticalScroll = ScrollController();
  final ScrollController _summaryHorizontalScroll = ScrollController();

  @override
  void initState() {
    super.initState();
    _loadMetadata();
  }

  @override
  void dispose() {
    _transformationController.dispose();
    _summaryVerticalScroll.dispose();
    _summaryHorizontalScroll.dispose();
    super.dispose();
  }

  void _loadMetadata() {
    final server = context.read<ServerService>();
    final bootstrap = server.status.bootstrapData;
    if (bootstrap != null) {
      setState(() {
        _calibratedCables = bootstrap.cableLossData.existingCables;
        _tsmData = bootstrap.tsmInternalLossData;
        _plannerList = bootstrap.plannerList;
        if (_plannerList.isEmpty) {
          _plannerList = ['plannerState'];
        }
        if (!_plannerList.contains(_selectedPlanner)) {
          _selectedPlanner = _plannerList.first;
        }

        server.fetchCableMeasuredDetails().then((resp) {
          if (resp != null && resp.ok) {
            setState(() => _allCalibratedRecords = resp.history);
          }
        });

        if (bootstrap.plannerData.isNotEmpty) {
          try {
            _hubNode = PlannerNode.fromJson(jsonDecode(bootstrap.plannerData));
          } catch (e) {
            _initializeHub();
          }
        } else {
          _initializeHub();
        }
      });
    }
  }

  Future<void> _loadPlanner(String name) async {
    final server = context.read<ServerService>();
    final data = await server.loadPlannerData(name);
    if (!mounted) return;
    if (data != null) {
      setState(() {
        if (data.isEmpty) {
          _hubNode = null;
          _initializeHub();
        } else {
          try {
            _hubNode = PlannerNode.fromJson(jsonDecode(data));
          } catch (e) {
            AppNotifications.showError(context, 'Failed to Load: $name');
            return;
          }
        }
        _selectedPlanner = name;
        _startNodeId = null;
        _endNodeId = null;
      });
      AppNotifications.showSuccess(context, 'Loaded: $name');
    }
  }

  Future<void> _saveDiagram({String? name}) async {
    if (_hubNode == null) return;
    final server = context.read<ServerService>();
    final saveName = name ?? _selectedPlanner ?? 'plannerState';
    final success = await server.savePlannerData(
      jsonEncode(_hubNode!.toJson()),
      name: saveName,
    );
    if (mounted) {
      if (success) {
        setState(() {
          _selectedPlanner = saveName;
          if (!_plannerList.contains(saveName)) {
            _plannerList.add(saveName);
          }
        });
        AppNotifications.show(
          context,
          'Configuration Saved: $saveName',
          type: NotificationType.success,
        );
      } else {
        AppNotifications.showError(context, 'Failed to Save Configuration');
      }
    }
  }

  void _initializeHub() {
    if (_hubNode != null) return;
    _hubNode = PlannerNode(
      id: 'tsm-hub',
      label: 'TSM Central Hub',
      type: NodeType.hub,
      direction: NodeDirection.center,
    );
    if (_tsmData != null && _tsmData!.ok) {
      final leftPorts = <String>{};
      final rightPorts = <String>{};
      for (var p in _tsmData!.measuredLoss.paths) {
        if (p.inputPort.isNotEmpty) {
          leftPorts.add(p.inputPort.replaceAll('-WithPad', ''));
        }
        if (p.outputPort.isNotEmpty) {
          rightPorts.add(p.outputPort.replaceAll('-WithPad', ''));
        }
      }
      for (var port in leftPorts) {
        _hubNode!.children.add(
          PlannerNode(
            id: 'port-$port',
            label: port,
            type: NodeType.branching,
            direction: NodeDirection.up,
          ),
        );
      }
      for (var port in rightPorts) {
        if (leftPorts.contains(port)) continue;
        _hubNode!.children.add(
          PlannerNode(
            id: 'port-$port',
            label: port,
            type: NodeType.branching,
            direction: NodeDirection.down,
          ),
        );
      }
    }
  }

  PlannerNode get _root => _hubNode!;

  void _addNode(PlannerNode parent, NodeType type, {bool insert = false}) {
    setState(() {
      final newNode = PlannerNode(
        id: 'node-${DateTime.now().millisecondsSinceEpoch}',
        label: 'New ${type.name}',
        type: type,
        direction: parent.direction,
        lossDb: type == NodeType.branching ? 1.0 : 0.0,
      );
      if (insert) {
        newNode.children.addAll(parent.children);
        parent.children.clear();
      }
      parent.children.add(newNode);
    });
  }

  void _deleteNode(PlannerNode parent, PlannerNode target) {
    setState(() {
      final idx = parent.children.indexOf(target);
      if (idx != -1) {
        parent.children.removeAt(idx);
        parent.children.insertAll(idx, target.children);
      }
    });
  }

  PlannerNode? _findParent(PlannerNode current, PlannerNode target) {
    for (var child in current.children) {
      if (child == target) return current;
      final found = _findParent(child, target);
      if (found != null) return found;
    }
    return null;
  }

  PlannerNode? _findHubChildOf(PlannerNode target) {
    if (_hubNode == null) return null;
    PlannerNode? curr = target;
    while (curr != null) {
      final p = _findParent(_root, curr);
      if (p == _hubNode) return curr;
      curr = p;
    }
    return null;
  }

  List<PlannerNode> _getTerminals(NodeType type) {
    final results = <PlannerNode>[];
    if (_hubNode == null) return results;
    void traverse(PlannerNode node) {
      if (node.type == type) results.add(node);
      for (var c in node.children) {
        traverse(c);
      }
    }

    for (var c in _hubNode!.children) {
      traverse(c);
    }
    return results;
  }

  double _interpolateCableLoss(String cableName, double freqMHz) {
    if (_allCalibratedRecords.isEmpty) return 0.0;
    final records =
        _allCalibratedRecords.where((r) => r.cableName == cableName).toList()
          ..sort((a, b) => b.slNo.compareTo(a.slNo));
    if (records.isEmpty || records.first.measurements.isEmpty) return 0.0;
    final sorted = List<MeasurementPoint>.from(records.first.measurements)
      ..sort((a, b) => a.frequency.compareTo(b.frequency));
    double res = 0.0;
    if (freqMHz <= sorted.first.frequency) {
      res = sorted.first.loss;
    } else if (freqMHz >= sorted.last.frequency) {
      res = sorted.last.loss;
    } else {
      for (int i = 0; i < sorted.length - 1; i++) {
        final p1 = sorted[i], p2 = sorted[i + 1];
        if (freqMHz >= p1.frequency && freqMHz <= p2.frequency) {
          final t = (freqMHz - p1.frequency) / (p2.frequency - p1.frequency);
          res = p1.loss + t * (p2.loss - p1.loss);
          break;
        }
      }
    }
    return res < 0 ? res * -1.0 : res;
  }

  SolveResult _solvePath() {
    return _calculatePath(_startNodeId, _endNodeId);
  }

  SolveResult _calculatePath(String? startId, String? endId) {
    if (startId == null || endId == null) {
      return SolveResult(totalLoss: 0.0, finalPower: 0.0, steps: []);
    }

    final sources = _getTerminals(NodeType.source);
    final sinks = _getTerminals(NodeType.instrument);

    // Find nodes or fallback to avoid crash
    PlannerNode? startNode, endNode;
    try {
      startNode = sources.firstWhere((t) => t.id == startId);
    } catch (_) {}
    try {
      endNode = sinks.firstWhere((t) => t.id == endId);
    } catch (_) {}
    if (startNode == null || endNode == null) {
      return SolveResult(totalLoss: 0.0, finalPower: 0.0, steps: []);
    }

    double currentFreq = startNode.sourceFrequency;
    double currentPower = startNode.lossDb;
    List<PathStep> steps = [];
    double totalLoss = 0.0;

    // Upstream (Source -> Hub Child)
    List<PlannerNode> upstream = [];
    PlannerNode? cu = startNode;
    while (cu != null && cu != _hubNode) {
      upstream.add(cu);
      cu = _findParent(_root, cu);
    }
    // Traverse from Source to HubChild
    for (var node in upstream) {
      if (node.type == NodeType.source) {
        steps.add(
          PathStep(
            nodeId: node.id,
            label: node.label,
            frequency: currentFreq,
            loss: 0.0,
            outputFrequency: currentFreq,
            inputPower: currentPower,
            outputPower: currentPower,
          ),
        );
      } else {
        double loss = node.lossDb;
        if (node.calibratedCableId != null) {
          loss = _interpolateCableLoss(node.calibratedCableId!, currentFreq);
        }
        double outFreq = (currentFreq + node.loOffset).abs();
        steps.add(
          PathStep(
            nodeId: node.id,
            label: node.label,
            frequency: currentFreq,
            loss: loss,
            outputFrequency: outFreq,
            inputPower: currentPower,
            outputPower: currentPower - loss,
          ),
        );
        totalLoss += loss;
        currentFreq = outFreq;
        currentPower -= loss;
      }
    }

    // Hub
    final startChild = _findHubChildOf(startNode);
    final endChild = _findHubChildOf(endNode);
    double hubLoss = 0.0;
    if (startChild != null && endChild != null && _tsmData != null) {
      for (var p in _tsmData!.measuredLoss.paths) {
        if (p.inputPort.replaceAll('-WithPad', '') == startChild.label &&
            p.outputPort.replaceAll('-WithPad', '') == endChild.label &&
            p.losses.isNotEmpty) {
          int cIdx = 0;
          double mDiff = double.infinity;
          for (int i = 0; i < p.frequencies.length; i++) {
            double diff = (p.frequencies[i] - currentFreq).abs();
            if (diff < mDiff) {
              mDiff = diff;
              cIdx = i;
            }
          }
          hubLoss = p.losses[cIdx];
          if (hubLoss < 0) hubLoss *= -1.0;
          break;
        }
      }
    }
    steps.add(
      PathStep(
        nodeId: 'hub-${startChild?.id}-${endChild?.id}',
        label:
            'TSM Hub (${startChild?.label ?? '?'} → ${endChild?.label ?? '?'})',
        frequency: currentFreq,
        loss: hubLoss,
        outputFrequency: currentFreq,
        inputPower: currentPower,
        outputPower: currentPower - hubLoss,
      ),
    );
    totalLoss += hubLoss;
    currentPower -= hubLoss;

    // Downstream (Hub Child -> Sink)
    List<PlannerNode> downstream = [];
    PlannerNode? cd = endNode;
    while (cd != null && cd != _hubNode) {
      downstream.add(cd);
      cd = _findParent(_root, cd);
    }
    // Traverse from HubChild down to Sink
    for (var node in downstream.reversed) {
      double loss = node.lossDb;
      if (node.calibratedCableId != null) {
        loss = _interpolateCableLoss(node.calibratedCableId!, currentFreq);
      }
      double outFreq = (currentFreq + node.loOffset).abs();
      steps.add(
        PathStep(
          nodeId: node.id,
          label: node.label,
          frequency: currentFreq,
          loss: loss,
          outputFrequency: outFreq,
          inputPower: currentPower,
          outputPower: currentPower - loss,
        ),
      );
      totalLoss += loss;
      currentFreq = outFreq;
      currentPower -= loss;
    }

    return SolveResult(
      totalLoss: totalLoss,
      finalPower: currentPower,
      steps: steps,
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Column(
        children: [
          _buildHeader(),
          Expanded(
            child: Row(
              children: [
                Expanded(flex: 3, child: _buildSchematicCanvas(theme)),
                Flexible(
                  flex: 1,
                  child: Container(
                    constraints: const BoxConstraints(minWidth: 400, maxWidth: 600),
                    padding: const EdgeInsets.all(24),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      border: Border(
                        left: BorderSide(color: Colors.grey.shade100),
                      ),
                    ),
                    child: _buildSummaryPanel(theme),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader() {
    return ScreenHeader(
      title: 'Path Loss Planner',
      subtitle: 'Visualize and calculate RF link budgets',
      icon: Icons.map_outlined,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            height: 36,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.grey.shade300),
            ),
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                value: _selectedPlanner,
                hint: const Text('Select Version'),
                style: GoogleFonts.inter(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: Colors.grey.shade800,
                ),
                items: _plannerList.map((String value) {
                  return DropdownMenuItem<String>(
                    value: value,
                    child: Text(value),
                  );
                }).toList(),
                onChanged: (val) {
                  if (val != null) _loadPlanner(val);
                },
              ),
            ),
          ),
          const SizedBox(width: 8),
          ElevatedButton.icon(
            onPressed: () {
              setState(() {
                _hubNode = null;
                _initializeHub();
                _transformationController.value = Matrix4.identity();
              });
              AppNotifications.show(
                context,
                'Planner has been reset to defaults',
              );
            },
            icon: const Icon(Icons.restart_alt),
            label: const Text('RESET'),
          ),
          const SizedBox(width: 8),
          ElevatedButton.icon(
            onPressed: () => _saveDiagram(),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.green.shade600,
              foregroundColor: Colors.white,
            ),
            icon: const Icon(Icons.save_outlined, size: 16),
            label: const Text('SAVE'),
          ),
          const SizedBox(width: 8),
          ElevatedButton.icon(
            onPressed: () {
              final controller = TextEditingController();
              showDialog(
                context: context,
                builder: (context) => AlertDialog(
                  title: const Text('Save Configuration As'),
                  content: TextField(
                    controller: controller,
                    decoration: const InputDecoration(
                      labelText: 'Configuration Name',
                      hintText: 'e.g. TestVariant_1',
                    ),
                  ),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.pop(context),
                      child: const Text('CANCEL'),
                    ),
                    ElevatedButton(
                      onPressed: () {
                        if (controller.text.isNotEmpty) {
                          _saveDiagram(name: controller.text);
                          Navigator.pop(context);
                        }
                      },
                      child: const Text('SAVE'),
                    ),
                  ],
                ),
              );
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.blue.shade600,
              foregroundColor: Colors.white,
            ),
            icon: const Icon(Icons.save_as_outlined, size: 16),
            label: const Text('SAVE AS'),
          ),
          const SizedBox(width: 8),
          ElevatedButton.icon(
            onPressed: () {
              if (_hubNode == null) return;
              try {
                final svgString = SvgExporter.generate(_hubNode!);
                final rawData = utf8.encode(svgString);
                final content = base64Encode(rawData);
                web.HTMLAnchorElement()
                  ..href = 'data:image/svg+xml;base64,$content'
                  ..download = 'PathLossDiagram_${DateTime.now().millisecondsSinceEpoch}.svg'
                  ..click();
                AppNotifications.showSuccess(
                  context,
                  'SVG layout exported successfully',
                );
              } catch (e) {
                AppNotifications.showError(
                  context,
                  'SVG layout export failed: $e',
                );
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.indigo.shade600,
              foregroundColor: Colors.white,
            ),
            icon: const Icon(Icons.download_rounded, size: 16),
            label: const Text('EXPORT SVG'),
          ),
          const SizedBox(width: 8),
          ElevatedButton.icon(
            onPressed: () {
              final sources = _getTerminals(NodeType.source);
              final sinks = _getTerminals(NodeType.instrument);
              showDialog(
                context: context,
                builder: (context) => CsvExportDialog(
                  sources: sources,
                  sinks: sinks,
                  onExport: (srcId, saId, pmId, scId) {
                    _exportToCsv(srcId, saId, pmId, scId);
                  },
                ),
              );
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.teal.shade600,
              foregroundColor: Colors.white,
            ),
            icon: const Icon(Icons.table_view_outlined, size: 16),
            label: const Text('EXPORT CSV'),
          ),
        ],
      ),
    );
  }

  void _exportToCsv(String srcId, String saId, String pmId, String? scId) {
    try {
      final saRes = _calculatePath(srcId, saId);
      final pmRes = _calculatePath(srcId, pmId);
      final scRes = scId != null && scId.isNotEmpty ? _calculatePath(srcId, scId) : null;

      final allSinks = [
        'SA',
        'PM',
        if (scRes != null) 'SC',
      ];

      final bool isUplink = scId != null && scId.isNotEmpty;

      // Step ID to Metadata (label, loss, etc)
      final Map<String, Map<String, dynamic>> stepData = {};
      // List to maintain unique node ordering
      final List<String> orderedNodeIds = [];

      void processPath(List<PathStep> steps, String type) {
        for (var step in steps) {
          if (!stepData.containsKey(step.nodeId)) {
            stepData[step.nodeId] = {
              'label': step.label,
              'loss': step.loss,
              'inputPower': step.inputPower,
              'types': <String>{},
            };
            orderedNodeIds.add(step.nodeId);
          }
          (stepData[step.nodeId]!['types'] as Set<String>).add(type);
          
          // Optimization: If a component is shared but has different losses (unlikely in common segments),
          // we stick to the first one discovered or could average. Here we stick to first.
        }
      }

      processPath(saRes.steps, 'SA');
      processPath(pmRes.steps, 'PM');
      if (scRes != null) {
        processPath(scRes.steps, 'SC');
      }

      // Build CSV content
      final buffer = StringBuffer();
      buffer.writeln('Sl. No, Description, Loss, Type');

      int slNo = 1;
      for (var nodeId in orderedNodeIds) {
        final data = stepData[nodeId]!;
        final double lossVal = data['loss'] as double;
        final bool isSource = nodeId == srcId;

        if (isSource) {
          if (!isUplink) continue;
        } else {
          if (lossVal == 0) continue;
        }

        final double displayVal = isSource ? -(data['inputPower'] as double) : lossVal;

        final Set<String> types = data['types'];
        
        String typeLabel;
        if (types.length == allSinks.length) {
          typeLabel = 'Common';
        } else if (types.length == 1) {
          typeLabel = types.first;
        } else {
          // Combination like "SA, PM"
          typeLabel = (types.toList()..sort()).join(', ');
        }

        final label = data['label'].toString().replaceAll('"', '""');
        final loss = displayVal.toStringAsFixed(2);

        buffer.writeln('$slNo, "$label", $loss, $typeLabel');
        slNo++;
      }

      // Filename logic
      String filename = 'Export';
      final sources = _getTerminals(NodeType.source);
      final sinks = _getTerminals(NodeType.instrument);
      
      if (scId != null && scId.isNotEmpty) {
        try {
          filename = sinks.firstWhere((s) => s.id == scId).label;
        } catch (_) {}
      } else {
        try {
          filename = sources.firstWhere((s) => s.id == srcId).label;
        } catch (_) {}
      }

      // Download
      final bytes = utf8.encode(buffer.toString());
      final base64 = base64Encode(bytes);
      web.HTMLAnchorElement()
        ..href = 'data:text/csv;base64,$base64'
        ..download = '$filename.csv'
        ..click();

      AppNotifications.showSuccess(context, 'CSV Exported Successfully');
    } catch (e) {
      AppNotifications.showError(context, 'Failed to export CSV: $e');
    }
  }

  Widget _buildSchematicCanvas(ThemeData theme) {
    return Stack(
      children: [
        Container(
          color: Colors.grey.shade50,
          child: InteractiveViewer(
            transformationController: _transformationController,
            constrained: false,
            boundaryMargin: const EdgeInsets.all(100),
            minScale: 0.1,
            maxScale: 2.5,
            child: SizedBox(
              width: 5000,
              height: 5000,
              child: Stack(
                children: [
                  Positioned.fill(
                    child: CustomPaint(
                      painter: GridPainter(
                        gridSize: 40,
                        color: Colors.grey.shade200,
                      ),
                    ),
                  ),
                  Positioned(
                    left: 2000,
                    top: 2000,
                    child: _hubNode != null
                        ? _buildStarNodeTree(_hubNode!, theme)
                        : const SizedBox(),
                  ),
                ],
              ),
            ),
          ),
        ),
        Positioned(
          bottom: 24,
          left: 24,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.9),
              borderRadius: BorderRadius.circular(30),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 10,
                ),
              ],
            ),
            child: Row(
              children: [
                Icon(
                  Icons.zoom_in_map_outlined,
                  size: 16,
                  color: theme.colorScheme.primary,
                ),
                const SizedBox(width: 12),
                Text(
                  'Pinch to Zoom • Drag to Pan',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: Colors.grey.shade700,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildStarNodeTree(PlannerNode node, ThemeData theme) {
    final topChildren = node.children
        .where((c) => c.direction == NodeDirection.up)
        .toList();
    final bottomChildren = node.children
        .where((c) => c.direction == NodeDirection.down)
        .toList();
    final lineColor = theme.colorScheme.primary.withValues(alpha: 0.3);

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        if (topChildren.isNotEmpty) ...[
          IntrinsicWidth(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: topChildren
                      .map(
                        (c) => Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Padding(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 20,
                              ),
                              child: _buildDirectionalTree(c, theme, true),
                            ),
                            Container(width: 2, height: 30, color: lineColor),
                          ],
                        ),
                      )
                      .toList(),
                ),
                Divider(color: lineColor, thickness: 2, height: 2),
              ],
            ),
          ),
          Container(width: 2, height: 30, color: lineColor),
        ],
        _buildNodeCard(node, theme),
        if (bottomChildren.isNotEmpty) ...[
          Container(width: 2, height: 30, color: lineColor),
          IntrinsicWidth(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Divider(color: lineColor, thickness: 2, height: 2),
                Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: bottomChildren
                      .map(
                        (c) => Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(width: 2, height: 30, color: lineColor),
                            Padding(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 20,
                              ),
                              child: _buildDirectionalTree(c, theme, false),
                            ),
                          ],
                        ),
                      )
                      .toList(),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildDirectionalTree(
    PlannerNode node,
    ThemeData theme,
    bool isUpward,
  ) {
    if (node.children.isEmpty) return _buildNodeCard(node, theme);
    final lineColor = theme.colorScheme.primary.withValues(alpha: 0.3);

    final childrenWidget = node.children.length == 1
        ? Column(
            mainAxisSize: MainAxisSize.min,
            children: isUpward
                ? [
                    _buildDirectionalTree(node.children.first, theme, isUpward),
                    Container(width: 2, height: 30, color: lineColor),
                  ]
                : [
                    Container(width: 2, height: 30, color: lineColor),
                    _buildDirectionalTree(node.children.first, theme, isUpward),
                  ],
          )
        : IntrinsicWidth(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: isUpward
                  ? [
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: node.children
                            .map(
                              (c) => Column(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Padding(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 20,
                                    ),
                                    child: _buildDirectionalTree(
                                      c,
                                      theme,
                                      true,
                                    ),
                                  ),
                                  Container(
                                    width: 2,
                                    height: 30,
                                    color: lineColor,
                                  ),
                                ],
                              ),
                            )
                            .toList(),
                      ),
                      Divider(color: lineColor, thickness: 2, height: 2),
                      Center(
                        child: Container(
                          width: 2,
                          height: 30,
                          color: lineColor,
                        ),
                      ),
                    ]
                  : [
                      Center(
                        child: Container(
                          width: 2,
                          height: 30,
                          color: lineColor,
                        ),
                      ),
                      Divider(color: lineColor, thickness: 2, height: 2),
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: node.children
                            .map(
                              (c) => Column(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Container(
                                    width: 2,
                                    height: 30,
                                    color: lineColor,
                                  ),
                                  Padding(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 20,
                                    ),
                                    child: _buildDirectionalTree(
                                      c,
                                      theme,
                                      false,
                                    ),
                                  ),
                                ],
                              ),
                            )
                            .toList(),
                      ),
                    ],
            ),
          );

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: isUpward
          ? [childrenWidget, _buildNodeCard(node, theme)]
          : [_buildNodeCard(node, theme), childrenWidget],
    );
  }

  Widget _buildNodeCard(PlannerNode node, ThemeData theme) {
    Color cardColor;
    IconData icon;
    switch (node.type) {
      case NodeType.hub:
        cardColor = Colors.deepOrange;
        icon = Icons.hub_outlined;
        break;
      case NodeType.source:
        cardColor = theme.colorScheme.primary;
        icon = Icons.satellite_alt;
        break;
      case NodeType.instrument:
        cardColor = Colors.teal;
        icon = Icons.analytics_outlined;
        break;
      case NodeType.branching:
        cardColor = Colors.indigo;
        icon = Icons.account_tree_outlined;
        break;
      case NodeType.converter:
        cardColor = Colors.purple;
        icon = Icons.published_with_changes;
        break;
      default:
        cardColor = Colors.grey.shade700;
        icon = Icons.cable;
    }

    final sol = _solvePath();
    double? appliedFreq, displayLoss;
    for (var s in sol.steps) {
      if (s.label == node.label) {
        appliedFreq = s.frequency;
        displayLoss = s.loss;
      }
    }

    return ContentCard(
      padding: EdgeInsets.zero,
      width: 180,
      borderRadius: 16,
      child: InkWell(
        onTap: () => _showEditDialog(node),
        borderRadius: BorderRadius.circular(16),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 12),
              decoration: BoxDecoration(
                color: cardColor,
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(16),
                ),
              ),
              child: Row(
                children: [
                  Icon(icon, color: Colors.white, size: 16),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      node.type == NodeType.source
                          ? 'SOURCE'
                          : node.type.name.toUpperCase(),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                  if (node.type != NodeType.instrument)
                    Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () => _showAddMenu(node),
                        borderRadius: BorderRadius.circular(12),
                        child: const Padding(
                          padding: EdgeInsets.all(4.0),
                          child: Icon(
                            Icons.add_circle_outline,
                            color: Colors.white,
                            size: 18,
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    node.label,
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 13,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        displayLoss != null
                            ? '${displayLoss.toStringAsFixed(2)} dB'
                            : (node.lossDb == 0 ? '0 dB' : '${node.lossDb} dB'),
                        style: TextStyle(
                          color: (displayLoss ?? node.lossDb) > 0
                              ? Colors.red
                              : ((displayLoss ?? node.lossDb) < 0
                                    ? Colors.green
                                    : Colors.grey),
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                      ),
                      if (appliedFreq != null) ...[
                        const SizedBox(width: 4),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 4,
                            vertical: 1,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.black.withValues(alpha: 0.05),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            '${appliedFreq.toInt()} MHz',
                            style: const TextStyle(
                              fontSize: 8,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                      if (node.type == NodeType.component ||
                          node.type == NodeType.branching ||
                          node.type == NodeType.converter)
                        const Icon(
                          Icons.edit_outlined,
                          size: 14,
                          color: Colors.grey,
                        ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showAddMenu(PlannerNode parent) {
    bool hasChild = parent.children.isNotEmpty;
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        ),
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              hasChild
                  ? 'Insert/Split after "${parent.label}"'
                  : 'Add Component after "${parent.label}"',
              style: GoogleFonts.outfit(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 20),
            _buildAddOp(
              Icons.satellite_alt,
              'Source',
              'Signal Generator / Spacecraft Tx',
              () {
                Navigator.pop(context);
                _addNode(parent, NodeType.source, insert: hasChild);
              },
            ),
            _buildAddOp(
              Icons.add_circle_outline,
              'Cable / Loss',
              'Insert a calibrated or fixed cable',
              () {
                Navigator.pop(context);
                _addNode(parent, NodeType.component, insert: hasChild);
              },
            ),
            _buildAddOp(
              Icons.account_tree_outlined,
              'Branching Point',
              'Add a Splitter/Coupler',
              () {
                Navigator.pop(context);
                _addNode(parent, NodeType.branching, insert: hasChild);
              },
            ),
            _buildAddOp(
              Icons.analytics_outlined,
              'Terminal Instrument',
              'Add a Sink',
              () {
                Navigator.pop(context);
                _addNode(parent, NodeType.instrument, insert: false);
              },
            ),
            _buildAddOp(
              Icons.published_with_changes,
              'Converter',
              'Add BUC/LNB/Mixer',
              () {
                Navigator.pop(context);
                _addNode(parent, NodeType.converter, insert: hasChild);
              },
            ),
            if (hasChild) ...[
              const Divider(height: 32),
              _buildAddOp(Icons.alt_route, 'New Branch', 'Add a side path', () {
                Navigator.pop(context);
                _addNode(parent, NodeType.component, insert: false);
              }),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildAddOp(
    IconData icon,
    String label,
    String desc,
    VoidCallback tap,
  ) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: Colors.indigo.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: Colors.indigo, size: 20),
      ),
      title: Text(label, style: const TextStyle(fontWeight: FontWeight.bold)),
      subtitle: Text(
        desc,
        style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
      ),
      onTap: tap,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    );
  }

  List<double> get _availableFrequencies {
    final freqs = <double>{};
    for (var r in _allCalibratedRecords) {
      for (var m in r.measurements) {
        freqs.add(m.frequency);
      }
    }
    if (freqs.isEmpty) return [1000.0];
    final sorted = freqs.toList()..sort();
    return sorted;
  }

  void _showEditDialog(PlannerNode node) {
    if (node.type == NodeType.hub) return;
    final lCtrl = TextEditingController(text: node.label);
    final lfCtrl = TextEditingController(text: node.lossDb.toString());
    final loCtrl = TextEditingController(text: node.loOffset.toString());
    double sFreq = node.sourceFrequency;
    String? sCable = node.calibratedCableId;
    NodeDirection nDir = node.direction;

    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 400),
          child: Container(
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
            child: StatefulBuilder(
              builder: (context, setDState) => Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: Colors.grey.shade50,
                      borderRadius: const BorderRadius.vertical(
                        top: Radius.circular(24),
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Edit ${node.type.name}',
                          style: GoogleFonts.outfit(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        IconButton(
                          onPressed: () => Navigator.pop(context),
                          icon: const Icon(Icons.close, size: 20),
                        ),
                      ],
                    ),
                  ),
                  Flexible(
                    child: SingleChildScrollView(
                      padding: const EdgeInsets.all(24),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _label('COMPONENT LABEL'),
                          TextField(
                            controller: lCtrl,
                            decoration: _dec('Enter name'),
                          ),
                          const SizedBox(height: 20),
                          if (_hubNode != null &&
                              _hubNode!.children.contains(node)) ...[
                            _label('SIDE (SPACECRAFT / INSTRUMENT)'),
                            DropdownButtonFormField<NodeDirection>(
                              initialValue: nDir,
                              decoration: _dec(''),
                              items: const [
                                DropdownMenuItem(
                                  value: NodeDirection.up,
                                  child: Text('Spacecraft Side (Top)'),
                                ),
                                DropdownMenuItem(
                                  value: NodeDirection.down,
                                  child: Text('Instrument Side (Bottom)'),
                                ),
                              ],
                              onChanged: (v) {
                                if (v != null) {
                                  setDState(() => nDir = v);
                                }
                              },
                            ),
                            const SizedBox(height: 20),
                          ],
                          if (node.type == NodeType.source) ...[
                            _label('SOURCE FREQUENCY (MHz)'),
                            DropdownButtonFormField<double>(
                              initialValue: _availableFrequencies.contains(sFreq)
                                  ? sFreq
                                  : _availableFrequencies.first,
                              decoration: _dec('Select Frequency'),
                              items: _availableFrequencies
                                  .map(
                                    (f) => DropdownMenuItem(
                                      value: f,
                                      child: Text(
                                        '${f.toStringAsFixed(1)} MHz',
                                      ),
                                    ),
                                  )
                                  .toList(),
                              onChanged: (v) {
                                if (v != null) setDState(() => sFreq = v);
                              },
                            ),
                            const SizedBox(height: 20),
                          ],
                          if (node.type == NodeType.component) ...[
                            _label('CALIBRATED CABLE'),
                            Row(
                              children: [
                                Expanded(
                                  child: DropdownButtonFormField<String>(
                                    initialValue: sCable,
                                    decoration: _dec(''),
                                    items: [
                                      const DropdownMenuItem(
                                        value: null,
                                        child: Text('Manual Entry'),
                                      ),
                                      ..._calibratedCables.map(
                                        (c) => DropdownMenuItem(
                                          value: c,
                                          child: Text(c),
                                        ),
                                      ),
                                    ],
                                    onChanged: (v) =>
                                        setDState(() => sCable = v),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                IconButton(
                                  onPressed: () => _showCableAssistant(
                                    (name, loss) => setDState(() {
                                      sCable = name;
                                      lfCtrl.text = loss.toStringAsFixed(2);
                                    }),
                                  ),
                                  icon: const Icon(
                                    Icons.auto_awesome,
                                    color: Colors.indigo,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 20),
                          ],
                          if (node.type == NodeType.converter) ...[
                            _label('LO OFFSET (MHz)'),
                            TextField(
                              controller: loCtrl,
                              keyboardType:
                                  const TextInputType.numberWithOptions(
                                    signed: true,
                                    decimal: true,
                                  ),
                              decoration: _dec('e.g. -16000', suf: 'MHz'),
                            ),
                            const SizedBox(height: 20),
                          ],
                          _label(
                            node.type == NodeType.source
                                ? 'OUTPUT POWER (dBm)'
                                : 'PATH LOSS (dB)',
                          ),
                          TextField(
                            controller: lfCtrl,
                            keyboardType: const TextInputType.numberWithOptions(
                              signed: true,
                              decimal: true,
                            ),
                            decoration: _dec(
                              node.type == NodeType.source ? 'Power' : 'Loss',
                              suf: node.type == NodeType.source ? 'dBm' : 'dB',
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.all(24),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        TextButton.icon(
                          onPressed: () {
                            final p = _findParent(_root, node);
                            if (p != null) {
                              _deleteNode(p, node);
                              Navigator.pop(context);
                            }
                          },
                          icon: const Icon(Icons.delete_outline, size: 18),
                          label: const Text('Delete'),
                          style: TextButton.styleFrom(
                            foregroundColor: Colors.red,
                          ),
                        ),
                        ElevatedButton(
                          onPressed: () {
                            setState(() {
                              node.label = lCtrl.text;
                              node.lossDb = double.tryParse(lfCtrl.text) ?? 0.0;
                              node.loOffset =
                                  double.tryParse(loCtrl.text) ?? 0.0;
                              node.sourceFrequency = sFreq;
                              node.direction = nDir;
                              node.calibratedCableId = sCable;
                            });
                            Navigator.pop(context);
                          },
                          child: const Text('Save Changes'),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  List<String> _getUsedCables() {
    List<String> res = [];
    void traverse(PlannerNode node) {
      if (node.calibratedCableId != null &&
          node.calibratedCableId!.isNotEmpty) {
        res.add(node.calibratedCableId!);
      }
      for (var c in node.children) {
        traverse(c);
      }
    }

    if (_hubNode != null) traverse(_hubNode!);
    return res;
  }

  void _showCableAssistant(Function(String, double) onSelect) {
    final usedCables = _getUsedCables();
    List<double> targets = [_availableFrequencies.first];
    final nCtrl = TextEditingController();
    final lCtrl = TextEditingController();
    bool loading = false;
    List<Map<String, dynamic>> cands = [];

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (ctx, setS) => Dialog(
          backgroundColor: Colors.transparent,
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 500),
            child: Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    'Cable Selection Assistant',
                    style: GoogleFonts.outfit(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        flex: 2,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            _label('NAME FILTER (Optional)'),
                            TextField(
                              controller: nCtrl,
                              decoration: _dec('e.g. SMA'),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        flex: 1,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            _label('LENGTH (m)'),
                            TextField(
                              controller: lCtrl,
                              keyboardType: TextInputType.number,
                              decoration: _dec('e.g. 1.5'),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  _label('TARGET FREQUENCIES (MHz)'),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: _availableFrequencies.map((f) {
                      final isSel = targets.contains(f);
                      return FilterChip(
                        label: Text(f.toStringAsFixed(1)),
                        selected: isSel,
                        onSelected: (val) {
                          setS(() {
                            if (val) {
                              targets.add(f);
                            } else {
                              targets.remove(f);
                            }
                            if (targets.isEmpty) {
                              targets.add(_availableFrequencies.first);
                            }
                          });
                        },
                      );
                    }).toList(),
                  ),
                  const SizedBox(height: 24),
                  ElevatedButton.icon(
                    onPressed: loading
                        ? null
                        : () async {
                            setS(() => loading = true);
                            final c = _allCalibratedRecords;
                            Map<String, CableLossRecord> unique = {};
                            final q = nCtrl.text.toLowerCase();
                            final filterLen = double.tryParse(lCtrl.text);
                            for (var r in c) {
                              if (q.isNotEmpty &&
                                  !r.cableName.toLowerCase().contains(q)) {
                                continue;
                              }
                              if (filterLen != null && r.length != filterLen) {
                                continue;
                              }
                              if (!unique.containsKey(r.cableName)) {
                                unique[r.cableName] = r;
                              }
                            }
                            List<Map<String, dynamic>> res = [];
                            for (var cab in unique.values) {
                              if (cab.measurements.isEmpty) continue;
                              cab.measurements.sort(
                                (a, b) => a.frequency.compareTo(b.frequency),
                              );
                              double maxL = -double.infinity;
                              for (var f in targets) {
                                double ml = 0;
                                if (f <= cab.measurements.first.frequency) {
                                  ml = cab.measurements.first.loss;
                                } else if (f >=
                                    cab.measurements.last.frequency) {
                                  ml = cab.measurements.last.loss;
                                } else {
                                  for (
                                    int i = 0;
                                    i < cab.measurements.length - 1;
                                    i++
                                  ) {
                                    final p1 = cab.measurements[i],
                                        p2 = cab.measurements[i + 1];
                                    if (f >= p1.frequency &&
                                        f <= p2.frequency) {
                                      ml =
                                          p1.loss +
                                          ((f - p1.frequency) /
                                                  (p2.frequency -
                                                      p1.frequency)) *
                                              (p2.loss - p1.loss);
                                      break;
                                    }
                                  }
                                }
                                double currentL = ml;
                                double parsedL = currentL < 0
                                    ? currentL * -1.0
                                    : currentL;
                                if (parsedL > maxL) maxL = parsedL;
                              }
                              res.add({
                                'name': cab.cableName,
                                'loss': maxL,
                                'length': cab.length,
                                'date': cab.date,
                              });
                            }
                            res.sort(
                              (a, b) => (a['loss'] as double).compareTo(
                                b['loss'] as double,
                              ),
                            );
                            setS(() {
                              cands = res.take(10).toList();
                              loading = false;
                            });
                          },
                    icon: loading
                        ? const SizedBox(
                            width: 14,
                            height: 14,
                            child: CircularProgressIndicator(
                              color: Colors.white,
                              strokeWidth: 2,
                            ),
                          )
                        : const Icon(Icons.search),
                    label: Text(loading ? 'Searching...' : 'Find Matches'),
                  ),
                  if (cands.isNotEmpty) ...[
                    const SizedBox(height: 16),
                    _label('TOP CANDIDATES'),
                    ...cands.map((c) {
                      final isUsed = usedCables.contains(c['name']);
                      return ListTile(
                        enabled: !isUsed,
                        title: Text(c['name']),
                        subtitle: Text(
                          targets.length > 1
                              ? '${(c['loss'] as double).toStringAsFixed(2)} dB Max • ${c['length'] ?? 0.0}m Length${isUsed ? " (Already Used)" : ""}'
                              : '${(c['loss'] as double).toStringAsFixed(2)} dB • ${c['length'] ?? 0.0}m Length${isUsed ? " (Already Used)" : ""}',
                        ),
                        onTap: isUsed
                            ? null
                            : () {
                                onSelect(c['name'], c['loss']);
                                Navigator.pop(ctx);
                              },
                      );
                    }),
                  ],
                  const SizedBox(height: 16),
                  TextButton(
                    onPressed: () => Navigator.pop(ctx),
                    child: const Text('Close'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSummaryPanel(ThemeData theme) {
    final sources = _getTerminals(NodeType.source);
    final sinks = _getTerminals(NodeType.instrument);
    if (_startNodeId != null && !sources.any((t) => t.id == _startNodeId)) {
      _startNodeId = null;
    }
    if (_endNodeId != null && !sinks.any((t) => t.id == _endNodeId)) {
      _endNodeId = null;
    }

    final sol = _solvePath();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Link Budget Summary',
          style: GoogleFonts.outfit(fontSize: 18, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 4),
        Text(
          'Select source and sink to analyze path',
          style: TextStyle(color: Colors.grey.shade500, fontSize: 13),
        ),
        const SizedBox(height: 24),
        _label('START TERMINAL (SOURCE)'),
        DropdownButtonFormField<String>(
          initialValue: _startNodeId,
          decoration: _dec(''),
          items: sources
              .map((s) => DropdownMenuItem(value: s.id, child: Text(s.label)))
              .toList(),
          onChanged: (v) => setState(() => _startNodeId = v),
        ),
        const SizedBox(height: 16),
        _label('END TERMINAL (SINK)'),
        DropdownButtonFormField<String>(
          initialValue: _endNodeId,
          decoration: _dec(''),
          items: sinks
              .map((s) => DropdownMenuItem(value: s.id, child: Text(s.label)))
              .toList(),
          onChanged: (v) => setState(() => _endNodeId = v),
        ),
        if (sol.steps.isNotEmpty) ...[
          const SizedBox(height: 24),
          _label('FREQUENCY-AWARE BREAKDOWN'),
          Expanded(
            child: Scrollbar(
              controller: _summaryVerticalScroll,
              thumbVisibility: true,
              child: SingleChildScrollView(
                controller: _summaryVerticalScroll,
                scrollDirection: Axis.vertical,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Scrollbar(
                      controller: _summaryHorizontalScroll,
                      thumbVisibility: true,
                      child: SingleChildScrollView(
                        controller: _summaryHorizontalScroll,
                        scrollDirection: Axis.horizontal,
                        child: DataTable(
                          columnSpacing: 16,
                          headingRowHeight: 40,
                          headingTextStyle: const TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 13,
                            color: Colors.indigo,
                          ),
                          dataTextStyle: const TextStyle(
                            fontSize: 12,
                            color: Colors.black87,
                          ),
                          columns: const [
                            DataColumn(label: Text('Component')),
                            DataColumn(label: Text('Freq (MHz)')),
                            DataColumn(label: Text('Loss (dB)')),
                            DataColumn(label: Text('In (dBm)')),
                            DataColumn(label: Text('Out (dBm)')),
                          ],
                          rows: sol.steps
                              .map(
                                (s) => DataRow(
                                  cells: [
                                    DataCell(Text(s.label)),
                                    DataCell(
                                      Text(s.frequency.toStringAsFixed(1)),
                                    ),
                                    DataCell(
                                      Text(
                                        s.loss.toStringAsFixed(2),
                                        style: TextStyle(
                                          color: s.loss > 0
                                              ? Colors.red
                                              : Colors.green,
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                    ),
                                    DataCell(
                                      Text(s.inputPower.toStringAsFixed(2)),
                                    ),
                                    DataCell(
                                      Text(
                                        s.outputPower.toStringAsFixed(2),
                                        style: const TextStyle(
                                          fontWeight: FontWeight.bold,
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              )
                              .toList(),
                        ),
                      ),
                    ),
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.blue.shade50,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.blue.shade200),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        const Text(
                          'Final Power Budget',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Colors.blue,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('Total Path Loss:'),
                            Text(
                              '${sol.totalLoss.toStringAsFixed(2)} dB',
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 4),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('Received Power:'),
                            Text(
                              '${sol.finalPower.toStringAsFixed(2)} dBm',
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                fontSize: 18,
                                color: Colors.green,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ] else
          const Spacer(),
      ],
    );
  }

  Widget _label(String text) => Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Text(
      text,
      style: TextStyle(
        fontSize: 10,
        fontWeight: FontWeight.bold,
        color: Colors.grey.shade600,
        letterSpacing: 1.2,
      ),
    ),
  );
  InputDecoration _dec(String hint, {String? suf}) => InputDecoration(
    hintText: hint,
    suffixText: suf,
    filled: true,
    fillColor: Colors.grey.shade50,
    border: OutlineInputBorder(
      borderRadius: BorderRadius.circular(12),
      borderSide: BorderSide.none,
    ),
    contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
  );
}

class GridPainter extends CustomPainter {
  final double gridSize;
  final Color color;
  GridPainter({required this.gridSize, required this.color});
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color
      ..strokeWidth = 1;
    for (double i = 0; i < size.width; i += gridSize) {
      canvas.drawLine(Offset(i, 0), Offset(i, size.height), paint);
    }
    for (double i = 0; i < size.height; i += gridSize) {
      canvas.drawLine(Offset(0, i), Offset(size.width, i), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
