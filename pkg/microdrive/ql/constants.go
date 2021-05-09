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

package ql

import (
	"fmt"
)

const HeaderLength = 28
const HeaderLengthMux = HeaderLength + 1
const RecordLength = 538
const RecordLengthMux = RecordLength + 1
const FormatExtraBytes = 86

const MaxSectorLength = HeaderLength + RecordLength + FormatExtraBytes

//
func toQLCheckSum(sum int) int {
	return (0x0f0f + sum) % 0x10000
}

//
func verifyQLCheckSum(sum, check int) error {
	if sum != check {
		return fmt.Errorf("want %d, got %d", check, sum)
	}
	return nil
}
