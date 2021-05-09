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
	"io"
)

//
type Sector struct {
	header Header
	record Record
	//
	err error
}

//
func NewSector(h Header, r Record) (*Sector, error) {
	s := &Sector{
		header: h,
		record: r,
	}
	s.validate()
	return s, s.err
}

//
func (s *Sector) Index() int {
	if s.header == nil {
		return -1
	}
	return s.header.Index()
}

// Name returns the name of the cartridge to which this sector belongs
func (s *Sector) Name() string {
	if s.header == nil {
		return ""
	}
	return s.header.Name()
}

//
func (s *Sector) Header() Header {
	return s.header
}

//
func (s *Sector) Record() Record {
	return s.record
}

//
func (s *Sector) SetRecord(r Record) {
	s.record = r
}

//
func (s *Sector) validate() {
}

// Emit emits this sector
func (s *Sector) Emit(w io.Writer) {
	s.header.Emit(w)
	s.record.Emit(w)
}
