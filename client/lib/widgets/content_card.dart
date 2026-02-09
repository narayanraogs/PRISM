import 'package:flutter/material.dart';

class ContentCard extends StatelessWidget {
  final Widget child;
  final double? width;
  final double? height;
  final EdgeInsetsGeometry padding;
  final double borderRadius;
  final bool isSidebar; // If true, uses slightly different shadow/radius
  final Color? color;

  const ContentCard({
    super.key,
    required this.child,
    this.width,
    this.height,
    this.padding = const EdgeInsets.all(24),
    this.borderRadius = 32,
    this.isSidebar = false,
    this.color,
    this.margin,
  });

  final EdgeInsetsGeometry? margin;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    // Sidebar style: Radius 24, lighter shadow, higher opacity
    // Main Content style: Radius 32, larger shadow, lower opacity

    final double br = isSidebar ? 24.0 : borderRadius;
    final double blur = isSidebar ? 20.0 : 30.0;
    final double offsetY = isSidebar ? 10.0 : 15.0;
    final double opacity = isSidebar ? 0.05 : 0.03;

    return Container(
      width: width,
      height: height,
      margin: margin,
      padding: padding,
      decoration: BoxDecoration(
        color: color ?? Colors.white,
        borderRadius: BorderRadius.circular(br),
        boxShadow: [
          BoxShadow(
            color: theme.colorScheme.primary.withOpacity(opacity),
            blurRadius: blur,
            offset: Offset(0, offsetY),
          ),
        ],
      ),
      child: child,
    );
  }
}
