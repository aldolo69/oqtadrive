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
	log "github.com/sirupsen/logrus"
)

//
func (c *command) driveMap(d *Daemon) error {

	d.conduit.hwGroupStart = int(c.arg(0))
	d.conduit.hwGroupEnd = int(c.arg(1))
	d.conduit.hwGroupLocked = c.arg(2) == 1

	log.WithFields(log.Fields{
		"start":  c.arg(0),
		"end":    c.arg(1),
		"locked": c.arg(2) == 1}).Info("MAP")

	return nil
}
