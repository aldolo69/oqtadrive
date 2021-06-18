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

package z80

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/if1"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/raw"
)

//
func writeUInt16(b *bytes.Buffer, i int) {
	b.WriteByte(byte(i))
	b.WriteByte(byte(i >> 8))
}

//
func readUInt16(r io.Reader) (int, error) {
	b := make([]byte, 2)
	if _, err := io.ReadFull(r, b); err != nil {
		return -1, err
	}
	return int(b[0]) + (int(b[1]) << 8), nil
}

//
func fill(r *bufio.Reader, target []byte, indexes []int) error {
	for _, ix := range indexes {
		if ix < 0 || ix >= len(target) {
			return fmt.Errorf("fill index out of range: %d", ix)
		}
		var err error
		if target[ix], err = r.ReadByte(); err != nil {
			return err
		}
	}
	return nil
}

//
func adjust(target []byte, ix int, val int) error {
	if ix < 0 || ix >= len(target) {
		return fmt.Errorf("adjustment index out of range: %d", ix)
	}
	target[ix] = byte(int(target[ix]) + val)
	return nil
}

//
func padCartridge(cart base.Cartridge) error {

	var b bytes.Buffer

	for ix := cart.AccessIx(); ix > 0; {

		ix = cart.AdvanceAccessIx(false)
		b.Reset()

		// sector header
		raw.WriteSyncPattern(&b)
		b.WriteByte(0x01)
		b.WriteByte(byte(ix + 1))
		b.WriteByte(0x00)
		b.WriteByte(0x00)
		b.WriteString(cart.Name())
		b.WriteByte(0x00)

		hd, _ := if1.NewHeader(b.Bytes(), false)
		if err := hd.FixChecksum(); err != nil {
			return err
		}

		// blank record
		rec, _ := if1.NewRecord(make([]byte, if1.RecordLength), false)
		if err := rec.FixChecksums(); err != nil {
			return err
		}

		if sec, err := base.NewSector(hd, rec); err != nil {
			return err
		} else {
			cart.SetSectorAt(ix, sec)
		}
	}

	return nil
}
