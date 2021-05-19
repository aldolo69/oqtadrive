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

	"github.com/xelalexv/oqtadrive/pkg/microdrive"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/if1"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/raw"
)

// MDR is a reader/writer for MDR format
// MDR files contain the sectors in replay order.
//
type MDR struct{}

//
func NewMDR() *MDR {
	return &MDR{}
}

//
func (m *MDR) Read(in io.Reader, strict, repair bool) (base.Cartridge, error) {

	cart := if1.NewCartridge()
	r := 0

	// TODO: possibly add switch to reassign or keep order from MDR file?
	for ; r < cart.SectorCount(); r++ {

		header := make([]byte, 27)
		ix := raw.CopySyncPattern(header)

		read, err := io.ReadFull(in, header[ix:])
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				if read == 1 {
					cart.SetWriteProtected(header[ix] > 0)
				} else {
					log.Warnf("expected one final byte, but got %d", read)
					cart.SetWriteProtected(false)
				}
				break
			}
			return nil, err
		}

		record := make([]byte, 540)
		ix = raw.CopySyncPattern(record)
		if _, err := io.ReadFull(in, record[ix:]); err != nil {
			return nil, err
		}

		hd, err := if1.NewHeader(header, false)
		if err != nil && repair {
			if e := hd.FixChecksum(); e != nil {
				log.Warnf("cannot fix checksum of header at index %d: %v", r, e)
			} else {
				log.Debugf("fixed checksum of header at index %d", r)
				err = nil
			}
		}
		if err != nil {
			msg := fmt.Sprintf("defective header at index %d: %v", r, err)
			if strict {
				return nil, fmt.Errorf(msg)
			}
			log.Warn(msg)
		}

		rec, err := if1.NewRecord(record, false)
		if err != nil && repair {
			if e := rec.FixChecksums(); e != nil {
				log.Warnf("cannot fix checksums of record at index %d: %v", r, e)
			} else {
				log.Debugf("fixed checksums of record at index %d", r)
				err = nil
			}
		}
		if err != nil {
			msg := fmt.Sprintf("defective record at index %d: %v", r, err)
			if strict {
				return nil, fmt.Errorf(msg)
			}
			log.Warn(msg)
		}

		sec, err := microdrive.NewSector(hd, rec)
		if err != nil {
			msg := fmt.Sprintf("defective sector at index %d: %v", r, err)
			if strict {
				return nil, fmt.Errorf(msg)
			} else {
				log.Warn(msg)
			}
		}

		cart.SetNextSector(sec)

		if log.IsLevelEnabled(log.TraceLevel) {
			sec.Emit(os.Stdout)
		}
	}

	if repair {
		RepairOrder(cart)
	}

	log.Debugf("%d sectors loaded", r)
	cart.SetModified(false)

	return cart, nil
}

//
func (m *MDR) Write(cart base.Cartridge, out io.Writer) error {

	cart.SeekToStart()

	for ix := 0; ix < cart.SectorCount(); ix++ {
		if sec := cart.GetNextSector(); sec != nil {
			if _, err := out.Write(
				sec.Header().Demuxed()[raw.SyncPatternLength:]); err != nil {
				return err
			}
			if _, err := out.Write(
				sec.Record().Demuxed()[raw.SyncPatternLength:]); err != nil {
				return err
			}
		}
	}

	var wp byte = 0x00
	if cart.IsWriteProtected() {
		wp = 0xff
	}
	if _, err := out.Write([]byte{wp}); err != nil {
		return err
	}

	return nil
}
