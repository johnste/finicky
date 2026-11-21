#ifndef SAFARI_MENU_H
#define SAFARI_MENU_H

#define SAFARI_MENU_OK 0
#define SAFARI_MENU_NOT_TRUSTED 1
#define SAFARI_MENU_NOT_RUNNING 2
#define SAFARI_MENU_NO_ITEM 3
#define SAFARI_MENU_TIMEOUT 4
#define SAFARI_MENU_NO_WINDOWS 5
#define SAFARI_MENU_PRESS_FAILED 6

// Maps the AXError of a menu press to the result it stands for before the wait
// for a new window: SAFARI_MENU_TIMEOUT when the press ran or may still run,
// SAFARI_MENU_PRESS_FAILED when it surely did not.
int safariPressResult(int axError);

// Brings Safari forward, presses the File menu item titled menuTitle through
// the Accessibility API and waits up to timeoutSeconds for a new window to
// appear. Returns one of the SAFARI_MENU_ codes above.
int safariOpenWindowFromMenu(const char *menuTitle, double timeoutSeconds);

#endif
