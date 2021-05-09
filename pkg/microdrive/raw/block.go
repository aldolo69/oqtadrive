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
func (b *Block) GetByte(key string) byte {
	if ix, ok := b.index[key]; ok {
		if 0 <= ix[0] && ix[0] < len(b.Data) && ix[1] == 1 {
			return b.Data[ix[0]]
		}
	}
	return 0
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
func (b *Block) GetInt(key string) int {
	bytes := b.GetSlice(key)
	if len(bytes) != 2 {
		return -1
	}
	return int(bytes[0]) | (int(bytes[1]) << 8)
}

//
func (b *Block) GetString(key string) string {
	return string(b.GetSlice(key))
}

//
func (b *Block) Sum(key string) int {
	sum := 0
	for _, s := range b.GetSlice(key) {
		sum += int(s)
	}
	return sum
}
