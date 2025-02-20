#import <Cocoa/Cocoa.h>

const char* resolveAppPath(const char* appName) {
    @autoreleasepool {
        NSString *name = [NSString stringWithUTF8String:appName];
        NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
        
        NSString *appPath = [workspace fullPathForApplication:name];
        if (!appPath) {
            return NULL;
        }
        return strdup([appPath UTF8String]);
    }
}

const char* resolveBundleId(const char* appName) {
    @autoreleasepool {
        const char* path = resolveAppPath(appName);
        if (!path) {
            return NULL;
        }
        
        NSString *appPath = [NSString stringWithUTF8String:path];
        free((void*)path);
        
        NSBundle *appBundle = [NSBundle bundleWithPath:appPath];
        if (!appBundle) {
            return NULL;
        }
        
        NSString *bundleId = [appBundle bundleIdentifier];
        if (!bundleId) {
            return NULL;
        }
        
        return strdup([bundleId UTF8String]);
    }
} 