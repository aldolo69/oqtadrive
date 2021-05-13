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

package daemon

import (
	"time"

	log "github.com/sirupsen/logrus"
)

//
func (c *command) get(d *Daemon) error {

	drive, err := c.drive()
	if err != nil {
		return err
	}

	if cart := d.getCartridge(drive); cart != nil {

		sec := cart.GetNextSector()
		if err := d.mru.setSector(sec); err != nil {
			return err
		}

		if sec != nil {
			toSend := d.conduit.fillBlock(sec)

			log.WithFields(log.Fields{
				"drive":  drive,
				"sector": sec.Index(),
			}).Debugf("GET")

			d.debugStart = time.Now()
			d.conduit.send([]byte{byte(toSend), byte(toSend >> 8)})

			return d.conduit.sendBlock(toSend)
		}
	}

	log.WithFields(log.Fields{"drive": drive, "sector": "(nil)"}).Debugf("GET")
	return d.conduit.send([]byte{0, 0})
}
