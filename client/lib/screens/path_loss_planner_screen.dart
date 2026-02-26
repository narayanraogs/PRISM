import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';
import 'package:prism_client/utils/notifications.dart';

enum NodeType { source, component, branching, instrument }

class PlannerNode {
  final String id;
  String label;
  double lossDb;
  NodeType type;
  String? physicalResourceId; // For shared assets
  List<PlannerNode> children;

  PlannerNode({
    required this.id,
    required this.label,
    this.lossDb = 0.0,
    required this.type,
    this.physicalResourceId,
    List<PlannerNode>? children,
  }) : children = children ?? [];

  PlannerNode copyWith({
    String? label,
    double? lossDb,
    NodeType? type,
    String? physicalResourceId,
  }) {
    return PlannerNode(
      id: id,
      label: label ?? this.label,
      lossDb: lossDb ?? this.lossDb,
      type: type ?? this.type,
      physicalResourceId: physicalResourceId ?? this.physicalResourceId,
      children: children,
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
  bool _isSpacecraftTx = true; // Toggle between TX and RX
  final Map<String, PlannerNode> _txPaths = {}; // Spacecraft TX Paths
  final Map<String, PlannerNode> _rxPaths = {}; // Spacecraft RX Paths
  String _activePathId = 'Primary Link';

  // Cache for calibrated cables from server
  List<String> _calibratedCables = [];

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
    final metadata = server.status.bootstrapData?.cableLossData;
    if (metadata != null) {
      setState(() {
        _calibratedCables = metadata.existingCables;
      });
    }
  }

  PlannerNode get _root =>
      _isSpacecraftTx ? _txPaths[_activePathId]! : _rxPaths[_activePathId]!;

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
    for (var path in _txPaths.values) {
      _syncSharedNodes(path, physicalId, loss, label);
    }
    for (var path in _rxPaths.values) {
      _syncSharedNodes(path, physicalId, loss, label);
    }
  }

