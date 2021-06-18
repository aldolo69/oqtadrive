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
	"bufio"
	"time"

	log "github.com/sirupsen/logrus"
)

// decompress z80 snapshot routine
func decompressZ80(r *bufio.Reader, out []byte) error {

	var c byte
	var k byte
	var j byte
	var err error

	for i := 0; i < len(out); {

		if c, err = r.ReadByte(); err != nil {
			return err
		}

		if c == 0xed { // is it 0xed [0]

			if c, err = r.ReadByte(); err != nil { // get next
				return err
			}

			if c == 0xed { // is 2nd 0xed then a sequence
				if j, err = r.ReadByte(); err != nil { // counter into j
					return err
				}
				if c, err = r.ReadByte(); err != nil {
					return err
				}
				for k = 0; k < j; k++ {
					out[i] = c
					i++
				}

			} else {
				out[i] = 0xed
				i++
				if err = r.UnreadByte(); err != nil { // back one
					return err
				}
			}

		} else {
			out[i] = c // just copy
			i++
		}
	}

	return nil
}

// zxsc modified lzf compressor
func zxsc(fload, store []byte, fileSize int, screen bool) int {

	tStart := time.Now()

	buffer := 0
	storeC := 0
	storeL := 0

	try := make([]*loj, fileSize)
	p := 0 // move pointer to start of storage
	c := 0

	i := 0
	j := 0
	costSum := 0.0

	// get max length & offset for each byte into try array, this also reorgs
	// a screen input file to a linear sequence
	if screen {
		buffer = 6144 // move screen check start to start of attr space
	} else {
		buffer = 0 // move screen check start to start of buffer
	}

	l := &loj{}
	try[p] = l
	l.length = 0
	l.offset = 0
	l.cost = 0.0
	l.byt = fload[buffer] // copy first as literal with control byte
	p++

	tMatch := time.Now()
	if screen {
		// screen version follows screen layout starting at attributes
		// move buffer start check on one and check not at end of the screen
		for {
			if buffer = zxLayout(buffer); buffer >= 6912 {
				break
			}
			try[p] = findMatch(fload, buffer)
			p++
		}
	} else {
		// normal version just linear
		// move screen start check on one and check not at end of the screen
		for {
			if buffer++; buffer >= fileSize {
				break
			}
			try[p] = findMatch2(fload, buffer, fileSize)
			p++
		}
	}
	log.Debugf("findmatch time: %v", time.Now().Sub(tMatch))

	// calculate cost to end for each byte, uses greedy parser, backwards
	// version with re-use for massive speed-up
	p = fileSize - 1 // move byte pointer to end
	try[p].cost = 1.0

	for p--; p > 0; p-- {
		c = p // count pointer to current byte pointer
		if try[c].length == 0 {
			costSum = 1.0 // literal needs 1 bytes
			c++
			// penalize literal followed by match by size of match, the longer
			// the match the smaller the penalty
			if try[c].length != 0 {
				costSum += (1.0 / float64((try[c].length))) / 10.0
			}
		} else {
			j = try[c].length
			if c+j < fileSize && j > MinLength {
				for i = MinLength; i < try[c].length; i++ {
					if try[c+i].cost < try[c+j].cost {
						j = i
					}
				}
				try[c].length = j // adjust if it can find a better route
			}
			if try[c].length < 9 {
				costSum = 2.0
			} else {
				costSum = 3.0 // if length 3-8 then 2 else 3 cost
			}
			c += try[c].length // move it on the match length
		}

		if c < fileSize {
			costSum += try[c].cost
		}
		try[p].cost = costSum // write cost to end for current byte
	}

	try[p].cost = 2.0 + try[p+1].cost
	p = 0 // move byte pointer to the start

	storeC = 0          // control byte pointer -> start of storage
	storeL = 1          // literal store pointer -> start of storage+1
	store[storeC] = 255 // set initial control byte to 255 (clear)

	for {
		// if not a literal then check for a lower cost alternative is available
		if try[p].length != 0 {
			// look over the full match length to see if there is a better match
			j = 0
			for i = 1; i < try[p].length; i++ {
				// check if adding literals makes a difference
				if i < MinLength {
					// also capture if it is a control byte 255
					if int(store[storeC])+i > 31 {
						if try[p+i].cost+float64(i)+1.0 < try[p+j].cost {
							j = i
						}
					} else if try[p+i].cost+float64(i) < try[p+j].cost {
						j = i
					}
				} else if i < 9 {
					if try[p+i].cost+2.0 < try[p+j].cost {
						j = i // add 2 to mimic storage of 3-8 match
					}
				} else {
					if try[p+i].cost+3.0 < try[p+j].cost {
						j = i // add 3 to mimic storage of 9+ match
					}
				}
			}
			if j != 0 { // if j=0 then nothing better found so just continue
				if j < MinLength { // is it 1 or 2 ahead?
					for i = 0; i < j; i++ {
						try[p+i].length = 0 // change to a literal
					}
				} else {
					// if j>2 then just change to new length
					try[p].length = j
				}
			}
		}

		// now store either an offset+length or a literal
		if try[p].length != 0 { // offset+length

			if !screen {
				try[p].offset-- // reduce offset by one for normal only
			}

			if store[storeC] != 255 {
				// if control is not clear move to literal store and
				// move that on one
				storeC = storeL
				storeL++
			}

			i = try[p].length - 1 // store length-1 for later jump
			// reduce by 2 so 3->1, 8->6 etc... and check if >6
			try[p].length -= 2

			if try[p].length > 6 {
				// if >6 then reduce max length by 7 to get length to store
				try[p].length -= 7
				// store 2nd part of length in literal store and move literal
				// byte on one
				store[storeL] = byte(try[p].length)
				storeL++
				try[p].length = 7 // make length 7 for offset control byte
			}

			// shift length 5 bits to left and bring in offset hi byte
			store[storeC] = (byte(try[p].length) << 5) + byte(try[p].offset>>8)
			// offset low byte in next byte and move on one
			store[storeL] = byte(try[p].offset)
			storeL++
			storeC = storeL // move control byte up, byte store on one
			storeL++
			store[storeC] = 255 // clear new control byte
			p += i              // jump forward to next byte

		} else { // store a literal
			// copy new literal into byte store and move byte store on one
			store[storeL] = try[p].byt
			storeL++

			// increase control byte by one and check if at max or at end of file
			store[storeC] += 1
			if store[storeC] == 31 || p == fileSize-1 {
				// move new control to literal and move literal on one
				storeC = storeL
				storeL++
				store[storeC] = 255 // clear new control byte
			}
		}

		// move start check on one and check not at end of the compression
		p++
		if p >= fileSize {
			break
		}
	}

	log.Debugf("compression time: %v", time.Now().Sub(tStart))

	return storeL
}

