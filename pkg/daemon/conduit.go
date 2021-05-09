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

package daemon

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/jacobsa/go-serial/serial"
	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/microdrive/pkg/microdrive"
	"github.com/xelalexv/microdrive/pkg/microdrive/abstract"
	"github.com/xelalexv/microdrive/pkg/microdrive/if1"
	"github.com/xelalexv/microdrive/pkg/microdrive/ql"
)

//
const commandLength = 4
const sendBufferLength = 1024
const receiveBufferLength = 1024

const headerFlagIndex = 12

//
var helloDaemon = []byte("hlod")
var helloIF1 = []byte("hloi")
var helloQL = []byte("hloq")

//
type conduit struct {
	//
	headerLengthMux int
	recordLengthMux int
	//
	client microdrive.Client
	port   io.ReadWriteCloser
	//
	sendBuf []byte
}

//
func newConduit(port string) (*conduit, error) {
	ret := &conduit{
		sendBuf: make([]byte, sendBufferLength),
	}
	var err error
	ret.port, err = openPort(port)
	return ret, err
}

//
func openPort(p string) (io.ReadWriteCloser, error) {
	return serial.Open(serial.OpenOptions{
		PortName:        p,
		BaudRate:        1000000,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	})
}

//
func (c *conduit) close() error {
	return c.port.Close()
}

//
func (c *conduit) syncOnHello() error {

	log.Info("syncing with adapter")
	hello := make([]byte, commandLength)

	for {
		if c.isHello(hello) {
			break
		}
		shiftLeft(hello)
		if err := c.receive(hello[len(hello)-1:]); err != nil {
			return err
		}
	}

	var cmd *command
	var err error

	for { // find last live hello from adapter
		start := time.Now()
		if cmd, err = c.receiveCommand(); err != nil {
			return err
		}
		if cmd.cmd() == CmdHello && time.Now().Sub(start) > 500*time.Millisecond {
			break
		}
		log.Debugf("discarding command: %v", cmd.data)
	}

	if err := c.send(helloDaemon); err != nil {
		return fmt.Errorf("error sending daemon hello: %v", err)
	}

	log.Infof("synced with %s", c.client)
	return nil
}

//
func (c *conduit) isHello(h []byte) bool {

	if bytes.Equal(h, helloIF1) {
		c.client = microdrive.IF1
		c.headerLengthMux = if1.HeaderLengthMux
		c.recordLengthMux = if1.RecordLengthMux

	} else if bytes.Equal(h, helloQL) {
		c.client = microdrive.QL
		c.headerLengthMux = ql.HeaderLengthMux
		c.recordLengthMux = ql.RecordLengthMux

	} else {
		return false
	}

	return true
}

//
func (c *conduit) receive(data []byte) error {
	_, err := io.ReadFull(c.port, data)
	return err
}

//
func (c *conduit) send(data []byte) error {
	_, err := c.port.Write(data)
	return err
}

//
func (c *conduit) receiveCommand() (*command, error) {
	data := make([]byte, commandLength)
	if err := c.receive(data); err != nil {
		return nil, err
	}
	return newCommand(data), nil
}

//
func (c *conduit) fillBlock(s *abstract.Sector) int {

	header := s.Header().Muxed()
	copy(c.sendBuf, header)

	record := s.Record().Muxed()
	copy(c.sendBuf[len(header):], record)

	return len(header) + len(record)
}

//
func (c *conduit) sendBlock(length int) error {
	if _, err := c.port.Write(c.sendBuf[0:length]); err != nil {
		return fmt.Errorf("error sending block: %v", err)
	}
	return nil
}

//
func (c *conduit) receiveBlock() ([]byte, error) {

	var raw []byte

	raw = make([]byte, receiveBufferLength)

	if err := c.receive(raw[c.fillPreamble(raw):c.headerLengthMux]); err != nil {
		return nil, fmt.Errorf("error reading block header: %v", err)
	}

	// unknown length, need to check what is being sent
	if rem := c.remainingBytes(raw); rem == 0 {
		log.Trace("header block received")
		raw = raw[:c.headerLengthMux]

	} else {
		log.Trace("record block received")
		end := c.headerLengthMux + rem
		if err := c.receive(raw[c.headerLengthMux:end]); err != nil {
			return nil, fmt.Errorf("error reading block: %v", err)
		}
		raw = raw[:end]
	}

	stop := make([]byte, 4)
	if err := c.receive(stop); err != nil {
		return nil, fmt.Errorf("error reading block stop: %v", err)
	}

	shift := stop[3]

	if shift > 3 {
		return nil, fmt.Errorf(
			"corrupted block, excessive stop shift '%d'", shift)
	} else if shift > 0 {
		if err := c.receive(stop[:shift]); err != nil {
			return nil, fmt.Errorf("error aligning to block end: %v", err)
		}
	}

	return raw, nil
}

