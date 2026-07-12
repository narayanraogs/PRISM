import 'package:flutter/material.dart';

import 'package:prism_client/utils/notifications.dart';
import 'package:prism_client/services/notification_service.dart';
import 'package:prism_client/widgets/screen_header.dart';
import 'package:prism_client/widgets/content_card.dart';

class GenericScreen extends StatelessWidget {
  final String title;
  const GenericScreen({super.key, required this.title});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(32.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ScreenHeader(
            title: title,
            subtitle: 'This screen is currently under development',
            icon: Icons.grid_view_rounded,
          ),
          const SizedBox(height: 24),
          Expanded(
            child: ContentCard(
              isSidebar: false,
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.construction_outlined,
                      size: 64,
                      color: Theme.of(
                        context,
                      ).colorScheme.primary.withValues(alpha: 0.2),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'Content for $title is under development',
                      style: TextStyle(
                        color: Colors.grey.shade400,
                        fontSize: 18,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    const SizedBox(height: 32),
                    ElevatedButton.icon(
                      onPressed: () {
                        final types = [
                          NotificationType.info,
                          NotificationType.success,
                          NotificationType.warning,
                          NotificationType.error,
                        ];
                        final type = (types..shuffle()).first;
                        AppNotifications.show(
                          context,
                          'Test notification for $title at ${DateTime.now().hour}:${DateTime.now().minute}',
                          type: type,
                        );
                      },
                      icon: const Icon(Icons.notification_add),
                      label: const Text('Send Test Notification'),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
