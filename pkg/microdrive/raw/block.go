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
	"fmt"
)

//
func NewBlock(index map[string][2]int, data []byte) *Block {
	return &Block{index: index, Data: data}
}

//
type Block struct {
	index map[string][2]int
	Data  []byte
}

//
func illegalKey(k string) error {
	return fmt.Errorf("illegal key %s", k)
}

//
func typeMismatch(key, typ string) error {
	return fmt.Errorf("key %s is not of type %s", key, typ)
}

//
func (b *Block) Length() int {
	return len(b.Data)
}

//
func (b *Block) GetByte(key string) byte {
	if ix, ok := b.index[key]; ok {
		if 0 <= ix[0] && ix[0] < len(b.Data) && ix[1] == 1 {
			return b.Data[ix[0]]
		}
	}
	return 0
}

//
func (b *Block) SetByte(key string, val byte) error {
	if ix, ok := b.index[key]; ok {
		if ix[1] != 1 {
			return typeMismatch(key, "byte")
		}
		if 0 <= ix[0] && ix[0] < len(b.Data) {
			b.Data[ix[0]] = val
			return nil
		}
		return fmt.Errorf("index out of range: %d > %d", ix[0], len(b.Data)-1)
	}
	return illegalKey(key)
}

//
func (b *Block) GetSlice(key string) []byte {
	if ix, ok := b.index[key]; ok {
		start := ix[0]
		end := start + ix[1]
		if 0 <= start && end <= len(b.Data) {
			return b.Data[start:end]
		}
	}
	return []byte{}
}

//
func (b *Block) SetSlice(key string, val []byte) error {
	bytes := b.GetSlice(key)
	if len(bytes) == 0 {
		return illegalKey(key)
	}
	if len(bytes) != len(val) {
		return fmt.Errorf(
			"wrong slice length: want %d, have %d", len(bytes), len(val))
	}
	copy(bytes, val)
	return nil
}

//
func (b *Block) GetInt(key string) int {
	bytes := b.GetSlice(key)
	if len(bytes) != 2 {
		return -1
	}
	return int(bytes[0]) | (int(bytes[1]) << 8)
}

//
func (b *Block) SetInt(key string, val int) error {
	bytes := b.GetSlice(key)
	if len(bytes) == 0 {
		return illegalKey(key)
	}
	if len(bytes) != 2 {
		return typeMismatch(key, "int")
	}
	bytes[0] = byte(val)
	bytes[1] = byte(val >> 8)
	return nil
}

//
func (b *Block) GetString(key string) string {
	return string(b.GetSlice(key))
}

//
func (b *Block) SetString(key, val string) error {
	return b.SetSlice(key, []byte(val))
}

//
func (b *Block) Sum(key string) int {
	sum := 0
	for _, s := range b.GetSlice(key) {
		sum += int(s)
	}
	return sum
}
