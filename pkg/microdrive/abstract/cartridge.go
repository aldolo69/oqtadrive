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

package abstract

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// sector numbers range from 1 through 254 for IF1, 0 through 254 for QL
const SectorCount = 255

//
type Cartridge struct {
	//
	name           string
	writeProtected bool
	//
	sectors  []*Sector
	accessIx int
	modified bool
	//
	lock chan bool
}

//
func NewCartridge() *Cartridge {
	return &Cartridge{
		sectors:  make([]*Sector, SectorCount),
		accessIx: SectorCount - 1,
		lock:     make(chan bool, 1),
	}
}

//
func (c *Cartridge) Name() string {
	return c.name
}

//
func (c *Cartridge) Lock(ctx context.Context) bool {
	select {
	case c.lock <- true:
		log.Debug("cartridge locked")
		return true
	case <-ctx.Done():
		log.Debug("cartridge lock timed out")
		return false
	}
}

//
func (c *Cartridge) Unlock() {
	select {
	case <-c.lock:
		log.Debug("cartridge unlocked")
	default:
		log.Debug("cartridge was already unlocked")
	}
}

//
func (c *Cartridge) SetSector(s *Sector) {
	c.SetSectorAt(s.Index(), s)
}

//
func (c *Cartridge) SetSectorAt(ix int, s *Sector) {
	if 0 <= ix && ix < len(c.sectors) {
		c.sectors[ix] = s
		c.name = s.Name()
		c.modified = true
	}
}

//
func (c *Cartridge) IsFormatted() bool {
	for _, s := range c.sectors {
		if s != nil {
			return true
		}
	}
	return false
}

//
func (c *Cartridge) IsWriteProtected() bool {
	return c.writeProtected
}

//
func (c *Cartridge) SetWriteProtected(p bool) {
	c.writeProtected = p
}

//
func (c *Cartridge) IsModified() bool {
	return c.modified
}

//
func (c *Cartridge) SetModified(m bool) {
	c.modified = m
}

//
func (c *Cartridge) GetSector() *Sector {
	return c.sectors[c.advanceAccessIx()]
}

//
func (c *Cartridge) GetSectorAt(ix int) *Sector {
	if 0 <= ix && ix < len(c.sectors) {
		return c.sectors[ix]
	}
	return nil
}

//
func (c *Cartridge) advanceAccessIx() int {
	ret := c.accessIx
	if c.IsFormatted() {
		for {
			c.accessIx--
			if c.accessIx < 0 {
				c.accessIx = SectorCount - 1
			}
			if c.sectors[c.accessIx] != nil {
				break
			}
		}
	}
	return ret
}
