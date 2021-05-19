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
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/format"
)

//
func NewList() *List {

	l := &List{}
	l.Runner = *NewRunner(
		"ls [-p|--port {port}] [-d|--drive {drive}] [-i|--input {file}]",
		"get cartridge list from daemon",
		`
Use the ls command to get a drive list from the daemon. If a drive number or input
file is given, the contents of that cartridge is listed.`,
		"", runnerHelpEpilogue, l.Run)

	l.AddBaseSettings()
	l.AddSetting(&l.File, "input", "i", "", nil, "cartridge file", false)
	l.AddSetting(&l.Drive, "drive", "d", "", 0, "drive number (1-8)", false)

	return l
}

//
type List struct {
	Runner
	//
	Drive int
	File  string
}

//
func (l *List) Run() error {

	l.ParseSettings()

	if l.File != "" {
		f, err := os.Open(l.File)
		if err != nil {
			return err
		}
		defer f.Close()

		form, err := format.NewFormat(getExtension(l.File))
		if err != nil {
			return err
		}

		cart, err := form.Read(bufio.NewReader(f), true, false)
		if err != nil {
			return err
		}

		cart.List(os.Stdout)

	} else if l.Drive > 0 {
		if err := validateDrive(l.Drive); err != nil {
			return err
		}

		resp, err := l.apiCall("GET", fmt.Sprintf("/drive/%d/list", l.Drive),
			false, nil)
		if err != nil {
			return err
		}
		defer resp.Close()

		if _, err := io.Copy(os.Stdout, resp); err != nil {
			return err
		}

	} else {
		resp, err := l.apiCall("GET", "/list", false, nil)
		if err != nil {
			return err
		}
		defer resp.Close()

		list, err := ioutil.ReadAll(resp)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", list)
	}

	return nil
}
