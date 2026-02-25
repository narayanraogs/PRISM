import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/utils/notifications.dart';
import 'package:prism_client/screens/test_progress_screen.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';

class TestScreen extends StatefulWidget {
  final bool isActive;
  const TestScreen({super.key, this.isActive = true});

  @override
  State<TestScreen> createState() => _TestScreenState();
}

class _TestScreenState extends State<TestScreen> {
  @override
  void didUpdateWidget(TestScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    // No WebSocket to close here yet, but added for consistency with IndexedStack pattern
  }

  int _selectedCategoryIndex = 0;
  bool _isHelpOpen = false;
  AllTests? _allTests;
  bool _isLoading = true;
  String? _selectedConfig;
  final Set<TestDescription> _selectedTests = {};
  final TextEditingController _remarkController = TextEditingController();
  bool _isStarting = false;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  void _fetchData() {
    setState(() => _isLoading = true);
    final serverService = Provider.of<ServerService>(context, listen: false);

    // Use bootstrapped data
    final tests = serverService.status.bootstrapData?.testData;

    if (tests != null) {
      debugPrint('TestScreen: Using Bootstrapped Metadata');
      setState(() {
        _allTests = tests;
        if (tests.categories.isNotEmpty) {
          _selectedCategoryIndex = 0;
          final category = tests.categories[0];
          if (tests.configurations[category]?.isNotEmpty ?? false) {
            _selectedConfig = tests.configurations[category]![0];
          }
        }
        _isLoading = false;
      });
    } else {
      debugPrint('TestScreen: Bootstrapped Metadata NOT FOUND');
      setState(() => _isLoading = false);
    }
  }

  void _onCategorySelected(int index) {
    setState(() {
      _selectedCategoryIndex = index;
      final category = _allTests!.categories[index];
      if (_allTests!.configurations[category]?.isNotEmpty ?? false) {
        _selectedConfig = _allTests!.configurations[category]![0];
      } else {
        _selectedConfig = null;
      }
      _selectedTests.clear();
    });
  }

  void _onConfigChanged(String? config) {
    setState(() {
      _selectedConfig = config;
      _selectedTests.clear();
    });
  }

