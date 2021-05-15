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
var headerIndex = map[string][2]int{
	"flags":    {12, 1},
	"number":   {13, 1},
	"spares":   {14, 2},
	"name":     {16, 10},
	"header":   {12, 14},
	"checksum": {26, 1},
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
		dmx = raw.Demux(data, false)
	} else {
		dmx = make([]byte, len(data))
		copy(dmx, data)
	}

	h.block = raw.NewBlock(headerIndex, dmx)
	h.muxed = raw.Mux(h.block.Data, false)

	return h, h.Validate()
}

//
func (h *header) Client() client.Client {
	return client.IF1
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
func (h *header) Flags() byte {
	return h.block.GetByte("flags")
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
func (h *header) Checksum() int {
	return int(h.block.GetByte("checksum"))
}

//
func (h *header) Validate() error {

	want := h.block.GetByte("checksum")
	got := byte(h.block.Sum("header") % 255)

	if want != got {
		return fmt.Errorf(
			"invalid sector header check sum, want %d, got %d", want, got)
	}
	return nil
}

//
func (h *header) Emit(w io.Writer) {
	io.WriteString(w, fmt.Sprintf("\nHEADER: %+q - flag: %X, index: %d\n",
		h.Name(), h.Flags(), h.Index()))
	d := hex.Dumper(w)
	defer d.Close()
	d.Write(h.block.Data)
}
