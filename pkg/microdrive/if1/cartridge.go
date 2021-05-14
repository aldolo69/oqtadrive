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

package if1

import (
	"fmt"
	"io"
	"sort"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
)

//
func NewCartridge() base.Cartridge {
	return &cartridge{base.NewCartridge(client.IF1, SectorCount)}
}

//
type cartridge struct {
	base.CartridgeBase
}

//
func (c *cartridge) List(w io.Writer) {

	fmt.Fprintf(w, "\n%s\n\n", c.Name())

	dir := make(map[string]int)
	used := 0

	c.SeekToStart()

	for ix := 0; ix < c.SectorCount(); ix++ {

		if sec := c.GetNextSector(); sec != nil {
			if rec := sec.Record(); rec != nil {

				if rec.Flag()&RecordFlagsUsed == 0 {
					continue
				}

				used++

				name := translate(rec.Name())
				if name == "" {
					continue
				}

				size, ok := dir[name]
				if ok {
					size += rec.Length()
				} else {
					size = rec.Length()
				}
				dir[name] = size
			}
		}
	}

	var files []string
	for f := range dir {
		files = append(files, f)
	}
	sort.Strings(files)

	for _, f := range files {
		fmt.Fprintf(w, "%-16s%d\n", f, dir[f])
	}

	fmt.Fprintf(w, "\n%d of %d sectors used (%dkb free)\n\n",
		used, c.SectorCount(), (c.SectorCount()-used)/2)
}
