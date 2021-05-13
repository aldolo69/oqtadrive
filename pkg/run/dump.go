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
	"path/filepath"
	"strings"

	"github.com/xelalexv/microdrive/pkg/microdrive/format"
)

//
func NewDump() *Dump {

	d := &Dump{}
	d.Runner = *NewRunner(
		"dump [-d|--drive {drive}] [-i|--input {file}] [-p|--port {port}]",
		"dump cartridge from file or daemon",
		"\nUse the dump command to output a hex dump for a cartridge from file or from daemon.",
		"", runnerHelpEpilogue, d.Run)

	d.AddBaseSettings()
	d.AddSetting(&d.File, "input", "i", "", nil, "cartridge input file", false)
	d.AddSetting(&d.Drive, "drive", "d", "", 1, "drive number (1-8)", false)

	return d
}

//
type Dump struct {
	//
	Runner
	//
	Drive int
	File  string
}

//
func (d *Dump) Run() error {

	d.ParseSettings()

	if d.File != "" {
		f, err := os.Open(d.File)
		if err != nil {
			return err
		}
		defer f.Close()

		ext := strings.TrimPrefix(filepath.Ext(d.File), ".")
		form, err := format.NewFormat(ext)
		if err != nil {
			return err
		}

		cart, err := form.Read(bufio.NewReader(f), false)
		if err != nil {
			return err
		}

		cart.Emit(os.Stdout)

	} else {
		if err := validateDrive(d.Drive); err != nil {
			return err
		}

		resp, err := d.apiCall("GET", fmt.Sprintf("/drive/%d/dump", d.Drive),
			false, nil)
		if err != nil {
			return err
		}
		defer resp.Close()

		if _, err := io.Copy(os.Stdout, resp); err != nil {
			return err
		}
	}

	fmt.Println()
	return nil
}
