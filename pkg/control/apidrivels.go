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

package control

import (
	"fmt"
	"io"
	"net/http"

	"github.com/xelalexv/oqtadrive/pkg/daemon"
)

//
func (a *api) dump(w http.ResponseWriter, req *http.Request) {
	a.driveInfo(w, req, "dump")
}

//
func (a *api) driveList(w http.ResponseWriter, req *http.Request) {
	a.driveInfo(w, req, "ls")
}

//
func (a *api) driveInfo(w http.ResponseWriter, req *http.Request, info string) {

	drive := getDrive(w, req)
	if drive == -1 {
		return
	}

	if a.daemon.GetStatus(drive) == daemon.StatusHardware {
		sendReply([]byte(fmt.Sprintf(
			"hardware drive mapped to slot %d", drive)),
			http.StatusOK, w)
		return
	}

	cart, ok := a.daemon.GetCartridge(drive)

	if !ok {
		handleError(fmt.Errorf("drive %d busy", drive), http.StatusLocked, w)
		return
	}

	if cart == nil {
		handleError(fmt.Errorf("no cartridge in drive %d", drive),
			http.StatusUnprocessableEntity, w)
		return
	}

	defer cart.Unlock()

	read, write := io.Pipe()

	go func() {
		switch info {
		case "dump":
			cart.Emit(write)
		case "ls":
			cart.List(write)
		}
		write.Close()
	}()

	sendStreamReply(read, http.StatusOK, w)
}
