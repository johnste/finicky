#include "main.h"
#include "util/info.h"
#import <Cocoa/Cocoa.h>
#import <stdlib.h>

// https://example.com

@implementation BrowseAppDelegate
- (void)applicationWillFinishLaunching:(NSNotification *)aNotification
{
    NSAppleEventManager *appleEventManager = [NSAppleEventManager sharedAppleEventManager];
    [appleEventManager setEventHandler:self
                       andSelector:@selector(handleGetURLEvent:withReplyEvent:)
                       forEventClass:kInternetEventClass andEventID:kAEGetURL];
}

- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
           withReplyEvent:(NSAppleEventDescriptor *)replyEvent {

    int32_t pid = [[event attributeDescriptorForKeyword:keySenderPIDAttr] int32Value];

    NSRunningApplication *application = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
    const char *url = [[[event paramDescriptorForKeyword:keyDirectObject] stringValue] UTF8String];
    const char *name = NULL;
    const char *bundleId = NULL;
    const char *path = NULL;

    if (application) {
        NSString *appName = [application localizedName];
        NSString *appBundleID = [application bundleIdentifier];
        NSString *appPath = [[application bundleURL] path];

        name = [appName UTF8String];
        bundleId = [appBundleID UTF8String];
        path = [appPath UTF8String];
    } else {
        NSLog(@"No running application found with PID: %d", pid);
    }

    HandleURL((char*)url, (char*)name, (char*)bundleId, (char*)path, pid);
}

@end

void RunApp() {
    [NSAutoreleasePool new];
    [NSApplication sharedApplication];
    BrowseAppDelegate *app = [BrowseAppDelegate alloc];
    [NSApp setDelegate:app];
    [NSApp run];
}

