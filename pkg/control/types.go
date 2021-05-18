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

	"github.com/xelalexv/oqtadrive/pkg/daemon"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
)

//
type Status struct {
	Drives []string `json:"drives"`
}

//
func (s *Status) Add(d string) {
	s.Drives = append(s.Drives, d)
}

//
func (s *Status) String() string {
	ret := "\n"
	for ix, d := range s.Drives {
		ret = fmt.Sprintf("%s%d: %s\n", ret, ix+1, d)
	}
	return ret
}

//
type Cartridge struct {
	Name           string `json:"name"`
	Status         string `json:"status"`
	Formatted      bool   `json:"formatted"`
	WriteProtected bool   `json:"writeProtected"`
	Modified       bool   `json:"modified"`
}

//
func (c *Cartridge) fill(cart base.Cartridge) {
	c.Name = strings.TrimSpace(cart.Name())
	c.Formatted = cart.IsFormatted()
	c.WriteProtected = cart.IsWriteProtected()
	c.Modified = cart.IsModified()
}

//
func (c *Cartridge) String() string {

	if c.Status != daemon.StatusIdle {
		return fmt.Sprintf("<%s>", c.Status)
	}

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
