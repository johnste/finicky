#include "main.h"
#include "util/info.h"
#import <Cocoa/Cocoa.h>
#import <stdlib.h>
#import <unistd.h>

@implementation BrowseAppDelegate

- (instancetype)initWithForceOpenWindow:(BOOL)forceOpenWindow {
    self = [super init];
    if (self) {
        _forceOpenWindow = forceOpenWindow;
        _receivedURL = false;
    }
    return self;
}

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    NSDictionary *dict = [notification userInfo];

    BOOL openInBackground = ![NSApp isActive];
    BOOL openWindow = self.forceOpenWindow;

    if (!openWindow) {
        // Even if we aren't forcing the window to open, we still want to open it if didn't receive a URL
        openWindow = !self.receivedURL;
    }

    NSLog(@"Madeleine openWindow: %d openInBackground: %d", openWindow, openInBackground);
    QueueWindowDisplay(openWindow, openInBackground);
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

    // If we recieve a url, we default to not showing the app in the dock
    [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    self.receivedURL = true;

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

- (BOOL)application:(NSApplication *)application willContinueUserActivityWithType:(NSString *)userActivityType {
    return [userActivityType isEqualToString:NSUserActivityTypeBrowsingWeb];
}

- (BOOL)application:(NSApplication *)application continueUserActivity:(NSUserActivity *)userActivity restorationHandler:(void (^)(NSArray<id<NSUserActivityRestoring>> * _Nullable))restorationHandler {
    if (![userActivity.activityType isEqualToString:NSUserActivityTypeBrowsingWeb]) {
        return NO;
    }

    NSURL *url = userActivity.webpageURL;
    if (!url) {
        return NO;
    }

    HandleURL((char*)[[url absoluteString] UTF8String], NULL, NULL, NULL);
    return YES;
}

- (void)application:(NSApplication *)application didFailToContinueUserActivityWithType:(NSString *)userActivityType error:(NSError *)error {
    // Handle failure if needed
}

@end

void RunApp(int forceOpenWindow) {
    @autoreleasepool {
        // Initialize on the main thread directly, not async
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

        BrowseAppDelegate *app = [[BrowseAppDelegate alloc] initWithForceOpenWindow:forceOpenWindow];
        [NSApp setDelegate:app];

        [NSApp finishLaunching];
        [NSApp run];
    }
}