  void _addNode(PlannerNode parent, NodeType type, {bool insert = false}) {
    setState(() {
      final newNode = PlannerNode(
        id: 'node-${DateTime.now().millisecondsSinceEpoch}',
        label: 'New ${type.name}',
        type: type,
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
    if (_txPaths.isNotEmpty || _rxPaths.isNotEmpty) return;

    _txPaths['Primary Link'] = PlannerNode(
      id: 'root-tx-1',
      label: 'Spacecraft Port (Ka-Band)',
      type: NodeType.source,
      children: [
        PlannerNode(
          id: 'trunk-1',
          label: 'Main Trunk Cable',
          lossDb: 1.5,
          physicalResourceId: 'COMMON-TRUNK-01',
          type: NodeType.component,
          children: [
            PlannerNode(
              id: 'split-1',
              label: '3-Way Splitter',
              lossDb: 4.8,
              type: NodeType.branching,
              children: [
                PlannerNode(
                  id: 'term-pm',
                  label: 'Power Meter',
                  type: NodeType.instrument,
                ),
                PlannerNode(
                  id: 'term-sa',
                  label: 'Spectrum Analyzer',
                  type: NodeType.instrument,
                ),
              ],
            ),
          ],
        ),
      ],
    );

    _txPaths['Secondary Link'] = PlannerNode(
      id: 'root-tx-2',
      label: 'Spacecraft Port (S-Band)',
      type: NodeType.source,
      children: [
        PlannerNode(
          id: 'trunk-shared',
          label: 'Main Trunk Cable',
          lossDb: 1.5,
          physicalResourceId: 'COMMON-TRUNK-01',
          type: NodeType.component,
          children: [
            PlannerNode(
              id: 'term-data-tx',
              label: 'Data Reception',
              type: NodeType.instrument,
            ),
          ],
        ),
      ],
    );

    _rxPaths['Primary Link'] = PlannerNode(
      id: 'root-rx-1',
      label: 'Ground Transmitter',
      type: NodeType.source,
      children: [
        PlannerNode(
          id: 'trunk-rx-1',
          label: 'TX Main Line',
          lossDb: 2.0,
          type: NodeType.component,
          children: [
            PlannerNode(
              id: 'split-rx-1',
              label: 'Directional Coupler',
              lossDb: 1.0,
              type: NodeType.branching,
              children: [
                PlannerNode(
                  id: 'term-sc',
                  label: 'Spacecraft Port (Ka)',
                  type: NodeType.instrument,
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }

  Map<String, double> _getTerminalLosses(PlannerNode node, double currentLoss) {
    double total = currentLoss + node.lossDb;
    if (node.type == NodeType.instrument) {
      return {node.label: total};
    }
    Map<String, double> results = {};
    for (var child in node.children) {
      results.addAll(_getTerminalLosses(child, total));
    }
    return results;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final terminalLosses = _getTerminalLosses(_root, 0.0);

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
                  child: _buildSummaryPanel(theme, terminalLosses),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(ThemeData theme) {
    final currentPaths = _isSpacecraftTx ? _txPaths : _rxPaths;

    return ScreenHeader(
      title: 'Path Loss Planner',
      subtitle: 'Visualize and calculate RF link budgets',
      icon: Icons.map_outlined,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildDirectionToggle(theme),
          const SizedBox(width: 16),
          _buildPathSelector(theme, currentPaths),
          const SizedBox(width: 16),
          ElevatedButton.icon(
            onPressed: () {
              setState(() {
                _txPaths.clear();
                _rxPaths.clear();
                _initializeDefaultTree();
                _activePathId = 'Primary Link';
                _transformationController.value = Matrix4.identity();
              });
              AppNotifications.show(
                context,
                'Planner Reset',
                type: NotificationType.info,
              );
            },
            icon: const Icon(Icons.restart_alt),
            label: const Text('Reset'),
          ),
        ],
      ),
    );
  }

  Widget _buildPathSelector(ThemeData theme, Map<String, PlannerNode> paths) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: theme.colorScheme.primary.withOpacity(0.2)),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          value: _activePathId,
          icon: const Icon(Icons.keyboard_arrow_down, size: 18),
          style: GoogleFonts.inter(
            color: theme.colorScheme.primary,
            fontWeight: FontWeight.bold,
            fontSize: 13,
          ),
          items: [
            ...paths.keys.map(
              (k) => DropdownMenuItem(value: k, child: Text(k)),
            ),
            const DropdownMenuItem(
              value: 'ADD_NEW',
              child: Row(
                children: [
                  Icon(Icons.add, size: 14, color: Colors.blue),
                  SizedBox(width: 8),
                  Text('Add New Port', style: TextStyle(color: Colors.blue)),
                ],
              ),
            ),
          ],
          onChanged: (val) {
            if (val == 'ADD_NEW') {
              _showAddPathDialog();
            } else if (val != null) {
              setState(() => _activePathId = val);
            }
          },
        ),
      ),
    );
  }

  Widget _buildDirectionToggle(ThemeData theme) {
    return Container(
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          _buildToggleButton(
            'Spacecraft TX (Downlink)',
            _isSpacecraftTx,
            () => setState(() {
              _isSpacecraftTx = true;
              if (!_txPaths.containsKey(_activePathId)) {
                _activePathId = _txPaths.keys.first;
              }
            }),
            theme,
          ),
          _buildToggleButton(
            'Spacecraft RX (Uplink)',
            !_isSpacecraftTx,
            () => setState(() {
              _isSpacecraftTx = false;
              if (!_rxPaths.containsKey(_activePathId)) {
                _activePathId = _rxPaths.keys.first;
              }
            }),
            theme,
          ),
        ],
      ),
    );
  }

  Widget _buildToggleButton(
    String label,
    bool isSelected,
    VoidCallback onTap,
    ThemeData theme,
  ) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.05),
                    blurRadius: 4,
                    offset: const Offset(0, 2),
                  ),
                ]
              : [],
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 13,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
            color: isSelected
                ? theme.colorScheme.primary
                : Colors.grey.shade600,
          ),
        ),
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
                    left: 100,
                    top: 200,
                    child: _buildNodeTree(_root, theme),
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

  Widget _buildNodeTree(PlannerNode node, ThemeData theme) {
    if (node.children.isEmpty) {
      return _buildNodeCard(node, theme);
    }

    // Optimization: If only one child, skip IntrinsicHeight and VerticalDivider
    if (node.children.length == 1) {
      return Row(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          _buildNodeCard(node, theme),
          Container(
            width: 40,
            height: 2,
            color: theme.colorScheme.primary.withOpacity(0.3),
          ),
          _buildNodeTree(node.children.first, theme),
        ],
      );
    }

    return Row(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        _buildNodeCard(node, theme),
        Container(
          width: 30,
          height: 2,
          color: theme.colorScheme.primary.withOpacity(0.3),
        ),
        IntrinsicHeight(
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              VerticalDivider(
                color: theme.colorScheme.primary.withOpacity(0.3),
                thickness: 2,
                width: 2,
              ),
              Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: node.children.map((child) {
                  return Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 30,
                        height: 2,
                        color: theme.colorScheme.primary.withOpacity(0.3),
                      ),
                      Padding(
                        padding: const EdgeInsets.symmetric(vertical: 20),
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
    );
  }

  Widget _buildNodeCard(PlannerNode node, ThemeData theme) {
    Color cardColor;
    IconData icon;
    switch (node.type) {
      case NodeType.source:
        cardColor = theme.colorScheme.primary;
        icon = _isSpacecraftTx
            ? Icons.satellite_alt
            : Icons.settings_input_antenna;
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

  Widget _buildSummaryPanel(ThemeData theme, Map<String, double> losses) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Live Path Summary',
          style: GoogleFonts.outfit(fontSize: 18, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 4),
        Text(
          'Calculated end-to-end losses',
          style: TextStyle(color: Colors.grey.shade500, fontSize: 13),
        ),
        const SizedBox(height: 24),
        ...losses.entries
            .map((e) => _buildSummaryItem(theme, e.key, e.value))
            .toList(),
        const Spacer(),
        _buildActionButtons(theme),
      ],
    );
  }

  Widget _buildSummaryItem(ThemeData theme, String label, double loss) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: theme.colorScheme.primary.withOpacity(0.05),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: theme.colorScheme.primary.withOpacity(0.1)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Expanded(
            child: Text(
              label,
              style: const TextStyle(fontWeight: FontWeight.w500, fontSize: 13),
            ),
          ),
          Text(
            '${loss.toStringAsFixed(2)} dB',
            style: TextStyle(
              fontWeight: FontWeight.bold,
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
                          if (node.type == NodeType.component) ...[
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
                                      _showCableAssistant((name, loss) {
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

    _txPaths.forEach((id, root) => checkPath(id, root, true));
    _rxPaths.forEach((id, root) => checkPath(id, root, false));
    return usages;
  }

  void _showCableAssistant(Function(String name, double loss) onSelect) {
    final lengthController = TextEditingController(text: '1.0');
    final freqController = TextEditingController(text: '1000');
    final nameFilterController = TextEditingController();
    List<CableLossRecord> allRecords = [];
    List<Map<String, dynamic>> candidates = [];
    bool loading = false;

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
              final targetFreq = double.tryParse(freqController.text) ?? 1000.0;
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

                // Simple interpolation
                cable.measurements.sort(
                  (a, b) => a.frequency.compareTo(b.frequency),
                );
                double measuredLoss = 0;
                if (targetFreq <= cable.measurements.first.frequency) {
                  measuredLoss = cable.measurements.first.loss;
                } else if (targetFreq >= cable.measurements.last.frequency) {
                  measuredLoss = cable.measurements.last.loss;
                } else {
                  for (int i = 0; i < cable.measurements.length - 1; i++) {
                    final p1 = cable.measurements[i];
                    final p2 = cable.measurements[i + 1];
                    if (targetFreq >= p1.frequency &&
                        targetFreq <= p2.frequency) {
                      final t =
                          (targetFreq - p1.frequency) /
                          (p2.frequency - p1.frequency);
                      measuredLoss = p1.loss + t * (p2.loss - p1.loss);
                      break;
                    }
                  }
                }

                // In metadata loss is negative, we want positive for the planner
                final correctedLoss = measuredLoss * -1.0;

                results.add({
                  'name': cable.cableName,
                  'loss': correctedLoss,
                  'date': cable.date,
                  'originalLength': cable.length,
                  'usages': _findAssetUsage(cable.cableName),
                });
              }

              results.sort(
                (a, b) => (a['loss'] as double).abs().compareTo(
                  (b['loss'] as double).abs(),
                ),
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
                                'FREQUENCY',
                                style: TextStyle(
                                  fontSize: 10,
                                  fontWeight: FontWeight.bold,
                                  letterSpacing: 1.2,
                                ),
                              ),
                              const SizedBox(height: 8),
                              TextField(
                                controller: freqController,
                                decoration: InputDecoration(
                                  suffixText: 'MHz',
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

  void _showAddPathDialog() {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add New Port/Link'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(hintText: 'e.g. Ka-Band Secondary'),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () {
              if (controller.text.isNotEmpty) {
                setState(() {
                  final name = controller.text;
                  final path = _isSpacecraftTx ? _txPaths : _rxPaths;
                  path[name] = PlannerNode(
                    id: 'root-${DateTime.now().millisecondsSinceEpoch}',
                    label: _isSpacecraftTx
                        ? 'Spacecraft Port ($name)'
                        : 'Ground Transmitter',
                    type: NodeType.source,
                  );
                  _activePathId = name;
                });
                Navigator.pop(context);
              }
            },
            child: const Text('Create'),
          ),
        ],
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
