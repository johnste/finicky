#ifndef INFO_H
#define INFO_H

#include <stdbool.h>

typedef struct {
    bool shift;
    bool option;
    bool command;
    bool control;
    bool capsLock;
    bool fn;
} ModifierKeys;

typedef struct {
    const char* localizedName;
    const char* name;
} SystemInfo;

ModifierKeys getModifierKeys(void);
SystemInfo getSystemInfo(void);

#endif /* INFO_H */