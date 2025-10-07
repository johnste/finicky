#import "window.h"

static WindowController* windowController = nil;
static NSString* htmlContent = nil;
static NSMutableDictionary* fileContents = nil;

void SetHTMLContent(const char* content) {
    if (content) {
        htmlContent = [NSString stringWithUTF8String:content];
    }
}

void SetFileContent(const char* path, const char* content) {
    if (!fileContents) {
        fileContents = [[NSMutableDictionary alloc] init];
    }
    if (path && content) {
        NSString* pathStr = [NSString stringWithUTF8String:path];
        if ([pathStr hasSuffix:@".png"]) {
            // For PNG files, use SetFileContentWithLength instead
            SetFileContentWithLength(path, content, strlen(content));
        } else {
            // For text files, store as string
            NSString* contentStr = [NSString stringWithUTF8String:content];
            fileContents[pathStr] = contentStr;
        }
    }
}

void SetFileContentWithLength(const char* path, const char* content, size_t length) {
    if (!fileContents) {
        fileContents = [[NSMutableDictionary alloc] init];
    }
    if (path && content) {
        NSString* pathStr = [NSString stringWithUTF8String:path];
        NSData* data = [[NSData alloc] initWithBytes:content length:length];
        fileContents[pathStr] = data;
    }
}

@implementation WindowController {
    NSWindow* window;
    WKWebView* webView;
}

