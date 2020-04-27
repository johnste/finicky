import AppKit
import Foundation

/*
 Show macOs notifications
 */
func showNotification(at center: NSUserNotificationCenter = .default, title: String, subtitle: String? = nil, informativeText: String? = nil, error: Bool = false) {
    let notification = NSUserNotification()
    notification.title = title
    notification.subtitle = subtitle
    notification.informativeText = informativeText
    notification.soundName = error ? NSUserNotificationDefaultSoundName : nil

    center.deliver(notification)
}
