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
	"encoding/hex"
	"fmt"
	"io"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/raw"
)

//
var recordIndex = map[string][2]int{
	"flag":              {12, 1},
	"number":            {13, 1},
	"header":            {12, 2},
	"headerChecksum":    {14, 2},
	"data":              {24, 512},
	"dataChecksum":      {536, 2},
	"extraData":         {538, 84},
	"extraDataChecksum": {622, 2},
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
		dmx = raw.Demux(data, true)
	} else {
		dmx = make([]byte, len(data))
		copy(dmx, data)
	}

	r.block = raw.NewBlock(recordIndex, dmx)
	r.muxed = raw.Mux(r.block.Data, true)

	return r, r.Validate()
}

//
func (r *record) Client() client.Client {
	return client.QL
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
	return -1
}

//
func (r *record) Name() string {
	return ""
}

//
func (r *record) CalculateHeaderChecksum() int {
	return toQLCheckSum(r.block.Sum("header"))
}

//
func (r *record) CalculateDataChecksum() int {
	return toQLCheckSum(r.block.Sum("data"))
}

//
func (r *record) CalculateExtraDataChecksum() int {
	return toQLCheckSum(r.block.Sum("extraData"))
}

//
func (r *record) Validate() error {

	if err := verifyQLCheckSum(r.CalculateHeaderChecksum(),
		r.block.GetInt("headerChecksum")); err != nil {
		return fmt.Errorf("invalid record header check sum: %v", err)
	}

	if err := verifyQLCheckSum(r.CalculateDataChecksum(),
		r.block.GetInt("dataChecksum")); err != nil {
		return fmt.Errorf("invalid record data check sum: %v", err)
	}

	if r.Flag() == 0xaa && r.Index() == 0x55 &&
		r.block.GetInt("headerChecksum") == 0x55aa {
		if err := verifyQLCheckSum(r.CalculateExtraDataChecksum(),
			r.block.GetInt("extraDataChecksum")); err != nil {
			return fmt.Errorf("invalid record extra data check sum: %v", err)
		}
	}

	return nil
}

//
func (r *record) Emit(w io.Writer) {
	io.WriteString(w, fmt.Sprintf(
		"\nRECORD: flag: %X, index: %d\n", r.Flag(), r.Index()))
	d := hex.Dumper(w)
	defer d.Close()
	d.Write(r.block.Data)
}
