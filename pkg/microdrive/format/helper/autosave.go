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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/oqtadrive/pkg/microdrive/base"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/format"
)

//
const FlagModified = 0x01
const FlagWriteProtected = 0x02
const AutoSaveVersion = 1

const ixVersion = 0
const ixClient = 1
const ixFlags = 2

//
func AutoSave(drive int, cart base.Cartridge) error {

	if cart == nil || !cart.IsFormatted() || cart.IsAutoSaved() {
		return nil
	}

	start := time.Now()
	log.Infof("auto-saving drive %d", drive)

	fm, err := format.NewFormat(cart.Client().DefaultFormat())
	if err != nil {
		return err
	}

	_, file, err := autoSavePath(drive, true)
	if err != nil {
		return err
	}

	tmp := fmt.Sprintf("%s_", file)

	fd, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	out := bufio.NewWriter(fd)

	preamble := make([]byte, 3)

	var flags byte = 0
	if cart.IsModified() {
		flags |= FlagModified
	}
	if cart.IsWriteProtected() {
		flags |= FlagWriteProtected
	}

	preamble[ixVersion] = AutoSaveVersion
	preamble[ixClient] = byte(cart.Client())
	preamble[ixFlags] = flags

	if err := writeRaw(preamble, out); err != nil {
		return err
	}

	if err := fm.Write(cart, out, nil); err != nil {
		return err
	}

	if err := out.Flush(); err != nil {
		return err
	}

	if err := fd.Sync(); err != nil {
		return err
	}

	if err := fd.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmp, file); err != nil {
		return err
	}

	cart.SetAutoSaved(true)

	log.Debugf("auto-save took %v", time.Now().Sub(start))
	return nil
}

//
func AutoLoad(drive int) (base.Cartridge, error) {

	log.Infof("loading auto-save for drive %d", drive)

	_, file, err := autoSavePath(drive, false)
	if err != nil {
		return nil, err
	}

	fd, err := os.Open(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		log.Infof("no auto-save file for drive %d", drive)
		return nil, nil
	}
	defer fd.Close()

	in := bufio.NewReader(fd)

	preamble, err := readRaw(in, 64)
	if err != nil {
		return nil, fmt.Errorf("error reading preamble: %v", err)
	}

	if preamble[ixVersion] != AutoSaveVersion {
		return nil, fmt.Errorf(
			"incompatible auto-save version, want %d, got %d",
			AutoSaveVersion, preamble[ixVersion])
	}

	cl := client.Client(preamble[ixClient])
	fm, err := format.NewFormat(cl.DefaultFormat())
	if err != nil {
		return nil, err
	}

	if cart, err := fm.Read(in, true, false, nil); err != nil {
		return nil, err

	} else {
		cart.SetModified(preamble[ixFlags]&FlagModified != 0)
		cart.SetWriteProtected(preamble[ixFlags]&FlagWriteProtected != 0)
		cart.SetAutoSaved(true)
		return cart, nil
	}
}

//
func AutoRemove(drive int) error {

	if _, file, err := autoSavePath(drive, false); err != nil {
		return err
	} else {
		if err := os.Remove(file); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			log.Infof("removed auto-save for drive %d", drive)
		}
	}

	return nil
}

//
func autoSavePath(drive int, create bool) (string, string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	dir := filepath.Join(home, ".oqtadrive", fmt.Sprintf("%d", drive))

	if create {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", "", err
		}
	}

	return dir, filepath.Join(dir, "cart"), nil
}

//
func readRaw(in io.Reader, maxLen int) ([]byte, error) {

	buf := []byte{0, 0}
	if _, err := in.Read(buf); err != nil {
		return nil, err
	}

	length := int(buf[0]) + 256*int(buf[1])

	if length > maxLen {
		return nil, fmt.Errorf("max length %d, but have %d", maxLen, length)
	}

	ret := make([]byte, length)
	if _, err := in.Read(ret); err != nil {
		return nil, err
	}

	return ret, nil
}

//
func writeRaw(data []byte, out io.Writer) error {

	buf := []byte{byte(len(data) % 256), byte((len(data) >> 8))}

	if _, err := out.Write(buf); err != nil {
		return err
	}

	if _, err := out.Write(data); err != nil {
		return err
	}

	return nil
}
