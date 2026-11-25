#ifndef PICKER_H
#define PICKER_H

// Shows a modal picker listing the given labels and returns the index of the
// chosen label, or -1 if the user cancelled. tooltip, when not NULL, is shown
// on hover and carries the untruncated form of message.
int showBrowserPicker(const char *title, const char *message, const char *tooltip, const char **labels, int count);

#endif
