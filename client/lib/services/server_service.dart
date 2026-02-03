import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:web_socket_channel/web_socket_channel.dart';
// Note: dart:html is for web only. Flutter web client specific.
// Using conditional import would be better if multi-platform, but this is a web project.
import 'dart:html' as html;

class ServerStatus {
  final String satelliteName;
  final String testPhase;
  final double memoryUsed;
  final double cpuUsed;
  final bool isConnected;

  ServerStatus({
    this.satelliteName = 'Unknown',
    this.testPhase = 'Unknown',
    this.memoryUsed = 0.0,
    this.cpuUsed = 0.0,
    this.isConnected = false,
  });

  factory ServerStatus.fromJson(Map<String, dynamic> json, bool connected) {
    return ServerStatus(
      satelliteName: json['SatelliteName'] ?? 'Unknown',
      testPhase: json['TestPhase'] ?? 'Unknown',
      memoryUsed: (json['MemoryUsed'] as num?)?.toDouble() ?? 0.0,
      cpuUsed: (json['CPUUsed'] as num?)?.toDouble() ?? 0.0,
      isConnected: connected,
    );
  }
}

class UplinkConfigInformation {
  final double powerAtSC;
  final double saLoss;
  final double scLoss;

  UplinkConfigInformation({
    required this.powerAtSC,
    required this.saLoss,
    required this.scLoss,
  });

