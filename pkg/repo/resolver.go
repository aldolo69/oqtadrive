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

package repo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

//
const PrefixRepoRef = "repo://"

//
func newFileSource(file string) (*fileSource, error) {
	if f, err := os.Open(file); err != nil {
		return nil, err
	} else {
		return &fileSource{file: f, reader: bufio.NewReader(f)}, nil
	}
}

//
type fileSource struct {
	file   *os.File
	reader io.Reader
}

//
func (fs *fileSource) Read(p []byte) (n int, err error) {
	return fs.reader.Read(p)
}

//
func (fs *fileSource) Close() error {
	return fs.file.Close()
}

//
func Resolve(ref, repo string) (io.ReadCloser, error) {

	log.WithFields(log.Fields{
		"reference":  ref,
		"repository": repo,
	}).Debug("resolving ref")

	if strings.HasPrefix(ref, PrefixRepoRef) {
		if repo == "" {
			return nil, fmt.Errorf("cartridge repository is not enbaled")
		}
		return newFileSource(filepath.Join(repo, ref[len(PrefixRepoRef):]))
	}

	return nil, fmt.Errorf("loading by reference not yet implemented")
}

//
func IsReference(r string) bool {
	return strings.HasPrefix(r, PrefixRepoRef)
}
