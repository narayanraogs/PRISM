import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:prism/app_icons.dart';
import 'package:prism/server_communication/classes.dart';
import 'package:prism/server_communication/messages.dart';
import 'package:prism/structures.dart';
import 'package:prism/theme.dart';
import 'package:web_socket_channel/io.dart';

class GtxMeasurement extends StatefulWidget {
  final Global global;

  const GtxMeasurement(this.global, {super.key});

  @override
  State<GtxMeasurement> createState() => _StateGtxMeasurement();
}

class _StateGtxMeasurement extends State<GtxMeasurement> {
  List<String> _profiles = [];
  String _selectedProfile = '';
  String _gtxName = '';
  String _saName = '';
  List<String> _components = [];
  String _selectedComponent = '';

  final TextEditingController _cableLoss = TextEditingController(text: "0.0");
  final TextEditingController _if = TextEditingController(text: "70000000");
  final TextEditingController _subCarFreq = TextEditingController(text: "8000");
  final List<String> _modulations = ["PM", "FM", "CDMA", "PSK", "FSK"];
  String _selectedModulation = "PM";
  final TextEditingController _modIndex = TextEditingController(text: "1.0");
  final TextEditingController _frequencyDeviation = TextEditingController(
    text: "200000",
  );

  final TextEditingController _powerSpan = TextEditingController(
    text: "1000000",
  );
  final TextEditingController _powerRBW = TextEditingController(text: "3000");
  final TextEditingController _powerVBW = TextEditingController(text: "1000");

  final TextEditingController _freqSpan = TextEditingController(
    text: "10000000",
  );
  final TextEditingController _freqRBW = TextEditingController(text: "10000");
  final TextEditingController _freqVBW = TextEditingController(text: "3000");

  final TextEditingController _inBandSpan = TextEditingController(
    text: "1000000",
  );
  final TextEditingController _inBandRBW = TextEditingController(text: "3000");
  final TextEditingController _inBandVBW = TextEditingController(text: "1000");

  final TextEditingController _outBandSpan = TextEditingController(
    text: "10000000",
  );
  final TextEditingController _outBandRBW = TextEditingController(
    text: "10000",
  );
  final TextEditingController _outBandVBW = TextEditingController(text: "3000");

  List<List<String>> _results = [];

  @override
  void initState() {
    super.initState();
    _components = ['IFM-1', 'IFM-2'];
    _selectedComponent = _components.first;
    _getDetails();
  }

  void _getDetails() async {
    GetResponse resp = await getParameters(widget.global, ["DeviceProfiles"]);
    if (!resp.ok) {
      return;
    }
    _profiles = resp.getValue("DeviceProfiles");
    _selectedProfile = _profiles.isNotEmpty ? _profiles.first : "";
    _getDeviceNames();
  }

  void _getDeviceNames() async {
    Acknowledgement ack = await setParameters(
      widget.global,
      ["SelectedDeviceProfile"],
      [_selectedProfile],
    );
    if (!ack.ok) {
      return;
    }
    GetResponse resp = await getParameters(widget.global, [
      "SANameFromDeviceProfile",
      "GTxNameFromDeviceProfile",
    ]);
    if (!resp.ok) {
      return;
    }
    List<String> sas = resp.getValue("SANameFromDeviceProfile");
    _saName = sas.isNotEmpty ? sas.first : "";
    List<String> gtx = resp.getValue("GTxNameFromDeviceProfile");
    _gtxName = gtx.isNotEmpty ? gtx.first : "";
    setState(() {});
  }

