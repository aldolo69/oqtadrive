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
	"runtime"
	"strings"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/format/helper"
)

//
func NewLoad() *Load {

	l := &Load{}
	l.Runner = *NewRunner(
		"load [-d|--drive {drive}] -i|--input {file} [-f|--force] [-r|--repair] [-p|--port {port}]",
		"load cartridge into daemon",
		"\nUse the load command to load a cartridge into the daemon.",
		"", `- If you have Z80onMDR installed on your system and added to PATH, you can
  directly load Z80 snapshot files into the daemon.

- Repair currently only recalculates checksums and reverts sector order, if needed.
  If the cartridge is really broken, it won't be fixed this way.

`+runnerHelpEpilogue, l.Run)

	l.AddBaseSettings()
	l.AddSetting(&l.File, "input", "i", "", nil, "cartridge input file", true)
	l.AddSetting(&l.Drive, "drive", "d", "", 1, "drive number (1-8)", false)
	l.AddSetting(&l.Force, "force", "f", "", false,
		"force replacing modified cartridge in daemon", false)
	l.AddSetting(&l.Repair, "repair", "r", "", false,
		"try to repair cartridge if corrupted", false)

	return l
}

//
type Load struct {
	//
	Runner
	//
	Drive  int
	File   string
	Force  bool
	Repair bool
}

//
func (l *Load) Run() error {

	l.ParseSettings()

	if err := validateDrive(l.Drive); err != nil {
		return err
	}

	if trapped, err := l.trapZ80(); err != nil {
		return err
	} else if trapped {
		defer os.Remove(l.File)
	}

	f, err := os.Open(l.File)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := l.apiCall("PUT",
		fmt.Sprintf("/drive/%d?type=%s&force=%v&repair=%v",
			l.Drive, getExtension(l.File), l.Force, l.Repair),
		false, bufio.NewReader(f))
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

//
func (l *Load) trapZ80() (bool, error) {

	ext := getExtension(l.File)

	if strings.ToLower(ext) != "z80" {
		return false, nil
	}

	if runtime.GOOS == "linux" && ext == "Z80" {
		if GetUserConfirmation(
			"Z80onMDR under Linux doesn't accept uppercase '.Z80' extension. Rename to '*.z80'?") {
			newPath := strings.TrimSuffix(l.File, ".Z80") + ".z80"
			if err := os.Rename(l.File, newPath); err != nil {
				return false, err
			}
			l.File = newPath
		}
	}

	if mdr, err := helper.Z80toMDR(l.File); err != nil {
		return false, fmt.Errorf("error converting Z80 file: %v", err)
	} else {
		l.File = mdr
	}

	return true, nil
}
