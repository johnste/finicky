#ifndef SAFARI_MENU_H
#define SAFARI_MENU_H

#define SAFARI_MENU_OK 0
#define SAFARI_MENU_NOT_TRUSTED 1
#define SAFARI_MENU_NOT_RUNNING 2
#define SAFARI_MENU_NO_ITEM 3
#define SAFARI_MENU_TIMEOUT 4

// Brings Safari forward, presses the File menu item titled menuTitle through
// the Accessibility API and waits up to timeoutSeconds for a new window to
// appear. Returns one of the SAFARI_MENU_ codes above.
int safariOpenWindowFromMenu(const char *menuTitle, double timeoutSeconds);

#endif
