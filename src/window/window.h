#ifndef WINDOW_H
#define WINDOW_H

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

@interface WindowController : NSObject <WKScriptMessageHandler, WKURLSchemeHandler, WKNavigationDelegate>
- (void)showWindow;
- (void)closeWindow;
- (void)sendMessageToWebView:(NSString *)message;
@end

void ShowWindow(void);
void CloseWindow(void);
void SendMessageToWebView(const char* message);
void SetHTMLContent(const char* content);
void SetFileContent(const char* path, const char* content);
void SetFileContentWithLength(const char* path, const char* content, size_t length);

extern void WindowDidClose(void);

#endif /* WINDOW_H */