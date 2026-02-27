import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/utils/notifications.dart';

enum NodeType { source, component, branching, instrument, hub }

enum NodeDirection { up, down, center }

class PlannerNode {
  final String id;
  String label;
  double lossDb;
  NodeType type;
  NodeDirection direction;
  String? physicalResourceId; // For shared assets
  List<PlannerNode> children;
  List<double> supportedFrequencies;

  PlannerNode({
    required this.id,
    required this.label,
    this.lossDb = 0.0,
    required this.type,
    this.direction = NodeDirection.down,
    this.physicalResourceId,
    List<PlannerNode>? children,
    this.supportedFrequencies = const [],
  }) : children = children ?? [];

  PlannerNode copyWith({
    String? label,
    double? lossDb,
    NodeType? type,
    NodeDirection? direction,
    String? physicalResourceId,
    List<double>? supportedFrequencies,
  }) {
    return PlannerNode(
      id: id,
      label: label ?? this.label,
      lossDb: lossDb ?? this.lossDb,
      type: type ?? this.type,
      direction: direction ?? this.direction,
      physicalResourceId: physicalResourceId ?? this.physicalResourceId,
      supportedFrequencies: supportedFrequencies ?? this.supportedFrequencies,
      children: children,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'label': label,
      'lossDb': lossDb,
      'type': type.index,
      'direction': direction.index,
      'physicalResourceId': physicalResourceId,
      'supportedFrequencies': supportedFrequencies,
      'children': children.map((c) => c.toJson()).toList(),
    };
  }

  factory PlannerNode.fromJson(Map<String, dynamic> json) {
    return PlannerNode(
      id: json['id'],
      label: json['label'],
      lossDb: (json['lossDb'] as num).toDouble(),
      type: NodeType.values[json['type']],
      direction: NodeDirection.values[json['direction']],
      physicalResourceId: json['physicalResourceId'],
      supportedFrequencies: List<double>.from(
        json['supportedFrequencies'] ?? [],
      ),
      children: (json['children'] as List?)
          ?.map((c) => PlannerNode.fromJson(c))
          .toList(),
    );
  }
}

class PathLossPlannerScreen extends StatefulWidget {
  final bool isActive;
  const PathLossPlannerScreen({super.key, this.isActive = false});

  @override
  State<PathLossPlannerScreen> createState() => _PathLossPlannerScreenState();
}

class _PathLossPlannerScreenState extends State<PathLossPlannerScreen> {
  // Multi-path State
  bool _isStarLayout = true; // New Direction: Hub-Centric
  PlannerNode? _hubNode; // The TSM Hub
  String? _startNodeId; // TSM Star Terminal Start
  String? _endNodeId; // TSM Star Terminal End

  // Cache for calibrated cables and TSM from server
  List<String> _calibratedCables = [];
  TSMInternalLossMetadata? _tsmData;

  final TransformationController _transformationController =
      TransformationController();

  @override
  void initState() {
    super.initState();
    _initializeDefaultTree();
    _loadMetadata();
  }

  void _loadMetadata() {
    final server = context.read<ServerService>();
    final bootstrap = server.status.bootstrapData;
    if (bootstrap != null) {
      setState(() {
        _calibratedCables = bootstrap.cableLossData.existingCables;
        _tsmData = bootstrap.tsmInternalLossData;

        // Restore saved diagram if available
        if (bootstrap.plannerData.isNotEmpty) {
          try {
            final Map<String, dynamic> data = jsonDecode(bootstrap.plannerData);
            _hubNode = PlannerNode.fromJson(data);
          } catch (e) {
            debugPrint('Failed to restore saved diagram: $e');
            _initializeDefaultTree();
          }
        } else {
          _initializeDefaultTree();
        }
      });
    }
  }

  Future<void> _saveDiagram() async {
    if (_hubNode == null) return;
    final server = context.read<ServerService>();
    final jsonString = jsonEncode(_hubNode!.toJson());
    final success = await server.savePlannerData(jsonString);

    if (mounted) {
      if (success) {
        AppNotifications.show(context, 'Configuration Saved Successfully');
      } else {
        AppNotifications.showError(context, 'Failed to Save Configuration');
      }
    }
  }

  PlannerNode get _root => _hubNode!;

  void _syncSharedNodes(
    PlannerNode root,
    String physicalId,
    double loss,
    String label,
  ) {
    void traverse(PlannerNode node) {
      if (node.physicalResourceId == physicalId) {
        node.lossDb = loss;
        node.label = label;
      }
      for (var child in node.children) {
        traverse(child);
      }
    }

    traverse(root);
  }

  void _globalSync(String physicalId, double loss, String label) {
    if (_hubNode != null) _syncSharedNodes(_hubNode!, physicalId, loss, label);
  }

