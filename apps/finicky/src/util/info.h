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

typedef struct {
    bool isConnected;
    bool isCharging;
    int percentage;
} PowerInfo;

ModifierKeys getModifierKeys(void);
SystemInfo getSystemInfo(void);
PowerInfo getPowerInfo(void);
_Bool isAppRunning(const char* identifier);
const char* getNSHomeDirectory(void);
const char* getNSCacheDirectory(void);

#endif /* INFO_H */