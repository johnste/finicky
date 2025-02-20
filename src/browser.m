
#include "browser.h"
#import <Cocoa/Cocoa.h>
#import <stdlib.h>

const char* getDefaultHandlerForURLScheme(const char* scheme) {
    @autoreleasepool {
        if (!scheme) return NULL;

        // Convert C string to NSString
        NSString *schemeStr = [NSString stringWithUTF8String:scheme];
        if (!schemeStr) return NULL;

        // Create an NSURL with the scheme
        NSURL *url = [NSURL URLWithString:[NSString stringWithFormat:@"%@://", schemeStr]];

        // Get the default application URL for the scheme
        NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
        NSURL *appURL = [workspace URLForApplicationToOpenURL:url];

        if (appURL) {
            const char *result = [appURL.path UTF8String];
            // NSLog(@"Default application URL for scheme '%@': %@", schemeStr, appURL.path);
            // Get the bundle identifier for the application at appURL
            NSBundle *appBundle = [NSBundle bundleWithURL:appURL];
            NSString *bundleId = [appBundle bundleIdentifier];
            if (bundleId) {
                // NSLog(@"Bundle ID for application: %@", bundleId);
                return strdup([bundleId UTF8String]); // Convert NSString to C string and return a copy
            } else {
                // NSLog(@"Failed to get Bundle ID for application at URL: %@", appURL.path);
            }
            return NULL;
        }

        return NULL;
    }
}

bool setDefaultHandlerForURLScheme(const char* bundleId, const char* scheme) {
    @autoreleasepool {
        if (!bundleId || !scheme) return false;

        // Convert C strings to NSString
        NSString *bundleIdStr = [NSString stringWithUTF8String:bundleId];
        NSString *schemeStr = [NSString stringWithUTF8String:scheme];

        // Create an NSURL with the scheme
        NSURL *url = [NSURL URLWithString:[NSString stringWithFormat:@"%@://", schemeStr]];

        // Get the URL for the application with the given bundle ID
        NSWorkspace *workspace = [NSWorkspace sharedWorkspace];
        // NSURL *appURL = [workspace URLForApplicationWithBundleIdentifier:bundleIdStr];
        NSBundle *mainBundle = [NSBundle mainBundle];
        NSURL *appURL = mainBundle.bundleURL;

        if (!appURL) {
            NSLog(@"Failed to find application with bundle ID: %@", bundleIdStr);
            return false;
        }
        
        // Check if appURL contains ".app"
        // if (![appURL.path containsString:@".app"]) {
        //     NSLog(@"The application URL does not contain '.app': %@", appURL.path);
        //     return false;
        // }

        NSLog(@"Setting default application: %@", appURL);
        NSLog(@"Setting default application for scheme: %@", schemeStr);
        [workspace setDefaultApplicationAtURL:appURL toOpenURLsWithScheme:schemeStr completionHandler:^(NSError *error) {
            if (error) {
                NSLog(@"Error setting default handler: %@", error);
            } else {
                NSLog(@"Successfully set default handler for scheme: %@", schemeStr);
            }
        }];
        
        return true;
    }
}
