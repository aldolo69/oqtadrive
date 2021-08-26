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
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/oqtadrive/pkg/control"
	"github.com/xelalexv/oqtadrive/pkg/daemon"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
)

//
func NewServe() *Serve {

	s := &Serve{}
	s.Runner = *NewRunner(
		`serve -d|--device {device} [-a|--address {address}]  [-c|--client {if1|ql}]
      [-r|--repo {repo base folder}]`,
		"daemon & API server command",
		`Use the serve command for running the adapter daemon and API server. Optionally, you
can specify whether the adapter should be configured for Interface 1 or QL after
connecting to it. Note however that if the adapter is forced to a particular client
in its configuration, then this cannot be changed.`,
		"", `- Logging can be configured with these environment variables:

  LOG_FORMAT		set to 'json' for JSON logging
  LOG_FORCE_COLORS	set to non-empty for forcing colorized log entries
  LOG_METHODS		set to non-empty for including methods in log
  LOG_LEVEL		panic, fatal, error, warn, info, debug, trace

`+runnerHelpEpilogue, s.Run)

	s.AddBaseSettings()
	s.AddSetting(&s.Device, "device", "d", "OQTADRIVE_DEVICE", nil,
		"serial port device for adapter", true)
	s.AddSetting(&s.Client, "client", "c", "", nil,
		"client type, 'if1' or 'ql'", false)
	s.AddSetting(&s.Repository, "repo", "r", "", nil,
		`cartridge repo base folder; when omitted, loading
cartridges from daemon host's file system is prohibited`, false)

	return s
}

//
type Serve struct {
	//
	Runner
	//
	Device     string
	Client     string
	Repository string
}

//
func (s *Serve) Run() error {

	s.ParseSettings()

	cl := client.UNKNOWN
	if s.Client != "" {
		if cl = client.GetClient(s.Client); cl == client.UNKNOWN {
			return fmt.Errorf("unknown client type: %s", s.Client)
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	d := daemon.NewDaemon(s.Device, cl)
	go func() {
		defer wg.Done()
		err := d.Serve()
		if err != nil && err != daemon.ErrDaemonStopped {
			log.Errorf("daemon closed with error: %v", err)
		} else {
			log.Info("daemon stopped")
		}
	}()

	api := control.NewAPIServer(s.Address, s.Repository, d)
	go func() {
		defer wg.Done()
		if err := api.Serve(); err != nil {
			log.Errorf("API server closed with error: %v", err)
		} else {
			log.Info("API server stopped")
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sigCount := 0
	done := make(chan bool)

	for {

		select {

		case sig := <-sigs: // interrupt signal
			log.WithField("signal", sig).Info("signal received")
			sigCount++

			switch sigCount {

			case 1:
				go func() {
					log.Info("shutting down, hit Ctrl-C twice to force exit...")
					api.Stop()
					d.Stop()
					wg.Wait()
					log.Info("OqtaDrive stopped")
					done <- true
				}()

			case 2:
				log.Warn("shutdown in progress, hit Ctrl-C again to force exit")

			default:
				log.Warn("forcing daemon to stop immediately")
				os.Exit(1)
			}

		case <-done: // shutdown sequence complete
			return nil
		}
	}
}
