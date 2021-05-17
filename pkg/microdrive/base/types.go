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

package base

import (
	"context"
	"io"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
)

//
type Cartridge interface {
	//
	CartridgeBase

	List(w io.Writer)
}

//
type CartridgeBase interface {
	//
	Lock(ctx context.Context) bool

	Unlock()

	Client() client.Client

	Name() string

	SectorCount() int

	SeekToStart()

	// GetNextSector gets the sector at the next access index, skipping slots
	// with nil sectors. Access index points to the slot of the returned sector
	// afterwards.
	GetNextSector() Sector

	// GetPreviousSector gets the sector at the previous access index, skipping
	// slots with nil sectors. Access index points to the slot of the returned
	// sector afterwards.
	GetPreviousSector() Sector

	// SetNextSector sets the provided sector at the next access index, whether
	// there is a sector present at that index or not. Access index points to the
	// slot of the set sector afterwards.
	SetNextSector(s Sector)

	// SetPreviousSector sets the provided sector at the previous access index,
	// whether there is a sector present at that index or not. Access index points
	// to the slot of the set sector afterwards.
	SetPreviousSector(s Sector)

	IsFormatted() bool

	IsWriteProtected() bool

	SetWriteProtected(p bool)

	IsModified() bool

	SetModified(m bool)

	IsAutoSaved() bool

	SetAutoSaved(a bool)

	AdvanceAccessIx(skipEmpty bool) int

	RewindAccessIx(skipEmpty bool) int

	Emit(w io.Writer)
}

//
type Sector interface {
	//
	Index() int

	// Name returns the name of the cartridge to which this sector belongs
	Name() string

	Header() Header

	Record() Record

	SetRecord(r Record)

	Emit(w io.Writer)
}

//
type Header interface {

	// Client returns the type of client for which the header is intended
	Client() client.Client

	// Muxed returns the muxed data bytes of the header as needed for replay
	Muxed() []byte

	// Demuxed returns the plain data bytes of the header
	Demuxed() []byte

	//
	Flags() byte

	//
	Index() int

	// Name returns the name of the cartridge the header belongs to
	Name() string

	// Emit emits the header
	Emit(w io.Writer)

	// Validate validates the header
	Validate() error
}

//
type Record interface {

	// Client returns the type of client for which the record is intended
	Client() client.Client

	// Muxed returns the muxed data bytes of the record as needed for replay
	Muxed() []byte

	// Demuxed returns the plain data bytes of the record
	Demuxed() []byte

	//
	Flags() byte

	//
	Index() int

	//
	Length() int

	// Name returns the name of the record, if applicable
	Name() string

	// Emit emits the record
	Emit(w io.Writer)

	// Validate validates the record
	Validate() error
}
