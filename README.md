# OqtaDrive

#### *Sinclair Microdrive* emulator for *Spectrum* & *QL*

## TL;DR
*OqtaDrive* emulates a bank of up to 8 *Microdrives* for use with a *Sinclair Spectrum* (with *Interface 1*) or *QL* machine. The goal is to functionally create a *faithful* reproduction of the original. That is, on the *Spectrum*/*QL* side, operating the emulated *Microdrives* should feel exactly the same as using the real thing.

*OqtaDrive* is built around an *Arduino Nano* that connects via its GPIO ports to the *Microdrive* interface and via serial connection to a daemon running on a host machine. This daemon host could be anything, ranging from your PC to a small embedded board such as a *RaspberryPi Zero*, as long as it can run a supported OS (*Linux*, *MacOS*, *Windows*). The same *Nano* can be used with both *Spectrum* and *QL*, without any reconfiguration. While the *Nano* is essentially a low-level protocol converter, the daemon takes care of storing and managing the *cartridges*. It additionally exposes an HTTP API endpoint. A few shell commands are provided that use this API and let you control the daemon, e.g. load and save cartridges into/from the virtual drives. A web UI is currently being developed.

## What Can I Do With This?
*OqtaDrive*'s architecture makes it very flexible, so many setups are possible. The simplest one would be just the *Nano* that connects your *Interface 1* or *QL* with your PC, and you manage everything from there. This configuration also [fits nicely into the case of an *Interface 1*](https://github.com/xelalexv/oqtadrive/discussions/15). If you're rather looking for a stand-alone solution, you could for example run the daemon on a *RaspberryPi Zero W*, [put it on a PCB together with the *Nano*](https://tomdalby.com/other/oqtadrive.html), and place this into a *Microdrive* or 3D printed case. The *Pi* would connect to your WiFi and you can control *OqtaDrive* from anywhere in your network.

Due to the minimal hardware required, *OqtaDrive* is also very cost-efficient. In the simplest setup, you only need an *Arduino Nano* and a few resistors and diodes. Additionally, if you own a *Spectrum* and a *QL*, you can use it with both, no need to have dedicated adapters. But above all, there's the fun involved in building this! If you've created your very own setup, then just head over to the discussion section and tell us about it!

### Features
- Supports all *Microdrive* operations on *Spectrum* with *Interface 1* and on *QL*, no modifications or additional software required
- Can co-exist with actual hardware *Microdrive* units, which can be mapped on demand to any slot in the drive chain or turned off
- Daemon can run on *Linux*, *MacOS*, and *Windows* (more community testing for the latter two needed!)
- Load & save from/to *MDR* and *MDV* formatted cartridge files
- For *Spectrum*, *Z80* snapshot files can be directly loaded, no additional software required. Big thanks to Tom Dalby for open-sourcing [Z80onMDR Lite](https://github.com/TomDDG/Z80onMDR_lite)!
- List virtual drives & contents of cartridges
- Hex dump cartridge contents for inspection

Here's a short [demo video](https://www.babbletower.net/forums/spectrum/microdrive/oqtadrive-demo.mp4) showing *OqtaDrive* & a *Spectrum* in action, doing a *Microdrive* test with the original *Sinclair* demo cartridge image, and a cartridge format.

## Warning & Disclaimer
If you want to build *OqtaDrive* yourself, please carefully read the hardware section below! It contains important instructions & notes. Not following these may break your vintage machine and/or the *Nano*! However, bear in mind that all the information in this project is published in good faith and for general information purpose only. I do not make any warranties about the completeness, reliability, and accuracy of this information. Any action you take upon the information you find here, is strictly at your own risk. I will not be liable for any losses and/or damages in connection with the use of *OqtaDrive*. 

## Status
*OqtaDrive* is currently in *alpha* stage, and under active development. Things may still get reworked quite considerably, which may introduce breaking changes.

### Caveats & Current Limitations

- Drive offset detection is only available for the *QL*. If you find that this is not working reliably, you can set a fixed value, i.e. `2` if the two internal drives on the *QL* are present. Have a look at the top of `oqtadrive.ino`. For the *Spectrum* it's technically not possible to offer offset auto detection, and it defaults to `0`. If you want to use an actual *Microdrive* between *Interface 1* and the adapter, you need to set that.

- When running more than one daemon under the same user, they will use the same auto-save directory and hence mutually overwrite auto-save states. If you need to run several instances, run them with different users. A better solution will be provided in the future.

## Hardware

### Circuit
![OqtaDrive](doc/schematic.png)

The circuit is straightforward. You only need to connect a few of the *Nano*'s GPIO pins to an edge connector plug, program `arduino/oqtadrive.ino` onto the board, and you're all set. Here are a few things to consider though, when building the adapter:

- The notch in the edge connector counts as pins 3A/3B.

- Keep the lines between the board and the plug as short as possible, to avoid interference. For the prototypes I built, I used ribbon cable no longer than 5 cm, which works well.

- The resistors in the data lines (`DATA1` & `DATA2`) and `WR.PROTECT` are not strictly required, the original *Microdrives* don't have them. I still recommend using them, since they will limit the current that can flow should there ever be a conflict between these outputs and the *Interface 1*, *QL*, or other *Microdrives*.

- The switching diodes (1N4148 or similar) in the `WR.PROTECT` and `/ERASE` lines are strictly required when using the adapter together with actual *Microdrives*. It protects the according GPIOs `D6` and `D5` on the *Nano* from over-voltage coming from the drives, and prevents `D5` from activating the erase head in an actual *Microdrive* unit when it is running.

- `COMMS_OUT` is only used when you want to daisy chain actual *Microdrives* behind the *OqtaDrive* adapter, instead of having the adapter at the end of the chain (see below for more details). `D7` needs to be connected to `COMMS_IN` of the first hardware drive in this case. By doing this, you can freely move the hardware drives as a group to wherever you need them in the chain, or turn them off completely.

- Connecting the 9V to `Vin` on the *Nano* is while not strictly required, still recommended. Without this, the *Nano* is only powered when connected to USB. If it's disconnected and the *Spectrum* or *QL* is powered on, current will be injected into the *Nano* via its GPIO pins. This may be outside the spec of the micro-controller on the *Nano*. So to be on the safe side, connect it, but don't skip the diode in that case! Any 1A diode such as a 1N4002 will do.

- You may also connect two LEDs for indicating read & write activity to pins `D12` and `D11`, respectively (don't forget resistors). By default, the LEDs are on during idle and start blinking during activity. If you want them to be off during idle, set `LED_RW_IDLE_ON` to `false` in `oqtadrive.ino`.

- In addition to the LEDs, you can also connect a small vibration motor, such as an *Adafruit 1201* to `D10`, to get a nice mechanical sound whenever a drive is active. Adds to the atmosphere ;-) However, you need a transistor to drive the motor, it's not advisable to connect it directly (see for example [here](http://learningaboutelectronics.com/Articles/Vibration-motor-circuit.php)). Pin `D10` is operated in *PWM* mode, to control the vibration level. With setting `RUMBLE_LEVEL` in `oqtadrive.ino` you can choose the default level after power on, and with `oqtactl config --rumble`, change this when the adapter is running.

- When designing a case for the adapter that should work with *Spectrum* and *QL*, keep in mind that on the *QL*, the edge connector is on the right hand side of the unit, while it is on the left for the *Interface 1*.

**My overall recommendation: Build the adapter as shown in the schematic above to minimize the risk of damaging your vintage machine!**

### Differences in Connector Pin-Outs
The pin-outs of the *Interface 1* and *QL* edge connectors are identical, so you can use the adapter with both. **Note however that the outgoing connector of a *Spectrum Microdrive* unit is different!** It is in fact upside down. That's why the cable for connecting a *Microdrive* unit to the *Interface 1* cannot be used to connect (i.e. daisy chain) two *Microdrive* units. If you want to use the adapter behind a *Microdrive* unit, you either need to wire it accordingly, or use an appropriate plug converter. Whichever you choose, fabricate it in a way that makes it mechanically impossible to accidentally plug it into an *Interface 1* or *QL*. **There will be damage otherwise!**

This table shows the respective pin-outs (A = component side, B = solder side):

| Pin | *Interface 1*, *QL* | *Microdrive* unit |
|-----|---------------------|-------------------|
| 1A  | `DATA1`             | `DATA2`           |
| 1B  | `DATA2`             | `DATA1`           |
| 2A  | `COMM CLK`          | `WR.PROTECT`      |
| 2B  | `WR.PROTECT`        | `COMM CLK`        |
| 3A  | (notch)             | (notch)           |
| 3B  | (notch)             | (notch)           |
| 4A  | `COMM`              | 9V                |
| 4B  | 9V                  | `COMM`            |
| 5A  | `/ERASE`            | `R/WR`            |
| 5B  | `R/WR`              | `/ERASE`          |
| 6A  | GND                 | GND               |
| 6B  | GND                 | GND               |
| 7A  | GND                 | GND               |
| 7B  | GND                 | GND               |
| 8A  | GND                 | GND               |
| 8B  | GND                 | GND               |

### Configuration
The adapter recognizes what it's plugged in to, i.e. *Interface 1* or *QL*. But it's also possible to force a particular machine. Have a look at the config section at the top of `oqtadrive.ino`. There are a few more settings that can be changed. If you need to maintain several configs for various adapters, you can alternatively place your settings in separate header files. The details for this are also explained in the config section. 

*Hint*: After turning on the *Spectrum*, the adapter sometimes erroneously detects the *Interface 1* as a *QL*. In that case, run `CAT 1` on the *Spectrum* and reset the adapter afterwards. That should fix the problem.

### Combination with Hardware *Microdrive* Units
If you're planning to use *OqtaDrive* together with actual hardware *Microdrive* units, then there are essentially two choices for placing the *OqtaDrive* adapter - either at the end of the drive chain or at the start. Here are a few considerations and pros & cons for both options.

#### *Last in Chain*

Pros:

- simple - adapter just plugs into the *Interface 1*, *QL*, or *Microdrive* unit edge connector
- requires just one edge connector plug
- no hardware modifications needed

Cons:

- hardware *Microdrive* units are always upstream of the adapter, and cannot be turned off or mapped into different slots

#### *First in Chain*

Pros:

- hardware *Microdrive* units can be freely moved as a group within the chain, or turned off completely

    **Note**: To take advantage of drive mapping, you need to route the `COMMS_OUT` signal to the first hardware drive (see above) and make a couple of settings in the config section at the top of `arduino/oqtadrive.ino`.

Cons:

- requires an additional edge connector (plug) for connecting hardware *Microdrive* units; alternatively, the adapter can be installed into an *Interface 1* or *QL*, but cannot be used with other machines in that case


### Using a Different *Arduino* Board
I haven't tried this out on anything other than a *Nano* (or compatible) board. It may work on other *Arduino* boards, but only if they use the same micro-controller running at the same clock speed. There are timing-sensitive sections in the code that would otherwise require tweaking. Also, stick to the GPIO pin assignments, the code relies on this.

## Installing
After building the adapter, the software needs to be installed. This comprises two separate tasks:

- Flashing the firmware onto the *Arduino Nano* - For this you can use for example the [*Arduino* IDE](https://www.arduino.cc/en/software).

- Copying the `oqtactl` binary - This is a single binary, which takes care of everything that needs to be done on the daemon host side. It can also be used to control the daemon, on the same host or over the network. In the *release* section of this project, there are binaries for *Linux*, *MacOS* and *Windows*, available for different architectures. Download, extract, and copy the appropriate binary onto the daemon host and any other system from which you want to use it. 

### Installation Script
In the `hack` folder, there's a `Makefile` that can be used to perform all of above steps, so you don't even have to install the *Arduino* IDE. Currently, this only supports installation on *Linux*. For more details, have a look at the [install guide](doc/install.md).

## Running
The `oqtactl` binary can run the daemon as well as several control actions. Just run `oqtactl -h` to get a list of the available actions, and `oqtactl {action} -h` for finding out more about a particular action.

### Daemon
Start the daemon with `oqtactl serve -d {serial device}`. It will look for the adapter at the specified serial port, and keep retrying if it's not yet present. You can also dis- and re-connect the adapter. The daemon should re-sync after a few seconds.

#### Cartridge Auto-Save
When a cartridge gets modified it is auto-saved as soon as the virtual drive in which it is located stops. It is also auto-saved when it is initially loaded into the drive. Whenever the daemon is restarted, the previously loaded cartridges are automatically reloaded from auto-saved state and are immediately available for use. Keep in mind however that auto-save does not write back to the file from which a cartridge was originally loaded. This is because the daemon is not aware of that location, and would possibly not even be able to reach it (you can load cartridges via network). Auto-saved states are instead located in `.oqtadrive` within the home directory of the user running the daemon (exact location depends on used OS). It is up to the user to decide whether and where a modified cartridge should be saved (see `save` action below).

#### Logging
Daemon logging behavior can be changed with these environment variables:

| variable     | function   | values                                            |
|--------------|------------|---------------------------------------------------|
| `LOG_LEVEL`  | log level; defaults to `info` | `fatal`, `error`, `warn`, `info`, `debug`, `trace`|
| `LOG_FORMAT` | log format; gets automatically switched to *JSON* when running without a TTY | `json` to force *JSON* log format, `text` to force text output |
| `LOG_FORCE_COLORS` | force colored log messages when running with a TTY | `true`, `false` |
| `LOG_METHODS` | include method names in log messages | `true`, `false` |

### Control Actions
The daemon also serves an HTTP control API on port `8888` (can be changed with `--address` option). This is the integration point for any tooling that may evolve in the future, e.g. a browser-based GUI. It is also used by the provided command line actions. The most important ones are:

- load cartridge: `oqtactl load -d {drive} -i {file}`
- save cartridge: `oqtactl save -d {drive} -o {file}`
- list drives: `oqtactl ls`
- list cartridge content: `oqtactl ls -d {drive}` or `oqtactl ls -i {file}`

`load` & `save` currently support `.mdr` and `.mdv` formatted files. I've only tested loading a very limited number of cartridge files available out there though, so there may be surprises. For the *Spectrum* `load` can also load *Z80* snapshot files into the daemon, converting them to *MDR* on the fly.

## Building
On *Linux* you can use the `Makefile` to build `oqtactl`, the *OqtaDrive* binary. Note that for consistency, building is done inside a *Golang* build container, so you will need *Docker* to build, but no other dependencies. Just run `make build`. You can also cross-compile for *MacOS* and *Windows*. Run `CROSS=y make build` in that case. If you want to build on *MacOS* or *Windows* directly, you would have to install the *Golang* SDK there and run the proper `go build` command manually. 

## Resources
- [Spectrum Microdrive Book](https://worldofspectrum.org/archive/books/spectrum-microdrive-book) by Ian Logan
- [QL Advanced User Guide](https://worldofspectrum.org/archive/books/ql-advanced-user-guide) by Adrian Dickens
