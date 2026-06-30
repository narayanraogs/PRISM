import 'package:flutter/material.dart';

class StatusBar extends StatelessWidget {
  final bool isConnected;
  final VoidCallback onReconnect;
  final VoidCallback onNotificationsTap;
  final VoidCallback onSearchTap;
  final String cpuUsage;
  final String memoryUsage;
  final String satelliteName;
  final String testPhase;
  
  final int notificationCount;
  final bool isOperationRunning;
  final VoidCallback? onMonitorTap;

  const StatusBar({
    super.key,
    required this.isConnected,
    required this.onReconnect,
    required this.onNotificationsTap,
    required this.onSearchTap,
    required this.satelliteName,
    required this.testPhase,
    this.cpuUsage = '12%',
    this.memoryUsage = '256 MB',
    this.notificationCount = 0,
    this.isOperationRunning = false,
    this.onMonitorTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final statusColor = isConnected ? Colors.green : Colors.red;

    return Container(
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        border: Border(top: BorderSide(color: Colors.grey.shade200, width: 1)),
      ),
      child: Row(
        children: [
          // Left Side: Satellite and Phase
          Row(
            children: [
              Text(
                satelliteName,
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: theme.colorScheme.primary,
                ),
              ),
              const SizedBox(width: 12),
              const VerticalDivider(width: 1, indent: 8, endIndent: 8),
              const SizedBox(width: 12),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: theme.colorScheme.primary.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  testPhase,
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: theme.colorScheme.primary,
                  ),
                ),
              ),
              if (isOperationRunning) ...[
                const SizedBox(width: 12),
                const VerticalDivider(width: 1, indent: 8, endIndent: 8),
                const SizedBox(width: 12),
                Material(
                  color: Colors.transparent,
                  child: InkWell(
                    onTap: onMonitorTap,
                    borderRadius: BorderRadius.circular(4),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          colors: [Colors.orange.shade600, Colors.red.shade600],
                        ),
                        borderRadius: BorderRadius.circular(4),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.orange.withOpacity(0.3),
                            blurRadius: 4,
                            spreadRadius: 1,
                          ),
                        ],
                      ),
                      child: const Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.radio_button_checked,
                            color: Colors.white,
                            size: 10,
                          ),
                          SizedBox(width: 6),
                          Text(
                            'LIVE REMOTE SESSION ACTIVE',
                            style: TextStyle(
                              fontSize: 9,
                              fontWeight: FontWeight.w900,
                              color: Colors.white,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ],
          ),

          const Spacer(),

          // Right Side: System Stats, Connection, Search, Notification
          Row(
            children: [
              _buildStatItem(Icons.memory, memoryUsage),
              const SizedBox(width: 16),
              _buildStatItem(Icons.developer_board, cpuUsage),
              const SizedBox(width: 24),
              const VerticalDivider(width: 1, indent: 8, endIndent: 8),
              const SizedBox(width: 24),

              // Search Button
              IconButton(
                onPressed: onSearchTap,
                icon: const Icon(Icons.search, size: 16),
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
                visualDensity: VisualDensity.compact,
                color: Colors.grey.shade600,
                tooltip: 'Search (Cmd+K)',
              ),
              const SizedBox(width: 16),
              const VerticalDivider(width: 1, indent: 8, endIndent: 8),
              const SizedBox(width: 24),

              // Server Status
              InkWell(
                onTap: onReconnect,
                child: Row(
                  children: [
                    Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        color: statusColor,
                        shape: BoxShape.circle,
                        boxShadow: [
                          BoxShadow(
                            color: statusColor.withValues(alpha: 0.4),
                            blurRadius: 4,
                            spreadRadius: 1,
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      isConnected ? 'CONNECTED' : 'DISCONNECTED',
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        color: Colors.grey.shade600,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              const VerticalDivider(width: 1, indent: 8, endIndent: 8),
              const SizedBox(width: 8),

              // Notifications
              Stack(
                clipBehavior: Clip.none,
                children: [
                  IconButton(
                    onPressed: onNotificationsTap,
                    icon: Icon(
                      notificationCount > 0
                          ? Icons.notifications_active
                          : Icons.notifications_none,
                      size: 16,
                    ),
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                    visualDensity: VisualDensity.compact,
                    color: notificationCount > 0
                        ? theme.colorScheme.primary
                        : Colors.grey.shade600,
                    tooltip: 'Notifications',
                  ),
                  if (notificationCount > 0)
                    Positioned(
                      right: -2,
                      top: -2,
                      child: Container(
                        padding: const EdgeInsets.all(2),
                        decoration: BoxDecoration(
                          color: Colors.red,
                          borderRadius: BorderRadius.circular(6),
                        ),
                        constraints: const BoxConstraints(
                          minWidth: 10,
                          minHeight: 10,
                        ),
                        child: Text(
                          notificationCount > 9 ? '9+' : '$notificationCount',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 7,
                            fontWeight: FontWeight.bold,
                          ),
                          textAlign: TextAlign.center,
                        ),
                      ),
                    ),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem(IconData icon, String value) {
    return Row(
      children: [
        Icon(icon, size: 12, color: Colors.grey.shade400),
        const SizedBox(width: 4),
        Text(
          value,
          style: TextStyle(
            fontSize: 10,
            color: Colors.grey.shade600,
            fontFamily: 'monospace',
          ),
        ),
      ],
    );
  }
}
