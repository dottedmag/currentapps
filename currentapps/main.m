#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import "AppDelegate.h"

int main(int argc, const char * argv[]) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        [NSApp setDelegate:[[AppDelegate alloc] init]];
        [NSApp run];
    }
    return 0;
}
