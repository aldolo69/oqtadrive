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
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/xelalexv/oqtadrive/pkg/daemon"
)

//
const runnerHelpPrologue = ""
const runnerHelpEpilogue = `- When a flag can be set via environment variable, the variable name is given
  in parenthesis at the end of the flag explanation. Note however that a flag,
  when specified overrides an environment variable.
`

/*
	NewRunner creates a base runner for commands to use. The parameters are
	passed to the base command wrapped by this runner.
*/
func NewRunner(use, short, long, helpPrologue, helpEpilogue string,
	exec func() error) *Runner {
	return &Runner{
		Command: *NewCommand(
			use, short, long, helpPrologue, helpEpilogue, exec),
	}
}

//
type Runner struct {
	//
	Command
	//
	Port int
}

//
func (r *Runner) AddBaseSettings() {
	// Implementation Note: This cannot be included in NewRunner, but rather has
	// to be called from the top level command type. Otherwise, we will confuse
	// Cobra/Viper and the settings will not be filled with their values.
	r.AddSetting(&r.Port, "port", "p", "OQTADRIVE_PORT", 8888,
		"port of daemon's API server", false)
}

//
func (r *Runner) apiCall(method, path string, json bool,
	body io.Reader) (io.ReadCloser, error) {

	client := &http.Client{}
	// FIXME: parameterize server
	req, err := http.NewRequest(
		method, fmt.Sprintf("http://127.0.0.1:%d%s", r.Port, path), body)
	if err != nil {
		return nil, err
	}

	if json {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
	} else {
		req.Header.Add("Content-Type", "text/plain")
		req.Header.Add("Accept", "text/plain")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

//
func validateDrive(d int) error {
	if d < 1 || d > daemon.DriveCount {
		return fmt.Errorf(
			"invalid drive number: %d; valid numbers are 1 through %d",
			d, daemon.DriveCount)
	}
	return nil
}

//
func getExtension(file string) string {
	return strings.TrimPrefix(filepath.Ext(file), ".")
}
