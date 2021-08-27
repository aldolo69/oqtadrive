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
)

//
func (a *api) config(w http.ResponseWriter, req *http.Request) {

	item, err := getArg(req, "item")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	arg1, err := getIntArg(req, "arg1")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	arg2, err := getIntArg(req, "arg2")
	if err != nil {
		arg2 = 0
	}

	if handleError(
		a.daemon.Configure(item, byte(arg1), byte(arg2)),
		http.StatusUnprocessableEntity, w) {
		return
	}

	sendReply([]byte("configuring"), http.StatusOK, w)
}
