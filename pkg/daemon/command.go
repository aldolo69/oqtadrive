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

	log "github.com/sirupsen/logrus"
)

//
const CmdHello = 'h'     // hello (send/receive to/from IF1/QL)
const CmdVersion = 'v'   // protocol version (send/receive to/from IF1/QL)
const CmdPing = 'P'      // ping/pong (send/receive to/from IF1/QL)
const CmdStatus = 's'    // get drive state (send to IF1/QL)
const CmdGet = 'g'       // get sector (send to IF1/QL)
const CmdPut = 'p'       // put sector (receive from IF1/QL)
const CmdVerify = 'y'    // verify sector bytes sent for previous get
const CmdTimeStart = 't' // start stop watch
const CmdTimeEnd = 'q'   // stop stop watch
const CmdMap = 'm'       // h/w drive mapping (receive from IF1/QL)
const CmdDebug = 'd'     // debug message (receive from IF1/QL)
const CmdResync = 'r'    // resync with adapter (send to IF1/QL)

const MaskIF1 = 1
const MaskQL = 2

var ping = []byte("Ping")
var pong = []byte("Pong")

//
func newCommand(data []byte) *command {
	return &command{data: data}
}

//
type command struct {
	data []byte
}

//
func (c *command) dispatch(d *Daemon) error {

	switch c.cmd() {

	case CmdHello:
		d.synced = false
		return nil

	case CmdPing:
		if bytes.Equal(c.data, ping) {
			log.Debugf("ping from %s", d.conduit.client)
			if err := d.conduit.send(pong); err != nil {
				return err
			}
			d.processControl()
		}
		return nil

	case CmdStatus:
		return c.status(d)

	case CmdGet:
		return c.get(d)

	case CmdPut:
		return c.put(d)

	case CmdDebug:
		return c.debug(d)

	case CmdTimeStart:
		return c.timer(true, d)

	case CmdTimeEnd:
		return c.timer(false, d)

	case CmdMap:
		return c.driveMap(d)

		// case CMD_VERIFY: FIXME
	}

	return fmt.Errorf("unknown command: %v", c.data)
}

//
func (c *command) cmd() byte {
	return c.data[0]
}

//
func (c *command) arg(ix int) byte {
	if 0 <= ix && ix < len(c.data)-1 {
		return c.data[ix+1]
	}
	return 0
}

// drive returns the 1-based drive number, or an error if the drive number
// contained in this command is not within 1 through 8
func (c *command) drive() (int, error) {
	drive := c.arg(0)
	if drive < 1 || drive > 8 {
		return -1, fmt.Errorf("illegal drive number: %d", drive)
	}
	return int(drive), nil
}