  factory UplinkConfigInformation.fromJson(Map<String, dynamic> json) {
    return UplinkConfigInformation(
      powerAtSC: (json['PowerAtSC'] as num?)?.toDouble() ?? 0.0,
      saLoss: (json['SALoss'] as num?)?.toDouble() ?? 0.0,
      scLoss: (json['SCLoss'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class ConfigPathInformation {
  final String path;
  final String mnemonic;

  ConfigPathInformation({required this.path, required this.mnemonic});

  factory ConfigPathInformation.fromJson(Map<String, dynamic> json) {
    return ConfigPathInformation(
      path: json['Path'] ?? '',
      mnemonic: json['Mnemonic'] ?? '',
    );
  }
}

class RFUplinkMetaData {
  final List<String> uplinkConfigs;
  final Map<String, UplinkConfigInformation> uplinkConfigInformation;
  final List<String> tsms;
  final List<String> allConfigs;
  final Map<String, List<ConfigPathInformation>> configPathInformation;
  final bool ok;
  final String message;

  RFUplinkMetaData({
    required this.uplinkConfigs,
    required this.uplinkConfigInformation,
    required this.tsms,
    required this.allConfigs,
    required this.configPathInformation,
    required this.ok,
    required this.message,
  });

  factory RFUplinkMetaData.fromJson(Map<String, dynamic> json) {
    var uplinkConfigInfoMap = <String, UplinkConfigInformation>{};
    if (json['UplinkConfigInformation'] != null) {
      (json['UplinkConfigInformation'] as Map<String, dynamic>).forEach((
        key,
        value,
      ) {
        uplinkConfigInfoMap[key] = UplinkConfigInformation.fromJson(value);
      });
    }

    var configPathInfoMap = <String, List<ConfigPathInformation>>{};
    if (json['ConfigPathInformation'] != null) {
      (json['ConfigPathInformation'] as Map<String, dynamic>).forEach((
        key,
        value,
      ) {
        configPathInfoMap[key] = (value as List)
            .map((i) => ConfigPathInformation.fromJson(i))
            .toList();
      });
    }

    return RFUplinkMetaData(
      uplinkConfigs: List<String>.from(json['UplinkConfigs'] ?? []),
      uplinkConfigInformation: uplinkConfigInfoMap,
      tsms: List<String>.from(json['TSMs'] ?? []),
      allConfigs: List<String>.from(json['AllConfigs'] ?? []),
      configPathInformation: configPathInfoMap,
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class LinkStatus {
  final List<String> removeConfigs;
  final bool tsmConnected;
  final List<String> switchStatus;
  final double attn1Value;
  final double attn2Value;
  final bool ok;
  final String message;

  LinkStatus({
    required this.removeConfigs,
    required this.tsmConnected,
    required this.switchStatus,
    required this.attn1Value,
    required this.attn2Value,
    required this.ok,
    required this.message,
  });

  factory LinkStatus.fromJson(Map<String, dynamic> json) {
    return LinkStatus(
      removeConfigs: List<String>.from(json['RemoveConfigs'] ?? []),
      tsmConnected: json['TSMConnected'] ?? false,
      switchStatus: List<String>.from(json['SwitchStatus'] ?? []),
      attn1Value: (json['Attn1Value'] as num?)?.toDouble() ?? 0.0,
      attn2Value: (json['Attn2Value'] as num?)?.toDouble() ?? 0.0,
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class Ack {
  final bool ok;
  final String message;

  Ack({required this.ok, required this.message});

  factory Ack.fromJson(Map<String, dynamic> json) {
    return Ack(ok: json['OK'] ?? false, message: json['Message'] ?? '');
  }
}

class TestStatus {
  final String testType;
  final String testCategory;
  final String config;
  final String testStatus;

  TestStatus({
    required this.testType,
    required this.testCategory,
    required this.config,
    required this.testStatus,
  });

  factory TestStatus.fromJson(Map<String, dynamic> json) {
    return TestStatus(
      testType: json['TestType'] ?? '',
      testCategory: json['TestCategory'] ?? '',
      config: json['Config'] ?? '',
      testStatus: json['TestStatus'] ?? '',
    );
  }
}

class SummaryCell {
  final String value;
  final bool error;
  final bool success;
  final bool warning;

  SummaryCell({
    required this.value,
    this.error = false,
    this.success = false,
    this.warning = false,
  });

  factory SummaryCell.fromJson(Map<String, dynamic> json) {
    return SummaryCell(
      value: (json['Value'] ?? json['value'])?.toString() ?? '',
      error: (json['Error'] ?? json['error']) == true,
      success: (json['Success'] ?? json['success']) == true,
      warning: (json['Warning'] ?? json['warning']) == true,
    );
  }
}

class SummaryTable {
  final List<String> header;
  final List<List<SummaryCell>> data;

  SummaryTable({required this.header, required this.data});

  factory SummaryTable.fromJson(Map<String, dynamic> json) {
    var headerData = json['Header'] ?? json['header'];
    var header = headerData != null
        ? List<String>.from(headerData)
        : <String>[];

    var dataRows = <List<SummaryCell>>[];
    var dataField = json['Data'] ?? json['data'];
    if (dataField != null && dataField is List) {
      for (var row in dataField) {
        if (row is List) {
          dataRows.add(
            row
                .map((c) => SummaryCell.fromJson(c as Map<String, dynamic>))
                .toList(),
          );
        }
      }
    }
    return SummaryTable(header: header, data: dataRows);
  }
}

class TestResult {
  final String testName;
  final String testCategory;
  final String configuration;
  final List<String> name;
  final List<SummaryTable> result;

  TestResult({
    required this.testName,
    required this.testCategory,
    required this.configuration,
    required this.name,
    required this.result,
  });

  factory TestResult.fromJson(Map<String, dynamic> json) {
    var nameData = json['Name'] ?? json['name'];
    var resultData = json['Result'] ?? json['result'];

    return TestResult(
      testName: (json['TestName'] ?? json['testName'])?.toString() ?? '',
      testCategory:
          (json['TestCategory'] ?? json['testCategory'])?.toString() ?? '',
      configuration:
          (json['Configuration'] ?? json['configuration'])?.toString() ?? '',
      name: nameData != null ? List<String>.from(nameData) : <String>[],
      result: (resultData != null && resultData is List)
          ? resultData
                .map((e) => SummaryTable.fromJson(e as Map<String, dynamic>))
                .toList()
          : <SummaryTable>[],
    );
  }
}

class SingleTestProgress {
  final String testName;
  final String testCategory;
  final String configuration;
  final String currentStep;
  final String errorMessage;
  final String dbValidationStatus;
  final List<String> instruments;
  final List<String> instrumentStatus;
  final List<String> preTestTMMnemonics;
  final List<String> preTestTMValues;
  final List<String> measurementSteps;
  final List<String> measurementValues;
  final List<String> measurementStatus;
  final List<String> postTestTMMnemonics;
  final List<String> postTestTMValues;

  SingleTestProgress({
    required this.testName,
    required this.testCategory,
    required this.configuration,
    required this.currentStep,
    required this.errorMessage,
    required this.dbValidationStatus,
    required this.instruments,
    required this.instrumentStatus,
    required this.preTestTMMnemonics,
    required this.preTestTMValues,
    required this.measurementSteps,
    required this.measurementValues,
    required this.measurementStatus,
    required this.postTestTMMnemonics,
    required this.postTestTMValues,
  });

  factory SingleTestProgress.fromJson(Map<String, dynamic> json) {
    return SingleTestProgress(
      testName: (json['TestName'] ?? json['testName'])?.toString() ?? '',
      testCategory:
          (json['TestCategory'] ?? json['testCategory'])?.toString() ?? '',
      configuration:
          (json['Configuration'] ?? json['configuration'])?.toString() ?? '',
      currentStep:
          (json['CurrentStep'] ?? json['currentStep'])?.toString() ?? '',
      errorMessage:
          (json['ErrorMessage'] ?? json['errorMessage'])?.toString() ?? '',
      dbValidationStatus:
          (json['DBValidationStatus'] ?? json['dbValidationStatus'])
              ?.toString() ??
          '',
      instruments: List<String>.from(
        json['Instruments'] ?? json['instruments'] ?? [],
      ),
      instrumentStatus: List<String>.from(
        json['InstrumentStatus'] ?? json['instrumentStatus'] ?? [],
      ),
      preTestTMMnemonics: List<String>.from(
        json['PreTestTMMnemonics'] ?? json['preTestTMMnemonics'] ?? [],
      ),
      preTestTMValues: List<String>.from(
        json['PreTestTMValues'] ?? json['preTestTMValues'] ?? [],
      ),
      measurementSteps: List<String>.from(
        json['MeasurementSteps'] ?? json['measurementSteps'] ?? [],
      ),
      measurementValues: List<String>.from(
        json['MeasurementValues'] ?? json['measurementValues'] ?? [],
      ),
      measurementStatus: List<String>.from(
        json['MeasurementStatus'] ?? json['measurementStatus'] ?? [],
      ),
      postTestTMMnemonics: List<String>.from(
        json['PostTestTMMnemonics'] ?? json['postTestTMMnemonics'] ?? [],
      ),
      postTestTMValues: List<String>.from(
        json['PostTestTMValues'] ?? json['postTestTMValues'] ?? [],
      ),
    );
  }
}

class UserInteraction {
  final bool userConfirmation;
  final bool userInput;
  final String prompt;
  final int timeoutSecs;
  final String defaultValue;

  UserInteraction({
    required this.userConfirmation,
    required this.userInput,
    required this.prompt,
    required this.timeoutSecs,
    required this.defaultValue,
  });

  factory UserInteraction.fromJson(Map<String, dynamic> json) {
    return UserInteraction(
      userConfirmation:
          (json['UserConfirmation'] ?? json['userConfirmation']) == true,
      userInput: (json['UserInput'] ?? json['userInput']) == true,
      prompt: (json['Prompt'] ?? json['prompt'])?.toString() ?? '',
      timeoutSecs: (json['TimeoutSecs'] ?? json['timeoutSecs']) ?? 0,
      defaultValue:
          (json['DefaultValue'] ?? json['defaultValue'])?.toString() ?? '',
    );
  }
}

class TestProgressResponse {
  final List<TestStatus> testStatus;
  final SingleTestProgress progress;
  final List<TestResult> summary;
  final UserInteraction ui;
  final bool ok;
  final String message;

  TestProgressResponse({
    required this.testStatus,
    required this.progress,
    required this.summary,
    required this.ui,
    required this.ok,
    required this.message,
  });

  factory TestProgressResponse.fromJson(Map<String, dynamic> json) {
    return TestProgressResponse(
      testStatus:
          ((json['TestStatus'] ?? json['testStatus']) as List?)
              ?.map((e) => TestStatus.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      progress: SingleTestProgress.fromJson(
        (json['Progress'] ?? json['progress']) ?? {},
      ),
      summary:
          ((json['Summary'] ?? json['summary']) as List?)
              ?.map((e) => TestResult.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      ui: UserInteraction.fromJson((json['UI'] ?? json['ui']) ?? {}),
      ok: (json['OK'] ?? json['ok']) == true,
      message: (json['Message'] ?? json['message'])?.toString() ?? '',
    );
  }
}

class TestDescription {
  final String testName;
  final String testCategory;
  final String? configuration;
  final String? remark;
  final List<String> extraParameters;

  TestDescription({
    required this.testName,
    required this.testCategory,
    this.configuration,
    this.remark,
    this.extraParameters = const [],
  });

  factory TestDescription.fromJson(Map<String, dynamic> json) {
    return TestDescription(
      testName: json['TestName'] ?? '',
      testCategory: json['TestCategory'] ?? '',
      configuration: json['Configuration'],
      remark: json['Remark'],
      extraParameters: List<String>.from(json['ExtraParameters'] ?? []),
    );
  }

  Map<String, dynamic> toJson() => {
    'TestName': testName,
    'TestCategory': testCategory,
    'Configuration': configuration,
    'Remark': remark,
    'ExtraParameters': extraParameters,
  };

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is TestDescription &&
          runtimeType == other.runtimeType &&
          testName == other.testName &&
          testCategory == other.testCategory &&
          configuration == other.configuration &&
          remark == other.remark &&
          listEquals(extraParameters, other.extraParameters);

  @override
  int get hashCode =>
      testName.hashCode ^
      testCategory.hashCode ^
      configuration.hashCode ^
      remark.hashCode ^
      extraParameters.hashCode;
}

class AllTests {
  final List<String> categories;
  final Map<String, List<String>> configurations;
  final Map<String, List<TestDescription>> tests;
  final Map<String, String> losses;
  final bool ok;
  final String message;

  AllTests({
    required this.categories,
    required this.configurations,
    required this.tests,
    required this.losses,
    required this.ok,
    required this.message,
  });

  factory AllTests.fromJson(Map<String, dynamic> json) {
    var categories = List<String>.from(json['Categories'] ?? []);
    var configurations = <String, List<String>>{};
    if (json['Configurations'] != null) {
      (json['Configurations'] as Map<String, dynamic>).forEach((key, value) {
        configurations[key] = List<String>.from(value);
      });
    }
    var tests = <String, List<TestDescription>>{};
    if (json['Tests'] != null) {
      (json['Tests'] as Map<String, dynamic>).forEach((key, value) {
        tests[key] = (value as List)
            .map((i) => TestDescription.fromJson(i))
            .toList();
      });
    }
    var losses = Map<String, String>.from(json['Losses'] ?? {});
    return AllTests(
      categories: categories,
      configurations: configurations,
      tests: tests,
      losses: losses,
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class ServerService extends ChangeNotifier {
  WebSocketChannel? _channel;
  ServerStatus _status = ServerStatus();
  bool _isReconnecting = false;

  ServerStatus get status => _status;

  ServerService() {
    _connect();
  }

  String _getWsUrl() {
    String host;
    if (kDebugMode) {
      host = 'localhost:8080';
    } else {
      host = html.window.location.host;
    }
    final protocol = html.window.location.protocol == 'https:' ? 'wss' : 'ws';
    return '$protocol://$host/serverStatus';
  }

  void _connect() {
    if (_isReconnecting) return;
    final url = _getWsUrl();
    debugPrint('Connecting to WebSocket at: $url');
    try {
      _channel = WebSocketChannel.connect(Uri.parse(url));
      _channel!.stream.listen(
        (data) {
          final Map<String, dynamic> json = jsonDecode(data);
          _status = ServerStatus.fromJson(json, true);
          notifyListeners();
        },
        onError: (error) {
          debugPrint('WebSocket Error: $error');
          _handleDisconnect();
        },
        onDone: () {
          debugPrint('WebSocket Connection Closed');
          _handleDisconnect();
        },
      );
    } catch (e) {
      debugPrint('WebSocket Connection failed: $e');
      _handleDisconnect();
    }
  }

  void _handleDisconnect() {
    _status = ServerStatus(isConnected: false);
    notifyListeners();
    _reconnect();
  }

  void _reconnect() async {
    if (_isReconnecting) return;
    _isReconnecting = true;
    await Future.delayed(const Duration(seconds: 5));
    _isReconnecting = false;
    _connect();
  }

  void manualReconnect() {
    _channel?.sink.close();
    _connect();
  }

  Future<RFUplinkMetaData?> fetchRFUplinkMetaData() async {
    final host = kDebugMode ? 'localhost:8080' : html.window.location.host;
    final protocol = html.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/getRFUplinkMetaData';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        return RFUplinkMetaData.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching RF Uplink MetaData: $e');
    }
    return null;
  }

  Future<LinkStatus?> fetchLinkStatus(String tsmSelected) async {
    final host = kDebugMode ? 'localhost:8080' : html.window.location.host;
    final protocol = html.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/getLinkStatus';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'TSMSelected': tsmSelected}),
      );

      if (response.statusCode == 200) {
        return LinkStatus.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching Link Status: $e');
    }
    return null;
  }

  Future<Ack?> setTSMRoute(String tsmSelected, String mnemonic) async {
    final host = kDebugMode ? 'localhost:8080' : html.window.location.host;
    final protocol = html.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/setTSMRoute';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'TSMSelected': tsmSelected, 'Mnemonic': mnemonic}),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error setting TSM Route: $e');
    }
    return null;
  }

  Future<Ack?> setTSMAttn(
    String tsmSelected,
    int attnNo,
    double attnValue,
  ) async {
    final host = kDebugMode ? 'localhost:8080' : html.window.location.host;
    final protocol = html.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/setTSMAttn';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'TSMSelected': tsmSelected,
          'AttnNo': attnNo,
          'AttnValue': attnValue,
        }),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error setting TSM Attenuation: $e');
    }
    return null;
  }

  Future<AllTests?> fetchAllTests() async {
    final host = kDebugMode ? 'localhost:8080' : html.window.location.host;
    final protocol = html.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/getAllTests';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({}),
      );
      if (response.statusCode == 200) {
        return AllTests.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching All Tests: $e');
    }
    return null;
  }

  Future<Ack?> startTests(List<TestDescription> tests, String remark) async {
    // Note: The user replaced startTests with testProgress endpoint.
    // However, they might still want an HTTP call or we might transition fully to WebSocket.
    // For now, I'll keep this but the user said "Start tests" will now move to "test progress screen".
    // So usually we will use connectTestProgress instead.
    return Ack(
      ok: true,
      message: "Use connectTestProgress for real-time tracking",
    );
  }

  WebSocketChannel? _progressChannel;

  Stream<TestProgressResponse> connectTestProgress(
    List<TestDescription> tests,
  ) {
    String host;
    if (kDebugMode) {
      host = 'localhost:8080';
    } else {
      host = html.window.location.host;
    }
    final protocol = html.window.location.protocol == 'https:' ? 'wss' : 'ws';
    final url = '$protocol://$host/testProgress';

    _progressChannel = WebSocketChannel.connect(Uri.parse(url));

    // Send initial request
    _progressChannel!.sink.add(
      jsonEncode({'Tests': tests.map((t) => t.toJson()).toList()}),
    );

    return _progressChannel!.stream.map((data) {
      return TestProgressResponse.fromJson(jsonDecode(data));
    });
  }

  void sendClientInput(List<String> parameters) {
    _progressChannel?.sink.add(jsonEncode({'Parameters': parameters}));
  }

  void closeTestProgress() {
    _progressChannel?.sink.close();
    _progressChannel = null;
  }

  @override
  void dispose() {
    _channel?.sink.close();
    super.dispose();
  }
}
