#include "main.h"
#include "util/info.h"
#import <Cocoa/Cocoa.h>
#import <ApplicationServices/ApplicationServices.h>
#import <stdlib.h>
#import <unistd.h>

#import "window/window.h"  // For ShowWindow()

// Extend BrowseAppDelegate to hold a status item and declare menu action
@interface BrowseAppDelegate ()
@property (nonatomic, strong) NSStatusItem *statusItem;
- (void)showWindowAction:(id)sender;
@end

@implementation BrowseAppDelegate

- (instancetype)initWithForceOpenWindow:(bool)forceOpenWindow initShow:(bool)showMenuItem keepRunning:(bool)keepRunning {
    self = [super init];
    if (self) {
        _forceOpenWindow = forceOpenWindow;
        _showMenuItem = showMenuItem;
        _keepRunning = keepRunning;
        _receivedURL = false;
    }
    return self;
}

// Use bool for openWindow and related logic
- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    bool openWindow = self.forceOpenWindow;
    if (!openWindow) {
        // Even if we aren't forcing the window to open, we still want to open it if didn't receive a URL
        openWindow = !self.receivedURL;
    }

    // Only show menu item if the option is enabled, and we either didn't receive a URL or we are keeping
    // the application running. We don't want to show the icon if Finicky is just receiving a url to open
    // and is expected to exit after
    if (self.showMenuItem && (self.keepRunning || !self.receivedURL)) {
        [self createStatusItem];
    }

    QueueWindowDisplay(openWindow);
}

- (BOOL)applicationShouldHandleReopen:(NSApplication *)sender hasVisibleWindows:(BOOL)flag {
    if (!flag) {
        // If there are no visible windows, we should open a new one
        [self showWindowAction:nil];
    }
    return YES;
}

- (void)createStatusItem {

    // Create menu bar status item
    self.statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSSquareStatusItemLength];
    // Load the template icon directly from the bundle
    NSString *iconPath = [[NSBundle mainBundle] pathForResource:@"menu-bar" ofType:@"icns"];
    NSImage *icon = [[NSImage alloc] initWithContentsOfFile:iconPath];
    if (icon) {
        icon.template = true;
        icon.size = NSMakeSize(18, 18);
        self.statusItem.button.image = icon;
    }

    self.statusItem.button.toolTip = @"Finicky";
    NSMenu *menu = [[NSMenu alloc] init];
    [menu addItemWithTitle:@"Show Window" action:@selector(showWindowAction:) keyEquivalent:@""];

    char* configPath = GetCurrentConfigPath();
    if (configPath) {
        [menu addItemWithTitle:@"Edit config" action:@selector(editConfigAction:) keyEquivalent:@""];
        [menu addItem:[NSMenuItem separatorItem]];
    }

    [menu addItemWithTitle:@"Quit" action:@selector(terminate:) keyEquivalent:@"q"];
    self.statusItem.menu = menu;
}

// Menu action to show the main window
- (void)showWindowAction:(id)sender {
    ShowConfigWindow();
}


-(void)editConfigAction:(id)sender {
    char* configPath = GetCurrentConfigPath();

    if (configPath) {
        NSString *path = [NSString stringWithUTF8String:configPath];
        free(configPath); // Free the C string after converting to NSString

        if (path && [path length] > 0) {
            NSURL *fileURL = [NSURL fileURLWithPath:path];
            [[NSWorkspace sharedWorkspace] openURL:fileURL];
        }
    }
}

- (void)applicationWillFinishLaunching:(NSNotification *)aNotification
{
    NSAppleEventManager *appleEventManager = [NSAppleEventManager sharedAppleEventManager];
    [appleEventManager setEventHandler:self
                    andSelector:@selector(handleGetURLEvent:withReplyEvent:)
                    forEventClass:kInternetEventClass andEventID:kAEGetURL];
}

