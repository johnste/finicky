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
    [self terminateOtherInstances];

    bool openWindow = self.forceOpenWindow;

    // Only show menu item if the option is enabled, and we either didn't receive a URL or we are keeping
    // the application running. We don't want to show the icon if Finicky is just receiving a url to open
    // and is expected to exit after
    if (self.showMenuItem && (self.keepRunning || !self.receivedURL)) {
        [self createStatusItem];
    }

    QueueWindowDisplay(openWindow);
}

// Ensure only one Finicky process is running. macOS's Launch Services normally
// routes GetURL events to an already-running instance, but that routing can fail
// (stale LS registration after app moves/updates, different bundle paths, SSO
// agents launching from another context) and silently spawn a second process
// that proceeds to create its own status bar icon. Without a guard, duplicates
// accumulate in the menu bar over time. Terminate any other instances we find
// with the same bundle identifier so we're the single surviving process.
- (void)terminateOtherInstances {
    NSString *selfBundleID = [[NSBundle mainBundle] bundleIdentifier];
    if (!selfBundleID) {
        return;
    }

    NSArray<NSRunningApplication *> *instances = [NSRunningApplication runningApplicationsWithBundleIdentifier:selfBundleID];
    pid_t myPID = [[NSRunningApplication currentApplication] processIdentifier];

    NSMutableArray<NSRunningApplication *> *duplicates = [NSMutableArray array];
    for (NSRunningApplication *app in instances) {
        if ([app processIdentifier] == myPID) continue;
        if ([app isTerminated]) continue;
        [duplicates addObject:app];
    }

    if (duplicates.count == 0) {
        return;
    }

    // -[NSRunningApplication terminate] returning YES only means the request was
    // sent, not that the target has exited. Fire all requests in parallel, then
    // share a single deadline across the poll so 16 hung duplicates don't cost
    // 16x the wait budget.
    for (NSRunningApplication *app in duplicates) {
        NSLog(@"Terminating duplicate Finicky instance (pid %d)", [app processIdentifier]);
        [app terminate];
    }

    NSDate *deadline = [NSDate dateWithTimeIntervalSinceNow:1.0];
    while ([deadline timeIntervalSinceNow] > 0) {
        BOOL anyAlive = NO;
        for (NSRunningApplication *app in duplicates) {
            if (![app isTerminated]) { anyAlive = YES; break; }
        }
        if (!anyAlive) return;

        [[NSRunLoop currentRunLoop] runMode:NSDefaultRunLoopMode
                                 beforeDate:[NSDate dateWithTimeIntervalSinceNow:0.05]];
    }

    for (NSRunningApplication *app in duplicates) {
        if ([app isTerminated]) continue;
        NSLog(@"Duplicate Finicky instance (pid %d) did not exit within 1s; force-terminating", [app processIdentifier]);
        [app forceTerminate];
    }
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

- (void)setErrorState:(bool)hasError {
    if (!self.statusItem) return;

    NSString *iconPath = [[NSBundle mainBundle] pathForResource:@"menu-bar" ofType:@"icns"];
    NSImage *baseIcon = [[NSBundle mainBundle] imageForResource:@"menu-bar"];
    if (!baseIcon) {
        // Fallback if imageForResource fails
        baseIcon = [[NSImage alloc] initWithContentsOfFile:iconPath];
    }
    if (!baseIcon) return;
    baseIcon.size = NSMakeSize(18, 18);

    if (!hasError) {
        // Normal state: Pure template icon (automatically flips black/white)
        baseIcon.template = YES;
        self.statusItem.button.image = baseIcon;
        return;
    }

    // Error state: Dynamic faded icon + SF Symbol warning triangle
    NSImage *badged = [NSImage imageWithSize:NSMakeSize(18, 18) flipped:NO drawingHandler:^BOOL(NSRect dstRect) {

        // 1. Determine current menu bar theme
        BOOL isDarkMode = NO;
        if (@available(macOS 10.14, *)) {
            NSAppearanceName bestMatch = [NSApp.effectiveAppearance bestMatchFromAppearancesWithNames:@[NSAppearanceNameDarkAqua, NSAppearanceNameAqua]];
            isDarkMode = [bestMatch isEqualToString:NSAppearanceNameDarkAqua];
        }

        // 2. Prepare and tint your base hand icon
        NSImage *tintedIcon = [baseIcon copy];
        [tintedIcon lockFocus];
        if (isDarkMode) {
            [[NSColor whiteColor] set];
        } else {
            [[NSColor controlTextColor] set];
        }
        NSRectFillUsingOperation(NSMakeRect(0, 0, 18, 18), NSCompositingOperationSourceAtop);
        [tintedIcon unlockFocus];

        // Draw your hand icon faded out (40% opacity / 0.4 alpha)
        [tintedIcon drawInRect:dstRect fromRect:NSZeroRect operation:NSCompositingOperationSourceOver fraction:0.4];

        // 3. Load a crisp vector warning triangle using SF Symbols (macOS 11.0+)
        NSImage *warningTriangle = nil;
        if (@available(macOS 11.0, *)) {
            warningTriangle = [NSImage imageWithSystemSymbolName:@"exclamationmark.triangle.fill" accessibilityDescription:@"Error"];
        }

        // Define badge placement (bottom-right corner makes the hand's pointer finger stay visible)
        CGFloat badgeSize = 13.0;
        NSRect badgeRect = NSMakeRect(18 - badgeSize, 0, badgeSize, badgeSize);

        if (warningTriangle) {
            // Tint the SF Symbol warning triangle to vibrant system red
            NSImage *redTriangle = [warningTriangle copy];
            [redTriangle lockFocus];
            [[NSColor systemRedColor] set];
            NSRectFillUsingOperation(NSMakeRect(0, 0, warningTriangle.size.width, warningTriangle.size.height), NSCompositingOperationSourceAtop);
            [redTriangle unlockFocus];

            // Draw a tiny dark/light mask cutout behind the triangle so it pops off the faded hand
            NSBezierPath *cutout = [NSBezierPath bezierPathWithOvalInRect:NSInsetRect(badgeRect, 1.0, 1.0)];
            if (isDarkMode) {
                [[NSColor colorWithWhite:0.12 alpha:1.0] setFill]; // Matches dark menu bar background
            } else {
                [[NSColor whiteColor] setFill];
            }
            [cutout fill];

            [redTriangle drawInRect:badgeRect];
        } else {
            // Fallback for older macOS versions: Draw a crisp vector red dot if SF Symbols aren't available
            NSBezierPath *dot = [NSBezierPath bezierPathWithOvalInRect:badgeRect];
            [[NSColor systemRedColor] setFill];
            [dot fill];
        }

        return YES;
    }];

    // Tell macOS to preserve our exact custom drawing pipeline
    badged.template = NO;

    self.statusItem.button.image = badged;
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
    self.receivedURL = true;

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

void SetStatusItemError(bool hasError) {
    dispatch_async(dispatch_get_main_queue(), ^{
        BrowseAppDelegate *app = (BrowseAppDelegate *)[NSApp delegate];
        [app setErrorState:hasError];
    });
}
