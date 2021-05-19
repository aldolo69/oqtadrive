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

package format

import (
	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
)

//
func RepairOrder(cart base.Cartridge) {

	if cart == nil {
		return
	}

	cmp := 0

	cart.SeekToStart()
	var last base.Sector

	for ix := 0; ix < cart.SectorCount(); ix++ {
		sec := cart.GetNextSector()
		if sec == nil {
			continue
		}
		if last != nil {
			if sec.Index() > last.Index() {
				cmp++
			}
			if sec.Index() < last.Index() {
				cmp--
			}
		}
		last = sec
	}

	if cmp < 0 {
		return
	}

	log.Debug("reverting sector order")
	cart.Revert()
}
