import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:web_socket_channel/web_socket_channel.dart';
// Note: dart:html is for web only. Flutter web client specific.
// Using conditional import would be better if multi-platform, but this is a web project.
import 'package:web/web.dart' as web;

class ServerStatus {
  final String satelliteName;
  final String testPhase;
  final double memoryUsed;
  final double cpuUsed;
  final bool isConnected;
  final BootstrapData? bootstrapData;

  ServerStatus({
    this.satelliteName = 'Unknown',
    this.testPhase = 'Unknown',
    this.memoryUsed = 0.0,
    this.cpuUsed = 0.0,
    this.isConnected = false,
    this.bootstrapData,
  });

  factory ServerStatus.fromJson(Map<String, dynamic> json, bool connected, {BootstrapData? bootstrap}) {
    return ServerStatus(
      satelliteName: json['SatelliteName'] ?? 'Unknown',
      testPhase: json['TestPhase'] ?? 'Unknown',
      memoryUsed: (json['MemoryUsed'] as num?)?.toDouble() ?? 0.0,
      cpuUsed: (json['CPUUsed'] as num?)?.toDouble() ?? 0.0,
      isConnected: connected,
      bootstrapData: bootstrap,
    );
  }
}

class BootstrapData {
  final RFUplinkMetaData rfuData;
  final AllTests testData;
  final StabilityMetadata stabilityData;
  final StabilityReportMetadataResponse stabilityReportsData;
  final SpectrumDumpMetadata spectrumDumpData;
  final MonitorMetadata monitorData;
  final TVACCableLossMetadata tvacCableLossData;
  final CableLossMetadata cableLossData;
  final DatabaseMetadata databaseData;
  final ReportsResponse reportsData;
  final TSMInternalLossMetadata tsmInternalLossData;
  final UCDCMetadata ucdcData;
  final AttnMetaData attnData;
  final GTxMeasurementMetadata gtxData;

  BootstrapData({
    required this.rfuData,
    required this.testData,
    required this.stabilityData,
    required this.stabilityReportsData,
    required this.spectrumDumpData,
    required this.monitorData,
    required this.tvacCableLossData,
    required this.cableLossData,
    required this.databaseData,
    required this.reportsData,
    required this.tsmInternalLossData,
    required this.ucdcData,
    required this.attnData,
    required this.gtxData,
  });

