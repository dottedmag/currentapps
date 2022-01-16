#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

NS_ASSUME_NONNULL_BEGIN

@interface AppDelegate : NSObject<NSApplicationDelegate> {
    bool idle;
    NSTimer *idleTimer;
}
- (void)applicationDidFinishLaunching:(NSNotification *)notification;
@end

NS_ASSUME_NONNULL_END
