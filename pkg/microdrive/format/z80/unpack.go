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
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

//
func (s *snapshot) unpack(in io.Reader) error {

	rd := bufio.NewReader(in)

	var c byte
	var err error

	s.launcher = make([]byte, len(launchMDRFull))
	copy(s.launcher, launchMDRFull)

	if err = fill(rd, s.launcher, []int{ // read in z80 starting with header
		ixA,  //      0   1    A register
		ixIF, //      1   1    F register
		ixBC, //      2   2    BC register pair(LSB, i.e.C, first)
		ixBC + 1,
		ixHL, //      4   2    HL register pair
		ixHL + 1,
		ixJP, //      6   2    Program counter (if zero then version 2 or 3 snapshot)
		ixJP + 1,
		ixSP, //      8   2    Stack pointer
		ixSP + 1,
		ixIF + 1, // 10   1    Interrupt register
		ixR,      // 11   1    Refresh register (Bit 7 is not significant!)
	}); err != nil {
		return err
	}

	// r, reduce by 6 so correct on launch
	if err = adjust(s.launcher, ixR, -6); err != nil {
		return err
	}

	//  12   1    Bit 0: Bit 7 of r register; Bit 1-3: Border colour;
	//            Bit 4=1: SamROM; Bit 5=1:v1 Compressed; Bit 6-7: N/A
	if c, err = rd.ReadByte(); err != nil {
		return err
	}

	s.compressed = (c&32)>>5 == 1 // 1 compressed, 0 not

	if c&1 == 1 || c > 127 {
		s.launcher[ixR] = s.launcher[ixR] | 128 // r high bit set
	} else {
		s.launcher[ixR] = s.launcher[ixR] & 127 // r high bit reset
	}

	bCol := ((c & 14) >> 1) + 0x30 //border/paper col

	if err = fill(rd, s.launcher, []int{
		ixDE, //      13   2    DE register pair
		ixDE + 1,
		ixBCA, //     15   2    BC' register pair
		ixBCA + 1,
		ixDEA, //     17   2    DE' register pair
		ixDEA + 1,
		ixHLA, //     19   2    HL' register pair
		ixHLA + 1,
		ixAFA + 1, // 21   1    A' register
		ixAFA,     // 22   1    F' register
		ixIY,      // 23   2    IY register (Again LSB first)
		ixIY + 1,
		ixIX, //      25   2    IX register
		ixIX + 1,
	}); err != nil {
		return err
	}

	// 27   1    Interrupt flip flop, 0 = DI, otherwise EI
	if c, err = rd.ReadByte(); err != nil {
		return err
	}
	if c == 0 {
		s.launcher[ixEI] = 0xf3 // di
	} else {
		s.launcher[ixEI] = 0xfb // ei
	}

	// 28   1    IFF2 [IGNORED]
	if _, err = rd.ReadByte(); err != nil {
		return err
	}

	// 29   1    Bit 0-1: IM(0, 1 or 2); Bit 2-7: N/A
	if c, err = rd.ReadByte(); err != nil {
		return err
	}
	c &= 3
	if c == 0 {
		s.launcher[ixIM] = 0x46 // im 0
	} else if c == 1 {
		s.launcher[ixIM] = 0x56 // im 1
	} else {
		s.launcher[ixIM] = 0x5e // im 2
	}

	// version 2 & 3 only
	addLen := 0 // 0 indicates v1, 23 for v2 otherwise v3
	s.otek = false

	if s.launcher[ixJP] == 0 && s.launcher[ixJP+1] == 0 {

		// 30   2    Length of additional header block
		if addLen, err = readUInt16(rd); err != nil {
			return err
		}

		// 32   2    Program counter
		if err = fill(rd, s.launcher, []int{ixJP, ixJP + 1}); err != nil {
			return err
		}

		// 34   1    Hardware mode
		if c, err = rd.ReadByte(); err != nil {
			return err
		}
		if c == 2 {
			return fmt.Errorf("SamRAM Z80 snapshots not supported")
		}
		if addLen == 23 && c > 2 {
			s.otek = true // v2 & c>2 then 128k, if v3 then c>3 is 128k
		} else if c > 3 {
			s.otek = true
		}

		// 35   1    If in 128 mode, contains last OUT to 0x7ffd
		if c, err = rd.ReadByte(); err != nil {
			return err
		}

		if s.otek {
			s.launcher[ixOUT] = c
		}

		// 36   1    Contains 0xff if Interface I rom paged [SKIPPED]
		// 37   1    Hardware Modify Byte [SKIPPED]
		// 38   1    Last OUT to port 0xfffd (soundchip register number) [SKIPPED]
		// 39  16    Contents of the sound chip registers [SKIPPED] *ideally for
		//           128k setting ay registers make sense, however in practise
		//           never found it is needed
		if _, err = rd.Discard(19); err != nil {
			return err
		}

		// following is only in v3 snapshots
		// 55   2    Low T state counter [SKIPPED]
		// 57   1    Hi T state counter [SKIPPED]
		// 58   1    Flag byte used by Spectator(QL spec.emulator) [SKIPPED]
		// 59   1    0xff if MGT Rom paged [SKIPPED]
		// 60   1    0xff if Multiface Rom paged.Should always be 0. [SKIPPED]
		// 61   1    0xff if 0 - 8191 is ROM, 0 if RAM [SKIPPED]
		// 62   1    0xff if 8192 - 16383 is ROM, 0 if RAM [SKIPPED]
		// 63  10    5 x keyboard mappings for user defined joystick [SKIPPED]
		// 73  10    5 x ASCII word : keys corresponding to mappings above [SKIPPED]
		// 83   1    MGT type : 0 = Disciple + Epson, 1 = Disciple + HP, 16 = Plus D [SKIPPED]
		// 84   1    Disciple inhibit button status : 0 = out, 0ff = in [SKIPPED]
		// 85   1    Disciple inhibit flag : 0 = rom pageable, 0ff = not [SKIPPED]
		if addLen > 23 {
			if _, err = rd.Discard(31); err != nil {
				return err
			}
		}

		// only if version 3 & 55 additional length
		// 86   1    Last OUT to port 0x1ffd, ignored for Microdrive as only
		//           applicable on +3/+2A machines [SKIPPED]
		if addLen == 55 {
			if c, err = rd.ReadByte(); err != nil {
				return err
			} else if c&1 == 1 {
				// special page mode so exit as not compatible with
				// earlier 128k machines
				return fmt.Errorf(
					"+3/2A snapshots with special RAM mode enabled not " +
						"supported. Microdrives do not work on +3/+2A hardware.")
			}
		}
	}

	// space for decompression of z80
	// 8 * 16384 = 131072 bytes
	// 0 - 49152 - Pages 5,2 & 0 (main memory)
	// *128k only - 49152 -  65536: Page 1
	//              65536 -  81920: Page 3
	//              81920 -  98304: Page 4
	//              98304 - 114688: Page 6
	//             114688 - 131072: Page 7
	//
	fullSize := 49152
	if s.otek {
		fullSize = 131072
	}
	s.main = make([]byte, fullSize)

	// which version of z80?
	length := 0
	s.bank = make([]int, 11)

	for i := range s.bank {
		s.bank[i] = 99 // default
	}

	if addLen == 0 { // version 1 snapshot & 48k only
		log.Debug("snapshot version: v1")
		s.version = 1
		if s.compressed {
			err = decompressZ80(rd, s.main)
		} else {
			_, err = io.ReadFull(rd, s.main)
		}
		if err != nil {
			return err
		}

	} else { // version 2 & 3
		if addLen == 23 {
			log.Debug("snapshot version: v2")
			s.version = 2
		} else {
			log.Debug("snapshot version: v3")
			s.version = 3
		}

		// Byte    Length  Description
		// -------------------------- -
		// 0       2       Length of compressed data(without this 3 - byte header)
		//                 If length = 0xffff, data is 16384 bytes longand not
		//                 compressed
		// 2       1       Page number of block
		//
		// for 48k snapshots the order is:
		//    0 48k ROM, 1, IF1/PLUSD/DISCIPLE ROM, 4 page 2, 5 page 0, 8 page 5,
		//    11 MF ROM only 4, 5 & 8 are valid for this usage, all others are
		//    just ignored
		// for 128k snapshots the order is:
		//    0 ROM, 1 ROM, 3 Page 0....10 page 7, 11 MF ROM.
		// all pages are saved and there is no end marker
		//
		if s.otek {
			s.bank[3] = 32768   // page 0
			s.bank[4] = 49152   // page 1
			s.bank[5] = 16384   // page 2
			s.bank[6] = 65536   // page 3
			s.bank[7] = 81920   // page 4
			s.bank[8] = 0       // page 5
			s.bank[9] = 98304   // page 6
			s.bank[10] = 114688 // page 7
			s.bankEnd = 10
		} else {
			s.bank[4] = 16384 // page 2
			s.bank[5] = 32768 // page 0
			s.bank[8] = 0     // page 5
			s.bankEnd = 8
		}

		for c = 0; c != s.bankEnd; {
			if length, err = readUInt16(rd); err != nil {
				return err
			}

			if c, err = rd.ReadByte(); err != nil {
				return err
			}

			addr := s.bank[c]

			if addr != 99 {
				target := s.main[addr : addr+16384]
				if length == 65535 {
					_, err = io.ReadFull(rd, target)
				} else {
					err = decompressZ80(rd, target)
				}
				if err != nil {
					return err
				}
			}
		}
	}

	if s.otek {
		log.Debug("snapshot size: 128k")
		s.code = make([]byte, len(mdrBl128k))
		copy(s.code, mdrBl128k)
		s.code[ix128kBrd] = bCol
		s.code[ix128kPap] = bCol
	} else {
		log.Debug("snapshot size: 48k")
		s.code = make([]byte, len(mdrBl48k))
		copy(s.code, mdrBl48k)
		s.code[ix48kBrd] = bCol
		s.code[ix48kPap] = bCol
	}

	return nil
}
