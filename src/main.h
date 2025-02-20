// browse.h
#ifndef MAIN_H
#define MAIN_H

#ifdef __OBJC__
#import <Cocoa/Cocoa.h>
#endif

#include <syslog.h>
#include <stdbool.h>

extern void HandleURL(char *url, char *name, char *bundleId, char *path);
extern void QueueWindowDisplay(int launchedByUser);

#ifdef __OBJC__
@interface BrowseAppDelegate: NSObject<NSApplicationDelegate>
    @property (nonatomic) BOOL forceOpenWindow;
    - (instancetype)initWithForceOpenWindow:(BOOL)forceOpenWindow;
    - (void)handleGetURLEvent:(NSAppleEventDescriptor *) event withReplyEvent:(NSAppleEventDescriptor *)replyEvent;
@end
#endif

void RunApp(int forceOpenWindow);

#endif /* MAIN_H */
