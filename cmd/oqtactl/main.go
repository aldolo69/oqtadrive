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

package main

import (
	"fmt"
	"os"

	"github.com/xelalexv/oqtadrive/pkg/run"
)

//
var OqtaDriveVersion string

//
func synopsis() {
	fmt.Print(`
synopsis: oqtactl {serve|load|unload|save|ls|dump|resync|version} ...

run 'oqtactl {action} -h|--help' to see detailed info

`)
}

//
func version() {
	fmt.Printf("\nOqtaDrive %s\n\n", OqtaDriveVersion)
}

//
func main() {

	var action string
	var args []string

	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	switch action {

	case "serve":
		version()
		run.DieOnError(run.NewServe().Execute(args))

	case "load":
		run.DieOnError(run.NewLoad().Execute(args))

	case "unload":
		run.DieOnError(run.NewUnload().Execute(args))

	case "save":
		run.DieOnError(run.NewSave().Execute(args))

	case "ls":
		run.DieOnError(run.NewList().Execute(args))

	case "dump":
		run.DieOnError(run.NewDump().Execute(args))

	case "map":
		run.DieOnError(run.NewMap().Execute(args))

	case "resync":
		run.DieOnError(run.NewResync().Execute(args))

	case "version":
		version()

	case "":
		fallthrough
	case "-h":
		fallthrough
	case "--help":
		synopsis()

	default:
		run.Die("unknown action: %s\n", action)
	}
}
