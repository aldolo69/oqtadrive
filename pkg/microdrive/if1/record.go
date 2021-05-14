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

package if1

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/raw"
)

//
var recordIndex = map[string][2]int{
	"flag":         {12, 1},
	"number":       {13, 1},
	"length":       {14, 2},
	"name":         {16, 10},
	"header":       {12, 14},
	"checksum":     {26, 1},
	"data":         {27, 512},
	"dataChecksum": {539, 1},
}

//
type record struct {
	muxed []byte
	block *raw.Block
}

//
func NewRecord(data []byte, isRaw bool) (*record, error) {

	r := &record{}
	var dmx []byte

	if isRaw {
		dmx = raw.Demux(data, false)
	} else {
		dmx = make([]byte, len(data))
		copy(dmx, data)
	}

	r.block = raw.NewBlock(recordIndex, dmx)
	r.muxed = raw.Mux(r.block.Data, false)

	return r, r.Validate()
}

//
func (r *record) Client() client.Client {
	return client.IF1
}

//
func (r *record) Muxed() []byte {
	return r.muxed
}

//
func (r *record) Demuxed() []byte {
	return r.block.Data
}

//
func (r *record) Flag() int {
	return int(r.block.GetByte("flag"))
}

//
func (r *record) Index() int {
	return int(r.block.GetByte("number"))
}

//
func (r *record) Length() int {
	return r.block.GetInt("length")
}

//
func (r *record) Name() string {
	return r.block.GetString("name")
}

//
func (r *record) Validate() error {

	want := r.block.GetByte("checksum")
	got := byte(r.block.Sum("header") % 255)

	if want != got {
		return fmt.Errorf(
			"invalid record descriptor check sum, want %d, got %d", want, got)
	}

	want = r.block.GetByte("dataChecksum")
	got = byte(r.block.Sum("data") % 255)

	if want != got {
		// FIXME: is this really correct?
		// calculate checksum only based on actual record data length
		// background: during ERASE, there always seems to be a bit set
		// somewhere, although all should be zero...
		if r.Flag() != 0 {
			return fmt.Errorf(
				"invalid record data check sum, want %d, got %d", want, got)
		}
	}

	return nil
}

//
func (r *record) Emit(w io.Writer) {
	io.WriteString(w, fmt.Sprintf(
		"\nRECORD: %+q - flag: %X, index: %d, length: %d\n",
		r.Name(), r.Flag(), r.Index(), r.Length()))
	d := hex.Dumper(w)
	defer d.Close()
	d.Write(r.block.Data)
}
