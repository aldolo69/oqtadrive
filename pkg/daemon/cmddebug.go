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
func (c *command) debug(d *Daemon) error {

	now := time.Now()

	log.Debugf("%c%c %3d  [ %08b ] - %v",
		c.arg(0), c.arg(1), c.arg(2), c.arg(2), now.Sub(d.debugStart))
	d.debugStart = now

	return nil
}

//
func (c *command) timer(start bool, d *Daemon) error {
	if start {
		d.debugStart = time.Now()
	} else {
		log.Debugf("%v", time.Now().Sub(d.debugStart))
	}
	return nil
}
