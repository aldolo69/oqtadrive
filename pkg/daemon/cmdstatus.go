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
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/format/helper"
)

//
const flagLoaded = 1
const flagFormated = 2
const flagReadonly = 4

//
func (c *command) status(d *Daemon) error {

	var cart base.Cartridge
	var state byte = 0x80

	drive, err := c.drive()
	msg := ""

	if err == nil {
		cart = d.getCartridge(drive)
		state = 0x00

		if cart == nil {
			msg = "empty"

		} else {
			if cart.IsFormatted() {
				msg = "formatted"
				state = flagLoaded | flagFormated

			} else {
				msg = "blank"
				state = flagLoaded
			}

			if cart.IsWriteProtected() {
				msg += ", write protected"
				state |= flagReadonly
			}
		}
	}

	action := "stopped"
	if c.arg(1) == 1 {
		action = "started"
	}

	d.mru.reset()

	log.WithFields(log.Fields{
		"drive": drive, "action": action, "state": msg}).Infof("STATUS")

	if c.arg(1) == 1 { // drive started, send cartridge state to adapter
		d.conduit.send([]byte{state})
		if cart != nil {
			ctx, cancel := context.WithTimeout(
				context.Background(), 5*time.Millisecond)
			defer cancel()
			if !cart.Lock(ctx) {
				return fmt.Errorf("could not lock cartridge in drive %d", drive)
			}
		}
	} else if cart != nil {
		if err := helper.AutoSave(drive, cart); err != nil {
			log.Errorf("auto-saving drive %d failed: %v", drive, err)
		}
		cart.Unlock()
	}

	return err
}
