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
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

//
func (a *api) watch(w http.ResponseWriter, req *http.Request) {

	timeout, err := strconv.Atoi(req.URL.Query().Get("timeout"))
	if err != nil || timeout < 0 || 1800 < timeout {
		timeout = 600
	}

	log.Infof("starting watch for %s, timeout %d", req.RemoteAddr, timeout)
	update := make(chan *Change)

	select {
	case a.longPollQueue <- update:
	case <-time.After(time.Duration(timeout) * time.Second):
		log.Infof("closing watch for %s after timeout", req.RemoteAddr)
		sendReply([]byte{}, http.StatusRequestTimeout, w)
		return
	}

	log.Infof("sending daemon change to %s", req.RemoteAddr)
	sendJSONReply(<-update, http.StatusOK, w)
}

//
func (a *api) watchDaemon() {

	log.Info("start watching for daemon changes")

	var client string
	var list []*Cartridge

	for a.server != nil {

		time.Sleep(2 * time.Second)
		change := &Change{}

		l := a.getCartridges()
		if !cartridgeListsEqual(l, list) {
			change.Drives = l
			list = l
		}

		c := a.daemon.GetClient()
		if c != client {
			change.Client = c
			client = c
		}

		if change.Drives == nil && change.Client == "" {
			continue
		}

		log.Info("daemon changes")

	Loop:
		for {
			select {
			case cl := <-a.longPollQueue:
				log.Info("notifying long poll client")
				cl <- change
			default:
				log.Info("all long poll clients notified")
				break Loop
			}
		}
	}

	log.Info("stopped watching for daemon changes")
}
