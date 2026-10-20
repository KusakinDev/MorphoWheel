#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
#include "bridge.h"
#include "_cgo_export.h"

static const CGFloat kSize = 380, kInner = 54, kOuter = 176;

@interface RingPanel : NSPanel
@end
@implementation RingPanel
- (BOOL)canBecomeKeyWindow { return YES; }
@end

@interface RingView : NSView <NSDraggingDestination>
@property (copy) NSString *group;
@property (copy) NSArray<NSString *> *paths;
@property (copy) NSArray<NSDictionary *> *items;
@property NSInteger hover;
@property BOOL dragMode;
@end

@implementation RingView

- (BOOL)acceptsFirstMouse:(NSEvent *)e { return YES; }
- (BOOL)acceptsFirstResponder { return YES; }

// Sector index under point p (view coordinates), or -1. Sector 0 starts at
// 12 o'clock and sectors proceed clockwise.
- (NSInteger)sectorAt:(NSPoint)p {
    if (self.items.count == 0) return -1;
    CGFloat dx = p.x - kSize / 2, dy = p.y - kSize / 2;
    CGFloat r = hypot(dx, dy);
    if (r < kInner || r > kOuter) return -1;
    CGFloat a = atan2(dx, dy); // 0 at top, clockwise positive
    if (a < 0) a += 2 * M_PI;
    return (NSInteger)(a / (2 * M_PI / self.items.count)) % self.items.count;
}

- (NSColor *)accent {
    if ([self.group isEqualToString:@"tools"]) return [NSColor systemBlueColor];
    if ([self.group isEqualToString:@"actions"]) return [NSColor systemGreenColor];
    return [NSColor systemRedColor];
}

- (void)drawRect:(NSRect)dirty {
    NSUInteger n = self.items.count;
    if (n == 0) return;
    NSPoint c = NSMakePoint(kSize / 2, kSize / 2);
    CGFloat step = 360.0 / n, gap = n > 1 ? 1.5 : 0;
    for (NSUInteger i = 0; i < n; i++) {
        // AppKit angles are counter-clockwise from 3 o'clock.
        CGFloat start = 90 - i * step - gap, end = 90 - (i + 1) * step + gap;
        NSBezierPath *p = [NSBezierPath bezierPath];
        [p appendBezierPathWithArcWithCenter:c radius:kOuter startAngle:start endAngle:end clockwise:YES];
        [p appendBezierPathWithArcWithCenter:c radius:kInner startAngle:end endAngle:start clockwise:NO];
        [p closePath];
        BOOL hot = (NSInteger)i == self.hover;
        [(hot ? [[self accent] colorWithAlphaComponent:0.95]
              : [NSColor colorWithWhite:0.1 alpha:0.82]) setFill];
        [p fill];

        CGFloat mid = (90 - (i + 0.5) * step) * M_PI / 180, rr = (kInner + kOuter) / 2;
        NSPoint lp = NSMakePoint(c.x + cos(mid) * rr, c.y + sin(mid) * rr);
        NSDictionary *attrs = @{
            NSFontAttributeName: [NSFont systemFontOfSize:13 weight:NSFontWeightMedium],
            NSForegroundColorAttributeName: [NSColor whiteColor]};
        NSString *t = self.items[i][@"title"];
        NSSize s = [t sizeWithAttributes:attrs];
        [t drawAtPoint:NSMakePoint(lp.x - s.width / 2, lp.y - s.height / 2) withAttributes:attrs];
    }
    // Centre badge: file count.
    NSString *badge = self.paths.count > 1
        ? [NSString stringWithFormat:@"%lu", (unsigned long)self.paths.count] : @"•";
    NSDictionary *b = @{NSFontAttributeName: [NSFont systemFontOfSize:20 weight:NSFontWeightBold],
                        NSForegroundColorAttributeName: [self accent]};
    NSSize bs = [badge sizeWithAttributes:b];
    [badge drawAtPoint:NSMakePoint(c.x - bs.width / 2, c.y - bs.height / 2) withAttributes:b];
}

- (void)setHoverIndex:(NSInteger)i {
    if (i != self.hover) { self.hover = i; [self setNeedsDisplay:YES]; }
}

- (void)runSector:(NSInteger)i {
    if (i < 0) return;
    NSData *d = [NSJSONSerialization dataWithJSONObject:self.paths options:0 error:nil];
    NSString *json = [[NSString alloc] initWithData:d encoding:NSUTF8StringEncoding];
    goDrop((char *)[self.items[i][@"id"] UTF8String], (char *)[json UTF8String]);
}

