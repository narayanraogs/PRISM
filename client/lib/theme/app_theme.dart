import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AppTheme {
  // Light Blue Color Palette
  static const Color primaryColor = Color(0xFF0D47A1);
  static const Color primaryLight = Color(0xFF5472D3);
  static const Color primaryDark = Color(0xFF002171);
  static const Color secondaryColor = Color(0xFFE3F2FD);
  static const Color accentColor = Color(0xFF2196F3);
  static const Color backgroundColor = Color(0xFFF8FAFC);
  static const Color surfaceColor = Colors.white;
  static const Color errorColor = Color(0xFFD32F2F);

  static ThemeData get lightTheme => _buildTheme(primaryColor);

  static ThemeData getDynamicTheme(String satelliteName) {
    if (satelliteName.isEmpty) return _buildTheme(primaryColor);

    // Curated professional palette
    final List<Color> palette = [
      const Color(0xFF0D47A1), // Deep Blue (Default)
      const Color(0xFF00695C), // Emerald Green
      const Color(0xFF4527A0), // Royal Purple
      const Color(0xFFC62828), // Crimson Red
      const Color(0xFFE65100), // Burnt Orange
      const Color(0xFF37474F), // Slate Grey
      const Color(0xFF2E7D32), // Forest Green
    ];

    // Dart's hashCode provides excellent distribution even for 1-char differences
    final int index = satelliteName.hashCode.abs() % palette.length;
    return _buildTheme(palette[index]);
  }

  static ThemeData _buildTheme(Color themeColor) {
    return ThemeData(
      useMaterial3: true,
      colorScheme: ColorScheme.fromSeed(
        seedColor: themeColor,
        primary: themeColor,
        secondary: accentColor,
        surface: surfaceColor,
        /* background: backgroundColor, */
        error: errorColor,
        onPrimary: Colors.white,
        onSecondary: Colors.white,
        onSurface: const Color(0xFF1A1C1E),
        /* onBackground: const Color(0xFF1A1C1E), */
      ),
      scaffoldBackgroundColor: backgroundColor,
      textTheme: GoogleFonts.interTextTheme().copyWith(
        displayLarge: GoogleFonts.outfit(
          fontWeight: FontWeight.bold,
          color: themeColor,
        ),
        displayMedium: GoogleFonts.outfit(
          fontWeight: FontWeight.bold,
          color: themeColor,
        ),
        headlineMedium: GoogleFonts.outfit(
          fontWeight: FontWeight.w600,
          color: themeColor,
        ),
        titleLarge: GoogleFonts.outfit(
          fontWeight: FontWeight.w600,
          color: themeColor,
        ),
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: surfaceColor,
        foregroundColor: themeColor,
        elevation: 0,
        centerTitle: false,
      ),
      drawerTheme: const DrawerThemeData(
        backgroundColor: surfaceColor,
        elevation: 0,
      ),
      navigationRailTheme: NavigationRailThemeData(
        backgroundColor: surfaceColor,
        selectedIconTheme: IconThemeData(color: themeColor),
        unselectedIconTheme: IconThemeData(color: Colors.grey.shade400),
        selectedLabelTextStyle: GoogleFonts.inter(
          color: themeColor,
          fontWeight: FontWeight.bold,
        ),
        unselectedLabelTextStyle: GoogleFonts.inter(
          color: Colors.grey.shade400,
        ),
      ),
      cardTheme: CardThemeData(
        color: surfaceColor,
        elevation: 2,
        shadowColor: themeColor.withValues(alpha: 0.1),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: themeColor,
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        ),
      ),
    );
  }
}
