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

package run

import (
	"github.com/xelalexv/microdrive/pkg/control"
	"github.com/xelalexv/microdrive/pkg/daemon"
)

//
func NewServe() *Serve {

	s := &Serve{}
	s.Runner = *NewRunner(
		"serve -d|--device {device} [-p|--port {port}]",
		"daemon & API server command",
		"Use the serve command for running the adapter daemon and API server.",
		"", `- Logging can be configured with these environment variables:

  LOG_FORMAT		set to 'json' for JSON logging
  LOG_FORCE_COLORS	set to non-empty for forcing colorized log entries
  LOG_METHODS		set to non-empty for including methods in log
  LOG_LEVEL		panic, fatal, error, warn, info, debug, trace

`+runnerHelpEpilogue, s.Run)

	s.AddBaseSettings()
	s.AddSetting(&s.Device, "device", "d", "OQTADRIVE_DEVICE", nil,
		"serial port device for adapter", true)

	return s
}

//
type Serve struct {
	//
	Runner
	//
	Device string
}

// FIXME: graceful shutdown
func (s *Serve) Run() error {

	s.ParseSettings()

	d := daemon.NewDaemon(s.Device)
	go d.Serve()

	api := control.NewAPIServer(s.Port, d)
	return api.Serve()
}
