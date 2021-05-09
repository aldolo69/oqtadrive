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
type Record interface {

	// Client returns the type of client for which the record is intended
	Client() microdrive.Client

	// Muxed returns the muxed data bytes of the record as needed for replay
	Muxed() []byte

	// Demuxed returns the plain data bytes of the record
	Demuxed() []byte

	Flag() int
	Index() int
	Length() int

	// Name returns the name of the record, if applicable
	Name() string

	// Emit emits the record
	Emit(w io.Writer)

	// Validate validates the record
	Validate() error
}

//
func NewRecord(cl microdrive.Client, data []byte, raw bool) (Record, error) {

	switch cl {

	case microdrive.IF1:
		return if1.NewRecord(data, raw)

	case microdrive.QL:
		return ql.NewRecord(data, raw)

	default:
		return nil, fmt.Errorf("unsupported client type for record: %d", cl)
	}
}
