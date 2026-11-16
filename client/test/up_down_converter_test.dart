import 'package:flutter_test/flutter_test.dart';
import 'package:prism_client/services/server_service.dart';

void main() {
  group('UCDCCableLoss and UCDCRequest Tests', () {
    test('UCDCCableLoss serialization and deserialization', () {
      final loss = UCDCCableLoss(
        inputCableLoss: 1.25,
        loCableLoss: 0.85,
        outputCableLoss: [0.35, 0.45, 0.55],
      );

      final jsonMap = loss.toJson();
      expect(jsonMap['InputCableLoss'], 1.25);
      expect(jsonMap['LOCableLoss'], 0.85);
      expect(jsonMap['OutputCableLoss'], [0.35, 0.45, 0.55]);

      final fromJson = UCDCCableLoss.fromJson(jsonMap);
      expect(fromJson.inputCableLoss, 1.25);
      expect(fromJson.loCableLoss, 0.85);
      expect(fromJson.outputCableLoss, [0.35, 0.45, 0.55]);
    });

    test('UCDCRequest serialization with multiple frequency pair cable losses', () {
      final pair1 = UCDCCableLoss(
        inputCableLoss: 1.0,
        loCableLoss: 0.5,
        outputCableLoss: [0.2, 0.3, 0.4],
      );
      final pair2 = UCDCCableLoss(
        inputCableLoss: 2.0,
        loCableLoss: 1.5,
        outputCableLoss: [0.6, 0.7, 0.8],
      );

      final request = UCDCRequest(
        converterName: 'UC-17-03-00',
        deviceProfile: 'Profile1',
        externalSGName: 'SG-KEYSIGHT-01',
        inputCableLoss: pair1.inputCableLoss,
        inputPower: -25.0,
        loCableLoss: pair1.loCableLoss,
        outputCableLoss: pair1.outputCableLoss,
        cableLosses: [pair1, pair2],
        powerSpectrum: GTxSpectrum(span: 1e6, rbw: 3e3, vbw: 1e3),
        frequencySpectrum: GTxSpectrum(span: 1e7, rbw: 1e4, vbw: 3e3),
        inBandSpectrum: GTxSpectrum(span: 1e6, rbw: 3e3, vbw: 1e3),
        outBandSpectrum: GTxSpectrum(span: 1e8, rbw: 1e4, vbw: 1e3),
        stepSize: 5.0,
        testsSelected: ['UCDC_GAIN_INT_CABLE'],
      );

      final json = request.toJson();
      expect(json['ConverterName'], 'UC-17-03-00');
      expect(json['CableLosses'], isNotNull);
      final cableLossesList = json['CableLosses'] as List;
      expect(cableLossesList.length, 2);
      expect(cableLossesList[0]['InputCableLoss'], 1.0);
      expect(cableLossesList[1]['InputCableLoss'], 2.0);
      expect(cableLossesList[0]['OutputCableLoss'], [0.2, 0.3, 0.4]);
      expect(cableLossesList[1]['OutputCableLoss'], [0.6, 0.7, 0.8]);
    });

    test('Loss entry validation helper logic', () {
      bool isValid(String s) {
        final t = s.trim();
        if (t.isEmpty) return false;
        final val = double.tryParse(t);
        return val != null && !val.isNaN && val >= 0;
      }

      // Empty string should be invalid
      expect(isValid(''), isFalse);
      expect(isValid('   '), isFalse);

      // Non-number string should be invalid
      expect(isValid('abc'), isFalse);
      expect(isValid('12.3.4'), isFalse);

      // Negative number should be invalid
      expect(isValid('-0.5'), isFalse);

      // Valid numbers
      expect(isValid('0'), isTrue);
      expect(isValid('0.0'), isTrue);
      expect(isValid('0.45'), isTrue);
      expect(isValid(' 1.25 '), isTrue);
    });
  });
}
