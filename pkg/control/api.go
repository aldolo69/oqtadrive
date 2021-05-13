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

package control

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/microdrive/pkg/daemon"
	"github.com/xelalexv/microdrive/pkg/microdrive/format"
)

//
type APIServer interface {
	Serve() error
}

//
func NewAPIServer(port int, d *daemon.Daemon) APIServer {
	return &api{port: port, daemon: d}
}

//
type api struct {
	port   int
	daemon *daemon.Daemon
}

//
func (a *api) Serve() error {

	router := mux.NewRouter().StrictSlash(true)
	addRoute(router, "load", "PUT", "/drive/{drive:[1-8]}", a.load)
	addRoute(router, "unload", "GET", "/drive/{drive:[1-8]}/unload", a.unload)
	addRoute(router, "save", "GET", "/drive/{drive:[1-8]}", a.save)
	addRoute(router, "dump", "GET", "/drive/{drive:[1-8]}/dump", a.dump)
	addRoute(router, "list", "GET", "/list", a.list)

	addr := fmt.Sprintf(":%d", a.port)
	log.Infof("OqtaDrive API starts listening on %s", addr)
	return http.ListenAndServe(addr, router)
}

//
func addRoute(r *mux.Router, name, method, pattern string,
	handler http.HandlerFunc) {
	r.Methods(method).
		Path(pattern).
		Name(name).
		Handler(requestLogger(handler, name))
}

//
func requestLogger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("API BEGIN | %s\t%s\t%s\t%s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			name,
		)
		start := time.Now()
		inner.ServeHTTP(w, r)
		log.Infof("API END   | %s\t%s\t%s\t%s\t%s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

//
func setHeaders(h http.Header, json bool) {
	if json {
		h.Set("Content-Type", "application/json; charset=UTF-8")
	} else {
		h.Set("Content-Type", "text/plain; charset=UTF-8")
	}
}

//
func handleError(e error, statusCode int, w http.ResponseWriter) bool {

	if e == nil {
		return false
	}

	log.Errorf("%v", e)

	setHeaders(w.Header(), false)
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(fmt.Sprintf("%v\n", e))); err != nil {
		log.Errorf("problem writing error: %v", err)
	}

	return true
}

//
func sendReply(body []byte, statusCode int, w http.ResponseWriter) {
	setHeaders(w.Header(), false)
	w.WriteHeader(statusCode)
	if _, err := fmt.Fprintf(w, "%s\n", body); err != nil {
		log.Errorf("problem sending reply: %v", err)
	}
}

//
func sendStreamReply(r io.Reader, statusCode int, w http.ResponseWriter) {
	setHeaders(w.Header(), false)
	w.WriteHeader(statusCode)
	if _, err := io.Copy(w, r); err != nil {
		log.Errorf("problem sending reply: %v", err)
	}
}

//
func sendJSONReply(obj interface{}, statusCode int, w http.ResponseWriter) {
	setHeaders(w.Header(), true)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		log.Errorf("problem writing error: %v", err)
	}
}

//
func (a *api) list(w http.ResponseWriter, req *http.Request) {

	var list []*Cartridge
	strList := "\nDRIVE CARTRIDGE       STATE"

	for drive := 1; drive <= daemon.DriveCount; drive++ {

		msg := "<no cartridge>"

		cart, ok := a.daemon.GetCartridge(drive)

		if cart != nil {
			c := NewCartridge(cart)
			list = append(list, c)
			msg = c.String()
			cart.Unlock()

		} else if !ok {
			msg = "<drive busy>"
		}

		strList += fmt.Sprintf("\n  %d   %s", drive, msg)
	}

	if req.Header.Get("Content-Type") == "application/json" {
		sendJSONReply(list, http.StatusOK, w)
	} else {
		sendReply([]byte(strList), http.StatusOK, w)
	}
}

//
func (a *api) load(w http.ResponseWriter, req *http.Request) {

	drive := getDrive(w, req)
	if drive == -1 {
		return
	}

	reader := getFormat(w, req)
	if reader == nil {
		return
	}

	cart, err := reader.Read(io.LimitReader(req.Body, 1048576), false)
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}
	if handleError(req.Body.Close(), http.StatusInternalServerError, w) {
		return
	}

	force := req.URL.Query().Get("force")

	if err := a.daemon.SetCartridge(drive, cart, force == "true"); err != nil {
		if strings.Contains(err.Error(), "could not lock") {
			handleError(fmt.Errorf("drive %d busy", drive), http.StatusLocked, w)
		} else if strings.Contains(err.Error(), "is modified") {
			handleError(fmt.Errorf(
				"cartridge in drive %d is modified", drive), http.StatusConflict, w)
		} else {
			handleError(err, http.StatusInternalServerError, w)
		}

	} else {
		sendReply([]byte(
			fmt.Sprintf("loaded data into drive %d", drive)), http.StatusOK, w)
	}
}

//
func (a *api) unload(w http.ResponseWriter, req *http.Request) {

	drive := getDrive(w, req)
	if drive == -1 {
		return
	}

	force := req.URL.Query().Get("force")

	if err := a.daemon.UnloadCartridge(drive, force == "true"); err != nil {
		if strings.Contains(err.Error(), "could not lock") {
			handleError(fmt.Errorf("drive %d busy", drive), http.StatusLocked, w)
		} else if strings.Contains(err.Error(), "is modified") {
			handleError(fmt.Errorf(
				"cartridge in drive %d is modified", drive), http.StatusConflict, w)
		} else {
			handleError(err, http.StatusInternalServerError, w)
		}

	} else {
		sendReply([]byte(
			fmt.Sprintf("unloaded drive %d", drive)), http.StatusOK, w)
	}
}

//
func (a *api) save(w http.ResponseWriter, req *http.Request) {

	drive := getDrive(w, req)
	if drive == -1 {
		return
	}

	cart, ok := a.daemon.GetCartridge(drive)

	if !ok {
		handleError(fmt.Errorf("drive %d busy", drive), http.StatusLocked, w)
		return
	}

	if cart == nil {
		handleError(fmt.Errorf("no cartridge in drive %d", drive),
			http.StatusUnprocessableEntity, w)
		return
	}

	defer cart.Unlock()

	writer := getFormat(w, req)
	if writer == nil {
		return
	}

	var out bytes.Buffer
	if handleError(writer.Write(cart, &out), http.StatusInternalServerError, w) {
		return
	}

	cart.SetModified(false)
	w.WriteHeader(http.StatusOK)
	w.Write(out.Bytes())
}

//
func (a *api) dump(w http.ResponseWriter, req *http.Request) {

	drive := getDrive(w, req)
	if drive == -1 {
		return
	}

	cart, ok := a.daemon.GetCartridge(drive)

	if !ok {
		handleError(fmt.Errorf("drive %d busy", drive), http.StatusLocked, w)
		return
	}

	if cart == nil {
		handleError(fmt.Errorf("no cartridge in drive %d", drive),
			http.StatusUnprocessableEntity, w)
		return
	}

	defer cart.Unlock()

	read, write := io.Pipe()

	go func() {
		cart.Emit(write)
		write.Close()
	}()

	sendStreamReply(read, http.StatusOK, w)
}

//
func getDrive(w http.ResponseWriter, req *http.Request) int {
	vars := mux.Vars(req)
	drive, err := strconv.Atoi(vars["drive"])
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return -1
	}
	return drive
}

//
func getFormat(w http.ResponseWriter, req *http.Request) format.ReaderWriter {
	ret, err := format.NewFormat(req.URL.Query().Get("type"))
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return nil
	}
	return ret
}
