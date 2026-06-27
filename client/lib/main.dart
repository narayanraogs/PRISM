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
import 'package:prism_client/screens/tsm_internal_path_loss_screen.dart';
import 'package:prism_client/screens/gtx_characterization_screen.dart';
import 'package:prism_client/screens/link_loss_screen.dart';
import 'package:prism_client/screens/up_down_converter_measurement_screen.dart';
import 'package:prism_client/screens/view_reports_screen.dart';
import 'package:prism_client/screens/stability_reports_screen.dart';
import 'package:prism_client/screens/splash_screen.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:prism_client/screens/scpi_commander_screen.dart';
import 'package:prism_client/screens/insights_screen.dart';
import 'package:prism_client/screens/path_loss_planner_screen.dart';

void main() {
  GoogleFonts.config.allowRuntimeFetching = false;
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
    return Selector<ServerService, String>(
      selector: (context, server) => server.status.satelliteName,
      builder: (context, satelliteName, child) {
        final appTitle =
            satelliteName.isNotEmpty ? '$satelliteName - PRISM' : 'PRISM';
        return MaterialApp(
          debugShowCheckedModeBanner: false,
          title: appTitle,
          theme: AppTheme.getDynamicTheme(satelliteName),
          home: const RootPage(),
        );
      },
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
  bool _isInitialized = false;

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
    'Path Loss Planner',
    'Attenuation',
    'TSM Internal Path Loss',
    'GTx Characterization',
    'Up Down converter',
    'Database Management',
    'View Reports',
    'Stability reports',
    'Insights',
    'PPT Generation',
  ];

  @override
  Widget build(BuildContext context) {
    if (!_isInitialized) {
      return SplashScreen(
        onLoaded: () {
          setState(() {
            _isInitialized = true;
          });
        },
      );
    }
    return MainLayout(
      selectedIndex: _selectedIndex,
      onDestinationSelected: (index) {
        setState(() {
          _selectedIndex = index;
        });
      },
      child: IndexedStack(
        index: _selectedIndex,
        children: List.generate(_titles.length, (index) {
          final title = _titles[index];
          return LazyLoadWidget(
            isActivated: _selectedIndex == index,
            child: Builder(
              builder: (context) {
                if (index == 0) return RFUplinkScreen(isActive: _selectedIndex == 0);
                if (index == 1) return TestScreen(isActive: _selectedIndex == 1);
                if (index == 2) return const ScheduleScreen();
                if (index == 3) return StabilityScreen(isActive: _selectedIndex == 3);
                if (index == 4)
                  return SpectrumDumpScreen(isActive: _selectedIndex == 4);
                if (index == 5) return MonitorScreen(isActive: _selectedIndex == 5);
                if (index == 6)
                  return TVACCableLossScreen(isActive: _selectedIndex == 6);
                if (index == 7)
                  return ScpiCommanderScreen(isActive: _selectedIndex == 7);
                if (index == 8) return CableLossScreen(isActive: _selectedIndex == 8);
                if (index == 9)
                  return PathLossPlannerScreen(isActive: _selectedIndex == 9);
                if (index == 10)
                  return AttenuationScreen(isActive: _selectedIndex == 10);
                if (index == 11)
                  return TSMInternalPathLossScreen(isActive: _selectedIndex == 11);
                if (index == 12)
                  return GTxCharacterizationScreen(isActive: _selectedIndex == 12);
                if (index == 13)
                  return UpDownConverterScreen(isActive: _selectedIndex == 13);
                if (index == 14) return const LinkLossScreen();
                if (index == 15)
                  return ViewReportsScreen(isActive: _selectedIndex == 15);
                if (index == 16) return const StabilityReportsScreen();
                if (index == 17) return const InsightsScreen();

                return GenericScreen(title: title);
              },
            ),
          );
        }),
      ),
    );
  }
}

class LazyLoadWidget extends StatefulWidget {
  final Widget child;
  final bool isActivated;

  const LazyLoadWidget({
    super.key,
    required this.child,
    required this.isActivated,
  });

  @override
  State<LazyLoadWidget> createState() => _LazyLoadWidgetState();
}

class _LazyLoadWidgetState extends State<LazyLoadWidget> {
  bool _initialized = false;

  @override
  Widget build(BuildContext context) {
    if (widget.isActivated) {
      _initialized = true;
    }
    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 300),
      child: _initialized 
          ? KeyedSubtree(
              key: const ValueKey('loaded'),
              child: widget.child,
            )
          : const SizedBox.shrink(key: ValueKey('empty')),
    );
  }
}

