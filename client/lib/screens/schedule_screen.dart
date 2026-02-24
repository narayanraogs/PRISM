import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/utils/notifications.dart';
import 'package:prism_client/screens/test_progress_screen.dart';
import 'dart:convert';
import 'dart:js_interop';
import 'package:web/web.dart' as web;
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';

class ScheduleScreen extends StatefulWidget {
  const ScheduleScreen({super.key});

  @override
  State<ScheduleScreen> createState() => _ScheduleScreenState();
}

class _ScheduleScreenState extends State<ScheduleScreen> {
  int _selectedCategoryIndex = 0;
  bool _isHelpOpen = false;
  AllTests? _allTests;
  bool _isLoading = true;
  String? _selectedConfig;
  final Set<TestDescription> _selectedTests = {};
  final TextEditingController _remarkController = TextEditingController();
  final TextEditingController _scheduleNameController = TextEditingController(
    text: 'Current Schedule',
  );

  final List<TestDescription> _scheduledTests = [];
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
      debugPrint('ScheduleScreen: Using Bootstrapped Metadata');
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
      debugPrint('ScheduleScreen: Bootstrapped Metadata NOT FOUND');
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

  void _addToSchedule() {
    if (_selectedTests.isEmpty) {
      AppNotifications.showError(context, 'Please select at least one test');
      return;
    }
    if (_selectedConfig == null) {
      AppNotifications.showError(context, 'Please select a configuration');
      return;
    }

    setState(() {
      for (var test in _selectedTests) {
        _scheduledTests.add(
          TestDescription(
            testName: test.testName,
            testCategory: test.testCategory,
            configuration: _selectedConfig,
            remark: _remarkController.text,
          ),
        );
      }
      _selectedTests.clear();
      _remarkController.clear();
    });
    AppNotifications.showSuccess(context, 'Added to schedule');
  }

  void _addPause() {
    if (_scheduledTests.isNotEmpty &&
        _scheduledTests.last.testName == 'Pause') {
      AppNotifications.showError(context, 'Consecutive pauses are not allowed');
      return;
    }
    setState(() {
      _scheduledTests.add(
        TestDescription(
          testName: 'Pause',
          testCategory: '',
          extraParameters: ['Type:Pause'],
        ),
      );
    });
  }

  void _saveSchedule() {
    if (_scheduledTests.isEmpty) {
      AppNotifications.showError(context, 'Schedule is empty');
      return;
    }

    final data = {
      'scheduleName': _scheduleNameController.text,
      'tests': _scheduledTests.map((e) => e.toJson()).toList(),
    };
    final jsonContent = jsonEncode(data);
    final bytes = utf8.encode(jsonContent);
    final blob = web.Blob(
      [bytes.toJS].toJS,
      web.BlobPropertyBag(type: 'application/json'),
    );
    final url = web.URL.createObjectURL(blob);
    final anchor = web.document.createElement('a') as web.HTMLAnchorElement;
    anchor.href = url;
    anchor.download =
        "${_scheduleNameController.text.trim().replaceAll(' ', '_')}.json";
    anchor.click();
    web.URL.revokeObjectURL(url);
    AppNotifications.showSuccess(context, 'Schedule exported');
  }

  void _loadSchedule() {
    final uploadInput =
        web.document.createElement('input') as web.HTMLInputElement;
    uploadInput.type = 'file';
    uploadInput.accept = '.json';
    uploadInput.click();

    uploadInput.onChange.listen((e) {
      final files = uploadInput.files;
      if (files == null || files.length == 0) return;
      final file = files.item(0)!;

      final reader = web.FileReader();
      reader.readAsText(file);
      reader.onLoadEnd.listen((e) {
        try {
          final result = reader.result as JSString;
          final dynamic decoded = jsonDecode(result.toDart);

          setState(() {
            if (decoded is Map && decoded.containsKey('tests')) {
              // New format with metadata
              _scheduleNameController.text =
                  decoded['scheduleName'] ?? 'Loaded Schedule';
              final List tests = decoded['tests'];
              _scheduledTests.clear();
              _scheduledTests.addAll(
                tests.map((e) => TestDescription.fromJson(e)).toList(),
              );
            } else if (decoded is List) {
              // Legacy format or raw list
              _scheduleNameController.text = file.name
                  .replaceAll('.json', '')
                  .replaceAll('_', ' ');
              _scheduledTests.clear();
              _scheduledTests.addAll(
                decoded.map((e) => TestDescription.fromJson(e)).toList(),
              );
            }
          });
          AppNotifications.showSuccess(context, 'Schedule loaded');
        } catch (e) {
          AppNotifications.showError(
            context,
            'Failed to load schedule: Invalid JSON',
          );
        }
      });
    });
  }

