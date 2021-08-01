/*
   OqtaDrive - Sinclair Microdrive emulator
   Copyright (c) 2021, Alexander Vollschwitz

   This file is part of OqtaDrive.

   OqtaDrive is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   OqtaDrive is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with OqtaDrive. If not, see <http://www.gnu.org/licenses/>.
*/

package run

import (
	"fmt"
	"io/ioutil"

	"github.com/xelalexv/oqtadrive/pkg/daemon"
)

//
func NewConfig() *Config {

	c := &Config{}
	c.Runner = *NewRunner(
		"config [-a|--address {address}] [-r|--rumble {level}]",
		"change configuration of daemon & adapter",
		`
Use the config command to change settings in the daemon and/or adapter. Currently,
these configuration changes are not persisted, and will be reverted once the daemon
or adapter restarts.`,
		"", runnerHelpEpilogue, c.Run)

	c.AddBaseSettings()
	c.AddSetting(&c.Rumble, "rumble", "r", "", -1, "rumble level (0-255)", false)

	return c
}

//
type Config struct {
	Runner
	//
	Rumble int
}

//
func (c *Config) Run() error {

	c.ParseSettings()

	if c.Rumble == -1 {
		fmt.Println("\nnothing to configure")
		return nil
	}

	resp, err := c.apiCall("PUT",
		fmt.Sprintf("/config?item=%s&arg1=%d",
			daemon.CmdConfigItemRumble, byte(c.Rumble)),
		false, nil)

	if err != nil {
		return err
	}
	defer resp.Close()

	msg, err := ioutil.ReadAll(resp)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", msg)
	return nil
}
