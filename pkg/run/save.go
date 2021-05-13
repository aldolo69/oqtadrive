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
	"os"
)

//
func NewSave() *Save {

	// FIXME: determine format by what's in the daemon

	s := &Save{}
	s.Runner = *NewRunner(
		"save [-d|--drive {drive}] -o|--output {file} [-f|--force] [-p|--port {port}]",
		"get cartridge from daemon and save",
		"\nUse the save command to get a cartridge from the daemon and save it to a file.",
		"", `- The format for saving the file is determined by the file extensions of the
  given file name. Currently supported formats are .mdr and .mdv

`+runnerHelpEpilogue, s.Run)

	s.AddBaseSettings()
	s.AddSetting(&s.File, "output", "o", "", nil, "cartridge output file", true)
	s.AddSetting(&s.Drive, "drive", "d", "", 1, "drive number (1-8)", false)
	s.AddSetting(&s.Force, "force", "f", "", false,
		"force overwriting output file", false)

	return s
}

//
type Save struct {
	//
	Runner
	//
	File  string
	Drive int
	Force bool
}

//
func (s *Save) Run() error {

	s.ParseSettings()

	if err := validateDrive(s.Drive); err != nil {
		return err
	}

	if !s.Force {
		if _, err := os.Stat(s.File); err == nil &&
			!GetUserConfirmation("File exists, overwrite?") {
			return nil
		}
	}

	resp, err := s.apiCall("GET",
		fmt.Sprintf("/drive/%d?type=%s", s.Drive, getExtension(s.File)),
		false, nil)
	if err != nil {
		return err
	}

	defer resp.Close()

	f, err := os.Create(s.File)
	if err != nil {
		return err
	}
	defer f.Close()

	out := bufio.NewWriter(f)
	defer out.Flush()

	if _, err := io.Copy(out, resp); err != nil {
		return err
	}

	fmt.Println("cartridge saved")
	return nil
}