  Future<void> _startTest() async {
    if (_selectedTests.isEmpty) {
      AppNotifications.showError(context, 'Please select at least one test');
      return;
    }

    final tests = _selectedTests.map((t) {
      return TestDescription(
        testName: t.testName,
        testCategory: t.testCategory,
        configuration: _selectedConfig,
        remark: _remarkController.text,
      );
    }).toList();

    await Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => TestProgressScreen(tests: tests)),
    );

    if (mounted) {
      setState(() {
        _selectedTests.clear();
        _remarkController.clear();
      });
    }
  }

  void _selectAll() {
    if (_selectedConfig == null) return;
    final tests = _allTests!.tests[_selectedConfig] ?? [];
    setState(() {
      _selectedTests.addAll(tests);
    });
  }

  void _clearSelection() {
    setState(() {
      _selectedTests.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildHeader(theme),
                const SizedBox(height: 32),
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSidebar(theme),
                      const SizedBox(width: 32),
                      _buildMainContent(theme),
                    ],
                  ),
                ),
              ],
            ),
          ),
          if (_isHelpOpen)
            GestureDetector(
              onTap: () => setState(() => _isHelpOpen = false),
              child: Container(color: Colors.black.withOpacity(0.1)),
            ),
          AnimatedPositioned(
            duration: const Duration(milliseconds: 300),
            curve: Curves.easeInOut,
            right: _isHelpOpen ? 0 : -400,
            top: 0,
            bottom: 0,
            child: _buildHelpPanel(theme),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(ThemeData theme) {
    return ScreenHeader(
      title: 'Test Executive',
      subtitle: _selectedTests.isEmpty
          ? 'Select category, configuration and tests to begin'
          : 'Selected ${_selectedTests.length} tests ready for execution',
      icon: Icons.assignment_outlined,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [_buildHelpTrigger(theme)],
      ),
    );
  }

  Widget _buildSidebar(ThemeData theme) {
    if (_isLoading) return const SizedBox(width: 280);
    final categories = _allTests?.categories ?? [];

    return ContentCard(
      width: 280,
      isSidebar: true,
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Row(
              children: [
                Text(
                  'CATEGORIES',
                  style: GoogleFonts.inter(
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                    color: theme.colorScheme.primary,
                    letterSpacing: 1.5,
                  ),
                ),
                const Spacer(),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 2,
                  ),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    categories.length.toString(),
                    style: TextStyle(
                      color: theme.colorScheme.primary,
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.all(12),
              itemCount: categories.length,
              separatorBuilder: (context, index) => const SizedBox(height: 4),
              itemBuilder: (context, index) {
                final isSelected = _selectedCategoryIndex == index;
                final category = categories[index];
                return InkWell(
                  onTap: () => _onCategorySelected(index),
                  borderRadius: BorderRadius.circular(16),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 16,
                    ),
                    decoration: BoxDecoration(
                      color: isSelected
                          ? theme.colorScheme.primary.withOpacity(0.08)
                          : Colors.transparent,
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: isSelected
                            ? theme.colorScheme.primary.withOpacity(0.2)
                            : Colors.transparent,
                      ),
                    ),
                    child: Row(
                      children: [
                        Icon(
                          isSelected ? Icons.folder : Icons.folder_open,
                          color: isSelected
                              ? theme.colorScheme.primary
                              : Colors.grey.shade400,
                          size: 20,
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Text(
                            category,
                            style: GoogleFonts.inter(
                              fontWeight: isSelected
                                  ? FontWeight.w900
                                  : FontWeight.bold,
                              color: isSelected
                                  ? theme.colorScheme.primary
                                  : Colors.black,
                              fontSize: 16,
                            ),
                          ),
                        ),
                        if (isSelected)
                          Icon(
                            Icons.chevron_right,
                            color: theme.colorScheme.primary,
                            size: 14,
                          ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMainContent(ThemeData theme) {
    if (_isLoading) {
      return Expanded(
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const CircularProgressIndicator(),
              const SizedBox(height: 24),
              Text(
                'Loading tests...',
                style: GoogleFonts.inter(
                  color: Colors.grey.shade600,
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      );
    }

    if (_allTests == null || _allTests!.categories.isEmpty) {
      return const Expanded(child: Center(child: Text('No tests available')));
    }

    final category = _allTests!.categories[_selectedCategoryIndex];
    final configs = _allTests!.configurations[category] ?? [];

    return Expanded(
      child: ContentCard(
        isSidebar: false,
        padding: EdgeInsets.zero,
        child: Column(
          children: [
            // Top Bar: Config Selection
            Padding(
              padding: const EdgeInsets.all(40.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'SELECT CONFIGURATION',
                              style: GoogleFonts.inter(
                                fontWeight: FontWeight.w900,
                                fontSize: 12,
                                color: theme.colorScheme.primary,
                                letterSpacing: 1.5,
                              ),
                            ),
                            const SizedBox(height: 16),
                            DropdownButtonFormField<String>(
                              decoration: InputDecoration(
                                prefixIcon: const Icon(Icons.settings_outlined),
                                border: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                contentPadding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 16,
                                ),
                              ),
                              value: _selectedConfig,
                              items: configs.map((c) {
                                return DropdownMenuItem(
                                  value: c,
                                  child: Text(c),
                                );
                              }).toList(),
                              onChanged: _onConfigChanged,
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 40),
                      Container(
                        width: 1,
                        height: 80,
                        color: Colors.grey.shade100,
                      ),
                      const SizedBox(width: 40),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'SUMMARY',
                                  style: GoogleFonts.inter(
                                    fontWeight: FontWeight.w900,
                                    fontSize: 12,
                                    color: theme.colorScheme.primary,
                                    letterSpacing: 1.5,
                                  ),
                                ),
                                if (_selectedConfig != null)
                                  Row(
                                    children: [
                                      TextButton(
                                        onPressed: _selectAll,
                                        style: TextButton.styleFrom(
                                          visualDensity: VisualDensity.compact,
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 8,
                                          ),
                                        ),
                                        child: const Text('SELECT ALL'),
                                      ),
                                      const Text(
                                        '|',
                                        style: TextStyle(color: Colors.grey),
                                      ),
                                      TextButton(
                                        onPressed: _clearSelection,
                                        style: TextButton.styleFrom(
                                          visualDensity: VisualDensity.compact,
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 8,
                                          ),
                                        ),
                                        child: const Text('CLEAR'),
                                      ),
                                    ],
                                  ),
                              ],
                            ),
                            const SizedBox(height: 16),
                            _buildSummaryItem(
                              'Selected Tests',
                              _selectedTests.length.toString(),
                              theme,
                            ),
                            if (_selectedConfig != null &&
                                _allTests!.losses[_selectedConfig] != null &&
                                _allTests!
                                    .losses[_selectedConfig]!
                                    .isNotEmpty) ...[
                              const SizedBox(height: 12),
                              _buildSummaryItem(
                                'Current Losses',
                                _allTests!.losses[_selectedConfig]!,
                                theme,
                              ),
                            ],
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            // Test Grid
            Expanded(
              child: _selectedConfig == null
                  ? const Center(
                      child: Text('Select a configuration to see tests'),
                    )
                  : _buildTestGrid(theme),
            ),
            const Divider(height: 1),
            // Bottom Bar: Remark & Start
            Padding(
              padding: const EdgeInsets.all(32.0),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _remarkController,
                      decoration: InputDecoration(
                        hintText: 'Add a remark (Optional)',
                        prefixIcon: const Icon(Icons.note_alt_outlined),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide(color: Colors.grey.shade200),
                        ),
                        filled: true,
                        fillColor: Colors.grey.shade50,
                      ),
                    ),
                  ),
                  const SizedBox(width: 24),
                  ElevatedButton.icon(
                    onPressed: _isStarting ? null : _startTest,
                    icon: _isStarting
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Icon(Icons.play_arrow_rounded),
                    label: const Text('START TEST'),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 32,
                        vertical: 20,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      elevation: 0,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSummaryItem(String label, String value, ThemeData theme) {
    return Row(
      children: [
        Text(
          label,
          style: TextStyle(color: Colors.grey.shade500, fontSize: 13),
        ),
        const Spacer(),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
          decoration: BoxDecoration(
            color: theme.colorScheme.primary.withOpacity(0.1),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Text(
            value,
            style: TextStyle(
              color: theme.colorScheme.primary,
              fontWeight: FontWeight.bold,
              fontSize: 14,
            ),
            softWrap: false,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }

  Widget _buildTestGrid(ThemeData theme) {
    final tests = _allTests!.tests[_selectedConfig] ?? [];

    if (tests.isEmpty) {
      return const Center(child: Text('No tests found for this configuration'));
    }

    return GridView.builder(
      padding: const EdgeInsets.all(32),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 5.5,
        mainAxisSpacing: 12,
        crossAxisSpacing: 12,
      ),
      itemCount: tests.length,
      itemBuilder: (context, index) {
        final test = tests[index];
        final isSelected = _selectedTests.contains(test);

        return InkWell(
          onTap: () {
            setState(() {
              if (isSelected) {
                _selectedTests.remove(test);
              } else {
                _selectedTests.add(test);
              }
            });
          },
          borderRadius: BorderRadius.circular(16),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
            decoration: BoxDecoration(
              color: isSelected
                  ? theme.colorScheme.primary.withOpacity(0.05)
                  : Colors.white,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: isSelected
                    ? theme.colorScheme.primary.withOpacity(0.3)
                    : Colors.grey.shade200,
                width: isSelected ? 2 : 1,
              ),
            ),
            child: Row(
              children: [
                Container(
                  width: 24,
                  height: 24,
                  decoration: BoxDecoration(
                    color: isSelected
                        ? theme.colorScheme.primary
                        : Colors.white,
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(
                      color: isSelected
                          ? theme.colorScheme.primary
                          : Colors.grey.shade400,
                    ),
                  ),
                  child: isSelected
                      ? const Icon(Icons.check, color: Colors.white, size: 16)
                      : null,
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        test.testName,
                        style: GoogleFonts.inter(
                          fontWeight: isSelected
                              ? FontWeight.w900
                              : FontWeight.bold,
                          color: isSelected
                              ? theme.colorScheme.primary
                              : Colors.black,
                          fontSize: 16,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      if (test.testCategory.isNotEmpty)
                        Text(
                          test.testCategory,
                          style: TextStyle(
                            color: isSelected
                                ? theme.colorScheme.primary
                                : Colors.black87,
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      },
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
      width: 400,
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
                  'Test Executive Help',
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
                  'How to use',
                  '1. Select a Category from the sidebar.\n'
                      '2. Choose a Configuration from the dropdown.\n'
                      '3. Select one or more tests from the grid.\n'
                      '4. Add an optional remark.\n'
                      '5. Click "START TEST" to execute selection.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Multi-selection',
                  'You can select multiple tests across different groups if they belong to the same configuration.',
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
