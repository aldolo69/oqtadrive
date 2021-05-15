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
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/format/helper"
)

//
func NewLoad() *Load {

	l := &Load{}
	l.Runner = *NewRunner(
		"load [-d|--drive {drive}] -i|--input {file} [-f|--force] [-p|--port {port}]",
		"load cartridge into daemon",
		"\nUse the load command to load a cartridge into the daemon.",
		"", `- If you have Z80onMDR installed on your system and added to PATH, you can
  directly load Z80 snapshot files into the daemon.

`+runnerHelpEpilogue, l.Run)

	l.AddBaseSettings()
	l.AddSetting(&l.File, "input", "i", "", nil, "cartridge input file", true)
	l.AddSetting(&l.Drive, "drive", "d", "", 1, "drive number (1-8)", false)
	l.AddSetting(&l.Force, "force", "f", "", false,
		"force replacing modified cartridge in daemon", false)

	return l
}

//
type Load struct {
	//
	Runner
	//
	Drive int
	File  string
	Force bool
}

//
func (l *Load) Run() error {

	l.ParseSettings()

	if err := validateDrive(l.Drive); err != nil {
		return err
	}

	ext := getExtension(l.File)
	in := l.File
	var err error

	if strings.ToLower(ext) == "z80" {
		in, err = helper.Z80toMDR(l.File)
		if err != nil {
			return fmt.Errorf("error converting Z80 file: %v", err)
		}
		defer os.Remove(in)
		ext = getExtension(in)
	}

	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := l.apiCall("PUT", fmt.Sprintf("/drive/%d?type=%s&force=%s",
		l.Drive, ext, strconv.FormatBool(l.Force)), false, bufio.NewReader(f))
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
