/*
    This is an example config file for the OqtaDrive adapter. This is optional,
    you can also make all settings in the `oqtadrive.ino` source file. Using
    config files makes uploading easier when you work with several adapters that
    use different settings. You only need to change the include statement in the
    source file in that case.
 */

#ifndef OQTADRIVE_CONFIG

// This is required to disable the config section in oqtadrive.ino
#define OQTADRIVE_CONFIG

// All settings need to be present, no merge with the settings from
// oqtadrive.ino will be done.
#define LED_RW_IDLE_ON    false
#define LED_SYNC_WAIT     true
#define RUMBLE_LEVEL      35
#define DRIVE_OFFSET_IF1  0
#define DRIVE_OFFSET_QL   0
#define HW_GROUP_START    7
#define HW_GROUP_END      8
#define HW_GROUP_LOCK     false
#define FORCE_IF1         false
#define FORCE_QL          true

#endif
