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
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/microdrive/pkg/microdrive/abstract"
)

//
func (c *command) put(d *Daemon) error {

	drive, err := c.drive()
	if err != nil {
		return err
	}

	if c.arg(2) != 0 { // ignore canceled PUT
		log.WithFields(
			log.Fields{"drive": drive, "code": c.arg(2)}).Debugf("PUT canceled")
		return nil
	}

	data, err := d.conduit.receiveBlock()
	if err != nil {
		return err
	}

	if len(data) < 200 {
		if hd, err := abstract.NewHeader(d.conduit.client, data, true); err != nil {
			return fmt.Errorf("error creating header: %v", err)
		} else if err = d.mru.setHeader(hd); err != nil {
			return err
		}

	} else {
		if rec, err := abstract.NewRecord(d.conduit.client, data, true); err != nil {
			return fmt.Errorf("error creating record: %v", err)
		} else if err = d.mru.setRecord(rec); err != nil {
			return err
		}

		if d.mru.isRecordUpdate() {
			defer d.mru.reset()
			if cart := d.getCartridge(drive); cart != nil {
				cart.SetModified(true)
				log.WithFields(log.Fields{
					"drive":  drive,
					"sector": d.mru.sector.Index(),
				}).Debugf("PUT record")
			} else {
				return fmt.Errorf("error updating record: no cartridge")
			}
		}
	}

	if d.mru.isNewSector() {
		sec, err := d.mru.createSector()
		if err != nil {
			return fmt.Errorf("error creating sector: %v", err)
		}

		if cart := d.getCartridge(drive); cart != nil {
			cart.SetSector(sec)
			log.WithFields(log.Fields{
				"drive":  drive,
				"sector": sec.Index(),
			}).Debugf("PUT sector complete")
		} else {
			return fmt.Errorf("error creating sector: no cartridge")
		}
	}

	return nil
}
