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
const RefSchemaRepo = "repo"
const RefSchemaHttp = "http"
const RefSchemaHttps = "https"

//
func Resolve(ref, repo string) (io.ReadCloser, error) {

	log.WithFields(log.Fields{
		"reference":  ref,
		"repository": repo,
	}).Debug("resolving ref")

	ok, parts, err := ParseReference(ref)

	if !ok {
		return nil, fmt.Errorf("not a reference: %s", ref)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid reference: %v", err)
	}

	switch parts[0] {

	case RefSchemaRepo:
		if repo == "" {
			return nil, fmt.Errorf("cartridge repository is not enbaled")
		}
		return newFileSource(filepath.Join(repo, parts[1]))

	case RefSchemaHttp:
		fallthrough
	case RefSchemaHttps:
		return nil, fmt.Errorf("loading by HTTP reference not yet implemented")
	}

	return nil, fmt.Errorf("invalid reference: %s", ref)
}

//
func ParseReference(ref string) (bool, []string, error) {

	parts := strings.SplitN(ref, "://", 2)

	if len(parts) < 2 {
		return false, nil, nil
	}

	var err error
	switch parts[0] {
	case RefSchemaRepo:
	case RefSchemaHttp:
	case RefSchemaHttps:
	default:
		err = fmt.Errorf("unsupported reference schema: %s", parts[0])
	}

	return true, parts, err
}

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
