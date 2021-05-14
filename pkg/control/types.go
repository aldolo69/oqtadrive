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

package control

import (
	"fmt"
	"strings"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
)

//
func NewCartridge(cart base.Cartridge) *Cartridge {
	return &Cartridge{
		Name:           strings.TrimSpace(cart.Name()),
		Formatted:      cart.IsFormatted(),
		WriteProtected: cart.IsWriteProtected(),
		Modified:       cart.IsModified(),
	}
}

//
type Cartridge struct {
	Name           string `json:"name"`
	Formatted      bool   `json:"formatted"`
	WriteProtected bool   `json:"writeProtected"`
	Modified       bool   `json:"modified"`
}

//
func (c *Cartridge) String() string {

	name := c.Name

	if name == "" {
		name = "<no name>"
	}

	format := 'b'
	if c.Formatted {
		format = 'f'
	}

	write := 'w'
	if c.WriteProtected {
		write = 'r'
	}

	mod := ' '
	if c.Modified {
		mod = '*'
	}

	return fmt.Sprintf("%-16s%c%c%c", name, format, write, mod)
}
