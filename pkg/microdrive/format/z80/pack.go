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
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/if1"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/raw"
)

//
func (s *snapshot) pack() error {

	s.cart = if1.NewCartridge()
	s.cart.SetName(s.name)

	// write 'run' file
	start := 23813
	param := 0

	length := len(s.code)

	log.Debugf("run file: %d", length)
	if err := s.addToCartridge(fmt.Sprintf("%-10s", "run"), s.code, length,
		start, param, 0x00); err != nil {
		return err
	}

	comp := make([]byte, 6912+216+109)

	// screen
	length = zxsc(s.main, comp[len(scrLoad):], 6912, true)
	length += copy(comp, scrLoad) // add m/c

	// write screen
	start = 25088
	param = 0xffff

	log.Debugf("screen file: %d", length)
	if err := s.addToCartridge(fmt.Sprintf("%-10s", "S"), comp, length, start,
		param, 0x03); err != nil {
		return err
	}

	// otek pages
	if s.otek {
		comp = make([]byte, 16384+512+len(unpack))
		length = zxsc(s.main[s.bank[4]:], comp[len(unpack):], 16384, false)
		copy(comp, unpack) // add in unpacker

		nameCount := 1

		length += len(unpack)
		log.Debugf("page file 1: %d", length)
		start = 32256 - len(unpack)
		param = 0xffff
		if err := s.addToCartridge(fmt.Sprintf("%-10d", nameCount), comp, length,
			start, param, 0x03); err != nil {
			return err
		}

		// page 3
		nameCount++
		comp[0] = 0x13
		length = zxsc(s.main[s.bank[6]:], comp[1:], 16384, false) + 1
		log.Debugf("page file 3: %d", length)
		start = 32255 // don't need to replace the unpacker, just the page number
		if err := s.addToCartridge(fmt.Sprintf("%-10d", nameCount), comp, length,
			start, param, 0x03); err != nil {
			return err
		}

		// page 4
		nameCount++
		comp[0] = 0x14
		length = zxsc(s.main[s.bank[7]:], comp[1:], 16384, false) + 1
		log.Debugf("page file 4: %d", length)
		if err := s.addToCartridge(fmt.Sprintf("%-10d", nameCount), comp, length,
			start, param, 0x03); err != nil {
			return err
		}

		// page 6
		nameCount++
		comp[0] = 0x16
		length = zxsc(s.main[s.bank[9]:], comp[1:], 16384, false) + 1
		log.Debugf("page file 6: %d", length)
		if err := s.addToCartridge(fmt.Sprintf("%-10d", nameCount), comp, length,
			start, param, 0x03); err != nil {
			return err
		}

		// page 7
		nameCount++
		comp[0] = 0x17
		length = zxsc(s.main[s.bank[10]:], comp[1:], 16384, false) + 1
		log.Debugf("page file 7: %d", length)
		if err := s.addToCartridge(fmt.Sprintf("%-10d", nameCount), comp, length,
			start, param, 0x03); err != nil {
			return err
		}
	}

	// main
	comp = make([]byte, 42240+1320)
	delta := 3

	for {
		//delta++
		// up to the full size - delta
		length = zxsc(s.main[6912:], comp, 42240-delta, false)
		i := decompressf(comp, length)
		delta += i
		if delta > BGap {
			return fmt.Errorf(
				"cannot compress main block, delta too large: %d > %d",
				delta, BGap)
		}
		if i < 1 {
			break
		}
	}

	maxSize := 40704 // 0x6100 lowest point
	if length > maxSize-delta {
		// too big to fit in Spectrum memory
		return fmt.Errorf(
			"cannot compress main block, max size exceeded: %d > %d",
			length, maxSize-delta)
	}

	// write main
	start = 65536 - length
	param = 0xffff
	log.Debugf("main file: %d (delta: %d)", length, delta)
	if err := s.addToCartridge(fmt.Sprintf("%-10s", "M"), comp, length, start,
		param, 0x03); err != nil {
		return err
	}

	//launcher
	length = 65536 - length         // start of compression
	s.launcher[ixLCS] = byte(delta) //adjust last copy for delta
	s.launcher[ixCP] = byte(length)
	s.launcher[ixCP+1] = byte(length >> 8)
	for i := 0; i < delta; i++ {
		//copy end delta*bytes to launcher
		s.launcher[launchMDRFullLen+i] = s.main[49152-delta+i]
	}

	// write launcher
	length = launchMDRFullLen + delta
	log.Debugf("launcher file: %d", length)
	start = 16384
	if err := s.addToCartridge(fmt.Sprintf("%-10s", "L"), s.launcher, length,
		start, param, 0x03); err != nil {
		return err
	}

	if err := padCartridge(s.cart); err != nil {
		return err
	}

	return nil
}

