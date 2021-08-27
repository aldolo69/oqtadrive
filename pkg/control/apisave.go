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
	"bytes"
	"fmt"
	"net/http"
)

//
func (a *api) save(w http.ResponseWriter, req *http.Request) {

	drive := getDrive(w, req)
	if drive == -1 {
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

	writer := getFormat(w, req)
	if writer == nil {
		return
	}

	var out bytes.Buffer
	if handleError(
		writer.Write(cart, &out, nil), http.StatusInternalServerError, w) {
		return
	}

	cart.SetModified(false)
	w.WriteHeader(http.StatusOK)
	w.Write(out.Bytes())
}
