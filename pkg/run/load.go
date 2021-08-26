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
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/xelalexv/oqtadrive/pkg/repo"
)

//
func NewLoad() *Load {

	l := &Load{}
	l.Runner = *NewRunner(
		`load [-d|--drive {drive}] -i|--input {file} [-f|--force] [-r|--repair]
       [-a|--address {address}] [-n|--name {cartridge name}]`,
		"load cartridge into daemon",
		"\nUse the load command to load a cartridge into the daemon.",
		"", `- You can directly load Z80 snapshot files into the daemon.

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
	l.AddSetting(&l.Name, "name", "n", "", "",
		"name to give to cartridge when loading a Z80 snapshot", false)

	return l
}

//
type Load struct {
	//
	Runner
	//
	Drive  int
	File   string
	Name   string
	Force  bool
	Repair bool
}

//
func (l *Load) Run() error {

	l.ParseSettings()

	if err := validateDrive(l.Drive); err != nil {
		return err
	}

	var name = l.Name
	if name == "" {
		_, name = filepath.Split(l.File)
		name = strings.TrimSuffix(strings.ToUpper(name), ".Z80")
	}

	path := fmt.Sprintf("/drive/%d?type=%s&force=%v&repair=%v&name=%s",
		l.Drive, getExtension(l.File), l.Force, l.Repair, url.QueryEscape(name))

	var in io.Reader

	if repo.IsReference(l.File) {
		path += fmt.Sprintf("&ref=true")
		in = strings.NewReader(l.File)

	} else {
		f, err := os.Open(l.File)
		if err != nil {
			return err
		}
		defer f.Close()
		in = bufio.NewReader(f)
	}

	resp, err := l.apiCall("PUT", path, false, in)
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
