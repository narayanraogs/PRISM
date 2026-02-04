import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:prism_client/utils/notifications.dart';
import 'package:prism_client/services/notification_service.dart';

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
          Text(
            title,
            style: GoogleFonts.outfit(
              fontSize: 28,
              fontWeight: FontWeight.w900,
              color: Colors.black,
            ),
          ),
          const SizedBox(height: 24),
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.02),
                    blurRadius: 20,
                    offset: const Offset(0, 10),
                  ),
                ],
              ),
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.construction_outlined,
                      size: 64,
                      color: Theme.of(
                        context,
                      ).colorScheme.primary.withOpacity(0.2),
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
