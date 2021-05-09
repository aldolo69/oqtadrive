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

/*
	Mux takes plain readable data bytes and transforms them into muxed data
	bytes that can be sent to the adapter for replay. Note that a mux-demux
	round-trip will not yield the same plain data, i.e.

		data -> mux -> demux -> data'

	This is because for replay, the bit order needs to be the reverse of the
	order in which the bytes were recorded.

	For the QL, track 1 (DATA1) is ahead of track 2 (DATA2), just the opposite
	of IF1. With invert set to true, muxing will be done as appropriate for the
	QL, i.e. tracks are switched.
*/
func Mux(data []byte, invert bool) []byte {

	if len(data) == 0 {
		return []byte{}
	}

	buf := make([]byte, len(data)+1)

	r := 0
	if invert {
		r = 1
	}

	for ix := 0; ix < len(data); {
		d := data[ix]
		if ix%2 == r {
			buf[ix] = buf[ix] | (d & 0x0f)
			ix++
			buf[ix] = buf[ix] | (d >> 4)
		} else {
			buf[ix] = buf[ix] | (d << 4)
			ix++
			buf[ix] = buf[ix] | (d & 0xf0)
		}
	}

	return buf
}

/*
	Demux takes raw bytes recorded by the adapter and transforms them into plain
	readable data.

	For the QL, track 1 (DATA1) is ahead of track 2 (DATA2), just the opposite
	of IF1. With invert set to true, demuxing will be done as appropriate for
	the QL, i.e. tracks are switched.

	Note that raw gets modified during demux.
*/
func Demux(raw []byte, invert bool) []byte {

	if len(raw) <= 1 {
		return []byte{}
	}

	for ix := range raw {
		raw[ix] = revertNibbles(raw[ix])
	}

	data := make([]byte, len(raw)-1)

	r := 0
	if invert {
		r = 1
	}

	for ix := range data {
		if ix%2 == r {
			data[ix] = (raw[ix] & 0x0f) | (raw[ix+1] << 4) // low nibble
		} else {
			data[ix] = (raw[ix] >> 4) | (raw[ix+1] & 0xf0) // high nibble
		}
	}

	return data
}

//
func revertByte(b byte) byte {
	var ret byte
	for x := 0; ; x++ {
		ret = (ret | 0x80) & (b | 0x7f)
		if x == 7 {
			return ret
		}
		b <<= 1
		ret >>= 1
	}
}

//
func revertNibbles(b byte) byte {
	var ret byte
	for x := 0; ; x++ {
		ret = (ret | 0x88) & (b | 0x77)
		if x == 3 {
			return ret
		}
		b <<= 1
		ret >>= 1
	}
}
