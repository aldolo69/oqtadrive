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

package daemon

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/microdrive/pkg/microdrive/abstract"
)

//
type mru struct {
	sector *abstract.Sector
	header abstract.Header
	record abstract.Record
}

//
func (m *mru) reset() {
	log.Trace("MRU reset")
	m.sector = nil
	m.header = nil
	m.record = nil
}

//
func (m *mru) setSector(s *abstract.Sector) error {

	if m.header != nil {
		log.Warn("processing next sector while pending header present")
		m.header = nil
	}

	if m.record != nil {
		log.Warn("processing next sector while pending record present")
		m.record = nil
	}

	if s != nil {
		log.WithField("sector", s.Index()).Trace("MRU")
	} else {
		log.WithField("sector", "(nil)").Trace("MRU")
	}

	m.sector = s
	return nil
}

//
func (m *mru) createSector() (*abstract.Sector, error) {
	defer m.reset()
	return abstract.NewSector(m.header, m.record)
}

//
func (m *mru) setHeader(h abstract.Header) error {

	if m.header != nil {
		return fmt.Errorf("processing next header while pending header present")
	}

	if m.record != nil {
		return fmt.Errorf("processing next header while pending record present")
	}

	m.sector = nil

	if h != nil {
		log.WithField("header", h.Index()).Trace("MRU")
	} else {
		log.WithField("header", "(nil)").Trace("MRU")
	}

	m.header = h
	return nil
}

//
func (m *mru) setRecord(r abstract.Record) error {

	if m.header == nil {
		if m.sector == nil {
			return fmt.Errorf("processing next record without sector or header")
		}
		m.sector.SetRecord(r)
	}

	if m.record != nil {
		return fmt.Errorf("processing next record while pending record present")
	}

	if r != nil {
		log.WithField("record", r.Index()).Trace("MRU")
	} else {
		log.WithField("record", "(nil)").Trace("MRU")
	}

	m.record = r
	return nil
}

//
func (m *mru) isNewSector() bool {
	return m.sector == nil && m.header != nil && m.record != nil
}

//
func (m *mru) isRecordUpdate() bool {
	return m.sector != nil && m.header == nil && m.record != nil
}
