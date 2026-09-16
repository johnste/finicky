#ifndef WINDOW_H
#define WINDOW_H

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

@interface WindowController : NSObject <WKURLSchemeHandler, WKNavigationDelegate>
- (void)showWindow;
- (void)closeWindow;
@end

void ShowWindow(void);
void CloseWindow(void);
void SetHTMLContent(const char* content);
void SetFileContent(const char* path, const char* content);
void SetFileContentWithLength(const char* path, const char* content, size_t length);

extern void WindowDidClose(void);
extern int GetAPIPort(void);
extern char* GetAPIToken(void);

#endif /* WINDOW_H */
