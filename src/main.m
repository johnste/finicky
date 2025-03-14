#include "main.h"
#include "util/info.h"
#import <Cocoa/Cocoa.h>
#import <stdlib.h>

@implementation BrowseAppDelegate

- (instancetype)initWithForceOpenWindow:(BOOL)forceOpenWindow {
    self = [super init];
    if (self) {
        _forceOpenWindow = forceOpenWindow;
    }
    return self;
}

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    NSDictionary *dict = [notification userInfo];

    // Use either the forced value or check if launched by user
    BOOL openWindow = self.forceOpenWindow;
    if (!openWindow) {
        NSNumber *launchedByUser = [dict objectForKey:@"NSApplicationLaunchIsDefaultLaunchKey"];
        openWindow = [launchedByUser boolValue];
    }

    QueueWindowDisplay(openWindow);
}

- (void)applicationWillFinishLaunching:(NSNotification *)aNotification
{
    NSAppleEventManager *appleEventManager = [NSAppleEventManager sharedAppleEventManager];
    [appleEventManager setEventHandler:self
                    andSelector:@selector(handleGetURLEvent:withReplyEvent:)
                    forEventClass:kInternetEventClass andEventID:kAEGetURL];
}

- (BOOL)application:(NSApplication *)sender openFile:(NSString *)filename {
    NSLog(@"Opening file: %@", filename);

    // Convert the file path to a file:// URL
    NSURL *fileURL = [NSURL fileURLWithPath:filename];
    NSString *urlString = [fileURL absoluteString];

    // Handle the file URL the same way we handle other URLs
    HandleURL((char*)[urlString UTF8String], NULL, NULL, NULL);

    return YES;
}

- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
        withReplyEvent:(NSAppleEventDescriptor *)replyEvent {

    // Get the application that opened the URL, if available
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

    HandleURL((char*)url, (char*)name, (char*)bundleId, (char*)path);
}
@end

void RunApp(int forceOpenWindow) {
    @autoreleasepool {
        // Initialize on the main thread directly, not async
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];

        BrowseAppDelegate *app = [[BrowseAppDelegate alloc] initWithForceOpenWindow:forceOpenWindow];
        [NSApp setDelegate:app];

        [NSApp finishLaunching];
        [NSApp run];
    }
}

