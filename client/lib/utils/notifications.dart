import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:prism_client/services/notification_service.dart';

class AppNotifications {
  static void show(
    BuildContext context,
    String message, {
    String? title,
    NotificationType type = NotificationType.info,
    bool showSnackBar = true,
  }) {
    // Add to notification center
    final notificationService = Provider.of<NotificationService>(
      context,
      listen: false,
    );
    notificationService.addNotification(
      message: message,
      title: title,
      type: type,
    );

    // Optionally show snackbar for immediate feedback
    if (showSnackBar) {
      final color = _getNotificationColor(type);

      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Row(
            children: [
              Icon(_getNotificationIcon(type), color: Colors.white, size: 20),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  message,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
          backgroundColor: color,
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          margin: const EdgeInsets.all(16),
          duration: const Duration(seconds: 4),
          action: SnackBarAction(
            label: 'VIEW',
            textColor: Colors.white,
            onPressed: () {
              // We can't easily trigger the notification panel from here
              // without a global key or state management, but adding to center is enough for now.
            },
          ),
        ),
      );
    }
  }

  static void showSuccess(
    BuildContext context,
    String message, {
    String? title,
  }) {
    show(context, message, title: title, type: NotificationType.success);
  }

  static void showError(BuildContext context, String message, {String? title}) {
    show(context, message, title: title, type: NotificationType.error);
  }

  static void showWarning(
    BuildContext context,
    String message, {
    String? title,
  }) {
    show(context, message, title: title, type: NotificationType.warning);
  }

  static void showInfo(BuildContext context, String message, {String? title}) {
    show(context, message, title: title, type: NotificationType.info);
  }

  static Color _getNotificationColor(NotificationType type) {
    switch (type) {
      case NotificationType.success:
        return Colors.green.shade600;
      case NotificationType.error:
        return Colors.red.shade600;
      case NotificationType.warning:
        return Colors.orange.shade600;
      case NotificationType.info:
        return Colors.blue.shade600;
    }
  }

  static IconData _getNotificationIcon(NotificationType type) {
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
