// browse.h
#ifndef MAIN_H
#define MAIN_H

#ifdef __OBJC__
#import <Cocoa/Cocoa.h>
#endif

#include <syslog.h>
#include <stdbool.h>

extern void HandleURL(char *url, char *name, char *bundleId, char *path, bool openInBackground);
extern void QueueWindowDisplay(int launchedByUser);
extern void ShowConfigWindow();

#ifdef __OBJC__
@interface BrowseAppDelegate: NSObject<NSApplicationDelegate>
    @property (nonatomic) bool forceOpenWindow;
    @property (nonatomic) bool receivedURL;
    @property (nonatomic) bool keepRunning;
    @property (nonatomic) bool showMenuItem;
    - (instancetype)initWithForceOpenWindow:(bool)forceOpenWindow initShow:(bool)showMenuItem keepRunning:(bool)keepRunning;
    - (void)handleGetURLEvent:(NSAppleEventDescriptor *) event withReplyEvent:(NSAppleEventDescriptor *)replyEvent;
    - (bool)application:(NSApplication *)sender openFile:(NSString *)filename;
@end
#endif

void RunApp(bool forceOpenWindow, bool showStatusItem, bool keepRunning);

#endif /* MAIN_H */
