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
	"io"
)

//
func NewSector(h Header, r Record) (Sector, error) {
	ret := &sector{
		header: h,
		record: r,
	}
	if err := ret.validate(); err != nil {
		return nil, err
	}
	return ret, nil
}

//
type sector struct {
	header Header
	record Record
	//
	err error
}

//
func (s *sector) Index() int {
	if s.header == nil {
		return -1
	}
	return s.header.Index()
}

//
func (s *sector) Name() string {
	if s.header == nil {
		return ""
	}
	return s.header.Name()
}

//
func (s *sector) Header() Header {
	return s.header
}

//
func (s *sector) Record() Record {
	return s.record
}

//
func (s *sector) SetRecord(r Record) {
	s.record = r
}

//
func (s *sector) validate() error {
	return nil
}

// Emit emits this sector
func (s *sector) Emit(w io.Writer) {
	s.header.Emit(w)
	s.record.Emit(w)
}
