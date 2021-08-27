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
)

//
func (a *api) list(w http.ResponseWriter, req *http.Request) {

	list := a.getCartridges()

	if wantsJSON(req) {
		sendJSONReply(list, http.StatusOK, w)

	} else {
		strList := "\nDRIVE CARTRIDGE       STATE"
		for ix, c := range list {
			strList += fmt.Sprintf("\n  %d   %s", ix+1, c.String())
		}
		sendReply([]byte(strList), http.StatusOK, w)
	}
}