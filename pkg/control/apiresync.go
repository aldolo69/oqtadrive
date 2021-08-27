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

	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
)

//
func (a *api) resync(w http.ResponseWriter, req *http.Request) {

	arg, err := getArg(req, "client")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	var cl client.Client = client.UNKNOWN

	if arg != "" {
		if cl = client.GetClient(arg); cl == client.UNKNOWN {
			handleError(fmt.Errorf("unknown client type: %s", arg),
				http.StatusUnprocessableEntity, w)
			return
		}
	}

	reset := isFlagSet(req, "reset")
	if handleError(
		a.daemon.Resync(cl, reset), http.StatusUnprocessableEntity, w) {
		return
	}

	msg := "re-syncing with adapter"
	if reset {
		msg = "resetting adapter"
	}
	sendReply([]byte(msg), http.StatusOK, w)
}
