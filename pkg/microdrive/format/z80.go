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
	"io"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/format/z80"
)

// Z80 is a format for loading Z80 snapshots. It is an asymmetrical format in
// the sense that it reads Z80 snapshots, but writes MDRs.
type Z80 struct{}

//
func NewZ80() *Z80 {
	return &Z80{}
}

//
func (z *Z80) Read(in io.Reader, strict, repair bool,
	params map[string]interface{}) (base.Cartridge, error) {

	name := ""
	if params != nil {
		if v, ok := params["name"]; ok && v != nil {
			if n, ok := v.(string); ok {
				name = n
			}
		}
	}

	cart, err := z80.LoadZ80(in, name)
	if err != nil {
		return nil, err
	}

	if repair {
		RepairOrder(cart)
	}

	cart.SetModified(false)
	return cart, nil
}

//
func (z *Z80) Write(cart base.Cartridge, out io.Writer,
	params map[string]interface{}) error {

	return NewMDR().Write(cart, out, params)
}
