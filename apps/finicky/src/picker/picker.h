#ifndef PICKER_H
#define PICKER_H

// Shows a modal picker listing the given labels and returns the index of the
// chosen label, or -1 if the user cancelled.
int showBrowserPicker(const char *title, const char *message, const char **labels, int count);

#endif
