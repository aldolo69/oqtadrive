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
	"flags":             {12, 1},
	"number":            {13, 1},
	"header":            {12, 2},
	"headerChecksum":    {14, 2},
	"data":              {24, 512},
	"length":            {24, 4},
	"name":              {38, 38},
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
	r.mux()

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
func (r *record) mux() {
	r.muxed = raw.Mux(r.block.Data, true)
}

//
func (r *record) Flags() byte {
	return r.block.GetByte("flags")
}

//
func (r *record) Index() int {
	return int(r.block.GetByte("number"))
}

//
func (r *record) Length() int {
	l := r.block.GetSlice("length")
	if len(l) != 4 {
		return -1
	}
	return int(int32(l[0])<<24 | int32(l[1])<<16 | int32(l[2])<<8 | int32(l[3]))
}

//
func (r *record) Name() string {
	name := r.block.GetSlice("name")
	if len(name) == 38 {
		l := (int(name[0]) << 8) | int(name[1])
		if 0 < l && l <= len(name)-2 {
			return string(name[2 : 2+l])
		}
	}
	return ""
}

//
func (r *record) HeaderChecksum() int {
	return r.block.GetInt("headerChecksum")
}

//
func (r *record) Data() []byte {
	return r.block.GetSlice("data")
}

//
func (r *record) DataChecksum() int {
	return r.block.GetInt("dataChecksum")
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
func (r *record) fixHeaderChecksum() error {
	if err := r.block.SetInt(
		"headerChecksum", r.CalculateHeaderChecksum()); err != nil {
		return err
	}
	return nil
}

//
func (r *record) fixDataChecksum() error {
	if err := r.block.SetInt(
		"dataChecksum", r.CalculateDataChecksum()); err != nil {
		return err
	}
	return nil
}

//
func (r *record) fixExtraDataChecksum() error {
	if err := r.block.SetInt(
		"extraDataChecksum", r.CalculateExtraDataChecksum()); err != nil {
		return err
	}
	return nil
}

//
func (r *record) FixChecksums() error {
	if err := r.fixHeaderChecksum(); err != nil {
		return err
	}
	if err := r.fixDataChecksum(); err != nil {
		return err
	}
	if err := r.fixExtraDataChecksum(); err != nil {
		return err
	}
	r.mux()
	return r.Validate()
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

	if r.Flags() == 0xaa && r.Index() == 0x55 &&
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
	io.WriteString(w, fmt.Sprintf("\nRECORD: flag: %X, index: %d\n",
		r.Flags(), r.Index()))
	d := hex.Dumper(w)
	defer d.Close()
	d.Write(r.block.Data)
}
