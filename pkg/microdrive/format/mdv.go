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
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/microdrive/pkg/microdrive"
	"github.com/xelalexv/microdrive/pkg/microdrive/abstract"
	"github.com/xelalexv/microdrive/pkg/microdrive/ql"
)

// Strangely, a sector in an MDV file is longer than what the QL actually writes
// during a format, which is 652 bytes (cf. Appendix D "Microdrive Format",
// Section 4 "Special Sector Structure" in "QL Advanced User Guide" by Adrian
// Dickens).
const MDVSectorLength = 686

// MDV is a reader/writer for MDV format
// MDV files contain the sectors in reverted replay order.
//
type MDV struct{}

//
func NewMDV() *MDV {
	return &MDV{}
}

func (m *MDV) Read(in io.Reader, strict bool) (*abstract.Cartridge, error) {

	cart := abstract.NewCartridge(microdrive.QL)

	for ix := 0; ix < cart.SectorCount(); ix++ {

		sector := make([]byte, MDVSectorLength)
		read, err := io.ReadFull(in, sector)

		if err != nil {
			if err == io.EOF && read == 0 {
				break
			}
			return nil, fmt.Errorf("error reading MDV file: %v", err)
		}

		hd, err := ql.NewHeader(sector[:ql.HeaderLength], false)
		if err != nil {
			if strict {
				log.Errorf("defective header, discarding sector: %v", err)
				continue
			} else {
				log.Warnf("defective header: %v", err)
			}
		}

		rec, err := ql.NewRecord(sector[ql.HeaderLength:ql.MaxSectorLength], false)
		if err != nil {
			if strict {
				log.Errorf("defective record, discarding sector: %v", err)
				continue
			} else {
				log.Warnf("defective record: %v", err)
			}
		}

		sec, err := abstract.NewSector(hd, rec)
		if err != nil {
			if strict {
				log.Errorf("defective sector, discarding: %v", err)
				continue
			} else {
				log.Warnf("defective sector: %v", err)
			}
		}

		log.Debugf("loaded sector with number %d", sec.Index())
		cart.SetPreviousSector(sec)

		if log.IsLevelEnabled(log.TraceLevel) {
			sec.Emit(os.Stdout)
		}
	}

	cart.SetWriteProtected(false)
	cart.SetModified(false)

	return cart, nil
}

//
func (m *MDV) Write(cart *abstract.Cartridge, out io.Writer) error {

	padding := make([]byte, 256)
	for ix := range padding {
		padding[ix] = 0x5a
	}

	cart.SeekToStart()
	cart.AdvanceAccessIx(false)

	for ix := 0; ix < cart.SectorCount(); ix++ {

		sec := cart.GetPreviousSector()

		if sec == nil { // MDV requires all sectors; FIXME: add blank one
			return fmt.Errorf("missing sector %d", ix)
		}

		missing := MDVSectorLength
		var written int
		var err error

		if written, err = out.Write(sec.Header().Demuxed()); err != nil {
			return err
		}
		missing -= written

		if written, err = out.Write(sec.Record().Demuxed()); err != nil {
			return err
		}
		missing -= written

		if missing > len(padding) {
			return fmt.Errorf("excessive padding, missing %d bytes", missing)
		}

		if missing > 0 {
			if _, err := out.Write(padding[0:missing]); err != nil {
				return err
			}
		}
	}

	return nil
}
