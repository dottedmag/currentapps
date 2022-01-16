#import "AppDelegate.h"

//
// Requesting privileges to capture screen (read window names):
//
// CGPreflightScreenCaptureAccess();
// CGRequestScreenCaptureAccess();
//

//
// List windows:
//
// CFArrayRef wis = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly, kCGNullWindowID);
// CFDictionaryGetValue(CFArrayGetValueAtIndex(wis, i), kCGWindowOwnerName);
// CFDictionaryGetValue(CFArrayGetValueAtIndex(wis, i), kCGWindowName); <- needs "Screen Capture" privileges
//

static NSDateFormatter *dateFormatter;

void out(const char *fmt, ...) {
    printf("%s ", [[dateFormatter stringFromDate: [NSDate now]] UTF8String]);

    va_list ap;
    va_start(ap, fmt);
    vprintf(fmt, ap);
    va_end(ap);
    putchar('\n');
}

#define MIN_IDLE 10.0
#define JITTER 1.0

@implementation AppDelegate

- (void)applicationDidFinishLaunching:(NSNotification *)notification
{
    setlinebuf(stdout);
    dateFormatter = [[NSDateFormatter alloc] init];
    [dateFormatter setDateFormat:@"yyyy-MM-dd HH:mm:ss.SSSS"];

    out("Started");

    NSNotificationCenter *nc = [[NSWorkspace sharedWorkspace] notificationCenter];
    [nc addObserver:self selector:@selector(onApplicationActivated:) name:NSWorkspaceDidActivateApplicationNotification object:NULL];
    [nc addObserver:self selector:@selector(onSessionActivated:) name:NSWorkspaceSessionDidBecomeActiveNotification object:NULL];
    [nc addObserver:self selector:@selector(onSessionDeactivated:) name:NSWorkspaceSessionDidResignActiveNotification object:NULL];
    [nc addObserver:self selector:@selector(onWakeUp:) name:NSWorkspaceDidWakeNotification object:NULL];
    [nc addObserver:self selector:@selector(onWillPowerOff:) name:NSWorkspaceWillPowerOffNotification object:NULL];
    [nc addObserver:self selector:@selector(onWillSleep:) name:NSWorkspaceWillSleepNotification object:NULL];
    [nc addObserver:self selector:@selector(onScreensDidSleep:) name:NSWorkspaceScreensDidSleepNotification object:NULL];
    [nc addObserver:self selector:@selector(onScreensWakeUp:) name:NSWorkspaceScreensDidWakeNotification object:NULL];

    self->idleTimer = [NSTimer scheduledTimerWithTimeInterval:MIN_IDLE
                                                 target:self selector:@selector(onIdleTimer:) userInfo:NULL repeats:YES];
    self->idleTimer.tolerance = 1.0;
}

- (void)adjustIdleTimer:(double)idle
{
    self->idleTimer.fireDate = [[NSDate now] dateByAddingTimeInterval:(MIN_IDLE-idle)];
}

- (void)action
{
    [self handleIdle:0.0];
}

- (void)handleIdle:(double)idle {
    if (self->idle) {
        if (idle < MIN_IDLE-JITTER) {
            self->idle = false;
            out("Not idle @ %s", [[dateFormatter stringFromDate: [[NSDate now] dateByAddingTimeInterval:-idle]] UTF8String]);
            [self adjustIdleTimer:idle];
        } else {
            // do nothing, idle and no input messages
        }
    } else {
        if (idle < MIN_IDLE-JITTER) {
            [self adjustIdleTimer:idle];
        } else {
            self->idle = true;
            out("Idle");
        }
    }
}

- (void)onIdleTimer:(NSTimer *)timer
{
    double idle = CGEventSourceSecondsSinceLastEventType(kCGEventSourceStateHIDSystemState, kCGAnyInputEventType);
    [self handleIdle:idle];
}

- (void)onApplicationActivated:(NSNotification *)notification
{
    NSRunningApplication *app = [[notification userInfo] objectForKey:NSWorkspaceApplicationKey];
    out("Application activated %s", [[app bundleIdentifier] UTF8String]);
    [self action];
}

- (void)onSessionActivated:(NSNotification *)notification
{
    out("Session activated");
    [self action];
}

- (void)onSessionDeactivated:(NSNotification *)notification
{
    out("Session deactivated");
    [self action];
}

- (void)onWakeUp:(NSNotification *)notification
{
    out("Wake up");
    [self action];
}

- (void)onWillPowerOff:(NSNotification *)notification
{
    out("Power off");
    [self action];
}

- (void)onWillSleep:(NSNotification *)notification
{
    out("Sleep");
    [self action];
}

- (void)onScreensDidSleep:(NSNotification *)notification
{
    out("Screen sleep");
    [self action];
}

- (void)onScreensWakeUp:(NSNotification *)notification
{
    out("Screen wake up");
    [self action];
}

@end
