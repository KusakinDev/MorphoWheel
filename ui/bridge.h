#ifndef MORPHOWHEEL_BRIDGE_H
#define MORPHOWHEEL_BRIDGE_H

// Starts NSApplication as a menu bar accessory and blocks until quit.
void MorphoWheelRun(void);

// Shows the ring at the mouse cursor for the given files (JSON array of
// paths). group is "convert", "tools" or "actions". Safe from any thread.
void MorphoWheelShowRing(const char *group, const char *pathsJSON);

// Shows a user notification banner. Safe from any thread.
void MorphoWheelNotify(const char *title, const char *body);

#endif
