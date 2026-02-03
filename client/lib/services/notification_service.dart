import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';

enum NotificationType { info, success, warning, error }

class NotificationItem {
  final String id;
  final String title;
  final String message;
  final DateTime timestamp;
  final NotificationType type;
  bool isRead;

  NotificationItem({
    required this.id,
    required this.title,
    required this.message,
    required this.timestamp,
    this.type = NotificationType.info,
    this.isRead = false,
  });
}

class NotificationService extends ChangeNotifier {
  final List<NotificationItem> _notifications = [];
  final _uuid = const Uuid();

  List<NotificationItem> get notifications => List.unmodifiable(_notifications);
  int get unreadCount => _notifications.where((n) => !n.isRead).length;

  void addNotification({
    required String message,
    String? title,
    NotificationType type = NotificationType.info,
  }) {
    final notification = NotificationItem(
      id: _uuid.v4(),
      title: title ?? _getDefaultTitle(type),
      message: message,
      timestamp: DateTime.now(),
      type: type,
    );
    _notifications.insert(0, notification);
    notifyListeners();
  }

  void markAsRead(String id) {
    final index = _notifications.indexWhere((n) => n.id == id);
    if (index != -1) {
      _notifications[index].isRead = true;
      notifyListeners();
    }
  }

  void markAllAsRead() {
    for (var n in _notifications) {
      n.isRead = true;
    }
    notifyListeners();
  }

  void removeNotification(String id) {
    _notifications.removeWhere((n) => n.id == id);
    notifyListeners();
  }

  void clearAll() {
    _notifications.clear();
    notifyListeners();
  }

  String _getDefaultTitle(NotificationType type) {
    switch (type) {
      case NotificationType.success:
        return 'Success';
      case NotificationType.error:
        return 'Error';
      case NotificationType.warning:
        return 'Warning';
      case NotificationType.info:
        return 'System Notification';
    }
  }
}
