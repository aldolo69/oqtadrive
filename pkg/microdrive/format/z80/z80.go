/*
   OqtaDrive - Sinclair Microdrive emulator
   Copyright (c) 2021, Alexander Vollschwitz

   This file is part of OqtaDrive.

   The Z80toMDR code is based on Z80onMDR_Lite, copyright (c) 2021 Tom Dalby,
   ported from C to Go by Alexander Vollschwitz. For the original C code, refer
   to:

        https://github.com/TomDDG/Z80onMDR_lite

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
	"fmt"
	"io"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
)

//
type snapshot struct {
	//
	compressed bool
	otek       bool
	version    int
	main       []byte
	launcher   []byte
	code       []byte
	bank       []int
	bankEnd    byte
	//
	name string
	cart base.Cartridge
}

//
func (s *snapshot) setName(n string) {
	if n == "" {
		n = "Z80onMDR"
	}
	s.name = fmt.Sprintf("%.10s", fmt.Sprintf("%-10s", n))
}

// reads Z80 snapshot and converts it into a cartridge on the fly
//
func LoadZ80(in io.Reader, name string) (base.Cartridge, error) {

	snap := &snapshot{}
	if err := snap.unpack(in); err != nil {
		return nil, fmt.Errorf("error unpacking Z80 snapshot: %v", err)
	}

	snap.setName(name)

	if err := snap.pack(); err != nil {
		return nil, fmt.Errorf(
			"error storing Z80 snapshot into cartridge: %v", err)
	}

	return snap.cart, nil
}
