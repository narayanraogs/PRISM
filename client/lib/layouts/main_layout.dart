import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/widgets/status_bar.dart';
import 'package:prism_client/services/server_service.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:intl/intl.dart';

class NavigationItem {
  final String label;
  final IconData icon;
  final IconData selectedIcon;
  final String category;

  NavigationItem({
    required this.label,
    required this.icon,
    required this.selectedIcon,
    required this.category,
  });
}

class MainLayout extends StatefulWidget {
  final Widget child;
  final int selectedIndex;
  final Function(int) onDestinationSelected;

  const MainLayout({
    super.key,
    required this.child,
    required this.selectedIndex,
    required this.onDestinationSelected,
  });

  @override
  State<MainLayout> createState() => _MainLayoutState();
}

class _MainLayoutState extends State<MainLayout> {
  bool _showNotifications = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          Column(
            children: [
              Expanded(
                child: Row(
                  children: [
                    SidebarNavigation(
                      selectedIndex: widget.selectedIndex,
                      onDestinationSelected: widget.onDestinationSelected,
                    ),
                    Expanded(
                      child: Container(
                        color: Theme.of(context).colorScheme.background,
                        child: ClipRRect(
                          borderRadius: const BorderRadius.only(
                            topLeft: Radius.circular(24),
                            bottomLeft: Radius.circular(24),
                          ),
                          child: widget.child,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Consumer2<ServerService, NotificationService>(
                builder: (context, server, notifications, child) {
                  return StatusBar(
                    satelliteName: server.status.satelliteName,
                    testPhase: server.status.testPhase,
                    isConnected: server.status.isConnected,
                    cpuUsage: '${server.status.cpuUsed.toStringAsFixed(1)}%',
                    memoryUsage:
                        '${server.status.memoryUsed.toStringAsFixed(1)}%',
                    notificationCount: notifications.unreadCount,
                    onReconnect: () => server.manualReconnect(),
                    onNotificationsTap: () {
                      setState(() {
                        _showNotifications = !_showNotifications;
                      });
                      if (!_showNotifications) {
                        notifications.markAllAsRead();
                      }
                    },
                  );
                },
              ),
            ],
          ),
          if (_showNotifications) ...[
            GestureDetector(
              onTap: () {
                setState(() {
                  _showNotifications = false;
                });
                context.read<NotificationService>().markAllAsRead();
              },
              child: Container(
                color: Colors.transparent,
                width: double.infinity,
                height: double.infinity,
              ),
            ),
            Positioned(
              right: 16,
              bottom: 44, // Just above the status bar
              child: _buildNotificationCenter(context),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildNotificationCenter(BuildContext context) {
    final notificationService = Provider.of<NotificationService>(context);
    final notifications = notificationService.notifications;

    return Container(
      width: 380,
      height: 500,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.15),
            blurRadius: 30,
            offset: const Offset(0, 10),
          ),
        ],
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(20.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Text(
                      'Notifications',
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                        fontSize: 18,
                      ),
                    ),
                    if (notificationService.unreadCount > 0) ...[
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: Theme.of(context).colorScheme.primary,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          '${notificationService.unreadCount} new',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 10,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
                IconButton(
                  onPressed: () => setState(() => _showNotifications = false),
                  icon: const Icon(Icons.close, size: 20),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                  style: IconButton.styleFrom(
                    backgroundColor: Colors.grey.shade100,
                    padding: const EdgeInsets.all(4),
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: notifications.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.notifications_none_outlined,
                          size: 48,
                          color: Colors.grey.shade300,
                        ),
                        const SizedBox(height: 16),
                        Text(
                          'No notifications yet',
                          style: TextStyle(
                            color: Colors.grey.shade500,
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  )
                : ListView.separated(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    itemCount: notifications.length,
                    separatorBuilder: (context, index) =>
                        const Divider(height: 1, indent: 20, endIndent: 20),
                    itemBuilder: (context, index) {
                      final item = notifications[index];
                      return _buildNotificationItem(context, item);
                    },
                  ),
          ),
          const Divider(height: 1),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                TextButton(
                  onPressed: notifications.isEmpty
                      ? null
                      : () => notificationService.clearAll(),
                  child: const Text(
                    'Clear all',
                    style: TextStyle(fontSize: 13),
                  ),
                ),
                TextButton(
                  onPressed: notificationService.unreadCount == 0
                      ? null
                      : () => notificationService.markAllAsRead(),
                  child: const Text(
                    'Mark all as read',
                    style: TextStyle(fontSize: 13),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNotificationItem(BuildContext context, NotificationItem item) {
    final color = _getNotificationColor(item.type);
    final icon = _getNotificationIcon(item.type);

    return InkWell(
      onTap: () {
        Provider.of<NotificationService>(
          context,
          listen: false,
        ).markAsRead(item.id);
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        color: item.isRead ? Colors.transparent : color.withOpacity(0.03),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, color: color, size: 18),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        item.title,
                        style: TextStyle(
                          fontWeight: item.isRead
                              ? FontWeight.w600
                              : FontWeight.bold,
                          fontSize: 14,
                          color: Colors.grey.shade800,
                        ),
                      ),
                      Text(
                        DateFormat('HH:mm').format(item.timestamp),
                        style: TextStyle(
                          fontSize: 11,
                          color: Colors.grey.shade500,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    item.message,
                    style: TextStyle(
                      fontSize: 13,
                      color: Colors.grey.shade600,
                      height: 1.4,
                    ),
                  ),
                ],
              ),
            ),
            if (!item.isRead)
              Container(
                margin: const EdgeInsets.only(left: 12, top: 4),
                width: 8,
                height: 8,
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.primary,
                  shape: BoxShape.circle,
                ),
              ),
          ],
        ),
      ),
    );
  }

  Color _getNotificationColor(NotificationType type) {
    switch (type) {
      case NotificationType.success:
        return Colors.green;
      case NotificationType.error:
        return Colors.red;
      case NotificationType.warning:
        return Colors.orange;
      case NotificationType.info:
        return Colors.blue;
    }
  }

  IconData _getNotificationIcon(NotificationType type) {
    switch (type) {
      case NotificationType.success:
        return Icons.check_circle_outline;
      case NotificationType.error:
        return Icons.error_outline;
      case NotificationType.warning:
        return Icons.warning_amber_rounded;
      case NotificationType.info:
        return Icons.info_outline;
    }
  }
}

class SidebarNavigation extends StatefulWidget {
  final int selectedIndex;
  final ValueChanged<int> onDestinationSelected;

  const SidebarNavigation({
    super.key,
    required this.selectedIndex,
    required this.onDestinationSelected,
  });

  @override
  State<SidebarNavigation> createState() => _SidebarNavigationState();
}

class _SidebarNavigationState extends State<SidebarNavigation> {
  bool _isExpanded = false;

  final List<NavigationItem> _navItems = [
    // Category 1: Actions
    NavigationItem(
      label: 'RF Uplink',
      icon: Icons.settings_input_antenna_outlined,
      selectedIcon: Icons.settings_input_antenna,
      category: 'ACTIONS',
    ),
    NavigationItem(
      label: 'Tests',
      icon: Icons.assignment_outlined,
      selectedIcon: Icons.assignment,
      category: 'ACTIONS',
    ),
    NavigationItem(
      label: 'Schedule',
      icon: Icons.calendar_today_outlined,
      selectedIcon: Icons.calendar_today,
      category: 'ACTIONS',
    ),

    // Category 2: Utilities
    NavigationItem(
      label: 'Stability',
      icon: Icons.speed_outlined,
      selectedIcon: Icons.speed,
      category: 'UTILITIES',
    ),
    NavigationItem(
      label: 'Spectrum Dump',
      icon: Icons.analytics_outlined,
      selectedIcon: Icons.analytics,
      category: 'UTILITIES',
    ),
    NavigationItem(
      label: 'Monitor',
      icon: Icons.monitor_heart_outlined,
      selectedIcon: Icons.monitor_heart,
      category: 'UTILITIES',
    ),
    NavigationItem(
      label: 'TVAC Cable Calibration',
      icon: Icons.settings_ethernet,
      selectedIcon: Icons.settings_ethernet,
      category: 'UTILITIES',
    ),
    NavigationItem(
      label: 'SCPI Commander',
      icon: Icons.terminal_outlined,
      selectedIcon: Icons.terminal,
      category: 'UTILITIES',
    ),

    // Category 3: T&E
    NavigationItem(
      label: 'Cable Loss Measurement',
      icon: Icons.linear_scale,
      selectedIcon: Icons.linear_scale,
      category: 'T&E',
    ),
    NavigationItem(
      label: 'Path Loss Planner',
      icon: Icons.map_outlined,
      selectedIcon: Icons.map,
      category: 'T&E',
    ),
    NavigationItem(
      label: 'Attenuation',
      icon: Icons.import_export,
      selectedIcon: Icons.import_export,
      category: 'T&E',
    ),
    NavigationItem(
      label: 'TSM Internal Path Loss',
      icon: Icons.router_outlined,
      selectedIcon: Icons.router,
      category: 'T&E',
    ),
    NavigationItem(
      label: 'GTx Characterization',
      icon: Icons.radar,
      selectedIcon: Icons.radar,
      category: 'T&E',
    ),
    NavigationItem(
      label: 'Up Down converter',
      icon: Icons.compare_arrows,
      selectedIcon: Icons.compare_arrows,
      category: 'T&E',
    ),

    // Category 4: Database
    NavigationItem(
      label: 'Database Management',
      icon: Icons.storage_outlined,
      selectedIcon: Icons.storage,
      category: 'DATABASE',
    ),

    // Category 5: Results
    NavigationItem(
      label: 'View Reports',
      icon: Icons.insert_drive_file_outlined,
      selectedIcon: Icons.insert_drive_file,
      category: 'RESULTS',
    ),
    NavigationItem(
      label: 'Stability reports',
      icon: Icons.assessment_outlined,
      selectedIcon: Icons.assessment,
      category: 'RESULTS',
    ),
    NavigationItem(
      label: 'Insights',
      icon: Icons.lightbulb_outline,
      selectedIcon: Icons.lightbulb,
      category: 'RESULTS',
    ),
    NavigationItem(
      label: 'PPT Generation',
      icon: Icons.slideshow_outlined,
      selectedIcon: Icons.slideshow,
      category: 'RESULTS',
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isExpanded = true),
      onExit: (_) => setState(() => _isExpanded = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        width: _isExpanded ? 240 : 72,
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.05),
              blurRadius: 10,
              offset: const Offset(2, 0),
            ),
          ],
        ),
        child: Column(
          children: [
            const SizedBox(height: 20),
            // Logo/Brand area
            Container(
              height: 50,
              alignment: Alignment.center,
              child: _isExpanded
                  ? Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        ClipRRect(
                          borderRadius: BorderRadius.circular(8),
                          child: Image.asset(
                            'assets/images/logo.jpg',
                            height: 36,
                            width: 36,
                            fit: BoxFit.cover,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'PRISM',
                              style: Theme.of(context)
                                  .textTheme
                                  .headlineMedium
                                  ?.copyWith(
                                    fontSize: 22,
                                    letterSpacing: 2,
                                    fontWeight: FontWeight.bold,
                                    color: const Color(0xFF050A14),
                                    height: 1.0,
                                  ),
                            ),
                            Consumer<ServerService>(
                              builder: (context, server, child) {
                                final satName = server.status.satelliteName;
                                if (satName.isEmpty) return const SizedBox.shrink();
                                return Padding(
                                  padding: const EdgeInsets.only(top: 2),
                                  child: Text(
                                    satName.toUpperCase(),
                                    style: TextStyle(
                                      fontSize: 10,
                                      fontWeight: FontWeight.w800,
                                      color: Theme.of(context).colorScheme.primary,
                                      letterSpacing: 1.2,
                                    ),
                                  ),
                                );
                              },
                            ),
                          ],
                        ),
                      ],
                    )
                  : ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Image.asset(
                        'assets/images/logo.jpg',
                        height: 32,
                        width: 32,
                        fit: BoxFit.cover,
                      ),
                    ),
            ),
            const SizedBox(height: 20),
            // Scrollable Navigation Items
            Expanded(
              child: ListView.builder(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                itemCount: _navItems.length,
                itemBuilder: (context, index) {
                  final item = _navItems[index];
                  final showCategory = _isExpanded &&
                      (index == 0 ||
                          _navItems[index - 1].category != item.category);

                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (showCategory)
                        Padding(
                          padding: const EdgeInsets.only(
                            left: 12,
                            top: 16,
                            bottom: 8,
                          ),
                          child: Text(
                            item.category,
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: Colors.grey.shade500,
                              letterSpacing: 1.2,
                            ),
                          ),
                        ),
                      _buildNavItem(
                        index,
                        item.icon,
                        item.selectedIcon,
                        item.label,
                      ),
                    ],
                  );
                },
              ),
            ),
            const SizedBox(height: 12),
          ],
        ),
      ),
    );
  }

  Widget _buildNavItem(
    int index,
    IconData icon,
    IconData selectedIcon,
    String label,
  ) {
    final isSelected = widget.selectedIndex == index;
    final theme = Theme.of(context);

    return InkWell(
      onTap: () {
        if (index != -1) {
          widget.onDestinationSelected(index);
        }
      },
      borderRadius: BorderRadius.circular(12),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        height: 50,
        margin: const EdgeInsets.symmetric(vertical: 2),
        padding: EdgeInsets.symmetric(horizontal: _isExpanded ? 12 : 0),
        decoration: BoxDecoration(
          color: isSelected
              ? theme.colorScheme.primary.withOpacity(0.1)
              : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          mainAxisAlignment: _isExpanded
              ? MainAxisAlignment.start
              : MainAxisAlignment.center,
          children: [
            Icon(
              isSelected ? selectedIcon : icon,
              color: isSelected
                  ? theme.colorScheme.primary
                  : Colors.grey.shade600,
              size: 22,
            ),
            if (_isExpanded) ...[
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  label,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: isSelected
                        ? theme.colorScheme.primary
                        : Colors.grey.shade700,
                    fontWeight: isSelected
                        ? FontWeight.bold
                        : FontWeight.normal,
                    fontSize: 13,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