// structure to store for each byte the max length, offset, the byte itself and
// the cost to end which is then used to optimize the compression. It is also
// used to create a linear version of the screen.
type loj struct {
	length int
	offset int
	byt    byte
	cost   float64
}

// screen version, attr then pixels char row then back to attr
func findMatch(buf []byte, ix int) *loj {

	// store length, offset, byte & cost; copy byte to compare to output, set
	// offset to 0, set session max match length to zero
	ret := &loj{byt: buf[ix]}

	sc := 0 // screen check position
	ds := 0 // dictionary start position
	dc := 0 // dictionary check position
	length := 0

	ds = 6144 // initial dictionary start to start of screen attr space

	for {
		length = 0 // reset current match length to zero
		dc = ds    // move dictionary check pos to dictionary start
		sc = ix    // move screen check pos to screen start

		// if the bytes match keep checking
		for buf[sc] == buf[dc] { // if screen check=dictionary keep checking
			if length++; length == MaxLength {
				break // increase length and if at max break out of inner loop
			}
			if sc = zxLayout(sc); sc == 6912 {
				// move screen check pos on one and if at end of screen
				// break out of inner loop
				break
			}
			// move dictionary on one, can go beyond current dictionary end as
			// extra dictionary would be built up before it got to this part
			dc = zxLayout(dc)
		}

		// check entire dictionary for max match, bigger than min size and
		// previous maximum?
		if length > 2 && length > ret.length {
			ret.length = length // new max found so store
			ret.offset = ds     // calc memory position from start for new max
		}
		if sc == 6912 || length == MaxLength {
			// check if end of screen or max size reached -> break out of loop
			break
		}

		// moves start of dictionary on one and checks if caught up
		if ds = zxLayout(ds); ds == ix {
			break
		}
	}

	return ret
}

