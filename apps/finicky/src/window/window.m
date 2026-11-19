#import "window.h"

static WindowController* windowController = nil;

@implementation WindowController {
    NSWindow* window;
    WKWebView* webView;
}

- (id)init {
    self = [super init];
    NSLog(@"Initialize window controller");
    if (self) {
        if ([NSThread isMainThread]) {
            [self setupWindow];
            [self setupMenu];
        } else {
            dispatch_sync(dispatch_get_main_queue(), ^{
                [self setupWindow];
                [self setupMenu];
            });
        }
    }
    return self;
}

- (void)setupWindow {
    window = [[NSWindow alloc] initWithContentRect:NSMakeRect(0, 0, 860, 600)
                                       styleMask:NSWindowStyleMaskTitled |
                                                NSWindowStyleMaskClosable |
                                                NSWindowStyleMaskMiniaturizable |
                                                NSWindowStyleMaskResizable
                                         backing:NSBackingStoreBuffered
                                           defer:NO];
    [window setTitle:@"Finicky"];
    [window center];
    [window setReleasedWhenClosed:NO];
    [window setBackgroundColor:[NSColor colorWithCalibratedWhite:0.1 alpha:1.0]];
    [window setMinSize:NSMakeSize(800, 500)];
    [window setMaxSize:NSMakeSize(1200, 900)];

    WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];

    // The UI is served by the same local HTTP server as the API (see
    // StartAPIServer in server.go), so its requests are same-origin. Inject the
    // per-process auth token so the UI can call the API. Both strings must be
    // freed since they cross the CGo boundary as copies (see window.go).
    char* urlCStr = GetUIURL();
    NSURL *uiURL = [NSURL URLWithString:[NSString stringWithUTF8String:urlCStr]];
    free(urlCStr);
    char* tokenCStr = GetAPIToken();
    NSString *token = [NSString stringWithUTF8String:tokenCStr];
    free(tokenCStr);
    NSString *apiScript = [NSString stringWithFormat:
        @"window.__FINICKY_API_TOKEN__ = '%@';", token];
    WKUserScript *apiUserScript = [[WKUserScript alloc]
        initWithSource:apiScript
         injectionTime:WKUserScriptInjectionTimeAtDocumentStart
      forMainFrameOnly:YES];
    [config.userContentController addUserScript:apiUserScript];

    webView = [[WKWebView alloc] initWithFrame:window.contentView.bounds configuration:config];
    webView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
    webView.navigationDelegate = self;
    webView.wantsLayer = true;
    webView.layer.backgroundColor = [NSColor colorWithCalibratedWhite:0.1 alpha:1.0].CGColor;

    [webView.configuration.preferences setValue:@true forKey:@"developerExtrasEnabled"];

    [webView loadRequest:[NSURLRequest requestWithURL:uiURL]];

    [[NSNotificationCenter defaultCenter] addObserver:self
                                         selector:@selector(windowWillClose:)
                                             name:NSWindowWillCloseNotification
                                           object:window];

    window.contentView = webView;
}

- (void)showWindow {
    if ([NSThread isMainThread]) {
        [window makeKeyAndOrderFront:nil];
        [NSApp activateIgnoringOtherApps:true];
    } else {
        dispatch_async(dispatch_get_main_queue(), ^{
            [window makeKeyAndOrderFront:nil];
            [NSApp activateIgnoringOtherApps:true];
        });
    }
}

- (void)closeWindow {
    if ([NSThread isMainThread]) {
        [window close];
    } else {
        dispatch_async(dispatch_get_main_queue(), ^{
            [window close];
        });
    }
}

#pragma mark - WKNavigationDelegate

- (void)webView:(WKWebView *)webView didFinishNavigation:(WKNavigation *)navigation {
}

- (void)webView:(WKWebView *)webView decidePolicyForNavigationAction:(WKNavigationAction *)navigationAction decisionHandler:(void (^)(WKNavigationActionPolicy))decisionHandler {
    NSURL *url = navigationAction.request.URL;
    if (navigationAction.navigationType == WKNavigationTypeLinkActivated) {
        [[NSWorkspace sharedWorkspace] openURL:url];
        decisionHandler(WKNavigationActionPolicyCancel);
        return;
    }
    decisionHandler(WKNavigationActionPolicyAllow);
}

- (void)windowWillClose:(NSNotification *)notification {
    extern void WindowDidClose(void);
    WindowDidClose();
}

- (void)setupMenu {
    NSMenu *mainMenu = [[NSMenu alloc] init];
    [NSApp setMainMenu:mainMenu];

    NSMenuItem *appMenuItem = [[NSMenuItem alloc] init];
    [mainMenu addItem:appMenuItem];
    NSMenu *appMenu = [[NSMenu alloc] init];
    [appMenuItem setSubmenu:appMenu];
    NSMenuItem *quitMenuItem = [[NSMenuItem alloc] initWithTitle:@"Quit"
                                                        action:@selector(terminate:)
                                                 keyEquivalent:@"q"];
    [quitMenuItem setTarget:NSApp];
    [appMenu addItem:quitMenuItem];

    NSMenuItem *fileMenuItem = [[NSMenuItem alloc] init];
    [mainMenu addItem:fileMenuItem];
    NSMenu *fileMenu = [[NSMenu alloc] initWithTitle:@"File"];
    [fileMenuItem setSubmenu:fileMenu];
    NSMenuItem *closeMenuItem = [[NSMenuItem alloc] initWithTitle:@"Close Window"
                                                         action:@selector(performClose:)
                                                  keyEquivalent:@"w"];
    [closeMenuItem setTarget:window];
    [fileMenu addItem:closeMenuItem];

    NSMenuItem *editMenuItem = [[NSMenuItem alloc] init];
    [mainMenu addItem:editMenuItem];
    NSMenu *editMenu = [[NSMenu alloc] initWithTitle:@"Edit"];
    [editMenuItem setSubmenu:editMenu];
    [editMenu addItem:[[NSMenuItem alloc] initWithTitle:@"Cut"    action:@selector(cut:)       keyEquivalent:@"x"]];
    [editMenu addItem:[[NSMenuItem alloc] initWithTitle:@"Copy"   action:@selector(copy:)      keyEquivalent:@"c"]];
    [editMenu addItem:[[NSMenuItem alloc] initWithTitle:@"Paste"  action:@selector(paste:)     keyEquivalent:@"v"]];
    [editMenu addItem:[NSMenuItem separatorItem]];
    [editMenu addItem:[[NSMenuItem alloc] initWithTitle:@"Undo"   action:@selector(undo:)      keyEquivalent:@"z"]];
    [editMenu addItem:[[NSMenuItem alloc] initWithTitle:@"Redo"   action:@selector(redo:)      keyEquivalent:@"Z"]];
    [editMenu addItem:[NSMenuItem separatorItem]];
    [editMenu addItem:[[NSMenuItem alloc] initWithTitle:@"Select All" action:@selector(selectAll:) keyEquivalent:@"a"]];
}

@end

void ShowWindow(void) {
    if (!windowController) {
        windowController = [[WindowController alloc] init];
    }
    [windowController showWindow];
}

void CloseWindow(void) {
    if (windowController) {
        [windowController closeWindow];
    }
}
