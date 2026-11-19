#ifndef WINDOW_H
#define WINDOW_H

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

@interface WindowController : NSObject <WKNavigationDelegate>
- (void)showWindow;
- (void)closeWindow;
@end

void ShowWindow(void);
void CloseWindow(void);

extern void WindowDidClose(void);
extern char* GetUIURL(void);
extern char* GetAPIToken(void);

#endif /* WINDOW_H */
