import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/theme/app_theme.dart';
import 'package:prism_client/layouts/main_layout.dart';
import 'package:prism_client/screens/generic_screen.dart';
import 'package:prism_client/screens/rf_uplink_screen.dart';
import 'package:prism_client/screens/test_screen.dart';
import 'package:prism_client/screens/schedule_screen.dart';
import 'package:prism_client/screens/stability_screen.dart';
import 'package:prism_client/screens/spectrum_dump_screen.dart';
import 'package:prism_client/screens/monitor_screen.dart';
import 'package:prism_client/screens/tvac_cable_loss_screen.dart';
import 'package:prism_client/screens/cable_loss_screen.dart';
import 'package:prism_client/screens/attenuation_screen.dart';
import 'package:prism_client/screens/link_loss_screen.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';

void main() {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (context) => ServerService()),
        ChangeNotifierProvider(create: (context) => NotificationService()),
      ],
      child: const MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'PRISM',
      theme: AppTheme.lightTheme,
      home: const RootPage(),
    );
  }
}

class RootPage extends StatefulWidget {
  const RootPage({super.key});

  @override
  State<RootPage> createState() => _RootPageState();
}

class _RootPageState extends State<RootPage> {
  int _selectedIndex = 0;

  final List<String> _titles = [
    'RF Uplink',
    'Tests',
    'Schedule',
    'Stability',
    'Spectrum Dump',
    'Monitor',
    'TVAC Cable Calibration',
    'SCPI Commander',
    'Cable Loss Measurement',
    'Attenuation',
    'TSM Internal Path Loss',
    'GTx Charecterization',
    'Up Down converter',
    'Database Management',
    'View Reports',
    'Generate reports',
    'Stability reports',
    'Insights',
    'PPT Generation',
  ];

  @override
  Widget build(BuildContext context) {
    return MainLayout(
      selectedIndex: _selectedIndex,
      onDestinationSelected: (index) {
        setState(() {
          _selectedIndex = index;
        });
      },
      child: IndexedStack(
        index: _selectedIndex,
        children: _titles.asMap().entries.map((entry) {
          final index = entry.key;
          final title = entry.value;
          if (index == 0) return const RFUplinkScreen();
          if (index == 1) return const TestScreen();
          if (index == 2) return const ScheduleScreen();
          if (index == 3) return const StabilityScreen();
          if (index == 4) return const SpectrumDumpScreen();
          if (index == 5) return MonitorScreen(isActive: _selectedIndex == 5);
          if (index == 6) return const TVACCableLossScreen();
          if (index == 8) return const CableLossScreen();
          if (index == 9) return const AttenuationScreen();
          if (index == 13) return const LinkLossScreen();

          return GenericScreen(title: title);
        }).toList(),
      ),
    );
  }
}
