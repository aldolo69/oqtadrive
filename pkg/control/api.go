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

	"github.com/xelalexv/oqtadrive/pkg/daemon"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/format"
)

//
type APIServer interface {
	Serve() error
}

//
func NewAPIServer(addr string, d *daemon.Daemon) APIServer {
	return &api{address: addr, daemon: d}
}

//
type api struct {
	address string
	daemon  *daemon.Daemon
}

//
func (a *api) Serve() error {

	router := mux.NewRouter().StrictSlash(true)

	addRoute(router, "status", "GET", "/status", a.status)
	addRoute(router, "ls", "GET", "/list", a.list)
	addRoute(router, "load", "PUT", "/drive/{drive:[1-8]}", a.load)
	addRoute(router, "unload", "GET", "/drive/{drive:[1-8]}/unload", a.unload)
	addRoute(router, "save", "GET", "/drive/{drive:[1-8]}", a.save)
	addRoute(router, "dump", "GET", "/drive/{drive:[1-8]}/dump", a.dump)
	addRoute(router, "map", "GET", "/map", a.getDriveMap)
	addRoute(router, "map", "PUT", "/map", a.setDriveMap)
	addRoute(router, "drivels", "GET", "/drive/{drive:[1-8]}/list", a.driveList)

	addr := a.address
	if len(strings.Split(addr, ":")) < 2 {
		addr = fmt.Sprintf("%s:8888", a.address)
	}
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

		log.WithFields(log.Fields{
			"remote": r.RemoteAddr,
			"method": r.Method,
			"path":   r.RequestURI,
		}).Debugf("API BEGIN | %s", name)

		start := time.Now()
		inner.ServeHTTP(w, r)

		log.WithFields(log.Fields{
			"remote":   r.RemoteAddr,
			"method":   r.Method,
			"path":     r.RequestURI,
			"duration": time.Since(start),
		}).Debugf("API END   | %s", name)
	})
}

//
func (a *api) status(w http.ResponseWriter, req *http.Request) {

	stat := &Status{}
	for drive := 1; drive <= daemon.DriveCount; drive++ {
		stat.Add(a.daemon.GetStatus(drive))
	}

	if wantsJSON(req) {
		sendJSONReply(stat, http.StatusOK, w)
	} else {
		sendReply([]byte(stat.String()), http.StatusOK, w)
	}
}

//
func (a *api) list(w http.ResponseWriter, req *http.Request) {

	var list []*Cartridge

	for drive := 1; drive <= daemon.DriveCount; drive++ {

		c := &Cartridge{Status: a.daemon.GetStatus(drive)}

		if c.Status == daemon.StatusIdle {
			if cart, ok := a.daemon.GetCartridge(drive); cart != nil {
				c.fill(cart)
				cart.Unlock()
			} else if !ok {
				c.Status = daemon.StatusBusy
			}
		}

		list = append(list, c)
	}

	if wantsJSON(req) {
		sendJSONReply(list, http.StatusOK, w)

	} else {
		strList := "\nDRIVE CARTRIDGE       STATE"
		for ix, c := range list {
			strList += fmt.Sprintf("\n  %d   %s", ix+1, c.String())
		}
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

	cart, err := reader.Read(io.LimitReader(req.Body, 1048576), true,
		isFlagSet(req, "repair"))
	if err != nil {
		handleError(fmt.Errorf("cartridge corrupted: %v", err),
			http.StatusUnprocessableEntity, w)
		return
	}
	if handleError(req.Body.Close(), http.StatusInternalServerError, w) {
		return
	}

	if err := a.daemon.SetCartridge(drive, cart, isFlagSet(req, "force")); err != nil {
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

	if err := a.daemon.UnloadCartridge(drive, isFlagSet(req, "force")); err != nil {
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
	a.driveInfo(w, req, "dump")
}

//
func (a *api) driveList(w http.ResponseWriter, req *http.Request) {
	a.driveInfo(w, req, "ls")
}

//
func (a *api) driveInfo(w http.ResponseWriter, req *http.Request, info string) {

	drive := getDrive(w, req)
	if drive == -1 {
		return
	}

	if a.daemon.GetStatus(drive) == daemon.StatusHardware {
		sendReply([]byte(fmt.Sprintf(
			"hardware drive mapped to slot %d", drive)),
			http.StatusOK, w)
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
		switch info {
		case "dump":
			cart.Emit(write)
		case "ls":
			cart.List(write)
		}
		write.Close()
	}()

	sendStreamReply(read, http.StatusOK, w)
}

// TODO: JSON response
func (a *api) getDriveMap(w http.ResponseWriter, req *http.Request) {

	start, end, locked := a.daemon.GetHardwareDrives()
	msg := ""

	if start == -1 || end == -1 {
		msg = "no hardware drives"

	} else {
		if start == 0 && end == 0 {
			msg = "hardware drives are off"
		} else {
			msg = fmt.Sprintf("hardware drives: start=%d, end=%d", start, end)
		}
		if locked {
			msg += " (locked)"
		}
	}
	sendReply([]byte(msg), http.StatusOK, w)
}

//
func (a *api) setDriveMap(w http.ResponseWriter, req *http.Request) {

	start, err := strconv.Atoi(getArg(req, "start"))
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	end, err := strconv.Atoi(getArg(req, "end"))
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	if handleError(a.daemon.MapHardwareDrives(start, end),
		http.StatusUnprocessableEntity, w) {
		return
	}

	sendReply([]byte(fmt.Sprintf(
		"mapped hardware drives: start=%d, end=%d", start, end)),
		http.StatusOK, w)
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
	ret, err := format.NewFormat(getArg(req, "type"))
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return nil
	}
	return ret
}

//
func isFlagSet(req *http.Request, flag string) bool {
	return getArg(req, flag) == "true"
}

//
func getArg(req *http.Request, arg string) string {
	return req.URL.Query().Get(arg)
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

// FIXME: make more tolerant
func wantsJSON(req *http.Request) bool {
	return req.Header.Get("Content-Type") == "application/json"
}
