// browse.h
#import <Cocoa/Cocoa.h>
#include <syslog.h>
#include <stdbool.h>

#ifndef MAIN_H
#define MAIN_H

extern void HandleURL(char* url, char* name, char* bundleId, char* path, int pid);

typedef struct {
    bool shift;
    bool option;
    bool command;
    bool control;
    bool capsLock;
    bool fn;
} ModifierKeys;

@interface BrowseAppDelegate: NSObject<NSApplicationDelegate>
    - (void)handleGetURLEvent:(NSAppleEventDescriptor *) event withReplyEvent:(NSAppleEventDescriptor *)replyEvent;    
@end

ModifierKeys getModifierKeys(void);
void RunApp();

typedef struct {
    const char* localizedName;
    const char* name;
} SystemInfo;

SystemInfo getSystemInfo(void);

#endif /* MAIN_H */