  void _getStatusUpdate() {
    String param = 'GTxMeasurementStatus';
    _results = [];
    String url = widget.global.url.replaceFirst("http", "ws");
    IOWebSocketChannel channel = IOWebSocketChannel.connect("$url/rtUpdate");
    ActionRequest request = ActionRequest();
    request.clientID = widget.global.clientID;
    request.action = param;
    channel.sink.add(jsonEncode(request.toJSON()));
    channel.stream.listen((data) {
      if (data.toLowerCase().contains("error")) {
        widget.global.updateNotification(data, NotificationType.failure, null);
        return;
      }
      ParameterValue value = ParameterValue();
      Map<String, dynamic> jsonData = jsonDecode(data);
      value.fromJSON(jsonData);

      String message = value.values[1];
      bool completed = value.values[2] == "Completed";
      bool ok = value.values[3] == "Success";
      if (!ok) {
        widget.global.updateNotification(
          message,
          NotificationType.failure,
          null,
        );
        setState(() {});
        return;
      }
      if (completed) {
        widget.global.updateNotification(
          "Measurement Completed",
          NotificationType.success,
          null,
        );
        return;
      }
      _results = [];
      String m = value.values[0];
      if (m.trim().isNotEmpty) {
        List<String> lines = m.split(";;;");
        for (String line in lines) {
          _results.add(line.split(","));
        }
      }
      widget.global.updateNotification(
        message,
        NotificationType.progress,
        null,
      );
      setState(() {});
    });
  }

