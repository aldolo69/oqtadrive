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
var headerIndex = map[string][2]int{
	"flag":     {12, 1},
	"number":   {13, 1},
	"name":     {14, 10},
	"random":   {24, 2},
	"header":   {12, 14},
	"checksum": {26, 2},
}

//
type header struct {
	muxed []byte
	block *raw.Block
}

//
func NewHeader(data []byte, isRaw bool) (*header, error) {

	h := &header{}
	var dmx []byte

	if isRaw {
		dmx = raw.Demux(data, true)
	} else {
		dmx = make([]byte, len(data))
		copy(dmx, data)
	}

	h.block = raw.NewBlock(headerIndex, dmx)
	h.muxed = raw.Mux(h.block.Data, true)

	return h, h.Validate()
}

//
func (h *header) Client() client.Client {
	return client.QL
}

//
func (h *header) Muxed() []byte {
	return h.muxed
}

//
func (h *header) Demuxed() []byte {
	return h.block.Data
}

//
func (h *header) Flag() int {
	return int(h.block.GetByte("flag"))
}

//
func (h *header) Index() int {
	return int(h.block.GetByte("number"))
}

//
func (h *header) Name() string {
	return h.block.GetString("name")
}

//
func (h *header) CalculateChecksum() int {
	return toQLCheckSum(h.block.Sum("header"))
}

//
func (h *header) Validate() error {
	if err := verifyQLCheckSum(
		h.CalculateChecksum(), h.block.GetInt("checksum")); err != nil {
		return fmt.Errorf("invalid sector header check sum: %v", err)
	}
	return nil
}

//
func (h *header) Emit(w io.Writer) {
	io.WriteString(w, fmt.Sprintf(
		"\nHEADER: %+q - flag: %X, index: %d\n", h.Name(), h.Flag(), h.Index()))
	d := hex.Dumper(w)
	defer d.Close()
	d.Write(h.block.Data)
}