// add data to the virtual cartridge
func (s *snapshot) addToCartridge(file string, data []byte,
	length, start, param int, dataType byte) error {

	log.WithFields(log.Fields{
		"file":   file,
		"length": length,
		"start":  start,
		"param":  param,
		"type":   dataType,
	}).Debug("adding to cartridge")

	var dataPos int
	var sPos int

	// work out how many sectors needed
	numSec := ((length + 9) / 512) + 1 // +9 for initial header

	for sequence := 0; sequence < numSec; sequence++ {

		var b bytes.Buffer

		// sector header
		raw.WriteSyncPattern(&b)
		b.WriteByte(0x01)
		secIx := s.cart.AdvanceAccessIx(false)
		b.WriteByte(byte(secIx + 1))
		b.WriteByte(0x00)
		b.WriteByte(0x00)
		b.WriteString(s.cart.Name())
		b.WriteByte(0x00)

		hd, _ := if1.NewHeader(b.Bytes(), false)
		if err := hd.FixChecksum(); err != nil {
			return fmt.Errorf("error creating header: %v", err)
		}

		// file header
		//	0x06 - for end of file and data, 0x04 for data if in numerous parts
		//	0x00 - sequence number (if file in many parts then this is the number)
		//	0x00 0x00 - length of this part 16bit
		//	0x00*10 - filename
		//	0x00 - header checksum
		b.Reset()
		raw.WriteSyncPattern(&b)
		if sequence == numSec-1 {
			b.WriteByte(0x06)
		} else {
			b.WriteByte(0x04)
		}
		b.WriteByte(byte(sequence))

		num := 0
		if length > 512 { // if length >512 then this is 512 until final part
			num = 512
		} else if numSec > 1 {
			num = length
		} else {
			num = length + 9 // add 9 for header info
		}
		writeUInt16(&b, num)

		b.WriteString(file)
		b.WriteByte(0x00)

		// data - 512 bytes of data
		//
		// *note first sequence of data must have the header in the format
		//
		//  (1)   0x00, 0x01, 0x02 or 0x03 - program, number array, character
		//        array or code file
		//  (2,3) 0x00 0x00 - total length
		//  (4,5) start address of the block (0x05 0x5d for basic 23813)
		//  (6,7) 0x00 0x00 - total length of program (same as above if
		//        basic of 0xff if code)
		//  (8,9) 0x00 0x00 - line number if LINE used
		//
		if sequence == 0 {
			b.WriteByte(dataType)
			writeUInt16(&b, length)
			writeUInt16(&b, start)

			if dataType == 0x00 { // basic
				writeUInt16(&b, length)
				writeUInt16(&b, param)
			} else {
				b.WriteByte(0xff)
				b.WriteByte(0xff)
				b.WriteByte(0xff)
				b.WriteByte(0xff)
			}

			sPos = 36

		} else {
			sPos = 27 // to cover the headers
		}

		j := length // copy code

		if j > 512 {
			j = 512
			if sequence == 0 {
				j -= 9
			}
		}

		for i := 0; i < j; i++ {
			b.WriteByte(data[dataPos])
			dataPos++
			sPos++
		}

		for ; sPos < if1.RecordLength; sPos++ { // padding on last sequence
			b.WriteByte(0x00)
		}

		if sequence == 0 {
			length -= 503
		} else {
			length -= 512
		}

		rec, _ := if1.NewRecord(b.Bytes(), false)
		if err := rec.FixChecksums(); err != nil {
			return fmt.Errorf("error creating record: %v", err)
		}

		if sec, err := base.NewSector(hd, rec); err != nil {
			return err
		} else {
			s.cart.SetSectorAt(secIx, sec)
		}
	}

	return nil
}
