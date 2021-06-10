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
	"io"
	"io/ioutil"
)

//
func NewMap() *Map {

	m := &Map{}
	m.Runner = *NewRunner(
		`map [-a|--address {address}] [-s|--start {first drive} -e|--end {last drive}]
      [-o|--off] [-y|--yes]`,
		"map group of hardware drives",
		`
Use the map command to move a group of hardware drives to the desired place within
the Microdrive daisy chain. Start and end denote the first and last drive of the
hardware drive group. Without any options, the current setting is shown.`,
		"", runnerHelpEpilogue, m.Run)

	m.AddBaseSettings()
	m.AddSetting(&m.Start, "start", "s", "", -1, "first hardware drive", false)
	m.AddSetting(&m.End, "end", "e", "", -1, "last hardware drive", false)
	m.AddSetting(&m.Off, "off", "o", "", false, "turn hardware drives off", false)
	m.AddSetting(&m.Yes, "yes", "y", "", false, "skip confirmation", false)

	return m
}

//
type Map struct {
	Runner
	//
	Start int
	End   int
	Off   bool
	Yes   bool
}

//
func (m *Map) Run() error {

	m.ParseSettings()

	if m.Off {
		m.Start = 0
		m.End = 0
	}

	var resp io.ReadCloser
	var err error

	if m.Start == -1 && m.End == -1 {
		resp, err = m.apiCall("GET", "/map", false, nil)
		fmt.Println()

	} else {
		if m.Start == 0 && m.End == 0 {
			fmt.Println("\nturning hardware drives off")

		} else if !m.Yes && !GetUserConfirmation(fmt.Sprintf(`
changing hardware drives

first drive: %d
last drive:  %d

Note: Specifying the wrong number of hardware drives will cause problems. If
      you set too many, you will block virtual drives, if you set too few,
      the excess hardware drives will conflict with virtual drives, causing
      bus contention. Proceed?`, m.Start, m.End)) {
			return nil
		}

		fmt.Println("\nreconfiguring adapter, this could take a moment...")
		resp, err = m.apiCall("PUT",
			fmt.Sprintf("/map?start=%d&end=%d", m.Start, m.End), false, nil)
	}

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