- (id)init {
    self = [super init];
    NSLog(@"Initialize window controller");
    if (self) {
        // Always setup window on main thread
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
    // Create window
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

    // Configure WKWebView
    WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];
    [config.userContentController addScriptMessageHandler:self name:@"finicky"];
    [config setURLSchemeHandler:self forURLScheme:@"finicky-assets"];

    // Create WKWebView
    webView = [[WKWebView alloc] initWithFrame:window.contentView.bounds configuration:config];
    webView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
    webView.navigationDelegate = self;
    webView.wantsLayer = true;
    webView.layer.backgroundColor = [NSColor colorWithCalibratedWhite:0.1 alpha:1.0].CGColor;

    [webView.configuration.preferences setValue:@true forKey:@"developerExtrasEnabled"];

    // Load HTML content
    if (htmlContent) {
        NSURL* baseURL = [NSURL URLWithString:@"finicky-assets://local/"];
        [webView loadHTMLString:htmlContent baseURL:baseURL];
    } else {
        NSLog(@"Warning: HTML content not set");
    }

    // Add window close notification observer
    [[NSNotificationCenter defaultCenter] addObserver:self
                                         selector:@selector(windowWillClose:)
                                             name:NSWindowWillCloseNotification
                                           object:window];

    // Set webView as content view
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

- (void)sendMessageToWebView:(NSString *)message {
    // The message is already JSON encoded from Go, pass it as a string literal, but escape the quotes and backslashes
    NSString *escapedMessage = [[message stringByReplacingOccurrencesOfString:@"\\" withString:@"\\\\"]
                                        stringByReplacingOccurrencesOfString:@"\"" withString:@"\\\""];
    NSString *js = [NSString stringWithFormat:@"finicky.receiveMessage(\"%@\")", escapedMessage];

    if ([NSThread isMainThread]) {
        if (webView && !webView.loading) {
            [webView evaluateJavaScript:js completionHandler:nil];
        }
    } else {
        dispatch_async(dispatch_get_main_queue(), ^{
            if (webView && !webView.loading) {
                [webView evaluateJavaScript:js completionHandler:nil];
            }
        });
    }
}

- (void)userContentController:(WKUserContentController *)userContentController
      didReceiveScriptMessage:(WKScriptMessage *)message {
    // Handle messages from JavaScript here
    NSLog(@"Received message from WebView: %@", message.body);
}

#pragma mark - WKURLSchemeHandler

- (void)webView:(WKWebView *)webView startURLSchemeTask:(id<WKURLSchemeTask>)urlSchemeTask {
    NSURLRequest *request = urlSchemeTask.request;
    NSString *path = request.URL.path;
    if ([path hasPrefix:@"/"]) {
        path = [path substringFromIndex:1];
    }

    // Remove 'local/' prefix if present
    if ([path hasPrefix:@"local/"]) {
        path = [path substringFromIndex:6];
    }

    id content = fileContents[path];
    if (content) {
        NSData *data;
        if ([content isKindOfClass:[NSData class]]) {
            data = (NSData *)content;
        } else {
            data = [(NSString *)content dataUsingEncoding:NSUTF8StringEncoding];
        }

        NSString *mimeType = @"text/plain";
        if ([path hasSuffix:@".css"]) {
            mimeType = @"text/css";
        } else if ([path hasSuffix:@".js"]) {
            mimeType = @"application/javascript";
        } else if ([path hasSuffix:@".png"]) {
            mimeType = @"image/png";
        }

        NSURLResponse *response = [[NSURLResponse alloc] initWithURL:request.URL
                                                          MIMEType:mimeType
                                             expectedContentLength:data.length
                                                  textEncodingName:nil];

        [urlSchemeTask didReceiveResponse:response];
        [urlSchemeTask didReceiveData:data];
        [urlSchemeTask didFinish];
    } else {
        NSLog(@"Asset not found: %@", path);
        [urlSchemeTask didFailWithError:[NSError errorWithDomain:NSURLErrorDomain
                                                          code:NSURLErrorResourceUnavailable
                                                      userInfo:nil]];
    }
}

- (void)webView:(WKWebView *)webView stopURLSchemeTask:(id<WKURLSchemeTask>)urlSchemeTask {
    // Nothing to do here
}

#pragma mark - WKNavigationDelegate

- (void)webView:(WKWebView *)webView didFinishNavigation:(WKNavigation *)navigation {
    // Notify Go that the window is ready to receive messages
    extern void WindowIsReady(void);
    WindowIsReady();
}

- (void)webView:(WKWebView *)webView decidePolicyForNavigationAction:(WKNavigationAction *)navigationAction decisionHandler:(void (^)(WKNavigationActionPolicy))decisionHandler {
    NSURL *url = navigationAction.request.URL;

    // Handle finicky-assets:// URLs internally
    if ([url.scheme isEqualToString:@"finicky-assets"]) {
        decisionHandler(WKNavigationActionPolicyAllow);
        return;
    }

    // If it's a regular link click (not a page load)
    if (navigationAction.navigationType == WKNavigationTypeLinkActivated) {
        // Open the URL in the default browser
        [[NSWorkspace sharedWorkspace] openURL:url];
        decisionHandler(WKNavigationActionPolicyCancel);
        return;
    }

    // Allow all other navigation
    decisionHandler(WKNavigationActionPolicyAllow);
}

// Add new method to handle window close
- (void)windowWillClose:(NSNotification *)notification {
    extern void WindowDidClose(void);
    WindowDidClose();
}

- (void)setupMenu {
    NSMenu *mainMenu = [[NSMenu alloc] init];
    [NSApp setMainMenu:mainMenu];

    // Application menu
    NSMenuItem *appMenuItem = [[NSMenuItem alloc] init];
    [mainMenu addItem:appMenuItem];
    NSMenu *appMenu = [[NSMenu alloc] init];
    [appMenuItem setSubmenu:appMenu];

    // Quit menu item (⌘Q)
    NSMenuItem *quitMenuItem = [[NSMenuItem alloc] initWithTitle:@"Quit"
                                                        action:@selector(terminate:)
                                                 keyEquivalent:@"q"];
    [quitMenuItem setTarget:NSApp];
    [appMenu addItem:quitMenuItem];

    // File menu
    NSMenuItem *fileMenuItem = [[NSMenuItem alloc] init];
    [mainMenu addItem:fileMenuItem];
    NSMenu *fileMenu = [[NSMenu alloc] initWithTitle:@"File"];
    [fileMenuItem setSubmenu:fileMenu];

    // Close window menu item (⌘W)
    NSMenuItem *closeMenuItem = [[NSMenuItem alloc] initWithTitle:@"Close Window"
                                                         action:@selector(performClose:)
                                                  keyEquivalent:@"w"];
    [closeMenuItem setTarget:window];
    [fileMenu addItem:closeMenuItem];

    // Edit menu
    NSMenuItem *editMenuItem = [[NSMenuItem alloc] init];
    [mainMenu addItem:editMenuItem];
    NSMenu *editMenu = [[NSMenu alloc] initWithTitle:@"Edit"];
    [editMenuItem setSubmenu:editMenu];

    // Add Cut menu item (⌘X)
    NSMenuItem *cutMenuItem = [[NSMenuItem alloc] initWithTitle:@"Cut"
                                                       action:@selector(cut:)
                                                keyEquivalent:@"x"];
    [editMenu addItem:cutMenuItem];

    // Add Copy menu item (⌘C)
    NSMenuItem *copyMenuItem = [[NSMenuItem alloc] initWithTitle:@"Copy"
                                                        action:@selector(copy:)
                                                 keyEquivalent:@"c"];
    [editMenu addItem:copyMenuItem];

    // Add Paste menu item (⌘V)
    NSMenuItem *pasteMenuItem = [[NSMenuItem alloc] initWithTitle:@"Paste"
                                                         action:@selector(paste:)
                                                  keyEquivalent:@"v"];
    [editMenu addItem:pasteMenuItem];

    // Add separator
    [editMenu addItem:[NSMenuItem separatorItem]];

    // Add Select All menu item (⌘A)
    NSMenuItem *selectAllMenuItem = [[NSMenuItem alloc] initWithTitle:@"Select All"
                                                            action:@selector(selectAll:)
                                                     keyEquivalent:@"a"];
    [editMenu addItem:selectAllMenuItem];
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

void SendMessageToWebView(const char* message) {
    if (windowController) {
        NSString *nsMessage = [NSString stringWithUTF8String:message];
        [windowController sendMessageToWebView:nsMessage];
    }
}