  void _addNode(PlannerNode parent, NodeType type, {bool insert = false}) {
    setState(() {
      final newNode = PlannerNode(
        id: 'node-${DateTime.now().millisecondsSinceEpoch}',
        label: 'New ${type.name}',
        type: type,
        direction: parent.direction, // Inherit direction from parent
        lossDb: type == NodeType.branching ? 1.0 : 0.0,
      );

      if (insert) {
        // Move all current children to the new node to insert it in between
        newNode.children.addAll(parent.children);
        parent.children.clear();
      }

      parent.children.add(newNode);
    });
    AppNotifications.show(
      context,
      insert ? 'Inserted ${type.name}' : 'Added ${type.name}',
      type: NotificationType.success,
    );
  }

  void _deleteNode(PlannerNode parent, PlannerNode nodeToDelete) {
    setState(() {
      parent.children.remove(nodeToDelete);
    });
    AppNotifications.show(
      context,
      'Node Removed',
      type: NotificationType.warning,
    );
  }

  PlannerNode? _findParent(PlannerNode current, PlannerNode target) {
    for (var child in current.children) {
      if (child == target) return current;
      final found = _findParent(child, target);
      if (found != null) return found;
    }
    return null;
  }

  void _showAddMenu(PlannerNode parent) {
    bool hasChildren = parent.children.isNotEmpty;

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
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  hasChildren
                      ? 'Insert/Split after "${parent.label}"'
                      : 'Add Component after "${parent.label}"',
                  style: GoogleFonts.outfit(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (hasChildren)
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.blue.shade50,
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      'Insert Mode Enabled',
                      style: TextStyle(
                        fontSize: 10,
                        color: Colors.blue.shade700,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 20),
            _buildAddOption(
              icon: Icons.add_circle_outline,
              label: 'Cable / Loss Component',
              desc: hasChildren
                  ? 'Insert a new cable between this and next components'
                  : 'Add a fixed loss or calibrated cable',
              onTap: () {
                Navigator.pop(context);
                _addNode(parent, NodeType.component, insert: hasChildren);
              },
            ),
            _buildAddOption(
              icon: Icons.account_tree_outlined,
              label: hasChildren ? 'Split with Branching' : 'Branching Point',
              desc: 'Add a Splitter/Coupler to create multiple paths',
              onTap: () {
                Navigator.pop(context);
                _addNode(parent, NodeType.branching, insert: hasChildren);
              },
            ),
            _buildAddOption(
              icon: Icons.analytics_outlined,
              label: 'Terminal Instrument',
              desc: 'Add a Power Meter, SA, or terminal spacecraft port',
              onTap: () {
                Navigator.pop(context);
                _addNode(parent, NodeType.instrument, insert: false);
              },
            ),
            if (hasChildren) ...[
              const Divider(height: 32),
              _buildAddOption(
                icon: Icons.alt_route,
                label: 'Create New Secondary Branch',
                desc: 'Add a separate path starting from here',
                onTap: () {
                  Navigator.pop(context);
                  _addNode(parent, NodeType.component, insert: false);
                },
              ),
            ],
            const SizedBox(height: 12),
          ],
        ),
      ),
    );
  }

  Widget _buildAddOption({
    required IconData icon,
    required String label,
    required String desc,
    required VoidCallback onTap,
  }) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: Colors.indigo.withOpacity(0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: Colors.indigo, size: 20),
      ),
      title: Text(label, style: const TextStyle(fontWeight: FontWeight.bold)),
      subtitle: Text(
        desc,
        style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
      ),
      onTap: onTap,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    );
  }

  void _initializeDefaultTree() {
    if (_isStarLayout) {
      if (_hubNode != null) return;
      _hubNode = PlannerNode(
        id: 'tsm-hub',
        label: 'TSM Central Hub',
        type: NodeType.hub,
        direction: NodeDirection.center,
      );

      final tsm = _tsmData;
      if (tsm != null && tsm.ok) {
        final ports = tsm.measuredLoss.paths;
        final leftPorts = <String, Set<double>>{};
        final rightPorts = <String, Set<double>>{};

        for (var p in ports) {
          if (p.inputPort.isNotEmpty) {
            final name = p.inputPort.replaceAll('-WithPad', '');
            leftPorts.putIfAbsent(name, () => {}).addAll(p.frequencies);
          }
          if (p.outputPort.isNotEmpty) {
            final name = p.outputPort.replaceAll('-WithPad', '');
            rightPorts.putIfAbsent(name, () => {}).addAll(p.frequencies);
          }
        }

        // Star layout: Top side nodes grow "upwards" from Hub
        leftPorts.forEach((port, freqs) {
          _hubNode!.children.add(
            PlannerNode(
              id: 'port-$port',
              label: port,
              type: NodeType.branching,
              direction: NodeDirection.up,
              supportedFrequencies: freqs.toList()..sort(),
            ),
          );
        });

        // Bottom side nodes grow "downwards" from Hub
        rightPorts.forEach((port, freqs) {
          // Avoid duplicates if a port is both In and Out
          if (leftPorts.containsKey(port)) return;
          _hubNode!.children.add(
            PlannerNode(
              id: 'port-$port',
              label: port,
              type: NodeType.branching,
              direction: NodeDirection.down,
              supportedFrequencies: freqs.toList()..sort(),
            ),
          );
        });
      }
      return;
    }
  }

  List<PlannerNode> _getTopTerminalSources() {
    final results = <PlannerNode>[];
    if (_hubNode == null) return results;

    void traverse(PlannerNode node) {
      if (node.type == NodeType.source || node.type == NodeType.instrument) {
        results.add(node);
      }
      for (var child in node.children) {
        traverse(child);
      }
    }

    for (var child in _hubNode!.children) {
      if (child.direction == NodeDirection.up) {
        traverse(child);
      }
    }
    return results;
  }

  List<PlannerNode> _getBottomTerminalInstruments() {
    final results = <PlannerNode>[];
    if (_hubNode == null) return results;

    void traverse(PlannerNode node) {
      if (node.type == NodeType.source || node.type == NodeType.instrument) {
        results.add(node);
      }
      for (var child in node.children) {
        traverse(child);
      }
    }

    for (var child in _hubNode!.children) {
      if (child.direction == NodeDirection.down) {
        traverse(child);
      }
    }
    return results;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Column(
        children: [
          _buildHeader(theme),
          Expanded(
            child: Row(
              children: [
                Expanded(flex: 3, child: _buildSchematicCanvas(theme)),
                Container(
                  width: 320,
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    border: Border(
                      left: BorderSide(color: Colors.grey.shade100),
                    ),
                  ),
                  child: _buildSummaryPanel(theme),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  PlannerNode? _findHubChildOf(PlannerNode target) {
    if (_hubNode == null) return null;
    PlannerNode? current = target;
    while (current != null) {
      final parent = _findParent(_root, current);
      if (parent == _hubNode) return current;
      current = parent;
    }
    return null;
  }

  double _getHubPathLoss(String? inputPort, String? outputPort) {
    if (_tsmData == null || inputPort == null || outputPort == null) return 0.0;
    // Hub internal path look-up
    for (var p in _tsmData!.measuredLoss.paths) {
      if (p.inputPort.replaceAll('WithPad', '') == inputPort &&
          p.outputPort.replaceAll('WithPad', '') == outputPort) {
        if (p.losses.isNotEmpty) {
          // Average for now? Or just first one?
          return p.losses.reduce((a, b) => a + b) / p.losses.length;
        }
      }
    }
    return 0.0;
  }

  double _calculateNodePathToHubLoss(PlannerNode? node) {
    if (node == null || _hubNode == null) return 0.0;
    double total = 0.0;
    PlannerNode? current = node;
    while (current != null && current != _hubNode) {
      total += current.lossDb;
      current = _findParent(_root, current);
    }
    return total;
  }

  Widget _buildHeader(ThemeData theme) {
    return ScreenHeader(
      title: 'Path Loss Planner',
      subtitle: 'Visualize and calculate RF link budgets',
      icon: Icons.map_outlined,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          ElevatedButton.icon(
            onPressed: () {
              setState(() {
                _hubNode = null;
                _initializeDefaultTree();
                _transformationController.value = Matrix4.identity();
              });
              // Also clear on server
              context.read<ServerService>().savePlannerData("");
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
            onPressed: _saveDiagram,
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.green.shade600,
              foregroundColor: Colors.white,
            ),
            icon: const Icon(Icons.save_outlined, size: 16),
            label: const Text('SAVE'),
          ),
        ],
      ),
    );
  }

  Widget _buildSchematicCanvas(ThemeData theme) {
    return Stack(
      children: [
        Container(
          color: Colors.grey.shade50,
          child: InteractiveViewer(
            transformationController: _transformationController,
            constrained: false, // Essential for large logic canvas
            boundaryMargin: const EdgeInsets.all(100),
            minScale: 0.1,
            maxScale: 2.5,
            child: SizedBox(
              width: 5000,
              height: 5000,
              child: Stack(
                children: [
                  // Massive Grid Background
                  Positioned.fill(
                    child: CustomPaint(
                      painter: GridPainter(
                        gridSize: 40,
                        color: Colors.grey.shade200,
                      ),
                    ),
                  ),
                  Positioned(
                    left: 2000, // Center of our massive canvas
                    top: 2000,
                    child: _isStarLayout
                        ? _buildStarNodeTree(_root, theme)
                        : _buildNodeTree(_root, theme),
                  ),
                ],
              ),
            ),
          ),
        ),
        // Infinite Canvas HUD (Fixed)
        Positioned(
          bottom: 24,
          left: 24,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.9),
              borderRadius: BorderRadius.circular(30),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.05),
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
    // Separate children by direction (up = spacecraft side, down = inst side)
    final topChildren = node.children
        .where((c) => c.direction == NodeDirection.up)
        .toList();
    final bottomChildren = node.children
        .where((c) => c.direction == NodeDirection.down)
        .toList();

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        // TOP SIDE: Grows UP from Hub (Signal flows DOWN towards Hub)
        if (topChildren.isNotEmpty) ...[
          IntrinsicWidth(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: topChildren.map((child) {
                    return Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 20),
                          child: _buildTopNodeTree(child, theme),
                        ),
                        Container(
                          width: 2,
                          height: 30,
                          color: theme.colorScheme.primary.withOpacity(0.3),
                        ),
                      ],
                    );
                  }).toList(),
                ),
                Divider(
                  color: theme.colorScheme.primary.withOpacity(0.3),
                  thickness: 2,
                  height: 2,
                ),
              ],
            ),
          ),
          Container(
            width: 2,
            height: 30,
            color: theme.colorScheme.primary.withOpacity(0.3),
          ),
        ],

        // CENTER: The Hub
        _buildNodeCard(node, theme),

        // BOTTOM SIDE: Grows DOWN from Hub (Signal flows DOWN from Hub)
        if (bottomChildren.isNotEmpty) ...[
          Container(
            width: 2,
            height: 30,
            color: theme.colorScheme.primary.withOpacity(0.3),
          ),
          IntrinsicWidth(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Divider(
                  color: theme.colorScheme.primary.withOpacity(0.3),
                  thickness: 2,
                  height: 2,
                ),
                Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: bottomChildren.map((child) {
                    return Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 2,
                          height: 30,
                          color: theme.colorScheme.primary.withOpacity(0.3),
                        ),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 20),
                          child: _buildNodeTree(child, theme),
                        ),
                      ],
                    );
                  }).toList(),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildTopNodeTree(PlannerNode node, ThemeData theme) {
    if (node.children.isEmpty) {
      return _buildNodeCard(node, theme);
    }

    // Grows UP (Children are above parent)
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        if (node.children.length == 1) ...[
          _buildTopNodeTree(node.children.first, theme),
          Container(
            width: 2,
            height: 30,
            color: theme.colorScheme.primary.withOpacity(0.3),
          ),
        ] else ...[
          IntrinsicWidth(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: node.children.map((child) {
                    return Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 20),
                          child: _buildTopNodeTree(child, theme),
                        ),
                        Container(
                          width: 2,
                          height: 30,
                          color: theme.colorScheme.primary.withOpacity(0.3),
                        ),
                      ],
                    );
                  }).toList(),
                ),
                Divider(
                  color: theme.colorScheme.primary.withOpacity(0.3),
                  thickness: 2,
                  height: 2,
                ),
                Center(
                  child: Container(
                    width: 2,
                    height: 30,
                    color: theme.colorScheme.primary.withOpacity(0.3),
                  ),
                ),
              ],
            ),
          ),
        ],
        _buildNodeCard(node, theme),
      ],
    );
  }

  Widget _buildNodeTree(PlannerNode node, ThemeData theme) {
    if (node.children.isEmpty) {
      return _buildNodeCard(node, theme);
    }

    // Grows DOWN (Children are below parent)
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        _buildNodeCard(node, theme),
        if (node.children.length == 1) ...[
          Container(
            width: 2,
            height: 30,
            color: theme.colorScheme.primary.withOpacity(0.3),
          ),
          _buildNodeTree(node.children.first, theme),
        ] else ...[
          IntrinsicWidth(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Center(
                  child: Container(
                    width: 2,
                    height: 30,
                    color: theme.colorScheme.primary.withOpacity(0.3),
                  ),
                ),
                Divider(
                  color: theme.colorScheme.primary.withOpacity(0.3),
                  thickness: 2,
                  height: 2,
                ),
                Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: node.children.map((child) {
                    return Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 2,
                          height: 30,
                          color: theme.colorScheme.primary.withOpacity(0.3),
                        ),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 20),
                          child: _buildNodeTree(child, theme),
                        ),
                      ],
                    );
                  }).toList(),
                ),
              ],
            ),
          ),
        ],
      ],
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
      default:
        cardColor = Colors.grey.shade700;
        icon = Icons.cable;
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
                  if (node.physicalResourceId != null)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 8.0),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 6,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.amber.shade100,
                          borderRadius: BorderRadius.circular(4),
                          border: Border.all(color: Colors.amber.shade300),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(
                              Icons.link,
                              size: 10,
                              color: Colors.amber,
                            ),
                            const SizedBox(width: 4),
                            Text(
                              'Shared: ${node.physicalResourceId}',
                              style: TextStyle(
                                fontSize: 9,
                                fontWeight: FontWeight.bold,
                                color: Colors.amber.shade900,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  if (node.supportedFrequencies.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 8.0),
                      child: Wrap(
                        spacing: 4,
                        runSpacing: 4,
                        children: [
                          ...node.supportedFrequencies
                              .take(2)
                              .map(
                                (f) => Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 6,
                                    vertical: 2,
                                  ),
                                  decoration: BoxDecoration(
                                    color: Colors.indigo.shade50,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: Text(
                                    '${f.toStringAsFixed(1)} GHz',
                                    style: TextStyle(
                                      fontSize: 8,
                                      color: Colors.indigo.shade700,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ),
                          if (node.supportedFrequencies.length > 2)
                            Text(
                              '+${node.supportedFrequencies.length - 2} more',
                              style: TextStyle(
                                fontSize: 8,
                                color: Colors.grey.shade600,
                              ),
                            ),
                        ],
                      ),
                    ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        node.lossDb == 0 ? '0 dB' : '${node.lossDb} dB',
                        style: TextStyle(
                          color: node.lossDb > 0
                              ? Colors.red
                              : (node.lossDb < 0 ? Colors.green : Colors.grey),
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                      ),
                      if (node.type == NodeType.component ||
                          node.type == NodeType.branching)
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

  Widget _buildSummaryPanel(ThemeData theme) {
    final startTerminals = _getTopTerminalSources();
    final endTerminals = _getBottomTerminalInstruments();

    // Ensure state is valid
    if (_startNodeId != null &&
        !startTerminals.any((t) => t.id == _startNodeId)) {
      _startNodeId = null;
    }
    if (_endNodeId != null && !endTerminals.any((t) => t.id == _endNodeId)) {
      _endNodeId = null;
    }

    final startNode = _startNodeId != null
        ? startTerminals.firstWhere((t) => t.id == _startNodeId)
        : null;
    final endNode = _endNodeId != null
        ? endTerminals.firstWhere((t) => t.id == _endNodeId)
        : null;

    final startPortChild = startNode != null
        ? _findHubChildOf(startNode)
        : null;
    final endPortChild = endNode != null ? _findHubChildOf(endNode) : null;

    final startLoss = _calculateNodePathToHubLoss(startNode);
    final endLoss = _calculateNodePathToHubLoss(endNode);
    final hubLoss = _getHubPathLoss(startPortChild?.label, endPortChild?.label);
    final totalLoss = startLoss + hubLoss + endLoss;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Live Path Summary',
          style: GoogleFonts.outfit(fontSize: 18, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 4),
        Text(
          'Select terminal endpoints to calculate loss',
          style: TextStyle(color: Colors.grey.shade500, fontSize: 13),
        ),
        const SizedBox(height: 24),

        _buildEndpointSelector(
          theme,
          'Start Terminal (Spacecraft Side)',
          _startNodeId,
          startTerminals,
          (val) => setState(() => _startNodeId = val),
          Icons.satellite_alt,
        ),
        const SizedBox(height: 12),
        _buildEndpointSelector(
          theme,
          'End Terminal (Instrument Side)',
          _endNodeId,
          endTerminals,
          (val) => setState(() => _endNodeId = val),
          Icons.analytics_outlined,
        ),

        if (startNode != null && endNode != null) ...[
          const SizedBox(height: 32),
          Text(
            'CALCULATION BREAKDOWN',
            style: TextStyle(
              fontSize: 10,
              fontWeight: FontWeight.bold,
              color: Colors.grey.shade600,
              letterSpacing: 1.2,
            ),
          ),
          const SizedBox(height: 16),
          _buildSummaryItem(
            theme,
            'Upstream: ${startNode.label} to Hub',
            startLoss,
          ),
          _buildSummaryItem(
            theme,
            'TSM Internal: ${startPortChild?.label} → ${endPortChild?.label}',
            hubLoss,
          ),
          _buildSummaryItem(
            theme,
            'Downstream: Hub to ${endNode.label}',
            endLoss,
          ),
          const Divider(height: 32),
          _buildSummaryItem(theme, 'Total Path Loss', totalLoss, isTotal: true),
        ],

        const Spacer(),
        _buildActionButtons(theme),
      ],
    );
  }

  Widget _buildEndpointSelector(
    ThemeData theme,
    String label,
    String? value,
    List<PlannerNode> terminals,
    Function(String?) onChanged,
    IconData icon,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFieldLabel(label.toUpperCase()),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: Colors.grey.shade50,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.grey.shade200),
          ),
          child: DropdownButtonHideUnderline(
            child: DropdownButton<String>(
              isExpanded: true,
              value: value,
              hint: Text('Select $label', style: const TextStyle(fontSize: 13)),
              items: terminals.map((t) {
                return DropdownMenuItem(value: t.id, child: Text(t.label));
              }).toList(),
              onChanged: onChanged,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSummaryItem(
    ThemeData theme,
    String label,
    double loss, {
    bool isTotal = false,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isTotal
            ? theme.colorScheme.primary.withOpacity(0.1)
            : theme.colorScheme.primary.withOpacity(0.05),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isTotal
              ? theme.colorScheme.primary.withOpacity(0.3)
              : theme.colorScheme.primary.withOpacity(0.1),
          width: isTotal ? 2 : 1,
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Expanded(
            child: Text(
              label,
              style: TextStyle(
                fontWeight: isTotal ? FontWeight.bold : FontWeight.w500,
                fontSize: isTotal ? 14 : 13,
              ),
            ),
          ),
          Text(
            '${loss.toStringAsFixed(2)} dB',
            style: TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: isTotal ? 16 : 14,
              color: loss > 0 ? Colors.red.shade700 : Colors.green.shade700,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildActionButtons(ThemeData theme) {
    return Column(
      children: [
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: () => AppNotifications.show(
              context,
              'Configuration Saved',
              type: NotificationType.success,
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: theme.colorScheme.primary,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.all(16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text('Export to System'),
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          width: double.infinity,
          child: OutlinedButton(
            onPressed: () => AppNotifications.show(
              context,
              'PDF Report generated',
              type: NotificationType.info,
            ),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.all(16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text('Download PDF'),
          ),
        ),
      ],
    );
  }

  void _showEditDialog(PlannerNode node) {
    final labelController = TextEditingController(text: node.label);
    final lossController = TextEditingController(text: node.lossDb.toString());
    String? selectedCable = node.physicalResourceId;

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
                  color: Colors.black.withOpacity(0.1),
                  blurRadius: 20,
                  offset: const Offset(0, 10),
                ),
              ],
            ),
            child: StatefulBuilder(
              builder: (context, setDialogState) => Column(
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
                          _buildFieldLabel('COMPONENT LABEL'),
                          TextField(
                            controller: labelController,
                            decoration: InputDecoration(
                              hintText: 'Enter name',
                              filled: true,
                              fillColor: Colors.grey.shade50,
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(12),
                                borderSide: BorderSide.none,
                              ),
                            ),
                          ),
                          const SizedBox(height: 20),
                          if (_isStarLayout && node.type != NodeType.hub) ...[
                            _buildFieldLabel('SIGNAL FLOW'),
                            ToggleButtons(
                              constraints: const BoxConstraints(
                                minHeight: 36,
                                minWidth: 100,
                              ),
                              borderRadius: BorderRadius.circular(12),
                              isSelected: [
                                node.direction == NodeDirection.up,
                                node.direction == NodeDirection.down,
                              ],
                              onPressed: (index) {
                                setDialogState(() {
                                  node.direction = index == 0
                                      ? NodeDirection.up
                                      : NodeDirection.down;
                                });
                              },
                              children: const [
                                Padding(
                                  padding: EdgeInsets.symmetric(horizontal: 12),
                                  child: Text('Spacecraft (Top)'),
                                ),
                                Padding(
                                  padding: EdgeInsets.symmetric(horizontal: 12),
                                  child: Text('Inst (Bottom)'),
                                ),
                              ],
                            ),
                            const SizedBox(height: 20),
                          ],
                          if (node.supportedFrequencies.isNotEmpty) ...[
                            _buildFieldLabel('SUPPORTED FREQUENCIES'),
                            Container(
                              padding: const EdgeInsets.all(12),
                              width: double.infinity,
                              decoration: BoxDecoration(
                                color: Colors.blue.shade50,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Wrap(
                                spacing: 8,
                                runSpacing: 8,
                                children: node.supportedFrequencies.map((f) {
                                  return Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 10,
                                      vertical: 4,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.white,
                                      borderRadius: BorderRadius.circular(8),
                                      boxShadow: [
                                        BoxShadow(
                                          color: Colors.black.withOpacity(0.05),
                                          blurRadius: 4,
                                        ),
                                      ],
                                    ),
                                    child: Text(
                                      '${f.toStringAsFixed(2)} GHz',
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: Colors.blue.shade900,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  );
                                }).toList(),
                              ),
                            ),
                            const SizedBox(height: 20),
                          ],
                          if (node.type == NodeType.component) ...[
                            _buildFieldLabel('CALIBRATED CABLE'),
                            Row(
                              children: [
                                Expanded(
                                  child: DropdownButtonFormField<String>(
                                    value: selectedCable,
                                    decoration: InputDecoration(
                                      filled: true,
                                      fillColor: Colors.grey.shade50,
                                      border: OutlineInputBorder(
                                        borderRadius: BorderRadius.circular(12),
                                        borderSide: BorderSide.none,
                                      ),
                                    ),
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
                                    onChanged: (val) {
                                      setDialogState(() => selectedCable = val);
                                    },
                                  ),
                                ),
                                const SizedBox(width: 8),
                                IconButton(
                                  onPressed: () =>
                                      _showCableAssistant(node, (name, loss) {
                                        setDialogState(() {
                                          selectedCable = name;
                                          lossController.text = loss
                                              .toStringAsFixed(2);
                                        });
                                      }),
                                  icon: const Icon(
                                    Icons.auto_awesome,
                                    color: Colors.indigo,
                                  ),
                                  tooltip: 'Help Me Select (AI Assistant)',
                                ),
                              ],
                            ),
                            const SizedBox(height: 20),
                          ],
                          _buildFieldLabel('PATH LOSS (dB)'),
                          TextField(
                            controller: lossController,
                            keyboardType: const TextInputType.numberWithOptions(
                              decimal: true,
                              signed: true,
                            ),
                            decoration: InputDecoration(
                              suffixText: 'dB',
                              filled: true,
                              fillColor: Colors.grey.shade50,
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(12),
                                borderSide: BorderSide.none,
                              ),
                            ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            'Enter positive value for loss, negative for gain.',
                            style: TextStyle(
                              fontSize: 11,
                              color: Colors.grey.shade500,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  Padding(
                    padding: const EdgeInsets.all(24),
                    child: Row(
                      children: [
                        if (node.type != NodeType.source)
                          TextButton.icon(
                            onPressed: () {
                              final parent = _findParent(_root, node);
                              if (parent != null) {
                                _deleteNode(parent, node);
                                Navigator.pop(context);
                              }
                            },
                            icon: const Icon(Icons.delete_outline, size: 18),
                            label: const Text('Delete'),
                            style: TextButton.styleFrom(
                              foregroundColor: Colors.red,
                            ),
                          ),
                        const Spacer(),
                        ElevatedButton(
                          onPressed: () {
                            setState(() {
                              node.label = labelController.text;
                              node.lossDb =
                                  double.tryParse(lossController.text) ?? 0.0;
                              node.physicalResourceId = selectedCable;
                              if (node.physicalResourceId != null) {
                                _globalSync(
                                  node.physicalResourceId!,
                                  node.lossDb,
                                  node.label,
                                );
                              }
                            });
                            Navigator.pop(context);
                          },
                          style: ElevatedButton.styleFrom(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 32,
                              vertical: 16,
                            ),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12),
                            ),
                          ),
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

  List<String> _findAssetUsage(String assetId) {
    if (assetId.isEmpty) return [];
    List<String> usages = [];

    void checkPath(String pathId, PlannerNode node, bool isTx) {
      if (node.physicalResourceId == assetId) {
        String dir = isTx ? "TX" : "RX";
        usages.add("$dir: $pathId");
      }
      for (var child in node.children) {
        checkPath(pathId, child, isTx);
      }
    }

    if (_hubNode != null) checkPath("HUB", _hubNode!, true);
    return usages;
  }

  Set<double> _getBranchFrequencies(PlannerNode node) {
    final freqs = <double>{};

    // 1. Collect from self and all descendants
    void collectDown(PlannerNode n) {
      freqs.addAll(n.supportedFrequencies);
      for (var child in n.children) {
        collectDown(child);
      }
    }

    collectDown(node);

    // 2. Walk up to Hub to collect from ancestors (e.g. the Port node)
    PlannerNode? current = _findParent(_root, node);
    while (current != null && current != _hubNode) {
      freqs.addAll(current.supportedFrequencies);
      current = _findParent(_root, current);
    }

    return freqs;
  }

  void _showCableAssistant(
    PlannerNode node,
    Function(String name, double loss) onSelect,
  ) {
    final lengthController = TextEditingController(text: '1.0');
    final nameFilterController = TextEditingController();
    List<CableLossRecord> allRecords = [];
    List<Map<String, dynamic>> candidates = [];
    bool loading = false;

    // Aggregate frequencies from the entire branch (ancestors and descendants)
    final branchFreqs = _getBranchFrequencies(node);
    final List<double> targetFrequencies = branchFreqs.isNotEmpty
        ? branchFreqs.map((f) => f * 1000.0).toList()
        : [1000.0];

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setAltState) {
          void searchCables() async {
            setAltState(() => loading = true);
            final server = context.read<ServerService>();
            final response = await server.fetchCableMeasuredDetails();
            if (response != null && response.ok) {
              allRecords = response.history;
              final targetLength =
                  double.tryParse(lengthController.text) ?? 1.0;
              final nameQuery = nameFilterController.text.trim().toLowerCase();

              // Group by cable name and get latest record
              Map<String, CableLossRecord> uniqueCables = {};
              for (var rec in allRecords) {
                // Filter by exact length ONLY
                if (rec.length != targetLength) continue;

                // Filter by name if provided
                if (nameQuery.isNotEmpty &&
                    !rec.cableName.toLowerCase().contains(nameQuery))
                  continue;

                if (!uniqueCables.containsKey(rec.cableName)) {
                  uniqueCables[rec.cableName] = rec;
                }
              }

              final List<Map<String, dynamic>> results = [];
              for (var cable in uniqueCables.values) {
                if (cable.measurements.isEmpty) continue;

                cable.measurements.sort(
                  (a, b) => a.frequency.compareTo(b.frequency),
                );

                double totalLoss = 0;
                for (var f in targetFrequencies) {
                  double mLoss = 0;
                  if (f <= cable.measurements.first.frequency) {
                    mLoss = cable.measurements.first.loss;
                  } else if (f >= cable.measurements.last.frequency) {
                    mLoss = cable.measurements.last.loss;
                  } else {
                    for (int i = 0; i < cable.measurements.length - 1; i++) {
                      final p1 = cable.measurements[i];
                      final p2 = cable.measurements[i + 1];
                      if (f >= p1.frequency && f <= p2.frequency) {
                        final t =
                            (f - p1.frequency) / (p2.frequency - p1.frequency);
                        mLoss = p1.loss + t * (p2.loss - p1.loss);
                        break;
                      }
                    }
                  }
                  totalLoss += mLoss * -1.0; // Corrected to positive
                }

                final avgLoss = totalLoss / targetFrequencies.length;

                results.add({
                  'name': cable.cableName,
                  'loss': avgLoss,
                  'date': cable.date,
                  'originalLength': cable.length,
                  'usages': _findAssetUsage(cable.cableName),
                });
              }

              results.sort(
                (a, b) => (a['loss'] as double).compareTo(b['loss'] as double),
              );
              candidates = results.take(5).toList();
            }
            setAltState(() => loading = false);
          }

          return Dialog(
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
                    const SizedBox(height: 8),
                    Text(
                      'Find the best cable for your specific requirements',
                      style: TextStyle(
                        color: Colors.grey.shade600,
                        fontSize: 13,
                      ),
                    ),
                    const SizedBox(height: 24),
                    const Text(
                      'NAME FILTER (e.g. SMA)',
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 1.2,
                      ),
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: nameFilterController,
                      decoration: InputDecoration(
                        hintText: 'Optional substring filter...',
                        filled: true,
                        fillColor: Colors.grey.shade50,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide.none,
                        ),
                      ),
                    ),
                    const SizedBox(height: 20),
                    Row(
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'TARGET LENGTH',
                                style: TextStyle(
                                  fontSize: 10,
                                  fontWeight: FontWeight.bold,
                                  letterSpacing: 1.2,
                                ),
                              ),
                              const SizedBox(height: 8),
                              TextField(
                                controller: lengthController,
                                decoration: InputDecoration(
                                  suffixText: 'm',
                                  filled: true,
                                  fillColor: Colors.grey.shade50,
                                  border: OutlineInputBorder(
                                    borderRadius: BorderRadius.circular(12),
                                    borderSide: BorderSide.none,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'AUTO FREQUENCIES',
                                style: TextStyle(
                                  fontSize: 10,
                                  fontWeight: FontWeight.bold,
                                  letterSpacing: 1.2,
                                ),
                              ),
                              const SizedBox(height: 8),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 12,
                                  vertical: 12,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.blue.shade50,
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Text(
                                  '${targetFrequencies.length} Bands',
                                  style: TextStyle(
                                    fontSize: 13,
                                    color: Colors.blue.shade900,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        onPressed: loading ? null : searchCables,
                        icon: loading
                            ? const SizedBox(
                                width: 14,
                                height: 14,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: Colors.white,
                                ),
                              )
                            : const Icon(Icons.search, size: 18),
                        label: Text(
                          loading
                              ? 'Searching Database...'
                              : 'Find Best Matches',
                        ),
                      ),
                    ),
                    if (candidates.isNotEmpty) ...[
                      const SizedBox(height: 24),
                      const Text(
                        'TOP 5 CANDIDATES',
                        style: TextStyle(
                          fontSize: 10,
                          fontWeight: FontWeight.bold,
                          letterSpacing: 1.2,
                        ),
                      ),
                      const SizedBox(height: 12),
                      ...candidates.map((c) {
                        final List<String> usages = List<String>.from(
                          c['usages'] ?? [],
                        );
                        return Container(
                          margin: const EdgeInsets.only(bottom: 8),
                          decoration: BoxDecoration(
                            color: usages.isNotEmpty
                                ? Colors.orange.shade50.withOpacity(0.3)
                                : Colors.grey.shade50,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: usages.isNotEmpty
                                  ? Colors.orange.shade200
                                  : Colors.grey.shade200,
                            ),
                          ),
                          child: ListTile(
                            title: Row(
                              children: [
                                Text(
                                  c['name'],
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                if (usages.isNotEmpty) ...[
                                  const SizedBox(width: 8),
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 6,
                                      vertical: 2,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.orange.shade100,
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                    child: const Text(
                                      'ALREADY USED',
                                      style: TextStyle(
                                        fontSize: 8,
                                        fontWeight: FontWeight.bold,
                                        color: Colors.orange,
                                      ),
                                    ),
                                  ),
                                ],
                              ],
                            ),
                            subtitle: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  'Last Calibrated: ${c['date']}',
                                  style: const TextStyle(fontSize: 11),
                                ),
                                if (usages.isNotEmpty)
                                  Padding(
                                    padding: const EdgeInsets.only(top: 4),
                                    child: Text(
                                      'Used in: ${usages.join(", ")}',
                                      style: TextStyle(
                                        fontSize: 10,
                                        color: Colors.orange.shade800,
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                  ),
                              ],
                            ),
                            trailing: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                Text(
                                  '${(c['loss'] as double).toStringAsFixed(3)} dB',
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                    color: Colors.indigo,
                                  ),
                                ),
                                const Text(
                                  'Loss @ Freq',
                                  style: TextStyle(fontSize: 9),
                                ),
                              ],
                            ),
                            onTap: () {
                              onSelect(c['name'], c['loss']);
                              Navigator.pop(context);
                            },
                          ),
                        );
                      }),
                    ],
                    const SizedBox(height: 24),
                    TextButton(
                      onPressed: () => Navigator.pop(context),
                      child: const Text('Close Assistant'),
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildFieldLabel(String label) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0, left: 4),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.bold,
          color: Colors.grey.shade600,
          letterSpacing: 1.2,
        ),
      ),
    );
  }
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
