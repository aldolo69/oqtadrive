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

	"github.com/xelalexv/microdrive/pkg/microdrive/abstract"
	"github.com/xelalexv/microdrive/pkg/microdrive/ql"
)

//
type QLDump struct{}

//
func NewQLDump() *QLDump {
	return &QLDump{}
}

func (d *QLDump) Read(in io.Reader, strict bool) (*abstract.Cartridge, error) {

	cart := abstract.NewCartridge()

	for {

		header := make([]byte, 28)
		_, err := io.ReadFull(in, header)
		if err != nil {
			cart.SetWriteProtected(false)
			break
		}

		record := make([]byte, 686)
		_, err = io.ReadFull(in, record)
		if err != nil {
			break
		}

		hd, err := ql.NewHeader(header, false)
		if err != nil {
			fmt.Printf("header error: %v\n", err)
			// ignore garbage sectors
			continue
		}

		rec, err := ql.NewRecord(record, false)
		if err != nil {
			fmt.Printf("record error: %v\n", err)
			// ignore garbage sectors
			continue
		}

		sec, err := abstract.NewSector(hd, rec)
		if err != nil {
			// ignore garbage sectors
			//	continue
		}

		cart.SetSector(sec)
	}

	cart.SetModified(false)
	return cart, nil
}

//
func (d *QLDump) Write(cart *abstract.Cartridge, out io.Writer) error {
	return nil
}