// linear version
func findMatch2(buf []byte, ix, fileSize int) *loj {

	// copy byte, set session max match length to zero
	ret := &loj{byt: buf[ix]}

	sc := 0
	ds := 0
	dc := 0
	length := 0

	ds = ix - 7936 // initial dictionary start to current pos - max offset
	if ds < 0 {
		ds = 0
	}

	for {
		length = 0 // reset current match length to zero
		dc = ds    // move dictionary check pos to dictionary start
		sc = ix    // move screen check pos to screen start

		for buf[sc] == buf[dc] {
			// increase length and check against max length possible
			// -> break out of inner loop
			if length++; length == MaxLength {
				break
			}
			// move screen check pos on one & check if at end of screen
			// -> break out out of inner loop
			if sc++; sc == fileSize {
				break
			}
			// can go beyond current dictionary end as extra dictionary would be
			// built up before it got to this part
			dc++
		}

		// bigger than min size and previous maximum?
		if length >= MinLength && length > ret.length {
			ret.length = length  // new max found so store
			ret.offset = ix - ds // calc offset
		}
		if sc == fileSize || length == MaxLength {
			// check at end of screen or max size reached -> break out of loop
			break
		}

		// moves start of dictionary on one and checks if caught up
		if ds++; ds == ix {
			break
		}
	}

	return ret
}

// follow the screen layout rather than linear
func zxLayout(pos int) int {

	p := uint16(pos) // current memory position

	if h := p & 0xff00; h >= 0x1800 {
		// if high byte >=24 then in attr space >=6144bytes
		// and %00000111, rotate hi byte to left x3 or move to pixel space
		// 	-> 6144 to 0 etc...
		p = (p & 0x00ff) | ((h & 0x0700) << 3)

	} else {
		// in pixel space so increment high byte to move down one char row
		p = (p & 0x00ff) | (h + 0x0100)
		// if and %00000111=0 then hi byte has just crossed into the next char
		// so need to move back to attr space
		if h = p >> 8; h&7 == 0 {
			// go back up one pixel row so conversion to attr space works
			h--
			// rotate hi byte to right x3
			h >>= 3
			// and %00000011 to keep first two bits
			h &= 3
			// or in %00011000 to move back to attr space
			h |= 24
			// preserving char column and row >=6144
			p = (p & 0x00ff) | (h << 8)
			p++ // move onto next char
		}
	}

	return int(p) // return byte position as an int
}

// check compression to ensure it can be decompressed within Spectrum memory
func decompressf(comp []byte, compSize int) int {

	var a byte

	deltaC := 42240 - compSize
	deltaN := 0
	c := 0
	j := 0

	for hl := 0; comp[hl] != 0xff; {
		if comp[hl] < 0x20 { // <32 simple literal copy
			j = int(comp[hl])
			hl++
			deltaC++
			j++ // inc c
			hl += j
			deltaC += j
			deltaN += j
		} else {
			a = comp[hl]
			hl++
			deltaC++
			a = a >> 5 // rotate hi byte to right x3
			a = a & 7  // and %00000011 to keep first two bits
			if a == 7 {
				c = int(a) + int(comp[hl])
				hl++
				deltaC++
			} else {
				c = int(a)
			}
			c += 2 // c now correct length
			deltaC++
			deltaN += c
			hl++
			if deltaC < deltaN {
				// caught up so delta not large enough, return gap
				return deltaN - deltaC
			}
		}
	}

	return 0
}
