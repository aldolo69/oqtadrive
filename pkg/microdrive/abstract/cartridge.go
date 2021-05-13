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
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/microdrive/pkg/microdrive"
)

// sector numbers range from 1 through 254 for IF1, 0 through 254 for QL
const SectorCountIF1 = 254
const SectorCountQL = 255

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
func NewCartridge(c microdrive.Client) *Cartridge {

	count := SectorCountIF1
	if c == microdrive.QL {
		count = SectorCountQL
	}

	return &Cartridge{
		sectors:  make([]*Sector, count),
		accessIx: count - 1,
		lock:     make(chan bool, 1),
	}
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
func (c *Cartridge) Name() string {
	return c.name
}

//
func (c *Cartridge) SectorCount() int {
	return len(c.sectors)
}

// SeekToStart sets the access index such that the next call to GetNextSector
// will retrieve the top-most sector, i.e. the sector with the highest sector
// number.
func (c *Cartridge) SeekToStart() {

	if !c.IsFormatted() {
		return
	}

	max := 0
	maxIx := -1

	for ix, sec := range c.sectors {
		if sec != nil && sec.Index() > max {
			max = sec.Index()
			maxIx = ix
		}
	}

	if maxIx > -1 {
		c.accessIx = maxIx
		c.RewindAccessIx(false)
	}
}

// GetNextSector gets the sector at the next access index, skipping slots
// with nil sectors. Access index points to the slot of the returned sector
// afterwards.
func (c *Cartridge) GetNextSector() *Sector {
	return c.getSectorAt(c.AdvanceAccessIx(true))
}

// GetPreviousSector gets the sector at the previous access index, skipping
// slots with nil sectors. Access index points to the slot of the returned
// sector afterwards.
func (c *Cartridge) GetPreviousSector() *Sector {
	return c.getSectorAt(c.RewindAccessIx(true))
}

//
func (c *Cartridge) getSectorAt(ix int) *Sector {
	if 0 <= ix && ix < len(c.sectors) {
		return c.sectors[ix]
	}
	return nil
}

// SetNextSector sets the provided sector at the next access index, whether
// there is a sector present at that index or not. Access index points to the
// slot of the set sector afterwards.
func (c *Cartridge) SetNextSector(s *Sector) {
	c.setSectorAt(c.AdvanceAccessIx(false), s)
}

// SetPreviousSector sets the provided sector at the previous access index,
// whether there is a sector present at that index or not. Access index points
// to the slot of the set sector afterwards.
func (c *Cartridge) SetPreviousSector(s *Sector) {
	c.setSectorAt(c.RewindAccessIx(false), s)
}

// setSector sets the provided sector in this cartridge at the given index.
func (c *Cartridge) setSectorAt(ix int, s *Sector) {
	if 0 <= ix && ix < len(c.sectors) {
		log.Debugf("setting sector at index %d", ix)
		c.sectors[ix] = s
		c.name = s.Name()
		c.modified = true
	} else {
		log.Errorf("trying to set sector at invalid index %d", ix)
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
func (c *Cartridge) AdvanceAccessIx(skipEmpty bool) int {
	return c.moveAccessIx(true, skipEmpty)
}

//
func (c *Cartridge) RewindAccessIx(skipEmpty bool) int {
	return c.moveAccessIx(false, skipEmpty)
}

//
func (c *Cartridge) moveAccessIx(forward, skipEmpty bool) int {

	from := c.accessIx

	if c.IsFormatted() {
		for {
			if forward {
				c.accessIx = c.ensureIx(c.accessIx - 1)
			} else {
				c.accessIx = c.ensureIx(c.accessIx + 1)
			}
			if !skipEmpty || c.sectors[c.accessIx] != nil {
				break
			}
		}
	}

	log.WithFields(
		log.Fields{"from": from, "to": c.accessIx}).Tracef("moving access ix")

	return c.accessIx
}

//
func (c *Cartridge) ensureIx(ix int) int {
	if ix < 0 {
		return c.SectorCount() - 1 - (-(ix + 1))%c.SectorCount()
	}
	return ix % c.SectorCount()
}

//
func (c *Cartridge) Emit(w io.Writer) {
	c.SeekToStart()
	for ix := 0; ix < c.SectorCount(); ix++ {
		sec := c.GetNextSector()
		if sec != nil {
			sec.Emit(w)
		}
	}
}