  void _removeFromSchedule(int index) {
    setState(() {
      _scheduledTests.removeAt(index);
    });
  }

  void _startSchedule() {
    if (_scheduledTests.isEmpty) {
      AppNotifications.showError(context, 'Schedule is empty');
      return;
    }

    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => TestProgressScreen(tests: _scheduledTests),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildHeader(theme),
                const SizedBox(height: 24),
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSidebar(theme),
                      const SizedBox(width: 24),
                      _buildSelectionContent(theme),
                      const SizedBox(width: 24),
                      _buildSchedulePanel(theme),
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
            right: _isHelpOpen ? 0 : -450,
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
      title: 'Test Scheduler',
      subtitle: 'Build and manage automatically executed test sequences',
      icon: Icons.schedule_rounded,
      trailing: _buildHelpTrigger(theme),
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
                return ListTile(
                  onTap: () => _onCategorySelected(index),
                  selected: isSelected,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  leading: Icon(
                    isSelected ? Icons.folder : Icons.folder_open,
                    size: 18,
                  ),
                  title: Text(
                    category,
                    style: GoogleFonts.inter(
                      fontSize: 16,
                      fontWeight: isSelected
                          ? FontWeight.w900
                          : FontWeight.bold,
                      color: isSelected
                          ? theme.colorScheme.primary
                          : Colors.black,
                    ),
                  ),
                  selectedTileColor: theme.colorScheme.primary.withOpacity(
                    0.08,
                  ),
                  selectedColor: theme.colorScheme.primary,
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSelectionContent(ThemeData theme) {
    if (_isLoading) {
      return const Expanded(child: Center(child: CircularProgressIndicator()));
    }

    if (_allTests == null || _allTests!.categories.isEmpty) {
      return const Expanded(child: Center(child: Text('No tests available')));
    }

    final category = _allTests!.categories[_selectedCategoryIndex];
    final configs = _allTests!.configurations[category] ?? [];

    return Expanded(
      flex: 3,
      child: ContentCard(
        isSidebar: false,
        padding: EdgeInsets.zero,
        child: Column(
          children: [
            // Top Bar: Config Selection
            Padding(
              padding: const EdgeInsets.all(32.0),
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
                      return DropdownMenuItem(value: c, child: Text(c));
                    }).toList(),
                    onChanged: _onConfigChanged,
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
            // Bottom Bar: Remark & Add
            Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _remarkController,
                          decoration: InputDecoration(
                            hintText: 'Add a remark (Optional)',
                            prefixIcon: const Icon(Icons.note_alt_outlined),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(12),
                            ),
                            filled: true,
                            fillColor: Colors.grey.shade50,
                          ),
                        ),
                      ),
                      const SizedBox(width: 16),
                      TextButton.icon(
                        onPressed: _addPause,
                        icon: const Icon(Icons.pause_circle_outline),
                        label: const Text('ADD PAUSE'),
                        style: TextButton.styleFrom(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 20,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      onPressed: _addToSchedule,
                      icon: const Icon(Icons.add_task),
                      label: const Text('ADD SELECTION TO SCHEDULE'),
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 20),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                      ),
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

  Widget _buildTestGrid(ThemeData theme) {
    final tests = _allTests!.tests[_selectedConfig] ?? [];
    return GridView.builder(
      padding: const EdgeInsets.all(16),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 4,
        mainAxisSpacing: 8,
        crossAxisSpacing: 8,
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
          borderRadius: BorderRadius.circular(12),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: isSelected
                  ? theme.colorScheme.primary.withOpacity(0.05)
                  : Colors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: isSelected
                    ? theme.colorScheme.primary
                    : Colors.grey.shade200,
              ),
            ),
            child: Row(
              children: [
                Icon(
                  isSelected ? Icons.check_box : Icons.check_box_outline_blank,
                  color: isSelected
                      ? theme.colorScheme.primary
                      : Colors.grey.shade400,
                  size: 20,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        test.testName,
                        style: GoogleFonts.inter(
                          fontSize: 16,
                          fontWeight: isSelected
                              ? FontWeight.w900
                              : FontWeight.bold,
                          color: isSelected
                              ? theme.colorScheme.primary
                              : Colors.black,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        test.testCategory,
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w500,
                          color: isSelected
                              ? theme.colorScheme.primary
                              : Colors.grey.shade600,
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

  Widget _buildSchedulePanel(ThemeData theme) {
    return ContentCard(
      width: 450,
      isSidebar: true, // Auxiliary panel
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(20.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _scheduleNameController,
                    decoration: const InputDecoration(
                      isDense: true,
                      contentPadding: EdgeInsets.zero,
                      border: InputBorder.none,
                      hintText: 'Schedule Name',
                    ),
                    style: GoogleFonts.inter(
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                      color: theme.colorScheme.primary,
                      letterSpacing: 1.5,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 2,
                  ),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    _scheduledTests.length.toString(),
                    style: TextStyle(
                      color: theme.colorScheme.primary,
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const Spacer(),
                IconButton(
                  onPressed: _loadSchedule,
                  icon: const Icon(Icons.file_upload_outlined, size: 20),
                  tooltip: 'Load Schedule',
                  color: theme.colorScheme.primary,
                ),
                IconButton(
                  onPressed: _saveSchedule,
                  icon: const Icon(Icons.download_outlined, size: 20),
                  tooltip: 'Save Schedule',
                  color: theme.colorScheme.primary,
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: _scheduledTests.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.event_note_outlined,
                          size: 48,
                          color: Colors.grey.shade300,
                        ),
                        const SizedBox(height: 16),
                        Text(
                          'Schedule is empty',
                          style: TextStyle(color: Colors.grey.shade400),
                        ),
                      ],
                    ),
                  )
                : ListView.separated(
                    padding: const EdgeInsets.all(12),
                    itemCount: _scheduledTests.length,
                    separatorBuilder: (context, index) =>
                        const SizedBox(height: 8),
                    itemBuilder: (context, index) {
                      final item = _scheduledTests[index];
                      if (item.testName == 'Pause') {
                        return _buildPauseRow(index, theme);
                      }
                      return _buildScheduleRow(index, item, theme);
                    },
                  ),
          ),
          const Divider(height: 1),
          Padding(
            padding: const EdgeInsets.all(20.0),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _isStarting ? null : _startSchedule,
                icon: _isStarting
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.play_circle_filled_rounded),
                label: const Text('START SCHEDULE'),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                  backgroundColor: theme.colorScheme.primary,
                  foregroundColor: Colors.white,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildScheduleRow(int index, TestDescription test, ThemeData theme) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: theme.colorScheme.primary.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  '#${index + 1}',
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: theme.colorScheme.primary,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  test.testName,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                    color: Colors.black,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              IconButton(
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
                onPressed: () => _removeFromSchedule(index),
                icon: Icon(
                  Icons.delete_outline,
                  color: Colors.red.shade300,
                  size: 18,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          _buildRowDetail(Icons.settings_outlined, test.configuration ?? 'N/A'),
          _buildRowDetail(Icons.category_outlined, test.testCategory),
          if (test.remark != null && test.remark!.isNotEmpty)
            _buildRowDetail(Icons.note_alt_outlined, test.remark!),
        ],
      ),
    );
  }

  Widget _buildPauseRow(int index, ThemeData theme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.orange.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.orange.shade100),
      ),
      child: Row(
        children: [
          Icon(
            Icons.pause_circle_outline,
            color: Colors.orange.shade700,
            size: 20,
          ),
          const SizedBox(width: 12),
          Text(
            'PAUSE SEQUENCE',
            style: TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: 12,
              color: Colors.orange.shade700,
              letterSpacing: 0.5,
            ),
          ),
          const Spacer(),
          IconButton(
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(),
            onPressed: () => _removeFromSchedule(index),
            icon: Icon(
              Icons.delete_outline,
              color: Colors.orange.shade300,
              size: 18,
            ),
          ),
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
                  'Scheduler Help',
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
                  'Building a Sequence',
                  '1. Select a Category and Configuration.\n'
                      '2. Choose tests from the grid.\n'
                      '3. Add a remark if needed.\n'
                      '4. Click "ADD SELECTION" to put them in the schedule.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Using Pauses',
                  'Pauses are useful when you need to perform manual actions between automated tests, '
                      'such as changing cable connections or restarting hardware.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Import & Export',
                  'Use the upload icon to load previously saved schedules (.json). '
                      'Use the download icon to save your current schedule for future use.',
                ),
                const SizedBox(height: 24),
                _buildHelpItem(
                  theme,
                  'Execution',
                  'Once you click "START SCHEDULE", the system will execute each test in the list sequentially. '
                      'If a pause is encountered, the sequence will wait for user confirmation before proceeding.',
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

  Widget _buildRowDetail(IconData icon, String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        children: [
          Icon(icon, size: 14, color: Colors.grey.shade500),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              text,
              style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}
