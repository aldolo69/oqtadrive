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

package raw

import (
	"bytes"
	"fmt"
	"io"
)

// maximum number of leading bytes in a sync allowed to be faulty
const syncErrorToleration = 3

// standard sync pattern
const SyncPatternLength = 12

var syncPattern = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

// data sync is used by the QL
const DataSyncPatternLength = 8

var dataSyncPattern = []byte{0, 0, 0, 0, 0, 0, 0xff, 0xff}

//
func CopySyncPattern(dest []byte) int {
	return copy(dest, syncPattern)
}

//
func WriteSyncPattern(wr io.Writer) (int, error) {
	return wr.Write(syncPattern)
}

//
func CopyDataSyncPattern(dest []byte) int {
	return copy(dest, dataSyncPattern)
}

//
func WriteDataSyncPattern(wr io.Writer) (int, error) {
	return wr.Write(dataSyncPattern)
}

//
type Sync struct {
	pattern []byte
	//
	err error
}

//
func NewSync(src *bytes.Reader) (*Sync, error) {
	return newSync(src, syncPattern, syncErrorToleration)
}

//
func NewDataSync(src *bytes.Reader) (*Sync, error) {
	return newSync(src, dataSyncPattern, 0)
}

//
func newSync(src *bytes.Reader, spec []byte, maxErrorToleration int) (*Sync, error) {
	s := &Sync{
		pattern: make([]byte, len(spec)),
	}
	if _, err := io.ReadFull(src, s.pattern); err != nil {
		return s, err
	}
	for ix := 0; ix < maxErrorToleration; ix++ {
		s.pattern[ix] = spec[ix]
	}
	s.validate(spec)
	return s, s.err
}

//
func (s *Sync) validate(spec []byte) {
	for ix := range s.pattern {
		if s.pattern[ix] != spec[ix] {
			s.err = fmt.Errorf(
				"invalid sync pattern, starting at index %d: %v", ix, s.pattern)
			return
		}
	}
}

//
func (s *Sync) Emit() {
	if s.err != nil {
		fmt.Printf("SYNC INVALID: %v\n", s.err)
	}
}
