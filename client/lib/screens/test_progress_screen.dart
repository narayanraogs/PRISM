import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/server_service.dart';

class TestProgressScreen extends StatefulWidget {
  final List<TestDescription> tests;

  const TestProgressScreen({super.key, required this.tests});

  @override
  State<TestProgressScreen> createState() => _TestProgressScreenState();
}

class _TestProgressScreenState extends State<TestProgressScreen> {
  late Stream<TestProgressResponse> _progressStream;
  final TextEditingController _inputController = TextEditingController();
  final FocusNode _confirmationFocusNode = FocusNode();

  bool _isCompleted = false;
  bool _isAborting = false;

  Timer? _countdownTimer;
  int _totalTimeout = 0;
  int _remainingSeconds = 0;

  @override
  void initState() {
    super.initState();
    final serverService = Provider.of<ServerService>(context, listen: false);
    _progressStream = serverService.connectTestProgress(widget.tests);
  }

  @override
  void dispose() {
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.closeTestProgress();
    _inputController.dispose();
    _confirmationFocusNode.dispose();
    _countdownTimer?.cancel();
    super.dispose();
  }

  void _startTimeoutCountdown(int seconds) {
    _countdownTimer?.cancel();
    if (seconds <= 0) return;

    _totalTimeout = seconds;
    _remainingSeconds = seconds;
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_remainingSeconds > 0) {
        setState(() {
          _remainingSeconds--;
        });
      } else {
        timer.cancel();
      }
    });
  }

  void _handleInput(String input) {
    _countdownTimer?.cancel();
    final serverService = Provider.of<ServerService>(context, listen: false);
    serverService.sendClientInput([input]);
    _inputController.clear();
    FocusScope.of(context).unfocus();
  }

  void _onDone() {
    if (!_isCompleted) {
      setState(() {
        _isCompleted = true;
        _isAborting = false;
      });
    }
  }

  void _handleAbort() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Abort Test Execution?'),
        content: const Text(
          'Are you sure you want to stop all currently running tests? This action cannot be undone.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('CANCEL'),
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
              serverService.sendAbort();
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red.shade600,
              foregroundColor: Colors.white,
            ),
            child: const Text('ABORT'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return PopScope(
      canPop: _isCompleted,
      child: Scaffold(
        backgroundColor: const Color(0xFFF8FAFC),
        appBar: AppBar(
          automaticallyImplyLeading: false,
          backgroundColor: Colors.white,
          elevation: 0,
          title: Text(
            'Test Execution Progress',
            style: GoogleFonts.outfit(
              color: Colors.black,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        body: StreamBuilder<TestProgressResponse>(
          stream: _progressStream,
          builder: (context, snapshot) {
            if (snapshot.hasError) {
              return Center(child: Text('Connection Error: ${snapshot.error}'));
            }

            if (snapshot.connectionState == ConnectionState.done) {
              WidgetsBinding.instance.addPostFrameCallback((_) => _onDone());
            }

            if (!snapshot.hasData) {
              return const Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    CircularProgressIndicator(),
                    SizedBox(height: 24),
                    Text('Starting test sequence...'),
                  ],
                ),
              );
            }

            final response = snapshot.data!;

            // Handle auto-fill and countdown for UI interactions
            if (response.ui.userInput || response.ui.userConfirmation) {
              if (_countdownTimer == null || !_countdownTimer!.isActive) {
                if (response.ui.userInput &&
                    response.ui.defaultValue.isNotEmpty &&
                    _inputController.text.isEmpty) {
                  _inputController.text = response.ui.defaultValue;
                }
                if (response.ui.timeoutSecs > 0) {
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (_countdownTimer == null || !_countdownTimer!.isActive) {
                      _startTimeoutCountdown(response.ui.timeoutSecs);
                    }
                  });
                }
              }
            } else {
              _countdownTimer?.cancel();
              _countdownTimer = null;
            }

            return Column(
              children: [
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Planned Tests Sidebar (20%)
                        Expanded(
                          flex: 2,
                          child: _buildPlannedTestsSidebar(
                            response.testStatus,
                            theme,
                          ),
                        ),
                        const SizedBox(width: 24),

                        // Main Execution Area (50%)
                        Expanded(
                          flex: 5,
                          child: Column(
                            children: [
                              Expanded(
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // Detailed Steps (Measurement/TM)
                                    Expanded(
                                      child: _buildStepContent(
                                        response.progress,
                                        theme,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 24),

                        // Summary Panel (30%)
                        Expanded(
                          flex: 3,
                          child: _buildSummarySidebar(response.summary, theme),
                        ),
                      ],
                    ),
                  ),
                ),
                // Persistent Interaction Bar
                _buildInteractionBar(response, theme),
              ],
            );
          },
        ),
      ),
    );
  }

  Widget _buildPlannedTestsSidebar(List<TestStatus> statuses, ThemeData theme) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(20),
            child: Text(
              'PLANNED TESTS',
              style: GoogleFonts.inter(
                fontSize: 12,
                fontWeight: FontWeight.w900,
                color: theme.colorScheme.primary,
                letterSpacing: 1.5,
              ),
            ),
          ),
          const Divider(height: 1, thickness: 1.5),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(12),
              itemCount: statuses.length,
              itemBuilder: (context, index) {
                final status = statuses[index];
                final isRunning =
                    status.testStatus.toLowerCase() == 'inprogress' ||
                    status.testStatus.toLowerCase() == 'running';
                final isSuccess =
                    status.testStatus.toLowerCase() == 'success' ||
                    status.testStatus.toLowerCase() == 'completed';
                final isFailure =
                    status.testStatus.toLowerCase() == 'failure' ||
                    status.testStatus.toLowerCase() == 'error';

                Color bgColor = Colors.transparent;
                IconData? icon;
                Color iconColor = Colors.grey;

                if (isRunning) {
                  bgColor = theme.colorScheme.primary.withOpacity(0.05);
                  icon = Icons.play_circle_filled;
                  iconColor = theme.colorScheme.primary;
                } else if (isSuccess) {
                  icon = Icons.check_circle;
                  iconColor = Colors.green;
                } else if (isFailure) {
                  icon = Icons.error;
                  iconColor = Colors.red;
                }

                return Container(
                  margin: const EdgeInsets.only(bottom: 4),
                  decoration: BoxDecoration(
                    color: bgColor,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: ListTile(
                    dense: true,
                    leading: Icon(
                      icon ?? Icons.circle_outlined,
                      size: 18,
                      color: iconColor,
                    ),
                    title: Text(
                      status.testType,
                      style: TextStyle(
                        fontWeight: isRunning
                            ? FontWeight.w900
                            : FontWeight.bold,
                        color: isRunning
                            ? theme.colorScheme.primary
                            : Colors.black,
                        fontSize: 16,
                      ),
                    ),
                    subtitle: Text(
                      '${status.testCategory}\n${status.config}',
                      style: TextStyle(
                        fontSize: 14,
                        color: isRunning
                            ? theme.colorScheme.primary
                            : Colors.black87,
                        height: 1.4,
                        fontWeight: FontWeight.w500,
                      ),
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

  Widget _buildStepContent(SingleTestProgress progress, ThemeData theme) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
              decoration: BoxDecoration(
                color: theme.colorScheme.primary,
                boxShadow: [
                  BoxShadow(
                    color: theme.colorScheme.primary.withOpacity(0.2),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              width: double.infinity,
              child: Row(
                children: [
                  const Icon(
                    Icons.rocket_launch,
                    color: Colors.white,
                    size: 24,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Text(
                      '${progress.configuration} : ${progress.testName} ${progress.testCategory.isNotEmpty ? "- ${progress.testCategory}" : ""}',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        fontSize: 20,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            // No divider needed between colored header and content
            Expanded(
              child: ListView(
                padding: const EdgeInsets.all(24),
                children: [
                  _buildSectionHeader(
                    'Database Validation',
                    progress.dbValidationStatus,
                    theme,
                  ),
                  const SizedBox(height: 12),
                  Padding(
                    padding: const EdgeInsets.only(left: 8.0),
                    child: Text(
                      progress.dbValidationStatus.isEmpty
                          ? 'Pending...'
                          : progress.dbValidationStatus,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                  const Divider(height: 48),

                  _buildSectionHeader('Instrument Status', '', theme),
                  const SizedBox(height: 16),
                  _buildInstrumentGrid(
                    progress.instruments,
                    progress.instrumentStatus,
                    theme,
                  ),
                  const Divider(height: 48),

                  _buildSectionHeader('Pre-Test Telemetry', '', theme),
                  const SizedBox(height: 16),
                  _buildTMTable(
                    progress.preTestTMMnemonics,
                    progress.preTestTMValues,
                    theme,
                  ),
                  const Divider(height: 48),

                  _buildSectionHeader('Measurement', '', theme),
                  const SizedBox(height: 16),
                  _buildMeasurementTable(
                    progress.measurementSteps,
                    progress.measurementValues,
                    progress.measurementStatus,
                    theme,
                  ),
                  const Divider(height: 48),

                  _buildSectionHeader('Post-Test Telemetry', '', theme),
                  const SizedBox(height: 16),
                  _buildTMTable(
                    progress.postTestTMMnemonics,
                    progress.postTestTMValues,
                    theme,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionHeader(String title, String status, ThemeData theme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
        border: Border.all(color: theme.colorScheme.primary.withOpacity(0.1)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: theme.colorScheme.primary.withOpacity(0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(
              Icons.label_important,
              size: 16,
              color: theme.colorScheme.primary,
            ),
          ),
          const SizedBox(width: 12),
          Text(
            title.toUpperCase(),
            style: GoogleFonts.inter(
              fontWeight: FontWeight.w900,
              fontSize: 16,
              color: Colors.black,
              letterSpacing: 1.2,
            ),
          ),
          const Spacer(),
          if (status.isNotEmpty) _getStatusIcon(status),
        ],
      ),
    );
  }

  Widget _getStatusIcon(String status) {
    final s = status.toLowerCase();
    if (s == 'success' || s == 'connected' || s == 'ok' || s == 'pass') {
      return const Icon(Icons.check_circle, color: Colors.green, size: 16);
    } else if (s == 'failed' ||
        s == 'notconnected' ||
        s == 'error' ||
        s == 'fail') {
      return const Icon(Icons.error, color: Colors.red, size: 16);
    } else {
      return const SizedBox(
        width: 14,
        height: 14,
        child: CircularProgressIndicator(strokeWidth: 2),
      );
    }
  }

  Widget _buildInstrumentGrid(
    List<String> names,
    List<String> statuses,
    ThemeData theme,
  ) {
    if (names.isEmpty)
      return const Text(
        'No instruments defined',
        style: TextStyle(color: Colors.grey, fontSize: 13),
      );
    return Wrap(
      spacing: 12,
      runSpacing: 12,
      children: List.generate(names.length, (i) {
        final status = i < statuses.length ? statuses[i] : 'pending';
        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          decoration: BoxDecoration(
            color: Colors.grey.shade50,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.grey.shade200),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                names[i],
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 4),
              _getStatusIcon(status),
            ],
          ),
        );
      }),
    );
  }

  Widget _buildTMTable(
    List<String> mnemonics,
    List<String> values,
    ThemeData theme,
  ) {
    if (mnemonics.isEmpty)
      return const Text(
        'No telemetry mnemonics',
        style: TextStyle(color: Colors.grey, fontSize: 13),
      );
    return Table(
      border: TableBorder.all(
        color: Colors.grey.shade200,
        borderRadius: BorderRadius.circular(8),
      ),
      columnWidths: const {0: FlexColumnWidth(2), 1: FlexColumnWidth(1)},
      children: [
        TableRow(
          decoration: BoxDecoration(color: Colors.grey.shade50),
          children: [
            const Padding(
              padding: EdgeInsets.all(10),
              child: Text(
                'Mnemonic',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ),
            const Padding(
              padding: EdgeInsets.all(10),
              child: Text(
                'Value',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ),
          ],
        ),
        ...List.generate(mnemonics.length, (i) {
          return TableRow(
            children: [
              Padding(
                padding: const EdgeInsets.all(10),
                child: Text(
                  mnemonics[i],
                  style: const TextStyle(
                    fontSize: 14,
                    color: Colors.black,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(10),
                child: Text(
                  i < values.length ? values[i] : '...',
                  style: GoogleFonts.robotoMono(
                    fontSize: 14,
                    color: Colors.blue.shade900,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          );
        }),
      ],
    );
  }

  Widget _buildMeasurementTable(
    List<String> steps,
    List<String> values,
    List<String> statuses,
    ThemeData theme,
  ) {
    if (steps.isEmpty)
      return const Text(
        'Measurement not started',
        style: TextStyle(color: Colors.grey, fontSize: 13),
      );
    return Table(
      border: TableBorder.all(
        color: Colors.grey.shade200,
        borderRadius: BorderRadius.circular(8),
      ),
      columnWidths: const {0: FlexColumnWidth(3), 1: FlexColumnWidth(2)},
      children: [
        TableRow(
          decoration: BoxDecoration(color: Colors.grey.shade50),
          children: [
            const Padding(
              padding: EdgeInsets.all(10),
              child: Text(
                'Action',
                style: TextStyle(
                  fontWeight: FontWeight.w900,
                  fontSize: 14,
                  color: Colors.black,
                ),
              ),
            ),
            const Padding(
              padding: EdgeInsets.all(10),
              child: Text(
                'Value',
                style: TextStyle(
                  fontWeight: FontWeight.w900,
                  fontSize: 14,
                  color: Colors.black,
                ),
              ),
            ),
          ],
        ),
        ...List.generate(steps.length, (i) {
          final status = i < statuses.length ? statuses[i].toLowerCase() : '';
          Color? rowColor;
          if (status == 'success' || status == 'pass')
            rowColor = Colors.green.shade50;
          if (status == 'inprogress') rowColor = Colors.blue.shade50;
          if (status == 'failed' || status == 'fail' || status == 'error')
            rowColor = Colors.red.shade50;

          return TableRow(
            decoration: BoxDecoration(color: rowColor),
            children: [
              Padding(
                padding: const EdgeInsets.all(10),
                child: Text(
                  steps[i],
                  style: const TextStyle(
                    fontSize: 14,
                    color: Colors.black,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(10),
                child: Text(
                  i < values.length ? values[i] : '...',
                  style: GoogleFonts.robotoMono(
                    fontSize: 14,
                    fontWeight: FontWeight.w800,
                    color: Colors.black,
                  ),
                ),
              ),
            ],
          );
        }),
      ],
    );
  }

  Widget _buildSummarySidebar(List<TestResult> results, ThemeData theme) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(20),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'SUMMARY',
                  style: GoogleFonts.inter(
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                    color: theme.colorScheme.primary,
                    letterSpacing: 1.5,
                  ),
                ),
                if (results.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.primary,
                      borderRadius: BorderRadius.circular(12),
                      boxShadow: [
                        BoxShadow(
                          color: theme.colorScheme.primary.withOpacity(0.3),
                          blurRadius: 4,
                        ),
                      ],
                    ),
                    child: Text(
                      '${results.length}',
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                      ),
                    ),
                  ),
              ],
            ),
          ),
          const Divider(height: 1, thickness: 1.5),
          Expanded(
            child: results.isEmpty
                ? Center(
                    child: Text(
                      'No results yet',
                      style: TextStyle(
                        color: Colors.grey.shade400,
                        fontSize: 13,
                      ),
                    ),
                  )
                : ListView.builder(
                    padding: const EdgeInsets.all(12),
                    itemCount: results.length,
                    itemBuilder: (context, index) {
                      final result = results[index];
                      return Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: Colors.grey.shade100),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.02),
                              blurRadius: 4,
                              offset: const Offset(0, 2),
                            ),
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              '${result.configuration} ${result.testName}',
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w900,
                                fontSize: 16,
                                color: const Color(0xFF478778),
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              result.testCategory,
                              style: TextStyle(
                                fontSize: 13,
                                color: Colors.grey.shade600,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            const Divider(height: 24, thickness: 1.5),
                            ...List.generate(result.name.length, (i) {
                              if (i >= result.result.length)
                                return const SizedBox();
                              return Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    result.name[i],
                                    style: const TextStyle(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  _buildSummaryTable(result.result[i]),
                                  const SizedBox(height: 12),
                                ],
                              );
                            }),
                          ],
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryTable(SummaryTable table) {
    if (table.header.isEmpty && table.data.isEmpty) {
      return const Text(
        'No data recorded',
        style: TextStyle(
          fontSize: 10,
          color: Colors.grey,
          fontStyle: FontStyle.italic,
        ),
      );
    }

    return Table(
      border: TableBorder.all(
        color: Colors.grey.shade200,
        borderRadius: BorderRadius.circular(4),
      ),
      children: [
        if (table.header.isNotEmpty)
          TableRow(
            decoration: BoxDecoration(color: Colors.grey.shade100),
            children: table.header
                .map(
                  (h) => Padding(
                    padding: const EdgeInsets.all(8),
                    child: Text(
                      h,
                      style: const TextStyle(
                        fontWeight: FontWeight.w900,
                        fontSize: 13,
                        color: Colors.black87,
                      ),
                    ),
                  ),
                )
                .toList(),
          ),
        ...table.data.map(
          (row) => TableRow(
            children: row
                .map(
                  (cell) => Padding(
                    padding: const EdgeInsets.all(8),
                    child: Text(
                      cell.value,
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: (cell.success || cell.error || cell.warning)
                            ? FontWeight.w900
                            : FontWeight.w500,
                        color: cell.success
                            ? Colors.green.shade700
                            : (cell.error
                                  ? Colors.red.shade700
                                  : (cell.warning
                                        ? Colors.amber.shade900
                                        : Colors.black87)),
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                )
                .toList(),
          ),
        ),
      ],
    );
  }

  Widget _buildInteractionBar(TestProgressResponse resp, ThemeData theme) {
    bool hasInteraction = resp.ui.userInput || resp.ui.userConfirmation;
    bool hasError = resp.progress.errorMessage.isNotEmpty;

    Color barColor = Colors.white;
    if (hasError)
      barColor = Colors.red.shade50;
    else if (hasInteraction)
      barColor = Colors.amber.shade50;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 20),
      decoration: BoxDecoration(
        color: barColor,
        border: Border(top: BorderSide(color: Colors.grey.shade200)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, -5),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (hasError) ...[
                  const Row(
                    children: [
                      Icon(Icons.error_outline, color: Colors.red, size: 20),
                      SizedBox(width: 8),
                      Text(
                        'TEST ERROR',
                        style: TextStyle(
                          color: Colors.red,
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    resp.progress.errorMessage,
                    style: const TextStyle(color: Colors.red, fontSize: 14),
                  ),
                ] else if (hasInteraction) ...[
                  Row(
                    children: [
                      Icon(
                        resp.ui.userConfirmation
                            ? Icons.help_outline
                            : Icons.edit_note,
                        color: Colors.amber.shade900,
                        size: 20,
                      ),
                      SizedBox(width: 8),
                      Text(
                        resp.ui.userConfirmation
                            ? 'CONFIRMATION REQUIRED'
                            : 'INPUT REQUIRED',
                        style: TextStyle(
                          color: Colors.amber.shade900,
                          fontWeight: FontWeight.bold,
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  if (resp.ui.userInput)
                    SizedBox(
                      width: 400,
                      child: TextFormField(
                        controller: _inputController,
                        autofocus: true,
                        decoration: InputDecoration(
                          hintText: resp.ui.prompt,
                          filled: true,
                          fillColor: Colors.white,
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                        ),
                        onFieldSubmitted: (v) => _handleInput(v),
                      ),
                    )
                  else
                    Text(
                      resp.ui.prompt,
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                ] else if (_isCompleted) ...[
                  const Row(
                    children: [
                      Icon(
                        Icons.check_circle_outline,
                        color: Colors.green,
                        size: 20,
                      ),
                      SizedBox(width: 8),
                      Text(
                        'TESTS COMPLETED',
                        style: TextStyle(
                          color: Colors.green,
                          fontWeight: FontWeight.bold,
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'All steps in the sequence have finished execution.',
                    style: TextStyle(fontSize: 14),
                  ),
                ] else ...[
                  const Row(
                    children: [
                      SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                      SizedBox(width: 12),
                      Text(
                        'EXECUTING...',
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 14,
                          color: Colors.blueGrey,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'System is running tests. Please wait for the next update or prompt.',
                    style: TextStyle(fontSize: 13, color: Colors.grey),
                  ),
                ],
              ],
            ),
          ),

          if (hasInteraction && _totalTimeout > 0) ...[
            const SizedBox(width: 24),
            _buildCountdownTimer(),
          ],

          const SizedBox(width: 24),
          Row(
            children: [
              if (hasInteraction) ...[
                ElevatedButton(
                  onPressed: () => _handleInput(
                    resp.ui.userConfirmation
                        ? 'CONFIRM'
                        : (_inputController.text.isEmpty
                              ? resp.ui.defaultValue
                              : _inputController.text),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: theme.colorScheme.primary,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 24,
                      vertical: 16,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: Text(resp.ui.userConfirmation ? 'CONFIRM' : 'SUBMIT'),
                ),
                if (resp.ui.userConfirmation) ...[
                  const SizedBox(width: 12),
                  OutlinedButton(
                    onPressed: () => _handleInput('REJECT'),
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
                    child: const Text('REJECT'),
                  ),
                ],
              ],
              if (!_isCompleted) ...[
                const SizedBox(width: 12),
                ElevatedButton.icon(
                  onPressed: _isAborting ? null : _handleAbort,
                  icon: _isAborting
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Colors.white,
                          ),
                        )
                      : const Icon(Icons.stop_circle_outlined, size: 18),
                  label: Text(_isAborting ? 'ABORTING...' : 'ABORT'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.red.shade600,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 24,
                      vertical: 16,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ] else ...[
                ElevatedButton.icon(
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(Icons.close),
                  label: const Text('DISMISS'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.grey.shade800,
                    foregroundColor: Colors.white,
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
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCountdownTimer() {
    return Container(
      width: 48,
      height: 48,
      child: Stack(
        fit: StackFit.expand,
        children: [
          CircularProgressIndicator(
            value: _totalTimeout > 0 ? _remainingSeconds / _totalTimeout : 0,
            strokeWidth: 4,
            backgroundColor: Colors.grey.shade200,
            color: _remainingSeconds < 10 ? Colors.red : Colors.amber,
          ),
          Center(
            child: Text(
              '$_remainingSeconds',
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
            ),
          ),
        ],
      ),
    );
  }
}