  @override
  Widget build(BuildContext context) {
    final selectedColor = Theme.of(context).colorScheme.inversePrimary;
    return Column(
      children: [
        Expanded(
          child: Row(
            children: [
              _getDevices(selectedColor),
              _getInformation(),
              _getSpectra(),
              _getResults(),
            ],
          ),
        ),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            ElevatedButton.icon(
              style: AppTheme.primaryButtonStyle(context),
              onPressed: () async {
                ActionRequest action = ActionRequest();
                action.action = "GTxMeasurement";
                action.addParameter("Information", [
                  _selectedProfile,
                  _selectedComponent,
                  _if.text,
                  _cableLoss.text,
                ]);
                action.addParameter("Modulation", [
                  _selectedModulation,
                  _subCarFreq.text,
                  _frequencyDeviation.text,
                  _modIndex.text,
                ]);
                action.addParameter("PowerSpectrum", [
                  _powerSpan.text,
                  _powerRBW.text,
                  _powerVBW.text,
                ]);
                action.addParameter("FrequencySpectrum", [
                  _freqSpan.text,
                  _freqRBW.text,
                  _freqVBW.text,
                ]);
                action.addParameter("InBandSpectrum", [
                  _inBandSpan.text,
                  _inBandRBW.text,
                  _inBandVBW.text,
                ]);
                action.addParameter("OutBandSpectrum", [
                  _outBandSpan.text,
                  _outBandRBW.text,
                  _outBandVBW.text,
                ]);
                Acknowledgement ack = await initiateServerAction(
                  widget.global,
                  action,
                );
                if (!ack.ok) {
                  return;
                }
                _getStatusUpdate();
              },
              label: Text('Start Measurement'),
              icon: Icon(AppIcons.start, color: Colors.green),
            ),
          ],
        ),
      ],
    );
  }

  Widget _getDevices(Color selectedColor) {
    List<Widget> colChildren = [];
    List<Widget> children = [];
    colChildren.add(
      ListTile(
        title: Text(
          'Select Devices',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
    );
    colChildren.add(Divider());
    for (String d in _profiles) {
      children.add(
        ListTile(
          title: Text(d),
          selectedTileColor: selectedColor,
          selected: _selectedProfile == d,
          onTap: () {
            _selectedProfile = d;
            _getDeviceNames();
            setState(() {});
          },
        ),
      );
    }
    colChildren.add(Expanded(child: ListView(children: children)));
    colChildren.add(Divider());
    colChildren.add(Text('GTx: $_gtxName'));
    colChildren.add(Text('SA: $_saName'));
    return Flexible(
      flex: 10,
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Column(children: colChildren),
      ),
    );
  }

  Widget _getInformation() {
    List<Widget> children = [];
    children.add(
      ListTile(
        title: Text('Details', style: TextStyle(fontWeight: FontWeight.bold)),
      ),
    );
    children.add(Divider());
    children.add(
      DropdownButtonFormField(
        items: _components
            .map((e) => DropdownMenuItem(value: e, child: Text(e)))
            .toList(),
        value: _selectedComponent,
        onChanged: (value) {
          _selectedComponent = value ?? _components.first;
          setState(() {});
        },
        decoration: InputDecoration(labelText: 'Component'),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _cableLoss,
        decoration: InputDecoration(labelText: "Cable Loss"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _if,
        decoration: InputDecoration(labelText: "Intermediate Frequency"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      DropdownButtonFormField(
        items: List.generate(_modulations.length, (i) {
          return DropdownMenuItem(
            value: _modulations[i],
            child: Text(_modulations[i]),
          );
        }),
        value: _selectedModulation,
        onChanged: (value) {
          _selectedModulation = value ?? _modulations.first;
          setState(() {});
        },
        decoration: InputDecoration(labelText: "Modulation"),
      ),
    );
    children.add(SizedBox(height: 16));
    if ((_selectedModulation == "PM") || (_selectedModulation == "FM")) {
      children.add(
        TextFormField(
          controller: _subCarFreq,
          decoration: InputDecoration(labelText: "Sub carrier Frequency"),
        ),
      );
      children.add(SizedBox(height: 16));
    }
    if (_selectedModulation == "PM") {
      children.add(
        TextFormField(
          controller: _modIndex,
          decoration: InputDecoration(labelText: "Modulation Index"),
        ),
      );
    }
    if (_selectedModulation == "FM") {
      children.add(
        TextFormField(
          controller: _frequencyDeviation,
          decoration: InputDecoration(labelText: "Frequency Deviation"),
        ),
      );
    }
    return Flexible(
      flex: 10,
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: ListView(children: children),
      ),
    );
  }

  Widget _getSpectra() {
    List<Widget> children = [];
    children.add(
      ListTile(
        title: Text('Spectrum', style: TextStyle(fontWeight: FontWeight.bold)),
      ),
    );
    children.add(Divider());
    children.add(
      ListTile(
        title: Text(
          'Power Measurement',
          style: TextStyle(fontStyle: FontStyle.italic),
        ),
      ),
    );
    children.add(
      TextFormField(
        controller: _powerSpan,
        decoration: InputDecoration(labelText: "Span"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _powerRBW,
        decoration: InputDecoration(labelText: "RBW"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _powerVBW,
        decoration: InputDecoration(labelText: "VBW"),
      ),
    );
    children.add(
      ListTile(
        title: Text(
          'Frequency Measurement',
          style: TextStyle(fontStyle: FontStyle.italic),
        ),
      ),
    );
    children.add(
      TextFormField(
        controller: _freqSpan,
        decoration: InputDecoration(labelText: "Span"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _freqRBW,
        decoration: InputDecoration(labelText: "RBW"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _freqVBW,
        decoration: InputDecoration(labelText: "VBW"),
      ),
    );
    children.add(
      ListTile(
        title: Text(
          'In-Band Spurious Measurement',
          style: TextStyle(fontStyle: FontStyle.italic),
        ),
      ),
    );
    children.add(
      TextFormField(
        controller: _inBandSpan,
        decoration: InputDecoration(labelText: "Span"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _inBandRBW,
        decoration: InputDecoration(labelText: "RBW"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _inBandVBW,
        decoration: InputDecoration(labelText: "VBW"),
      ),
    );
    children.add(
      ListTile(
        title: Text(
          'Out of Band Spurious Measurement',
          style: TextStyle(fontStyle: FontStyle.italic),
        ),
      ),
    );
    children.add(
      TextFormField(
        controller: _outBandSpan,
        decoration: InputDecoration(labelText: "Span"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _outBandRBW,
        decoration: InputDecoration(labelText: "RBW"),
      ),
    );
    children.add(SizedBox(height: 16));
    children.add(
      TextFormField(
        controller: _outBandVBW,
        decoration: InputDecoration(labelText: "VBW"),
      ),
    );
    return Flexible(
      flex: 10,
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: ListView(children: children),
      ),
    );
  }

  Widget _getResults() {
    List<Widget> children = [];
    children.add(
      ListTile(
        title: Text('Results', style: TextStyle(fontWeight: FontWeight.bold)),
      ),
    );
    children.add(Divider());
    Table table = Table(
      children: _results.map((row) {
        return TableRow(children: row.map((cell) => Text(cell)).toList());
      }).toList(),
    );
    children.add(table);

    return Flexible(
      flex: 20,
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: ListView(children: children),
      ),
    );
  }
}
