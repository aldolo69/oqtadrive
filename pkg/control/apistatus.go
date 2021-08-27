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
	"net/http"

	"github.com/xelalexv/oqtadrive/pkg/daemon"
)

//
func (a *api) status(w http.ResponseWriter, req *http.Request) {

	stat := &Status{Client: a.daemon.GetClient()}
	for drive := 1; drive <= daemon.DriveCount; drive++ {
		stat.Add(a.daemon.GetStatus(drive))
	}

	if wantsJSON(req) {
		sendJSONReply(stat, http.StatusOK, w)
	} else {
		sendReply([]byte(stat.String()), http.StatusOK, w)
	}
}
