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

package microdrive

import (
	"fmt"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/if1"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/ql"
)

//
func NewCartridge(cl client.Client) (base.Cartridge, error) {

	switch cl {

	case client.IF1:
		return if1.NewCartridge(), nil

	case client.QL:
		return ql.NewCartridge(), nil

	default:
		return nil, fmt.Errorf("unsupported client type for cartridge: %d", cl)
	}
}

//
func NewSector(h base.Header, r base.Record) (base.Sector, error) {
	return base.NewSector(h, r)
}

//
func NewHeader(cl client.Client, data []byte, raw bool) (base.Header, error) {

	switch cl {

	case client.IF1:
		return if1.NewHeader(data, raw)

	case client.QL:
		return ql.NewHeader(data, raw)

	default:
		return nil, fmt.Errorf("unsupported client type for header: %d", cl)
	}
}

//
func NewRecord(cl client.Client, data []byte, raw bool) (base.Record, error) {

	switch cl {

	case client.IF1:
		return if1.NewRecord(data, raw)

	case client.QL:
		return ql.NewRecord(data, raw)

	default:
		return nil, fmt.Errorf("unsupported client type for record: %d", cl)
	}
}
