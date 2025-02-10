// browse.h
#import <Cocoa/Cocoa.h>
#include <syslog.h>
#include <stdbool.h>

#ifndef MAIN_H
#define MAIN_H

extern void HandleURL(char* url, char* name, char* bundleId, char* path, int pid);

@interface BrowseAppDelegate: NSObject<NSApplicationDelegate>
    - (void)handleGetURLEvent:(NSAppleEventDescriptor *) event withReplyEvent:(NSAppleEventDescriptor *)replyEvent;
@end

void RunApp();

#endif /* MAIN_H */