// Hotkey mode: click and hover.
- (void)mouseMoved:(NSEvent *)e { [self setHoverIndex:[self sectorAt:[self convertPoint:e.locationInWindow fromView:nil]]]; }
- (void)mouseDown:(NSEvent *)e {
    NSInteger i = [self sectorAt:[self convertPoint:e.locationInWindow fromView:nil]];
    [self.window orderOut:nil];
    [self runSector:i];
}
- (void)keyDown:(NSEvent *)e {
    NSString *k = e.charactersIgnoringModifiers;
    if (e.keyCode == kVK_Escape) { [self.window orderOut:nil]; return; }
    NSDictionary *map = @{@"1": @"convert", @"2": @"tools", @"3": @"actions"};
    if (map[k]) MorphoWheelShowRing([map[k] UTF8String],
        [[[NSString alloc] initWithData:[NSJSONSerialization dataWithJSONObject:self.paths options:0 error:nil]
                               encoding:NSUTF8StringEncoding] UTF8String]);
}

// Drag mode: the panel is a drop target.
- (NSDragOperation)draggingEntered:(id<NSDraggingInfo>)s { return [self draggingUpdated:s]; }
- (NSDragOperation)draggingUpdated:(id<NSDraggingInfo>)s {
    NSInteger i = [self sectorAt:[self convertPoint:s.draggingLocation fromView:nil]];
    [self setHoverIndex:i];
    return i >= 0 ? NSDragOperationCopy : NSDragOperationNone;
}
- (void)draggingExited:(id<NSDraggingInfo>)s {
    [self setHoverIndex:-1];
    NSWindow *w = self.window;
    dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.5 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
        if (self.hover < 0 && self.dragMode) [w orderOut:nil];
    });
}
- (BOOL)performDragOperation:(id<NSDraggingInfo>)s {
    NSInteger i = [self sectorAt:[self convertPoint:s.draggingLocation fromView:nil]];
    if (i < 0) return NO;
    [self runSector:i];
    return YES;
}
- (void)concludeDragOperation:(id<NSDraggingInfo>)s { [self.window orderOut:nil]; }
- (void)draggingEnded:(id<NSDraggingInfo>)s { [self.window orderOut:nil]; }
@end

@interface AppDelegate : NSObject <NSApplicationDelegate>
@property (strong) NSStatusItem *item;
@property (strong) RingPanel *panel;
@property (strong) RingView *view;
@property NSInteger lastDragCount;
@end

static AppDelegate *gDelegate;

static NSString *GroupForFlags(NSEventModifierFlags f) {
    if (f & NSEventModifierFlagShift) return @"convert";
    if (f & NSEventModifierFlagControl) return @"tools";
    if (f & NSEventModifierFlagOption) return @"actions";
    return nil;
}

static NSArray<NSString *> *DraggedFiles(void) {
    NSPasteboard *pb = [NSPasteboard pasteboardWithName:NSPasteboardNameDrag];
    NSArray<NSURL *> *urls = [pb readObjectsForClasses:@[NSURL.class]
        options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];
    NSMutableArray *res = [NSMutableArray array];
    for (NSURL *u in urls) [res addObject:u.path];
    return res;
}

static void ShowRing(NSString *group, NSArray<NSString *> *paths, BOOL dragMode) {
    AppDelegate *d = gDelegate;
    if (paths.count == 0) return;
    char *js = goSectors((char *)group.UTF8String, (char *)paths[0].UTF8String);
    NSData *data = [NSData dataWithBytes:js length:strlen(js)];
    free(js);
    NSArray *items = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
    if (items.count == 0) {
        if (!dragMode) MorphoWheelNotify("MorphoWheel", "Для этого файла нет операций");
        return;
    }
    d.view.group = group; d.view.paths = paths; d.view.items = items;
    d.view.hover = -1; d.view.dragMode = dragMode;
    [d.view setNeedsDisplay:YES];

    NSPoint m = [NSEvent mouseLocation];
    NSRect f = NSMakeRect(m.x - kSize / 2, m.y - kSize / 2, kSize, kSize);
    NSRect vis = NSScreen.mainScreen.visibleFrame;
    for (NSScreen *s in NSScreen.screens) if (NSPointInRect(m, s.frame)) vis = s.visibleFrame;
    f.origin.x = MAX(NSMinX(vis), MIN(f.origin.x, NSMaxX(vis) - kSize));
    f.origin.y = MAX(NSMinY(vis), MIN(f.origin.y, NSMaxY(vis) - kSize));
    [d.panel setFrame:f display:YES];
    [d.panel orderFrontRegardless];
    if (!dragMode) { [NSApp activateIgnoringOtherApps:YES]; [d.panel makeKeyWindow]; [d.panel makeFirstResponder:d.view]; }
}

