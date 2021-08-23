# Change Log

## 0.2.0
This release requires re-flashing the adapter. Protocol version remains at `3`.

### Overall
- added installer (for *Linux*)
- bug fixes
- documentation updates

### Adapter
- support for including config files in `oqtadrive.ino`
- support for connecting a *rumble motor* on *Arduino* pin `D10` (see notes in `README.md`)

### Daemon / `oqtactl`
- added initial version of web UI
- added `--reset` flag to `resync` command
- switched to *Go* 1.17

## 0.1.4
This release requires re-flashing the adapter. Protocol version is now at `3`.

### Adapter
- indicate sync wait by alternating read/write LEDs

### Daemon
- added `resync` command
- added `--client` option to `serve` command
- graceful shutdown
- integrated *Z80onMDR* into *OqtaDrive* code base; Big thanks to Tom Dalby for open-sourcing [Z80onMDR Lite](https://github.com/TomDDG/Z80onMDR_lite)! You now no longer need to install *Z80onMDR* on your system, it's available wherever *OqtaDrive* can run.
- less aggressive retry back-off when daemon fails to open serial port; improves boot time in stand-alone setups
- documentation updates

## 0.1.3
This release requires re-flashing the adapter. Protocol version is now at `2`.

### Adapter
- GPIOs are now switched to `INPUT` mode without pull-ups when adapter is inactive, to minimize load on signal lines and that way reduce interference with hardware *Microdrives*.

### Daemon
- freely move hardware *Microdrives* as a group within the drive chain, or turn them off completely; requires adapter to be first in chain; refer to README for more details
- changed `--port` option of `oqtactl` to `--address`, to specify full listen address, i.e. host & port
- bug fix for saving cartridges formatted with earlier *Spectrum* ROMs - instead of dropping sectors with overlong format records, their records are now reduced to standard length when saving to *MDR*
- added *MacOS* *ARM64* build to release binaries
- switched to *Go* 1.16

## 0.1.2
This release contains important bug fixes, so upgrading to this is strongly recommended. It requires re-flashing the adapter. The circuit also has small but important changes.

### Adapter
- circuit update: resistor + diode in `WR.PROTECT` line, diode in `/ERASE` line, 9V to `Vin` via 1A diode
- fixed `WR.PROTECT` output pin control (this potentially put the *Nano* at risk)

### Daemon
- auto-save cartridges in daemon
- support `FORMAT` for *Spectrums* with early ROMs (*Issue 3* and older)
- versioning of serial protocol
- reject loading of corrupted cartridges; added repair flag to `load` action
- offer renaming of `.Z80` files to `.z80` during load on Linux (`.Z80` suffix is not accepted by *Z80onMDR* under Linux)
- added status API
- doc updates & corrections

## 0.1.1

### Adapter
This release does not require re-flashing the adapter. There were no changes to the firmware.

### Daemon
- Support *Spectrum Z80* snapshot files via [*Z80onMDR*](https://www.tomdalby.com/other/z80onmdr.html). You can now directly load *Z80* snapshot files into the daemon. They get converted to *MDR* on the fly by calling *Z80onMDR*. This requires *Z80onMDR* to be installed on your system and set in `PATH`.
- `list` command can now also list the contents of cartridges. Just specify a drive with `-d` or an input file with `-i`.
- Added new `dump` command. This lets you inspect the sectors of a cartridge `hexdump -C` style:
    ```
    $ oqtactl dump -d 1 | more

    HEADER: "INTRO2    " - flag: 21, index: 248
    00000000  00 00 00 00 00 00 00 00  00 00 ff ff 21 f8 69 6e  |............!.in|
    00000010  49 4e 54 52 4f 32 20 20  20 20 32                 |INTRO2    2|

    RECORD: "Database  " - flag: 0, index: 1, length: 512
    00000000  00 00 00 00 00 00 00 00  00 00 ff ff 00 01 00 02  |................|
    00000010  44 61 74 61 62 61 73 65  20 20 5b 75 66 66 65 72  |Database  [uffer|
    00000020  20 77 69 6c 6c 20 62 65  20 73 65 6e 74 2e 20 20  | will be sent.  |
    00000030  20 20 20 20 20 20 20 20  20 20 53 74 72 65 61 6d  |          Stream|
    00000040  73 20 30 2d 33 20 72 65  76 65 72 74 20 74 6f 20  |s 0-3 revert to |
    00000050  74 68 65 69 72 20 20 20  20 20 69 6e 69 74 69 61  |their     initia|
    00000060  6c 20 63 68 61 6e 6e 65  6c 73 20 20 20 20 20 20  |l channels      |
    ...
    ```

- Added *ARM* build
- Refactorings, minor bug fixes, doc updates

## 0.1.0
- First alpha release