- (bool)application:(NSApplication *)sender openFile:(NSString *)filename {
    NSLog(@"Opening file: %@", filename);

    // Convert the file path to a file:// URL
    NSURL *fileURL = [NSURL fileURLWithPath:filename];
    NSString *urlString = [fileURL absoluteString];

    // Handle the file URL the same way we handle other URLs
    HandleURL((char*)[urlString UTF8String], NULL, NULL, NULL, NULL, false);

    return true;
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

    self.receivedURL = true;

    NSRunningApplication *frontApp = [[NSWorkspace sharedWorkspace] frontmostApplication];

    // Assume Finicky is in front if we are not keeping running, since there's no good way
    // to detect if Finicky was launched in the background
    bool finickyIsInFront =  !self.keepRunning || [frontApp isEqual:[NSRunningApplication currentApplication]];

    char *windowTitle = NULL;

    if (application) {
        NSString *appName = [application localizedName];
        NSString *appBundleID = [application bundleIdentifier];
        NSString *appPath = [[application bundleURL] path];

        name = [appName UTF8String];
        bundleId = [appBundleID UTF8String];
        path = [appPath UTF8String];

        // Try to get the focused window title via Accessibility API
        AXUIElementRef appElement = AXUIElementCreateApplication(pid);
        if (appElement) {
            AXUIElementRef focusedWindow = NULL;
            AXError err = AXUIElementCopyAttributeValue(appElement, kAXFocusedWindowAttribute, (CFTypeRef *)&focusedWindow);
            if (err == kAXErrorSuccess && focusedWindow) {
                CFTypeRef titleValue = NULL;
                AXError titleErr = AXUIElementCopyAttributeValue(focusedWindow, kAXTitleAttribute, &titleValue);
                if (titleErr == kAXErrorSuccess && titleValue && CFGetTypeID(titleValue) == CFStringGetTypeID()) {
                    // strdup to keep a copy alive after CFRelease (no ARC in this project)
                    windowTitle = strdup([(NSString *)titleValue UTF8String]);
                }
                if (titleValue) CFRelease(titleValue);
                CFRelease(focusedWindow);
            }
            CFRelease(appElement);
        }
    } else {
        NSLog(@"No running application found with PID: %d", pid);
    }

    // If Finicky isn't frontmost, we take that to mean that the browser should, by default, be opened in the background
    HandleURL((char*)url, (char*)name, (char*)bundleId, (char*)path, windowTitle, !finickyIsInFront);
    free(windowTitle);
}

- (bool)application:(NSApplication *)application willContinueUserActivityWithType:(NSString *)userActivityType {
    return [userActivityType isEqualToString:NSUserActivityTypeBrowsingWeb];
}

- (bool)application:(NSApplication *)application continueUserActivity:(NSUserActivity *)userActivity restorationHandler:(void (^)(NSArray<id<NSUserActivityRestoring>> * _Nullable))restorationHandler {
    if (![userActivity.activityType isEqualToString:NSUserActivityTypeBrowsingWeb]) {
        return false;
    }

    NSURL *url = userActivity.webpageURL;
    if (!url) {
        return false;
    }

    HandleURL((char*)[[url absoluteString] UTF8String], NULL, NULL, NULL, NULL, false);
    return true;
}

- (void)application:(NSApplication *)application didFailToContinueUserActivityWithType:(NSString *)userActivityType error:(NSError *)error {
    // Handle failure if needed
}

@end

void RunApp(bool forceOpenWindow, bool showStatusItem, bool keepRunning) {
    @autoreleasepool {
        // Initialize on the main thread directly, not async
        [NSApplication sharedApplication];

        BrowseAppDelegate *app = [[BrowseAppDelegate alloc] initWithForceOpenWindow:forceOpenWindow initShow:showStatusItem keepRunning:keepRunning];
        [NSApp setDelegate:app];

        [NSApp finishLaunching];
        [NSApp run];
    }
}
