// browse.h
#import <Cocoa/Cocoa.h>
#include <syslog.h>

const char* getDefaultHandlerForURLScheme(const char* scheme);

bool setDefaultHandlerForURLScheme(const char* bundleId, const char* scheme);
