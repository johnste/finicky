//
//  Chrome.swift
//  Finicky
//
//  Created by John Sterling on 2019-10-29.
//  Copyright © 2019 John sterling. All rights reserved.
//

import Foundation
import AppKit
import Cocoa

public func StartChromeIncognito(app: BrowserOpts, url: URL, defaultOpenInBackground: Bool) {

    let apps = NSRunningApplication.runningApplications(withBundleIdentifier: app.bundleId)

    // If Chrome isn't running we open it from the shell to avoid opening a normal window as well as the incognito window we need
    if apps.isEmpty {
        var command = ["open", "-b", app.bundleId]
        if app.openInBackground ?? defaultOpenInBackground {
               command.append("-g")
           }

        command.append(contentsOf: ["--args", url.absoluteString, "--incognito"])

        print(command.joined(separator: " "))
        shell(command)

    } else {

        // If Chrome is running we do one of two things.
        // a. If there is a incognito window opened already, we append a new tab to it with the wanted url
        // b. If there is no incognito window we open one and set the url of the first tab to the wanted url

        // FIXME: Currently using string interpolation to add the url to the script, since I didn't get a prompt when trying to send as parameters. Should be able to fix this in time.
        let chromeScript = """
            set myLink to "\(url.absoluteString.replacingOccurrences(of: "\"", with: "\\\""))"

            tell application "Google Chrome"
                set incognitoWindows to (every window whose mode is "incognito")
                if (count of incognitoWindows) > 0 then
                    set incognitoWindow to first window whose mode is "incognito"
                    tell incognitoWindow to make new tab at after (get active tab) with properties {URL:myLink}
                else
                    set incognitoWindow to make new window with properties {mode:"incognito"}
                    tell incognitoWindow to set URL of active tab to myLink
                end if
            end tell
        """
        executeScript(chromeScript)
        return
    }



}

func executeScript(_ script: String) {
    var error: NSDictionary?
    if let scriptObject = NSAppleScript(source: script) {
        if let output: NSAppleEventDescriptor = scriptObject.executeAndReturnError(
                                                                           &error) {
            print(output.stringValue)
        } else if (error != nil) {
            showNotification(title: "error: \(error)")
        }
    }
}
