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
	"strconv"
)

//
func NewUnload() *Unload {

	u := &Unload{}
	u.Runner = *NewRunner(
		"unload [-d|--drive {drive}] [-f|--force] [-p|--port {port}]",
		"unload cartridge from daemon",
		`
Use the unload command to unload a cartridge from the daemon, replacing
it with a blank, unformatted one`,
		"", runnerHelpEpilogue, u.Run)

	u.AddBaseSettings()
	u.AddSetting(&u.Drive, "drive", "d", "", 1, "drive number (1-8)", false)
	u.AddSetting(&u.Force, "force", "f", "", false,
		"force unloading modified cartridge from daemon", false)

	return u
}

//
type Unload struct {
	//
	Runner
	//
	Drive int
	Force bool
}

//
func (u *Unload) Run() error {

	u.ParseSettings()

	if err := validateDrive(u.Drive); err != nil {
		return err
	}

	resp, err := u.apiCall("GET", fmt.Sprintf("/drive/%d/unload?force=%s",
		u.Drive, strconv.FormatBool(u.Force)), false, nil)
	if err != nil {
		return err
	}
	defer resp.Close()

	msg, err := ioutil.ReadAll(resp)
	if err != nil {
		return err
	}

	fmt.Printf("%s", msg)
	return nil
}
