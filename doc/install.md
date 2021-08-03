# Install Guide (*Linux*)

This guide shows you how to set up *OqtaDrive* using the installation `Makefile`. At the current stage of development, this only supports *Linux*. It *may* work on *MacOS*, and possibly in the *Linux* sub-system on *Windows*, but has not been tested there at all. Contributions in this area are very welcome! For the hardware, we'll be using [Tom Dalby's stand-alone setup](https://tomdalby.com/other/oqtadrive.html), which is based on a *RaspberryPi Zero W*. However, by skipping the steps that are particular for the *RaspberryPI*, the guide can essentially be used for any other setup using *Linux*. This guide has borrowed some input from Tom's guide linked above. Check it out to find out more about fine-tuning the *Pi* OS setup.

## Preparing the *RaspberryPi*

- [Download the *Raspian* image](https://www.raspberrypi.org/software/operating-systems/), pick the *Raspberry Pi OS Lite* version, since you won't be needing a fully-fledged desktop system.

- Use an image writer (e.g. *gnome-disk-utility* on *Ubuntu*) to write the image to a suitable micro SD card.

- *Optional*: Use a partition editor such as *gparted* to increase the `root` partition to full size.

- Create file `ssh` in the `boot` partition on the SD card. This file enables login via *ssh*, and can be empty ([details](https://www.raspberrypi.org/documentation/remote-access/ssh/)).

- For letting the *Pi* access your wireless network, create file `wpa-supplicant.conf`, also in the `boot` partition. The contents of this file [is documented here](https://www.raspberrypi.org/documentation/configuration/wireless/headless.md).

- Place the SD card in the *Pi* and boot it up. Check your wireless router to find out which IP address it received, and log in via ssh, e.g. `ssh pi@192.168.1.12`.

- Change the password with `passwd`. The initial password is `raspberry`.

- Edit `/boot/config.txt` to enable the serial port:

    `sudo nano /boot/config.txt`

    Look for the `[all]` section. If it doesn't exist, create it at the end of the file, then add `enable_uart=1`, so you get something similar to this:

    ```
    [all]
    enable_uart=1
    ```

- Reboot:

    `sudo reboot`

## Running the Installer
All steps in this section are performed on the target system, i.e. the *RaspberryPi* in our example.

### The Short Version
For the impatient, here are all the steps that you would perform on the *Pi* setup after you installed the OS:

```
sudo apt install jq gawk
cd
curl -fsSL https://github.com/xelalexv/oqtadrive/raw/master/hack/Makefile -o Makefile
PORT=/dev/ttyS0 make install patch_avrdude flash service_on
```

### The Long Version
And here's the same with a bit more background information:

- Install the *curl*, *jq*, and *gawk* OS packages, if they're not present. E.g. on *Debian* based systems such as *Raspian*, run:

    `sudo apt install curl jq gawk`

- Create and/or change into the folder where you want to install *OqtaDrive*. For the *Pi* setup, we're using `/home/pi`.

- Download the install `Makefile`:

    `curl -fsSL https://github.com/xelalexv/oqtadrive/raw/master/hack/Makefile -o Makefile`

- Using the `Makefile`, download *OqtaDrive*'s `oqtactl` binary & firmware, and install the [*Arduino CLI*](https://github.com/arduino/arduino-cli) for compiling and flashing it:

    `make install`

    The installation of the *Arduino CLI* can take quite a bit, so some patience is required ;-)

- *Optional*: In our *RaspberryPi* setup, the serial connection between *Arduino Nano* and *Pi* is done via GPIO pins, not USB. This requires applying a small patch to the *avrdude* flash program ([details](https://siytek.com/raspberry-pi-gpio-arduino/)). This only has to be run once: 

    `make patch_avrdude`

- Now we're ready to compile & flash the firmware. Note that for this step it's important to specify the serial port device to which the adapter is connected. For the *Pi*, that's `/dev/ttyS0`. Compiling and flashing can again take quite a bit:

    `PORT=/dev/ttyS0 make flash`

- If you want to automatically start the *OqtaDrive* daemon whenever the system boots, you can enable it as a *systemd* service. This of course only works if your system is using *systemd* as its init system, which is the case with *Raspian* on the *Pi*. Note that also in this step, it's necessary to specify the serial port device:

    `PORT=/dev/ttyS0 make service_on`

    To check on the state of the service, you can run:

    `systemctl status oqtadrive.service`

- To upgrade to the latest *OqtaDrive* release, run:

    `make upgrade`

The installer `Makefile` has a few more targets you can invoke, and environment variables you may specify for configuration. Simply run `make` to get an online help with detailed explanations.

## Hints for *Pi* Setup

- Fix the IP address of the *Pi* in your router, so you don't have to look it up each time you want to `ssh` into it.
- For password-less login, you can also place your public *ssh* key in `/home/pi/.ssh/authorized_keys`.
- You can add the environment variables you want to set such as `PORT` to the `.bashrc` of the `pi` user, so they're already set up when `ssh` into the *Pi*.
