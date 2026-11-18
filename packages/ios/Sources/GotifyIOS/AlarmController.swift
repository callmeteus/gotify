import Foundation
import UserNotifications

/// Full-screen alarm actions. Critical-alert sound needs an Apple entitlement;
/// without it the notification falls back to a high-priority alert.
enum AlarmController {
    static let categoryIdentifier = "GOTIFY_ALARM"
    static let actionUp = "GOTIFY_ALARM_UP"
    static let actionDesk = "GOTIFY_ALARM_DESK"

    static func registerCategories() {
        let up = UNNotificationAction(
            identifier: actionUp,
            title: NSLocalizedString("I'm up", comment: ""),
            options: [.foreground]
        )
        let desk = UNNotificationAction(
            identifier: actionDesk,
            title: NSLocalizedString("I'm at my computer", comment: ""),
            options: [.foreground]
        )
        let category = UNNotificationCategory(
            identifier: categoryIdentifier,
            actions: [up, desk],
            intentIdentifiers: [],
            options: []
        )
        UNUserNotificationCenter.current().setNotificationCategories([category])
    }
}
