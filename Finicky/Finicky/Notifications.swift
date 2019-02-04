import Foundation
import AppKit

func showNotification(title: String, subtitle: String? = nil, informativeText: String? = nil) -> Void {
    let notification = NSUserNotification()

    notification.title = title
    if subtitle != nil {
        notification.subtitle = subtitle!
    }

    if informativeText != nil {
        notification.informativeText = informativeText!
    }

    notification.soundName = NSUserNotificationDefaultSoundName
    NSUserNotificationCenter.default.deliver(notification)    
}