  factory BootstrapData.fromJson(Map<String, dynamic> json) {
    return BootstrapData(
      rfuData: RFUplinkMetaData.fromJson(json['RFUplinkData'] ?? {}),
      testData: AllTests.fromJson(json['TestData'] ?? {}),
      stabilityData: StabilityMetadata.fromJson(json['StabilityData'] ?? {}),
      stabilityReportsData:
          StabilityReportMetadataResponse.fromJson(json['StabilityReportsData'] ?? {}),
      spectrumDumpData: SpectrumDumpMetadata.fromJson(json['SpectrumDumpData'] ?? {}),
      monitorData: MonitorMetadata.fromJson(json['MonitorData'] ?? {}),
      tvacCableLossData: TVACCableLossMetadata.fromJson(json['TVACCableLossData'] ?? {}),
      cableLossData: CableLossMetadata.fromJson(json['CableLossData'] ?? {}),
      databaseData: DatabaseMetadata.fromJson(json['DatabaseData'] ?? {}),
      reportsData: ReportsResponse.fromJson(json['ReportsData'] ?? {}),
      tsmInternalLossData: TSMInternalLossMetadata.fromJson(json['TSMInternalLossData'] ?? {}),
      ucdcData: UCDCMetadata.fromJson(json['UCDCData'] ?? {}),
      attnData: AttnMetaData.fromJson(json['AttnData'] ?? {}),
      gtxData: GTxMeasurementMetadata.fromJson(json['GTxData'] ?? {}),
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

class StabilityMetadata {
  final List<String> instrumentTypes;
  final Map<String, List<String>> instruments;
  final Map<String, List<String>> parameters;
  final List<SpectrumProfile> profiles;
  final List<String> plConfigs;
  final List<String> pulseProfiles;
  final List<String> ppmChannels;
  final bool ok;
  final String message;

  StabilityMetadata({
    required this.instrumentTypes,
    required this.instruments,
    required this.parameters,
    required this.profiles,
    required this.plConfigs,
    required this.pulseProfiles,
    required this.ppmChannels,
    required this.ok,
    required this.message,
  });

  factory StabilityMetadata.fromJson(Map<String, dynamic> json) {
    return StabilityMetadata(
      instrumentTypes: List<String>.from(json['InstrumentTypes'] ?? []),
      instruments:
          (json['Instruments'] as Map<String, dynamic>?)?.map(
            (key, value) => MapEntry(key, List<String>.from(value)),
          ) ??
          {},
      parameters:
          (json['Parameters'] as Map<String, dynamic>?)?.map(
            (key, value) => MapEntry(key, List<String>.from(value)),
          ) ??
          {},
      profiles:
          (json['Profiles'] as List?)
              ?.map((e) => SpectrumProfile.fromJson(e))
              .toList() ??
          [],
      plConfigs: List<String>.from(json['PLConfigs'] ?? []),
      pulseProfiles: List<String>.from(json['PulseProfiles'] ?? []),
      ppmChannels: List<String>.from(json['PPMChannels'] ?? []),
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class DeviceProfileDetails {
  final String gtxName;
  final String saName;
  final String vsaName;
  final String pmName;
  final String ppmName;
  final String sgName;
  final String tsmName;

  DeviceProfileDetails({
    required this.gtxName,
    required this.saName,
    required this.vsaName,
    required this.pmName,
    required this.ppmName,
    required this.sgName,
    required this.tsmName,
  });

  factory DeviceProfileDetails.fromJson(Map<String, dynamic> json) {
    return DeviceProfileDetails(
      gtxName: json['GTxName'] ?? '',
      saName: json['SAName'] ?? '',
      vsaName: json['VSAName'] ?? '',
      pmName: json['PMName'] ?? '',
      ppmName: json['PPMName'] ?? '',
      sgName: json['SGName'] ?? '',
      tsmName: json['TSMName'] ?? '',
    );
  }
}

class UCDCDetails {
  final double inputFrequency;
  final double outputFrequency;
  final double loFrequency;

  UCDCDetails({
    required this.inputFrequency,
    required this.outputFrequency,
    required this.loFrequency,
  });

  factory UCDCDetails.fromJson(Map<String, dynamic> json) {
    return UCDCDetails(
      inputFrequency: (json['InputFrequency'] as num?)?.toDouble() ?? 0.0,
      outputFrequency: (json['OutputFrequency'] as num?)?.toDouble() ?? 0.0,
      loFrequency: (json['LOFrequency'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class UCDCMetadata {
  final List<String> converters;
  final Map<String, UCDCDetails> converterDetails;
  final List<String> deviceProfiles;
  final Map<String, DeviceProfileDetails> deviceMapping;
  final List<String> signalGenerators;
  final bool ok;
  final String message;

  UCDCMetadata({
    required this.converters,
    required this.converterDetails,
    required this.deviceProfiles,
    required this.deviceMapping,
    required this.signalGenerators,
    required this.ok,
    required this.message,
  });

  factory UCDCMetadata.fromJson(Map<String, dynamic> json) {
    var converterDetailsMap = <String, UCDCDetails>{};
    if (json['ConverterDetails'] != null) {
      (json['ConverterDetails'] as Map<String, dynamic>).forEach((key, value) {
        converterDetailsMap[key] = UCDCDetails.fromJson(value);
      });
    }

    var deviceMappingMap = <String, DeviceProfileDetails>{};
    if (json['DeviceMapping'] != null) {
      (json['DeviceMapping'] as Map<String, dynamic>).forEach((key, value) {
        deviceMappingMap[key] = DeviceProfileDetails.fromJson(value);
      });
    }

    return UCDCMetadata(
      converters: List<String>.from(json['Converters'] ?? []),
      converterDetails: converterDetailsMap,
      deviceProfiles: List<String>.from(json['DeviceProfiles'] ?? []),
      deviceMapping: deviceMappingMap,
      signalGenerators: List<String>.from(json['SignalGenerators'] ?? []),
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class GTxMeasurementMetadata {
  final List<String> deviceProfile;
  final Map<String, DeviceProfileDetails> deviceMapping;
  final bool ok;
  final String message;

  GTxMeasurementMetadata({
    required this.deviceProfile,
    required this.deviceMapping,
    required this.ok,
    required this.message,
  });

  factory GTxMeasurementMetadata.fromJson(Map<String, dynamic> json) {
    var deviceMappingMap = <String, DeviceProfileDetails>{};
    if (json['DeviceMapping'] != null) {
      (json['DeviceMapping'] as Map<String, dynamic>).forEach((key, value) {
        deviceMappingMap[key] = DeviceProfileDetails.fromJson(value);
      });
    }

    return GTxMeasurementMetadata(
      deviceProfile: List<String>.from(json['DeviceProfile'] ?? []),
      deviceMapping: deviceMappingMap,
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class SpectrumProfile {
  final String profileName;
  final double centerFrequency;
  final double span;
  final double rbw;
  final double vbw;

  SpectrumProfile({
    required this.profileName,
    required this.centerFrequency,
    required this.span,
    required this.rbw,
    required this.vbw,
  });

  factory SpectrumProfile.fromJson(Map<String, dynamic> json) {
    return SpectrumProfile(
      profileName: json['ProfileName'] ?? '',
      centerFrequency: (json['CenterFrequency'] as num?)?.toDouble() ?? 0.0,
      span: (json['Span'] as num?)?.toDouble() ?? 0.0,
      rbw: (json['RBW'] as num?)?.toDouble() ?? 0.0,
      vbw: (json['VBW'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class SpectrumDumpMetadata {
  final List<String> spectrumDumpMode;
  final Map<String, List<String>> instruments;
  final List<SpectrumProfile> spectrumProfiles;
  final List<String> screenshotProfiles;
  final bool ok;
  final String message;

  SpectrumDumpMetadata({
    required this.spectrumDumpMode,
    required this.instruments,
    required this.spectrumProfiles,
    required this.screenshotProfiles,
    required this.ok,
    required this.message,
  });

  factory SpectrumDumpMetadata.fromJson(Map<String, dynamic> json) {
    return SpectrumDumpMetadata(
      spectrumDumpMode: List<String>.from(json['SpectrumDumpMode'] ?? []),
      instruments:
          (json['Instruments'] as Map<String, dynamic>?)?.map(
            (key, value) => MapEntry(key, List<String>.from(value)),
          ) ??
          {},
      spectrumProfiles:
          (json['SpectrumProfiles'] as List?)
              ?.map((e) => SpectrumProfile.fromJson(e))
              .toList() ??
          [],
      screenshotProfiles: List<String>.from(json['ScreenshotProfiles'] ?? []),
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class MonitorMetadata {
  final List<String> instrumentTypes;
  final Map<String, List<String>> instruments;
  final bool ok;
  final String message;

  MonitorMetadata({
    required this.instrumentTypes,
    required this.instruments,
    required this.ok,
    required this.message,
  });

  factory MonitorMetadata.fromJson(Map<String, dynamic> json) {
    return MonitorMetadata(
      instrumentTypes: List<String>.from(json['InstrumentTypes'] ?? []),
      instruments:
          (json['Instruments'] as Map<String, dynamic>?)?.map(
            (key, value) => MapEntry(key, List<String>.from(value)),
          ) ??
          {},
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class MonitorResponse {
  final String image;
  final double pmChannelA;
  final double pmChannelB;
  final double ppmChannelAPeakPower;
  final double ppmChannelBPeakPower;
  final double ppmChannelAAvgPower;
  final double ppmChannelBAvgPower;
  final bool ok;
  final String message;

  MonitorResponse({
    required this.image,
    required this.pmChannelA,
    required this.pmChannelB,
    required this.ppmChannelAPeakPower,
    required this.ppmChannelBPeakPower,
    required this.ppmChannelAAvgPower,
    required this.ppmChannelBAvgPower,
    required this.ok,
    required this.message,
  });

  factory MonitorResponse.fromJson(Map<String, dynamic> json) {
    return MonitorResponse(
      image: json['Image'] ?? '',
      pmChannelA: (json['PMChannelA'] as num?)?.toDouble() ?? 0.0,
      pmChannelB: (json['PMChannelB'] as num?)?.toDouble() ?? 0.0,
      ppmChannelAPeakPower:
          (json['PPMChannelAPeakPower'] as num?)?.toDouble() ?? 0.0,
      ppmChannelBPeakPower:
          (json['PPMChannelBPeakPower'] as num?)?.toDouble() ?? 0.0,
      ppmChannelAAvgPower:
          (json['PPMChannelAAvgPower'] as num?)?.toDouble() ?? 0.0,
      ppmChannelBAvgPower:
          (json['PPMChannelBAvgPower'] as num?)?.toDouble() ?? 0.0,
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class MeasurementPoint {
  final double frequency;
  final double loss;
  final double delta;

  MeasurementPoint({
    required this.frequency,
    required this.loss,
    this.delta = 0.0,
  });

  factory MeasurementPoint.fromJson(Map<String, dynamic> json) {
    return MeasurementPoint(
      frequency: ((json['frequency'] as num?)?.toDouble() ?? 0.0) / 1e6,
      loss: (json['loss'] as num?)?.toDouble() ?? 0.0,
      delta: (json['delta'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class TVACCableLossRecord {
  final int slNo;
  final String cableName;
  final String cycleName;
  final String phase;
  final String date;
  final String time;
  final bool isReference;
  final List<MeasurementPoint> measurements;

  TVACCableLossRecord({
    required this.slNo,
    required this.cableName,
    required this.cycleName,
    required this.phase,
    required this.date,
    required this.time,
    required this.isReference,
    required this.measurements,
  });

  factory TVACCableLossRecord.fromJson(Map<String, dynamic> json) {
    return TVACCableLossRecord(
      slNo: json['slNo'] ?? 0,
      cableName: json['cableName'] ?? '',
      cycleName: json['cycleName'] ?? '',
      phase: json['phase'] ?? '',
      date: json['date'] ?? '',
      time: json['time'] ?? '',
      isReference: json['isReference'] ?? false,
      measurements:
          (json['measurements'] as List?)
              ?.map((e) => MeasurementPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class TVACCableLossMetadata {
  final List<double> frequencies;
  final List<String> deviceProfiles;
  final List<String> existingCables;
  final bool isPmZeroed;
  final bool ok;
  final String message;

  TVACCableLossMetadata({
    required this.frequencies,
    required this.deviceProfiles,
    required this.existingCables,
    required this.isPmZeroed,
    required this.ok,
    required this.message,
  });

  factory TVACCableLossMetadata.fromJson(Map<String, dynamic> json) {
    return TVACCableLossMetadata(
      frequencies:
          (json['frequencies'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      deviceProfiles: List<String>.from(json['deviceProfiles'] ?? []),
      existingCables: List<String>.from(json['existingCables'] ?? []),
      isPmZeroed: json['isPmZeroed'] ?? false,
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
    );
  }
}

class CableLossMetadata {
  final List<String> frequencies;
  final List<String> deviceProfiles;
  final List<String> existingCables;
  final bool isPmZeroed;
  final bool ok;
  final String message;

  CableLossMetadata({
    required this.frequencies,
    required this.deviceProfiles,
    required this.existingCables,
    required this.isPmZeroed,
    required this.ok,
    required this.message,
  });

  factory CableLossMetadata.fromJson(Map<String, dynamic> json) {
    return CableLossMetadata(
      frequencies: List<String>.from(json['frequencies'] ?? []),
      deviceProfiles: List<String>.from(json['deviceProfiles'] ?? []),
      existingCables: List<String>.from(json['existingCables'] ?? []),
      isPmZeroed: json['isPmZeroed'] ?? false,
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
    );
  }
}

class InternalLossPMOrCableEntry {
  final List<double> frequencies;
  final List<double> losses;
  final bool measured;

  InternalLossPMOrCableEntry({
    required this.frequencies,
    required this.losses,
    required this.measured,
  });

  factory InternalLossPMOrCableEntry.fromJson(Map<String, dynamic> json) {
    return InternalLossPMOrCableEntry(
      frequencies:
          (json['Frequencies'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      losses:
          (json['Losses'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      measured: json['Measured'] ?? false,
    );
  }
}

class InternalLossEntry {
  final String inputPort;
  final String outputPort;
  final String pathMnemonic;
  final List<double> frequencies;
  final List<double> losses;
  final bool measured;

  InternalLossEntry({
    required this.inputPort,
    required this.outputPort,
    required this.pathMnemonic,
    required this.frequencies,
    required this.losses,
    required this.measured,
  });

  factory InternalLossEntry.fromJson(Map<String, dynamic> json) {
    return InternalLossEntry(
      inputPort: json['InputPort'] ?? '',
      outputPort: json['OutputPort'] ?? '',
      pathMnemonic: json['PathMnemonic'] ?? '',
      frequencies:
          (json['Frequencies'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      losses:
          (json['Losses'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      measured: json['Measured'] ?? false,
    );
  }
}

class TSMInternalLossMeasured {
  final InternalLossPMOrCableEntry pm;
  final InternalLossPMOrCableEntry cable;
  final List<InternalLossEntry> paths;

  TSMInternalLossMeasured({
    required this.pm,
    required this.cable,
    required this.paths,
  });

  factory TSMInternalLossMeasured.fromJson(Map<String, dynamic> json) {
    return TSMInternalLossMeasured(
      pm: InternalLossPMOrCableEntry.fromJson(json['PM'] ?? {}),
      cable: InternalLossPMOrCableEntry.fromJson(json['Cable'] ?? {}),
      paths:
          (json['Paths'] as List?)
              ?.map((e) => InternalLossEntry.fromJson(e))
              .toList() ??
          [],
    );
  }
}

class TSMInternalLossMetadata {
  final List<String> deviceProfiles;
  final TSMInternalLossMeasured measuredLoss;
  final bool ok;
  final String message;

  TSMInternalLossMetadata({
    required this.deviceProfiles,
    required this.measuredLoss,
    required this.ok,
    required this.message,
  });

  factory TSMInternalLossMetadata.fromJson(Map<String, dynamic> json) {
    return TSMInternalLossMetadata(
      deviceProfiles: List<String>.from(json['DeviceProfile'] ?? []),
      measuredLoss: TSMInternalLossMeasured.fromJson(
        json['MeasuredLoss'] ?? {},
      ),
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class InternalLossMeasurementRequest {
  final String deviceProfile;
  final String pmChannel;
  final String mode;
  final String inputPort;
  final String outputPort;

  InternalLossMeasurementRequest({
    required this.deviceProfile,
    required this.pmChannel,
    required this.mode,
    required this.inputPort,
    required this.outputPort,
  });

  Map<String, dynamic> toJson() {
    return {
      'DeviceProfile': deviceProfile,
      'PMChannel': pmChannel,
      'Mode': mode,
      'InputPort': inputPort,
      'OutputPort': outputPort,
    };
  }
}

class AttnRange {
  final double max;
  final double min;
  final double stepSize;

  AttnRange({required this.max, required this.min, required this.stepSize});

  factory AttnRange.fromJson(Map<String, dynamic> json) {
    return AttnRange(
      max: (json['Max'] as num?)?.toDouble() ?? 0.0,
      min: (json['Min'] as num?)?.toDouble() ?? 0.0,
      stepSize: (json['StepSize'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class AttnMetaData {
  final List<String> deviceProfile;
  final List<String> receiver;
  final List<String> sprectrumProfile;
  final List<String> tsmConfig;
  final List<String> gtxComponents;
  final Map<String, AttnRange> attnRanges;
  final bool ok;
  final String message;

  AttnMetaData({
    required this.deviceProfile,
    required this.receiver,
    required this.sprectrumProfile,
    required this.tsmConfig,
    required this.gtxComponents,
    required this.attnRanges,
    required this.ok,
    required this.message,
  });

  factory AttnMetaData.fromJson(Map<String, dynamic> json) {
    var attnRangesMap = <String, AttnRange>{};
    if (json['AttnRanges'] != null) {
      (json['AttnRanges'] as Map<String, dynamic>).forEach((key, value) {
        attnRangesMap[key] = AttnRange.fromJson(value);
      });
    }

    return AttnMetaData(
      deviceProfile: List<String>.from(json['DeviceProfile'] ?? []),
      receiver: List<String>.from(json['Receiver'] ?? []),
      sprectrumProfile: List<String>.from(json['SprectrumProfile'] ?? []),
      tsmConfig: List<String>.from(json['TSMConfig'] ?? []),
      gtxComponents: List<String>.from(json['GTxComponents'] ?? []),
      attnRanges: attnRangesMap,
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class AttnMeasurementStatus {
  final int slNo;
  final double setAttn;
  final double measuredAttn;
  final double deviation;
  final bool hasData;
  final bool completed;
  final bool error;
  final String message;
  final bool plotDeviation;

  AttnMeasurementStatus({
    required this.slNo,
    required this.setAttn,
    required this.measuredAttn,
    required this.deviation,
    required this.hasData,
    required this.completed,
    required this.error,
    required this.message,
    required this.plotDeviation,
  });

  factory AttnMeasurementStatus.fromJson(Map<String, dynamic> json) {
    return AttnMeasurementStatus(
      slNo: json['SlNo'] ?? 0,
      setAttn: (json['SetAttn'] as num?)?.toDouble() ?? 0.0,
      measuredAttn: (json['MeasuredAttn'] as num?)?.toDouble() ?? 0.0,
      deviation: (json['Deviation'] as num?)?.toDouble() ?? 0.0,
      hasData: json['HasData'] ?? false,
      completed: json['Completed'] ?? false,
      error: json['Error'] ?? false,
      message: json['Message'] ?? '',
      plotDeviation: json['PlotDeviation'] ?? false,
    );
  }
}

class CorrectedDeviation {
  final double setValue;
  final double measuredDeviation;
  final double correctedDeviation;

  CorrectedDeviation({
    required this.setValue,
    required this.measuredDeviation,
    required this.correctedDeviation,
  });

  factory CorrectedDeviation.fromJson(Map<String, dynamic> json) {
    return CorrectedDeviation(
      setValue: (json['SetValue'] as num?)?.toDouble() ?? 0.0,
      measuredDeviation: (json['MeasuredDeviation'] as num?)?.toDouble() ?? 0.0,
      correctedDeviation:
          (json['CorrectedDeviation'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class AttnProgressResponse {
  final AttnMeasurementStatus? measurementStatus;
  final List<CorrectedDeviation>? deviations;
  final bool ok;
  final String message;

  AttnProgressResponse({
    this.measurementStatus,
    this.deviations,
    required this.ok,
    required this.message,
  });

  factory AttnProgressResponse.fromJson(Map<String, dynamic> json) {
    return AttnProgressResponse(
      measurementStatus: json['measurementStatus'] != null
          ? AttnMeasurementStatus.fromJson(json['measurementStatus'])
          : null,
      deviations: (json['deviations'] as List?)
          ?.map((e) => CorrectedDeviation.fromJson(e))
          .toList(),
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
    );
  }
}

class CableLossRecord {
  final int slNo;
  final String cableName;
  final double length;
  final String date;
  final String time;
  final List<MeasurementPoint> measurements;

  CableLossRecord({
    required this.slNo,
    required this.cableName,
    required this.length,
    required this.date,
    required this.time,
    required this.measurements,
  });

  factory CableLossRecord.fromJson(Map<String, dynamic> json) {
    return CableLossRecord(
      slNo: json['slNo'] ?? 0,
      cableName: json['cableName'] ?? '',
      length: (json['length'] as num?)?.toDouble() ?? 0.0,
      date: json['date'] ?? '',
      time: json['time'] ?? '',
      measurements:
          (json['measurements'] as List?)
              ?.map((e) => MeasurementPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class CableLossResponse {
  final List<CableLossRecord> history;
  final bool ok;
  final String message;

  CableLossResponse({
    required this.history,
    required this.ok,
    required this.message,
  });

  factory CableLossResponse.fromJson(Map<String, dynamic> json) {
    return CableLossResponse(
      history:
          (json['history'] as List?)
              ?.map((e) => CableLossRecord.fromJson(e))
              .toList() ??
          [],
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
    );
  }
}

class TVACCableLossResponse {
  final TVACCableLossRecord? latestRecord;
  final List<TVACCableLossRecord> history;
  final bool isPmZeroed;
  final bool ok;
  final String message;

  TVACCableLossResponse({
    this.latestRecord,
    required this.history,
    required this.isPmZeroed,
    required this.ok,
    required this.message,
  });

  factory TVACCableLossResponse.fromJson(Map<String, dynamic> json) {
    return TVACCableLossResponse(
      latestRecord: json['latestRecord'] != null
          ? TVACCableLossRecord.fromJson(json['latestRecord'])
          : null,
      history:
          (json['history'] as List?)
              ?.map((e) => TVACCableLossRecord.fromJson(e))
              .toList() ??
          [],
      isPmZeroed: json['isPmZeroed'] ?? false,
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
    );
  }
}

class ReportMetadata {
  final String phase;
  final String config;
  final String testType;
  final String testCategory;
  final String date;
  final String time;
  final String remarks;
  final bool vsaUsed;
  final bool ppmUsed;

  ReportMetadata({
    required this.phase,
    required this.config,
    required this.testType,
    required this.testCategory,
    required this.date,
    required this.time,
    required this.remarks,
    required this.vsaUsed,
    required this.ppmUsed,
  });

  factory ReportMetadata.fromJson(Map<String, dynamic> json) {
    return ReportMetadata(
      phase: json['phase'] ?? '',
      config: json['config'] ?? '',
      testType: json['testType'] ?? '',
      testCategory: json['testCategory'] ?? '',
      date: json['date'] ?? '',
      time: json['time'] ?? '',
      remarks: json['remarks'] ?? '',
      vsaUsed: json['vsaUsed'] ?? false,
      ppmUsed: json['ppmUsed'] ?? false,
    );
  }
}

class ReportsResponse {
  final bool ok;
  final String message;
  final List<ReportMetadata> reports;
  final List<String> allVsaParams;
  final List<String> selectedVsaParams;
  final List<String> allPpmParams;
  final List<String> selectedPpmParams;

  ReportsResponse({
    required this.ok,
    required this.message,
    required this.reports,
    required this.allVsaParams,
    required this.selectedVsaParams,
    required this.allPpmParams,
    required this.selectedPpmParams,
  });

  factory ReportsResponse.fromJson(Map<String, dynamic> json) {
    return ReportsResponse(
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
      reports:
          (json['reports'] as List?)
              ?.map((e) => ReportMetadata.fromJson(e))
              .toList() ??
          [],
      allVsaParams: (json['allVsaParams'] as List?)?.cast<String>() ?? [],
      selectedVsaParams:
          (json['selectedVsaParams'] as List?)?.cast<String>() ?? [],
      allPpmParams: (json['allPpmParams'] as List?)?.cast<String>() ?? [],
      selectedPpmParams:
          (json['selectedPpmParams'] as List?)?.cast<String>() ?? [],
    );
  }
}

class MeasurementStatus {
  final String message;
  final bool error;
  final bool completed;
  final bool success;

  MeasurementStatus({
    required this.message,
    required this.error,
    this.completed = false,
    this.success = false,
  });

  factory MeasurementStatus.fromJson(Map<String, dynamic> json) {
    return MeasurementStatus(
      message: json['message'] ?? '',
      error: json['error'] ?? false,
      completed: json['completed'] ?? false,
      success: json['success'] ?? false,
    );
  }
}

class RTStatus {
  final String message;
  final bool completed;
  final bool success;
  final bool error;

  RTStatus({
    required this.message,
    required this.completed,
    required this.success,
    required this.error,
  });

  factory RTStatus.fromJson(Map<String, dynamic> json) {
    return RTStatus(
      message: json['message'] ?? '',
      completed: json['completed'] ?? false,
      success: json['success'] ?? false,
      error: json['error'] ?? false,
    );
  }
}

class GTxSpectrum {
  final double span;
  final double rbw;
  final double vbw;

  GTxSpectrum({required this.span, required this.rbw, required this.vbw});

  Map<String, dynamic> toJson() => {'Span': span, 'RBW': rbw, 'VBW': vbw};
}

class GTxTneRequest {
  final String deviceProfile;
  final String component;
  final double intermediateFrequency;
  final double cableLoss;
  final String modulationScheme;
  final double subCarrierFrequency;
  final double modIndex;
  final double frequencyDeviation;
  final GTxSpectrum frequencySpectrum;
  final GTxSpectrum powerSpectrum;
  final GTxSpectrum inBandSpectrum;
  final GTxSpectrum outBandSpectrum;

  GTxTneRequest({
    required this.deviceProfile,
    required this.component,
    required this.intermediateFrequency,
    required this.cableLoss,
    required this.modulationScheme,
    required this.subCarrierFrequency,
    required this.modIndex,
    required this.frequencyDeviation,
    required this.frequencySpectrum,
    required this.powerSpectrum,
    required this.inBandSpectrum,
    required this.outBandSpectrum,
  });

  Map<String, dynamic> toJson() => {
    'DeviceProfile': deviceProfile,
    'Component': component,
    'IntermediateFrequency': intermediateFrequency,
    'CableLoss': cableLoss,
    'ModulationScheme': modulationScheme,
    'SubCarrierFrequency': subCarrierFrequency,
    'ModIndex': modIndex,
    'FrequencyDeviation': frequencyDeviation,
    'FrequencySpectrum': frequencySpectrum.toJson(),
    'PowerSpectrum': powerSpectrum.toJson(),
    'InBandSpectrum': inBandSpectrum.toJson(),
    'OutBandSpectrum': outBandSpectrum.toJson(),
  };
}

class GTxResult {
  final double powerSpec;
  final double powerMeasured;
  final double powerDeviation;
  final bool powerMeasurementCompleted;
  final double freqSpecMHz;
  final double freqMeasuredMHz;
  final double freqDeviationkHz;
  final bool freqMeasurementCompleted;
  final List<double> inBandSpuriousFreqOffsetskHz;
  final List<double> inBandPowerOffsets;
  final bool inBandSpuriousMeasurementCompleted;
  final List<double> outBandSpuriousFreqOffsetskHz;
  final List<double> outBandPowerOffsets;
  final bool outBandSpuriousMeasurementCompleted;
  final List<double> harmonicsFreqMHz;
  final List<double> harmonicsMeasureddBm;
  final List<bool> harmonicsPresent;
  final List<double> harmonicsNoiseFloor;
  final bool harmonicsMeasurementCompleted;
  final bool modIndexApplicable;
  final double modIndexSet;
  final double modIndexMeasured;
  final double modIndexDeviation;
  final bool modIndexMeasurementCompleted;
  final bool frequencyDeviationApplicable;
  final double frequencyDeviationSet;
  final double frequencyDeviationMeasured;
  final double frequencyDeviationDeviation;
  final bool frequencyDeviationMeasurementCompleted;
  final double phaseNoiseAt1Khz;
  final double phaseNoiseAt10Khz;
  final double phaseNoiseAt100Khz;
  final double phaseNoiseAt1Mhz;
  final bool phaseNoiseMeasurementCompleted;

  GTxResult({
    required this.powerSpec,
    required this.powerMeasured,
    required this.powerDeviation,
    required this.powerMeasurementCompleted,
    required this.freqSpecMHz,
    required this.freqMeasuredMHz,
    required this.freqDeviationkHz,
    required this.freqMeasurementCompleted,
    required this.inBandSpuriousFreqOffsetskHz,
    required this.inBandPowerOffsets,
    required this.inBandSpuriousMeasurementCompleted,
    required this.outBandSpuriousFreqOffsetskHz,
    required this.outBandPowerOffsets,
    required this.outBandSpuriousMeasurementCompleted,
    required this.harmonicsFreqMHz,
    required this.harmonicsMeasureddBm,
    required this.harmonicsPresent,
    required this.harmonicsNoiseFloor,
    required this.harmonicsMeasurementCompleted,
    required this.modIndexApplicable,
    required this.modIndexSet,
    required this.modIndexMeasured,
    required this.modIndexDeviation,
    required this.modIndexMeasurementCompleted,
    required this.frequencyDeviationApplicable,
    required this.frequencyDeviationSet,
    required this.frequencyDeviationMeasured,
    required this.frequencyDeviationDeviation,
    required this.frequencyDeviationMeasurementCompleted,
    required this.phaseNoiseAt1Khz,
    required this.phaseNoiseAt10Khz,
    required this.phaseNoiseAt100Khz,
    required this.phaseNoiseAt1Mhz,
    required this.phaseNoiseMeasurementCompleted,
  });

  factory GTxResult.fromJson(Map<String, dynamic> json) {
    return GTxResult(
      powerSpec: (json['PowerSpec'] as num?)?.toDouble() ?? 0.0,
      powerMeasured: (json['PowerMeasured'] as num?)?.toDouble() ?? 0.0,
      powerDeviation: (json['PowerDeviation'] as num?)?.toDouble() ?? 0.0,
      powerMeasurementCompleted: json['PowerMeasurementCompleted'] ?? false,
      freqSpecMHz: (json['FreqSpecMHz'] as num?)?.toDouble() ?? 0.0,
      freqMeasuredMHz: (json['FreqMeasuredMHz'] as num?)?.toDouble() ?? 0.0,
      freqDeviationkHz: (json['FreqDeviationkHz'] as num?)?.toDouble() ?? 0.0,
      freqMeasurementCompleted: json['FreqMeasurementCompleted'] ?? false,
      inBandSpuriousFreqOffsetskHz:
          (json['InBandSpuriousFreqOffsetskHz'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      inBandPowerOffsets:
          (json['InBandPowerOffsets'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      inBandSpuriousMeasurementCompleted:
          json['InBandSpuriousMeasurementCompleted'] ?? false,
      outBandSpuriousFreqOffsetskHz:
          (json['OutBandSpuriousFreqOffsetskHz'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      outBandPowerOffsets:
          (json['OutBandPowerOffsets'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      outBandSpuriousMeasurementCompleted:
          json['OutBandSpuriousMeasurementCompleted'] ?? false,
      harmonicsFreqMHz:
          (json['HarmonicsFreqMHz'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      harmonicsMeasureddBm:
          (json['HarmonicsMeasureddBm'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      harmonicsPresent:
          (json['HarmonicsPresent'] as List?)?.cast<bool>() ?? [],
      harmonicsNoiseFloor:
          (json['HarmonicsNoiseFloor'] as List?)
              ?.map((e) => (e as num).toDouble())
              .toList() ??
          [],
      harmonicsMeasurementCompleted:
          json['HarmonicsMeasurementCompleted'] ?? false,
      modIndexApplicable: json['ModIndexApplicable'] ?? false,
      modIndexSet: (json['ModIndexSet'] as num?)?.toDouble() ?? 0.0,
      modIndexMeasured: (json['ModIndexMeasured'] as num?)?.toDouble() ?? 0.0,
      modIndexDeviation: (json['ModIndexDeviation'] as num?)?.toDouble() ?? 0.0,
      modIndexMeasurementCompleted:
          json['ModIndexMeasurementCompleted'] ?? false,
      frequencyDeviationApplicable:
          json['FrequencyDeviationApplicable'] ?? false,
      frequencyDeviationSet:
          (json['FrequencyDeviationSet'] as num?)?.toDouble() ?? 0.0,
      frequencyDeviationMeasured:
          (json['FrequencyDeviationMeasured'] as num?)?.toDouble() ?? 0.0,
      frequencyDeviationDeviation:
          (json['FrequencyDeviationDeviation'] as num?)?.toDouble() ?? 0.0,
      frequencyDeviationMeasurementCompleted:
          json['FrequencyDeviationMeasurementCompleted'] ?? false,
      phaseNoiseAt1Khz: (json['PhaseNoiseAt1Khz'] as num?)?.toDouble() ?? 0.0,
      phaseNoiseAt10Khz: (json['PhaseNoiseAt10Khz'] as num?)?.toDouble() ?? 0.0,
      phaseNoiseAt100Khz:
          (json['PhaseNoiseAt100Khz'] as num?)?.toDouble() ?? 0.0,
      phaseNoiseAt1Mhz: (json['PhaseNoiseAt1Mhz'] as num?)?.toDouble() ?? 0.0,
      phaseNoiseMeasurementCompleted:
          json['PhaseNoiseMeasurementCompleted'] ?? false,
    );
  }
}

class DatabaseMetadata {
  final List<String> testPhases;
  final bool ok;
  final String message;

  DatabaseMetadata({
    required this.testPhases,
    required this.ok,
    required this.message,
  });

  factory DatabaseMetadata.fromJson(Map<String, dynamic> json) {
    return DatabaseMetadata(
      testPhases: List<String>.from(json['TestPhases'] ?? []),
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class ConfigsForLossResponse {
  final List<String> configs;
  final bool ok;
  final String message;

  ConfigsForLossResponse({
    required this.configs,
    required this.ok,
    required this.message,
  });

  factory ConfigsForLossResponse.fromJson(Map<String, dynamic> json) {
    return ConfigsForLossResponse(
      configs: List<String>.from(json['Configs'] ?? []),
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class LossProfileResponse {
  final String profile;
  final bool ok;
  final String message;

  LossProfileResponse({
    required this.profile,
    required this.ok,
    required this.message,
  });

  factory LossProfileResponse.fromJson(Map<String, dynamic> json) {
    return LossProfileResponse(
      profile: json['Profile'] ?? '',
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
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

class StabilityDataUpdate {
  final String description;
  final double value;
  final DateTime timestamp;

  StabilityDataUpdate({
    required this.description,
    required this.value,
    required this.timestamp,
  });

  factory StabilityDataUpdate.fromJson(Map<String, dynamic> json) {
    return StabilityDataUpdate(
      description: json['Description'] ?? '',
      value: (json['Value'] as num?)?.toDouble() ?? 0.0,
      timestamp: json['Timestamp'] != null
          ? DateTime.parse(json['Timestamp'])
          : DateTime.now(),
    );
  }
}

class StabilityResponse {
  final List<StabilityDataUpdate> updates;
  final bool ok;
  final String message;

  StabilityResponse({
    required this.updates,
    required this.ok,
    required this.message,
  });

  factory StabilityResponse.fromJson(Map<String, dynamic> json) {
    return StabilityResponse(
      updates:
          (json['Updates'] as List?)
              ?.map(
                (e) => StabilityDataUpdate.fromJson(e as Map<String, dynamic>),
              )
              .toList() ??
          [],
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class StabilityParameterSelection {
  final String description;
  final String instrumentType;
  final String instrument;
  final String parameter;
  final String details;
  final Map<String, dynamic>? extraDetails;

  StabilityParameterSelection({
    required this.description,
    required this.instrumentType,
    required this.instrument,
    required this.parameter,
    required this.details,
    this.extraDetails,
  });

  Map<String, dynamic> toJson() => {
    'description': description,
    'instrumentType': instrumentType,
    'instrument': instrument,
    'parameter': parameter,
    'details': details,
    'extraDetails': extraDetails,
  };

  factory StabilityParameterSelection.fromJson(Map<String, dynamic> json) =>
      StabilityParameterSelection(
        description: json['description'] ?? '',
        instrumentType: json['instrumentType'] ?? '',
        instrument: json['instrument'] ?? '',
        parameter: json['parameter'] ?? '',
        details: json['details'] ?? '',
        extraDetails: json['extraDetails'] != null
            ? Map<String, dynamic>.from(json['extraDetails'])
            : null,
      );
}

class StabilityReportMetadataResponse {
  final List<int> id;
  final List<String> date;
  final List<String> time;
  final List<List<String>> parameters;
  final bool ok;
  final String message;

  StabilityReportMetadataResponse({
    required this.id,
    required this.date,
    required this.time,
    required this.parameters,
    required this.ok,
    required this.message,
  });

  factory StabilityReportMetadataResponse.fromJson(Map<String, dynamic> json) {
    return StabilityReportMetadataResponse(
      id: List<int>.from(json['id'] ?? []),
      date: List<String>.from(json['date'] ?? []),
      time: List<String>.from(json['time'] ?? []),
      parameters: (json['parameters'] as List?)
          ?.map((e) => List<String>.from(e))
          .toList() ?? [],
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
    );
  }
}

class StabilityPointsResponse {
  final List<StabilityPointData> points;
  final bool ok;
  final String message;

  StabilityPointsResponse({
    required this.points,
    required this.ok,
    required this.message,
  });

  factory StabilityPointsResponse.fromJson(Map<String, dynamic> json) {
    return StabilityPointsResponse(
      points: (json['points'] as List?)
          ?.map((e) => StabilityPointData.fromJson(e))
          .toList() ?? [],
      ok: json['ok'] ?? false,
      message: json['message'] ?? '',
    );
  }
}

class StabilityPointData {
  final int tsInt;
  final String timeStamp;
  final String description;
  final double value;

  StabilityPointData({
    required this.tsInt,
    required this.timeStamp,
    required this.description,
    required this.value,
  });

  factory StabilityPointData.fromJson(Map<String, dynamic> json) {
    return StabilityPointData(
      tsInt: json['TimeStampInt'] ?? 0,
      timeStamp: json['TimeStamp'] ?? '',
      description: json['Description'] ?? '',
      value: (json['Value'] as num?)?.toDouble() ?? 0.0,
    );
  }
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

class ReadSpectrumResponse {
  final double centerFrequency;
  final double span;
  final double rbw;
  final double vbw;
  final double referenceLevel;
  final bool ok;
  final String message;

  ReadSpectrumResponse({
    required this.centerFrequency,
    required this.span,
    required this.rbw,
    required this.vbw,
    required this.referenceLevel,
    required this.ok,
    required this.message,
  });

  factory ReadSpectrumResponse.fromJson(Map<String, dynamic> json) {
    return ReadSpectrumResponse(
      centerFrequency: (json['CenterFrequency'] as num?)?.toDouble() ?? 0.0,
      span: (json['Span'] as num?)?.toDouble() ?? 0.0,
      rbw: (json['RBW'] as num?)?.toDouble() ?? 0.0,
      vbw: (json['VBW'] as num?)?.toDouble() ?? 0.0,
      referenceLevel: (json['ReferenceLevel'] as num?)?.toDouble() ?? 0.0,
      ok: json['OK'] ?? false,
      message: json['Message'] ?? '',
    );
  }
}

class ServerService extends ChangeNotifier {
  WebSocketChannel? _channel;
  WebSocketChannel? _attnChannel;
  WebSocketChannel? _cableLossChannel;
  WebSocketChannel? _tvacCableLossChannel;
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
      host = web.window.location.host;
    }
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
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

  Future<BootstrapData?> fetchBootstrapData() async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/bootstrapData';

    try {
      final response = await http.get(Uri.parse(url));

      if (response.statusCode == 200) {
        final data = BootstrapData.fromJson(jsonDecode(response.body));
        _status = ServerStatus(
          satelliteName: _status.satelliteName,
          testPhase: _status.testPhase,
          memoryUsed: _status.memoryUsed,
          cpuUsed: _status.cpuUsed,
          isConnected: _status.isConnected,
          bootstrapData: data,
        );
        notifyListeners();
        return data;
      }
    } catch (e) {
      debugPrint('Error fetching Bootstrap Data: $e');
    }
    return null;
  }


  Future<LinkStatus?> fetchLinkStatus(String tsmSelected) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
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
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
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
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
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
      host = web.window.location.host;
    }
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
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

  void sendAbort() {
    _progressChannel?.sink.add(
      jsonEncode({
        'Parameters': ['abort'],
      }),
    );
  }

  void closeTestProgress() {
    _progressChannel?.sink.close();
    _progressChannel = null;
  }

  WebSocketChannel? _monitorChannel;

  Stream<MonitorResponse> connectMonitor(String type, String instrument) {
    String host;
    if (kDebugMode) {
      host = 'localhost:8080';
    } else {
      host = web.window.location.host;
    }
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
    final url = '$protocol://$host/monitor';

    _monitorChannel = WebSocketChannel.connect(Uri.parse(url));

    // Send initial request
    _monitorChannel!.sink.add(
      jsonEncode({'InstrumentType': type, 'Instrument': instrument}),
    );

    return _monitorChannel!.stream.map((data) {
      return MonitorResponse.fromJson(jsonDecode(data));
    });
  }

  void closeMonitor() {
    _monitorChannel?.sink.close();
    _monitorChannel = null;
  }



  Future<CableLossResponse?> fetchCableMeasuredDetails() async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/getCableMeasuredDetails';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        return CableLossResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching Cable Loss Details: $e');
    }
    return null;
  }


  Future<Ack?> fetchReportPDF(String date, String time) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/getReportPDF';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'date': date, 'time': time}),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching Report PDF: $e');
    }
    return null;
  }

  Future<Ack?> regenerateReport({
    required String date,
    required String time,
    required List<String> ppmParameters,
    required List<String> vsaParameters,
  }) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/regenerateReport';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'date': date,
          'time': time,
          'ppmParameters': ppmParameters,
          'vsaParameters': vsaParameters,
        }),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error regenerating report: $e');
    }
    return null;
  }

  Stream<MeasurementStatus> streamCableLossAction(
    Map<String, dynamic> request,
  ) {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
    final url = '$protocol://$host/measureCableLoss';

    _cableLossChannel = WebSocketChannel.connect(Uri.parse(url));
    _cableLossChannel!.sink.add(jsonEncode(request));

    return _cableLossChannel!.stream
        .map((event) {
          return MeasurementStatus.fromJson(jsonDecode(event));
        })
        .handleError((error) {
          debugPrint('Cable Action Stream Error: $error');
          return MeasurementStatus(message: error.toString(), error: true);
        });
  }

  void abortCableLossMeasurement() {
    if (_cableLossChannel != null) {
      _cableLossChannel!.sink.add('abort');
    }
  }


  WebSocketChannel? _stabilityChannel;

  Stream<StabilityResponse> connectStability(
    List<StabilityParameterSelection> parameters,
    String profileName,
  ) {
    String host;
    if (kDebugMode) {
      host = 'localhost:8080';
    } else {
      host = web.window.location.host;
    }
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
    final url = '$protocol://$host/stability';

    _stabilityChannel = WebSocketChannel.connect(Uri.parse(url));

    // Send initial request
    _stabilityChannel!.sink.add(
      jsonEncode({
        'ProfileName': profileName,
        'Parameters': parameters.map((p) => p.toJson()).toList(),
      }),
    );

    return _stabilityChannel!.stream.map((data) {
      return StabilityResponse.fromJson(jsonDecode(data));
    });
  }

  void closeStability() {
    _stabilityChannel?.sink.close();
    _stabilityChannel = null;
  }

  void sendAbortStability() {
    _stabilityChannel?.sink.add(jsonEncode({'Action': 'abort'}));
  }



  Future<Ack?> setSpectrum({
    required String sa,
    required double centerFrequency,
    required double span,
    required double rbw,
    required double vbw,
    required bool autoReference,
    required double referenceLevel,
    required String mode,
  }) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/setSpectrum';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'SA': sa,
          'CenterFrequency': centerFrequency,
          'Span': span,
          'RBW': rbw,
          'VBW': vbw,
          'AutoReference': autoReference,
          'ReferenceLevel': referenceLevel,
          'Mode': mode,
        }),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error setting Spectrum: $e');
    }
    return null;
  }

  Future<ReadSpectrumResponse?> readSpectrum(String sa) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/readSpectrum';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'SA': sa}),
      );

      if (response.statusCode == 200) {
        return ReadSpectrumResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error reading Spectrum: $e');
    }
    return null;
  }

  Stream<AttnProgressResponse> streamAttnAction(Map<String, dynamic> request) {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
    final url = '$protocol://$host/measureAttn';

    _attnChannel = WebSocketChannel.connect(Uri.parse(url));
    _attnChannel!.sink.add(jsonEncode(request));

    return _attnChannel!.stream
        .map((event) {
          return AttnProgressResponse.fromJson(jsonDecode(event));
        })
        .handleError((error) {
          debugPrint('Attn Action Stream Error: $error');
          return AttnProgressResponse(ok: false, message: error.toString());
        });
  }

  void abortAttnMeasurement() {
    if (_attnChannel != null) {
      _attnChannel!.sink.add('abort');
    }
  }


  Future<ReadSpectrumResponse?> dumpSpectrum(String sa) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/dumpSpectrum';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'SA': sa}),
      );

      if (response.statusCode == 200) {
        return ReadSpectrumResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error dumping Spectrum: $e');
    }
    return null;
  }

  Future<ReadSpectrumResponse?> dumpTrace({
    required String sa,
    required int tracePoints,
  }) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/dumpTrace';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'SA': sa, 'TracePoints': tracePoints}),
      );

      if (response.statusCode == 200) {
        return ReadSpectrumResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error dumping Trace: $e');
    }
    return null;
  }

  Future<ReadSpectrumResponse?> dumpScreenshot({
    required String vsa,
    required String mode,
  }) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/dumpScreenshot';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'VSA': vsa, 'Mode': mode}),
      );

      if (response.statusCode == 200) {
        return ReadSpectrumResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error dumping Screenshot: $e');
    }
    return null;
  }

  Future<ReadSpectrumResponse?> saveSpectrum({
    required String spectrumBase64,
    required String remark,
  }) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/saveSpectrum';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'Spectrum': spectrumBase64, 'Remark': remark}),
      );

      if (response.statusCode == 200) {
        return ReadSpectrumResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error saving Spectrum: $e');
    }
    return null;
  }

  Future<TVACCableLossResponse?> fetchTVACCableMeasuredDetails() async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/getTVACCableMeasuredDetails';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
      );

      if (response.statusCode == 200) {
        return TVACCableLossResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching TVAC Cable Loss Details: $e');
    }
    return null;
  }


  WebSocketChannel? _tsmInternalLossChannel;

  Stream<dynamic> streamTSMInternalLossAction(
    InternalLossMeasurementRequest request,
  ) {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
    final url = '$protocol://$host/measureTSMInternalLoss';

    _tsmInternalLossChannel = WebSocketChannel.connect(Uri.parse(url));
    _tsmInternalLossChannel!.sink.add(jsonEncode(request.toJson()));

    return _tsmInternalLossChannel!.stream
        .map((event) {
          final data = jsonDecode(event);
          if (data['PM'] != null ||
              data['Cable'] != null ||
              data['Paths'] != null) {
            return TSMInternalLossMeasured.fromJson(data);
          }
          return MeasurementStatus.fromJson(data);
        })
        .handleError((error) {
          debugPrint('TSM Internal Loss Stream Error: $error');
          return MeasurementStatus(
            message: error.toString(),
            error: true,
            completed: true,
          );
        });
  }

  void abortTSMInternalLossMeasurement() {
    if (_tsmInternalLossChannel != null) {
      _tsmInternalLossChannel!.sink.add('abort');
    }
  }

  Stream<MeasurementStatus> streamTVACCableLossAction(
    Map<String, dynamic> request,
  ) {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
    final url = '$protocol://$host/measureTVACCableLoss';

    _tvacCableLossChannel = WebSocketChannel.connect(Uri.parse(url));
    _tvacCableLossChannel!.sink.add(jsonEncode(request));

    return _tvacCableLossChannel!.stream
        .map((event) {
          return MeasurementStatus.fromJson(jsonDecode(event));
        })
        .handleError((error) {
          debugPrint('TVAC Action Stream Error: $error');
          return MeasurementStatus(message: error.toString(), error: true);
        });
  }

  void abortTVACCableLossMeasurement() {
    if (_tvacCableLossChannel != null) {
      _tvacCableLossChannel!.sink.add('abort');
    }
  }


  Future<ConfigsForLossResponse?> fetchConfigsForLoss(
    String phase,
    bool isUplink,
  ) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final endpoint = isUplink
        ? '/getConfigsForUplink'
        : '/getConfigsForDownlink';
    final url = '$protocol://$host$endpoint';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'TestPhase': phase}),
      );

      if (response.statusCode == 200) {
        return ConfigsForLossResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching Configs for Loss: $e');
    }
    return null;
  }

  Future<LossProfileResponse?> fetchLossProfile(
    String phase,
    String config,
    bool isUplink,
  ) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final endpoint = isUplink
        ? '/getUplinkLossProfile'
        : '/getDownlinkLossProfile';
    final url = '$protocol://$host$endpoint';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'TestPhase': phase, 'Config': config}),
      );

      if (response.statusCode == 200) {
        return LossProfileResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching Loss Profile: $e');
    }
    return null;
  }

  Future<Ack?> saveLossProfile(
    String phase,
    String config,
    String profile,
    bool isUplink,
  ) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final endpoint = isUplink
        ? '/saveUplinkLossProfile'
        : '/saveDownlinkLossProfile';
    final url = '$protocol://$host$endpoint';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'TestPhase': phase,
          'Config': config,
          'Profile': profile,
        }),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error saving Loss Profile: $e');
    }
    return null;
  }

  Future<Ack?> selectTestPhase(String phase) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/selectTestPhase';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'TestPhase': phase}),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error selecting Test Phase: $e');
    }
    return null;
  }

  Future<Ack?> addNewTestPhase(String newPhase, String copyFrom) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:'
        ? 'https'
        : 'http';
    final url = '$protocol://$host/addNewTestPhase';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'NewPhase': newPhase, 'CopyFrom': copyFrom}),
      );

      if (response.statusCode == 200) {
        return Ack.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error adding new Test Phase: $e');
    }
    return null;
  }


  Future<StabilityPointsResponse?> fetchStabilityPoints(int id, String parameter) async {
    final host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    final protocol = web.window.location.protocol == 'https:' ? 'https' : 'http';
    final url = '$protocol://$host/getStabilityPoints';

    try {
      final response = await http.post(
        Uri.parse(url),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'id': id,
          'parameter': parameter,
        }),
      );

      if (response.statusCode == 200) {
        return StabilityPointsResponse.fromJson(jsonDecode(response.body));
      }
    } catch (e) {
      debugPrint('Error fetching Stability Points: $e');
    }
    return null;
  }

  WebSocketChannel? _gtxTneChannel;

  Stream<dynamic> connectGTxTne(GTxTneRequest request) {
    String host = kDebugMode ? 'localhost:8080' : web.window.location.host;
    String protocol = web.window.location.protocol == 'https:' ? 'wss' : 'ws';
    String url = '$protocol://$host/conductGTxTne';

    _gtxTneChannel = WebSocketChannel.connect(Uri.parse(url));
    _gtxTneChannel!.sink.add(jsonEncode(request.toJson()));

    return _gtxTneChannel!.stream.map((data) {
      final decoded = jsonDecode(data);
      if (decoded.containsKey('message')) {
        return RTStatus.fromJson(decoded);
      } else {
        return GTxResult.fromJson(decoded);
      }
    });
  }

  void abortGTxTne() {
    _gtxTneChannel?.sink.add('abort');
  }

  void closeGTxTne() {
    _gtxTneChannel?.sink.close();
    _gtxTneChannel = null;
  }



  @override
  void dispose() {
    _channel?.sink.close();
    super.dispose();
  }
}