//
func (c *conduit) fillPreamble(raw []byte) int {
	if len(raw) < 12 {
		return 0
	}
	for ix := 0; ix < 10; ix++ {
		raw[ix] = 0
	}
	if c.client == microdrive.QL {
		raw[10] = 0xf0
	} else {
		raw[10] = 0x0f
	}
	raw[11] = 0xff
	return 12
}

/*
	The section flag byte is at position 12, right after the 12 bytes lead in.
	For a header section, this flag byte has a particular value, depending on
	the client to which we're connected. We use that to detect what kind of
	section we're currently receiving. If it's a header, we're done reading,
	if it's a record, we still need to read the record data.
*/
func (c *conduit) remainingBytes(raw []byte) int {

	// The bytes come in with DATA2 bits in high nibble, DATA1 bits in low
	// nibble, in reversed bit order.

	// for QL, track 1 is ahead of track 2, just the opposite of IF1.
	if c.client == microdrive.QL {
		// QL
		// raw byte | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10| 11| 12| 13|
		//          --------------------------------------------------------
		//   DATA1: |l0 |h0 |l2 |h2 |l4 |h4 |l6 |h6 |l8 |h8 |l10|h10|l12|h12| high
		//   DATA2: |   |l1 |h1 |l3 |h3 |l5 |h5 |l7 |h7 |l9 |h9 |l11|h11|   | low
		//

		// Note: bit order in flag and sum is reversed
		flag := (raw[headerFlagIndex] & 0xf0) | ((raw[headerFlagIndex+1] & 0xf0) >> 4)

		if flag == 0xff {
			return 0 // header - nothing more to read
		}

		ret := c.recordLengthMux - c.headerLengthMux // standard record

		// pos 24  hex:  5f  5a  5a  5a  5a                '_ZZZZ'
		flag = (raw[24] & 0xf0) | ((raw[25] & 0xf0) >> 4)
		num := ((raw[25] & 0x0f) << 4) | (raw[26] & 0x0f)
		chL := (raw[26] & 0xf0) | ((raw[27] & 0xf0) >> 4)
		chH := ((raw[27] & 0x0f) << 4) | (raw[28] & 0x0f)

		if flag == 0x55 && num == 0xaa && chL == 0x55 && chH == 0xaa {
			// 0xAA55 in both flag+num and the two byte checksum of a record
			// header signify a record written during format, which is longer
			// than a standard record.
			ret += ql.FormatExtraBytes
		}
		return ret

	} else {
		// IF1
		// raw byte | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10| 11| 12| 13|
		//          --------------------------------------------------------
		//   DATA1: |   |l1 |h1 |l3 |h3 |l5 |h5 |l7 |h7 |l9 |h9 |l11|h11|   | high
		//   DATA2: |l0 |h0 |l2 |h2 |l4 |h4 |l6 |h6 |l8 |h8 |l10|h10|l12|h12| low
		//
		if (raw[headerFlagIndex] & 0x0f) == 0x08 {
			return 0 // header - nothing more to read
		}
		return c.recordLengthMux - c.headerLengthMux // record
	}
}

// FIXME: validate
func (c *conduit) verifyBlock(expected []byte) {

	//l := c.blockLength()
	l := 0 // FIXME
	if l != len(expected) {
		log.Errorf("length mismatch, want %d, got %d", len(expected), l)
		return
	}

	raw := make([]byte, l)
	if _, err := io.ReadFull(c.port, raw); err != nil {
		log.Errorf("read error: %v", err)
		return
	}

	errors := 0
	for ix := range expected {
		if expected[ix] != raw[ix] {
			errors++
		}
	}

	if errors == 0 {
		log.Info("OK")
	} else {
		log.Errorf("NG: %d", errors)
	}
}

//
func shiftLeft(buf []byte) {
	if len(buf) > 1 {
		for ix := 0; ix < len(buf)-1; ix++ {
			buf[ix] = buf[ix+1]
		}
	}
}