void MorphoWheelShowRing(const char *group, const char *pathsJSON) {
    NSString *g = [NSString stringWithUTF8String:group];
    NSData *data = [NSData dataWithBytes:pathsJSON length:strlen(pathsJSON)];
    NSArray *paths = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
    dispatch_async(dispatch_get_main_queue(), ^{ ShowRing(g, paths, NO); });
}

void MorphoWheelNotify(const char *title, const char *body) {
    NSString *script = [NSString stringWithFormat:@"display notification %@ with title %@",
        [NSString stringWithFormat:@"\"%@\"", [[NSString stringWithUTF8String:body] stringByReplacingOccurrencesOfString:@"\"" withString:@"'"]],
        [NSString stringWithFormat:@"\"%@\"", [NSString stringWithUTF8String:title]]];
    dispatch_async(dispatch_get_global_queue(QOS_CLASS_UTILITY, 0), ^{
        NSTask *t = [NSTask launchedTaskWithLaunchPath:@"/usr/bin/osascript" arguments:@[@"-e", script]];
        [t waitUntilExit];
    });
}

static OSStatus HotkeyHandler(EventHandlerCallRef next, EventRef ev, void *ud) {
    goHotkey();
    return noErr;
}

@implementation AppDelegate

- (void)applicationDidFinishLaunching:(NSNotification *)n {
    gDelegate = self;

    self.view = [[RingView alloc] initWithFrame:NSMakeRect(0, 0, kSize, kSize)];
    [self.view registerForDraggedTypes:@[NSPasteboardTypeFileURL]];
    self.panel = [[RingPanel alloc] initWithContentRect:self.view.frame
        styleMask:NSWindowStyleMaskBorderless | NSWindowStyleMaskNonactivatingPanel
        backing:NSBackingStoreBuffered defer:NO];
    self.panel.opaque = NO;
    self.panel.backgroundColor = NSColor.clearColor;
    self.panel.level = NSStatusWindowLevel;
    self.panel.hasShadow = NO;
    self.panel.acceptsMouseMovedEvents = YES;
    self.panel.collectionBehavior = NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary;
    self.panel.hidesOnDeactivate = NO;
    self.panel.contentView = self.view;

    // Menu bar item.
    self.item = [NSStatusBar.systemStatusBar statusItemWithLength:NSSquareStatusItemLength];
    self.item.button.image = [NSImage imageWithSystemSymbolName:@"circle.hexagongrid.fill" accessibilityDescription:@"MorphoWheel"];
    NSMenu *menu = [[NSMenu alloc] init];
    NSMenuItem *sel = [menu addItemWithTitle:@"Конвертировать выбранное в Finder   ⌃⌥Пробел" action:@selector(fromFinder) keyEquivalent:@""];
    sel.target = self;
    [menu addItem:NSMenuItem.separatorItem];
    [menu addItemWithTitle:@"Выйти" action:@selector(terminate:) keyEquivalent:@"q"];
    self.item.menu = menu;

    // Global hotkey: Control+Option+Space.
    EventTypeSpec spec = {kEventClassKeyboard, kEventHotKeyPressed};
    InstallApplicationEventHandler(&HotkeyHandler, 1, &spec, NULL, NULL);
    EventHotKeyID hid = {'swch', 1};
    EventHotKeyRef ref;
    RegisterEventHotKey(kVK_Space, controlKey | optionKey, hid, GetApplicationEventTarget(), 0, &ref);

    // Drag detection. A new drag session bumps the drag pasteboard's
    // changeCount; the ring appears once a modifier is held during the drag.
    NSPasteboard *dp = [NSPasteboard pasteboardWithName:NSPasteboardNameDrag];
    self.lastDragCount = dp.changeCount;
    __weak AppDelegate *ws = self;
    [NSEvent addGlobalMonitorForEventsMatchingMask:(NSEventMaskLeftMouseDragged | NSEventMaskFlagsChanged)
        handler:^(NSEvent *e) {
            AppDelegate *s = ws;
            if (!s || s.panel.isVisible) return;
            NSString *g = GroupForFlags(NSEvent.modifierFlags);
            if (!g || !(NSEvent.pressedMouseButtons & 1)) return;
            NSPasteboard *p = [NSPasteboard pasteboardWithName:NSPasteboardNameDrag];
            NSArray *files = DraggedFiles();
            if (files.count == 0 || p.changeCount == 0) return;
            ShowRing(g, files, YES);
        }];
}

- (void)fromFinder {
    goHotkey();
}

@end

void MorphoWheelRun(void) {
    @autoreleasepool {
        NSApplication *app = NSApplication.sharedApplication;
        [app setActivationPolicy:NSApplicationActivationPolicyAccessory];
        AppDelegate *d = [[AppDelegate alloc] init];
        app.delegate = d;
        [app run];
    }
}
