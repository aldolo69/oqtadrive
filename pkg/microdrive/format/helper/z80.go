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

package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

//
const BinaryLinux = "z80onmdr"
const BinaryWindows = "Z80onMDR.exe"

// reads Z80 snapshot file and converts it on the fly using the z80onmdr tool
//
func Z80toMDR(in string) (string, error) {

	f, err := os.Stat(in)
	if err != nil {
		return "", err
	}
	if f.IsDir() {
		return "", fmt.Errorf("%s is not a file", in)
	}

	tmp, err := ioutil.TempFile("", "oqtadrive.*")
	if err != nil {
		return "", err
	}
	tmp.Close()

	name := strings.TrimSuffix(strings.ToUpper(in), ".Z80")
	if len(name) > 10 {
		name = name[:10]
	}

	if err = runZ80onMDR("-f", tmp.Name(), "-m", name, in,
		"-n", strings.ToLower(name)); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return fmt.Sprintf("%s.mdr", tmp.Name()), nil
}

//
func runZ80onMDR(args ...string) error {

	var bin string

	switch runtime.GOOS {
	case "linux":
		bin = BinaryLinux
	case "windows":
		bin = BinaryWindows
	default:
		return fmt.Errorf("Z80onMDR not supported on platform %s", runtime.GOOS)
	}

	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf(`
You need to install Z80onMDR on your system and add it to PATH.
To get Z80onMDR, visit https://www.tomdalby.com/other/z80onmdr.html`)
	}

	fmt.Printf("\ninvoking %s...\n\n", bin)
	defer fmt.Println()

	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
