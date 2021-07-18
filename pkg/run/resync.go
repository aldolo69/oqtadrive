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
func NewResync() *Resync {

	r := &Resync{}
	r.Runner = *NewRunner(
		`resync [-a|--address {address}] [-c|--client {if1|ql}] [-r|--reset]`,
		"resync with the adapter",
		`
Use the resync command to re-synchronize with the adapter. Optionally, you can force
whether the adapter should be re-configured for Interface 1 or QL during the resync.
Note however that if the adapter is forced to a particular client in its configuration,
then this cannot be changed. Otherwise, if the client is not specified, the adapter
uses auto-detect during resync.

When setting the reset flag, the daemon will try to reset the adapter by closing and
opening the serial connection. This will stop any actions currently performed and may
therefore lead to data loss if a write operation is in progress!`,
		"", runnerHelpEpilogue, r.Run)

	r.AddBaseSettings()
	r.AddSetting(&r.Client, "client", "c", "", nil,
		"client type, 'if1' or 'ql'", false)
	r.AddSetting(&r.Reset, "reset", "r", "", false, "reset adapter", false)

	return r
}

//
type Resync struct {
	Runner
	//
	Client string
	Reset  bool
}

//
func (r *Resync) Run() error {

	r.ParseSettings()

	resp, err := r.apiCall("PUT",
		fmt.Sprintf("/resync?client=%s&reset=%v", r.Client, r.Reset), false, nil)

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
