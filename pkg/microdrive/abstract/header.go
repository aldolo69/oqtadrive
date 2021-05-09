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

package abstract

import (
	"fmt"
	"io"

	"github.com/xelalexv/microdrive/pkg/microdrive"
	"github.com/xelalexv/microdrive/pkg/microdrive/if1"
	"github.com/xelalexv/microdrive/pkg/microdrive/ql"
)

//
type Header interface {

	// Client returns the type of client for which the header is intended
	Client() microdrive.Client

	// Muxed returns the muxed data bytes of the header as needed for replay
	Muxed() []byte

	// Demuxed returns the plain data bytes of the header
	Demuxed() []byte

	Flag() int
	Index() int

	// Name returns the name of the cartridge the header belongs to
	Name() string

	// Emit emits the header
	Emit(w io.Writer)

	// Validate validates the header
	Validate() error
}

//
func NewHeader(cl microdrive.Client, data []byte, raw bool) (Header, error) {

	switch cl {

	case microdrive.IF1:
		return if1.NewHeader(data, raw)

	case microdrive.QL:
		return ql.NewHeader(data, raw)

	default:
		return nil, fmt.Errorf("unsupported client type for header: %d", cl)
	}
}
