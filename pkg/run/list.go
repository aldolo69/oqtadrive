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
)

//
func NewList() *List {

	l := &List{}
	l.Runner = *NewRunner(
		"ls [-p|--port {port}]",
		"get cartridge list from daemon",
		"\nUse the ls command to get a drive list from the daemon.",
		"", runnerHelpEpilogue, l.Run)

	l.AddBaseSettings()

	return l
}

//
type List struct {
	Runner
}

//
func (l *List) Run() error {

	l.ParseSettings()

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
	return nil
}
