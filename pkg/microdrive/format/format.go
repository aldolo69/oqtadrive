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
	"strings"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
)

// Reader interface for reading in a cartridge
type Reader interface {
	// when setting strict, invalid sectors are discarded
	Read(in io.Reader, strict, repair bool,
		params map[string]interface{}) (base.Cartridge, error)
}

// Writer interface for writing out a cartridge
type Writer interface {
	Write(cart base.Cartridge, out io.Writer,
		params map[string]interface{}) error
}

// ReaderWriter interface for reading/writing a cartridge
type ReaderWriter interface {
	Reader
	Writer
}

//
func NewFormat(typ string) (ReaderWriter, error) {

	switch strings.ToLower(typ) {

	case "mdr":
		return NewMDR(), nil

	case "mdv":
		return NewMDV(), nil

	case "z80":
		return NewZ80(), nil

	default:
		return nil, fmt.Errorf("unsupported cartridge format: %s", typ)
	}
}
