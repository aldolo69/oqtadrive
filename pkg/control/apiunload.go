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
	"net/http"
	"strings"
)

//
func (a *api) unload(w http.ResponseWriter, req *http.Request) {

	drive := getDrive(w, req)
	if drive == -1 {
		return
	}

	if err := a.daemon.UnloadCartridge(drive, isFlagSet(req, "force")); err != nil {
		if strings.Contains(err.Error(), "could not lock") {
			handleError(fmt.Errorf("drive %d busy", drive), http.StatusLocked, w)
		} else if strings.Contains(err.Error(), "is modified") {
			handleError(fmt.Errorf(
				"cartridge in drive %d is modified", drive), http.StatusConflict, w)
		} else {
			handleError(err, http.StatusInternalServerError, w)
		}

	} else {
		sendReply([]byte(
			fmt.Sprintf("unloaded drive %d", drive)), http.StatusOK, w)
	}
}
