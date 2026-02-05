import 'package:flutter/material.dart';
import 'package:prism/server_communication/classes.dart';
import 'package:prism/server_communication/messages.dart';
import 'package:prism/structures.dart';

class TestPhase extends StatefulWidget {
  final Global global;

  const TestPhase(this.global, {super.key});

  @override
  State<TestPhase> createState() => _StateTestPhase();
}

class _StateTestPhase extends State<TestPhase> {
  final List<String> _modes = ['Add New', 'Select Existing'];
  String _selectedMode = '';
  final TextEditingController _newPhase = TextEditingController();
  bool _copyLossesFrom = false;
  List<String> _testPhases = [];
  String _selectedTestPhase = '';

  @override
  void initState() {
    super.initState();
    _selectedMode = _modes.first;
    _getTestPhases();
  }

  Future<void> _getTestPhases() async {
    GetResponse resp = await getParameters(widget.global, ["TestPhases"]);
    if (!resp.ok) {
      widget.global.updateNotification(
        "Unable to get Test Phases from Server",
        NotificationType.failure,
        null,
      );
      return;
    }
    List<String> values = resp.getValue("TestPhases");
    _testPhases = [];
    _testPhases.addAll(values);
    _selectedTestPhase = _testPhases.isEmpty ? '' : _testPhases.first;
    setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final selectedColor = Theme.of(context).colorScheme.inversePrimary;
    Widget secondChild = _getAddNewCard(selectedColor);
    if (_selectedMode.contains("Select")) {
      secondChild = _getSelectCard(selectedColor);
    }
    return Row(
      spacing: 8,
      children: [
        _getTypeSelection(selectedColor),
        secondChild,
        _getEmptySpace(selectedColor),
      ],
    );
  }

  Widget _getTypeSelection(Color selectedColor) {
    List<Widget> configs = [];
    for (String cfg in _modes) {
      configs.add(
        ListTile(
          title: Text(cfg),
          selected: _selectedMode == cfg,
          onTap: () {
            _selectedMode = cfg;
            setState(() {});
          },
          selectedTileColor: selectedColor,
        ),
      );
    }
    List<Widget> children = [];
    children.add(
      Text('Add/Select', style: TextStyle(fontWeight: FontWeight.bold)),
    );
    children.add(Divider());
    children.add(Expanded(child: ListView(children: configs)));
    children.add(Divider());
    return Flexible(
      flex: 10,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Column(children: children),
        ),
      ),
    );
  }

  Widget _getAddNewCard(Color selectedColor) {
    List<Widget> configs = [];

    for (String cfg in _testPhases) {
      configs.add(
        ListTile(
          title: Text(cfg),
          selected: _selectedTestPhase == cfg,
          onTap: () {
            _selectedTestPhase = cfg;
            setState(() {});
          },
          selectedTileColor: selectedColor,
        ),
      );
    }
    List<Widget> children = [];
    children.add(
      Text('Add New', style: TextStyle(fontWeight: FontWeight.bold)),
    );
    children.add(Divider());
    children.add(
      TextFormField(
        controller: _newPhase,
        decoration: InputDecoration(labelText: 'New Phase Name'),
      ),
    );
    children.add(
      CheckboxListTile(
        value: _copyLossesFrom,
        onChanged: (value) {
          _copyLossesFrom = value ?? false;
          setState(() {});
        },
        title: Text("Copy Losses"),
        controlAffinity: ListTileControlAffinity.leading,
      ),
    );
    children.add(Expanded(child: ListView(children: configs)));
    children.add(Divider());
    children.add(
      ElevatedButton.icon(
        onPressed: () async {
          if (_selectedMode == 'Select Existing') {
            ActionRequest action = ActionRequest();
            action.action = "SelectExistingTestPhase";
            action.addParameter("TestPhase", [_selectedTestPhase]);
            Acknowledgement ack = await initiateServerAction(
              widget.global,
              action,
            );
            if (!ack.ok) {
              return;
            }
            _getTestPhases();
            widget.global.updateNotification(
              "Test Phase updated",
              NotificationType.success,
              null,
            );
          } else {
            if (_newPhase.text.trim().isEmpty) {
              widget.global.updateNotification(
                "New test phase name cannot be empty",
                NotificationType.failure,
                null,
              );
              return;
            }
            ActionRequest action = ActionRequest();
            action.action = "AddNewTestPhase";
            List<String> testPhaseParams = [_newPhase.text];
            if (_copyLossesFrom) {
              testPhaseParams.add(_selectedTestPhase);
            } else {
              testPhaseParams.add('');
            }
            action.addParameter("TestPhase", testPhaseParams);
            Acknowledgement ack = await initiateServerAction(
              widget.global,
              action,
            );
            if (!ack.ok) {
              return;
            }
            _getTestPhases();
            widget.global.updateNotification(
              "Test Phase updated",
              NotificationType.success,
              null,
            );
          }
        },
        label: Text('Submit'),
        icon: Icon(Icons.play_circle),
      ),
    );
    return Flexible(
      flex: 10,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Column(children: children),
        ),
      ),
    );
  }

  Widget _getSelectCard(Color selectedColor) {
    List<Widget> configs = [];

    for (String cfg in _testPhases) {
      configs.add(
        ListTile(
          title: Text(cfg),
          selected: _selectedTestPhase == cfg,
          onTap: () {
            _selectedTestPhase = cfg;
            setState(() {});
          },
          selectedTileColor: selectedColor,
        ),
      );
    }
    List<Widget> children = [];
    children.add(
      Text('Select Existing', style: TextStyle(fontWeight: FontWeight.bold)),
    );
    children.add(Divider());
    children.add(Expanded(child: ListView(children: configs)));
    children.add(Divider());
    children.add(
      ElevatedButton.icon(
        onPressed: () {},
        label: Text('Submit'),
        icon: Icon(Icons.play_circle),
      ),
    );
    return Flexible(
      flex: 10,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Column(children: children),
        ),
      ),
    );
  }

  Widget _getEmptySpace(Color selectedColor) {
    return Flexible(flex: 50, child: Column(children: []));
  }
}
