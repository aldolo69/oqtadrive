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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/xelalexv/oqtadrive/pkg/daemon"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/client"
	"github.com/xelalexv/oqtadrive/pkg/microdrive/format"
)

//
type APIServer interface {
	Serve() error
	Stop() error
}

//
func NewAPIServer(addr string, d *daemon.Daemon) APIServer {
	return &api{address: addr, daemon: d}
}

//
type api struct {
	address string
	daemon  *daemon.Daemon
	server  *http.Server
	//
	longPollQueue chan chan *Change
}

//
func (a *api) Serve() error {

	router := mux.NewRouter().StrictSlash(true)

	addRoute(router, "status", "GET", "/status", a.status)
	addRoute(router, "watch", "GET", "/watch", a.watch)
	addRoute(router, "ls", "GET", "/list", a.list)
	addRoute(router, "load", "PUT", "/drive/{drive:[1-8]}", a.load)
	addRoute(router, "unload", "GET", "/drive/{drive:[1-8]}/unload", a.unload)
	addRoute(router, "save", "GET", "/drive/{drive:[1-8]}", a.save)
	addRoute(router, "dump", "GET", "/drive/{drive:[1-8]}/dump", a.dump)
	addRoute(router, "map", "GET", "/map", a.getDriveMap)
	addRoute(router, "map", "PUT", "/map", a.setDriveMap)
	addRoute(router, "drivels", "GET", "/drive/{drive:[1-8]}/list", a.driveList)
	addRoute(router, "resync", "PUT", "/resync", a.resync)
	addRoute(router, "config", "PUT", "/config", a.config)

	router.PathPrefix("/").Handler(
		requestLogger(http.FileServer(http.Dir("./ui/web/")), "webui"))

	addr := a.address
	if len(strings.Split(addr, ":")) < 2 {
		addr = fmt.Sprintf("%s:8888", a.address)
	}

	log.Infof("OqtaDrive API starts listening on %s", addr)
	a.server = &http.Server{Addr: addr, Handler: router}

	a.longPollQueue = make(chan chan *Change)
	go a.watchDaemon()

	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

//
func (a *api) Stop() error {
	if a.server != nil {
		log.Info("API server stopping...")
		err := a.server.Shutdown(context.Background())
		a.server = nil
		return err
	}
	return nil
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

	stat := &Status{Client: a.daemon.GetClient()}
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
func (a *api) watch(w http.ResponseWriter, req *http.Request) {

	timeout, err := strconv.Atoi(req.URL.Query().Get("timeout"))
	if err != nil || timeout < 0 || 1800 < timeout {
		timeout = 600
	}

	log.Infof("starting watch for %s, timeout %d", req.RemoteAddr, timeout)
	update := make(chan *Change)

	select {
	case a.longPollQueue <- update:
	case <-time.After(time.Duration(timeout) * time.Second):
		log.Infof("closing watch for %s after timeout", req.RemoteAddr)
		sendReply([]byte{}, http.StatusRequestTimeout, w)
		return
	}

	log.Infof("sending daemon change to %s", req.RemoteAddr)
	sendJSONReply(<-update, http.StatusOK, w)
}

//
func (a *api) watchDaemon() {

	log.Info("start watching for daemon changes")

	var client string
	var list []*Cartridge

	for a.server != nil {

		time.Sleep(2 * time.Second)
		change := &Change{}

		l := a.getCartridges()
		if !cartridgeListsEqual(l, list) {
			change.Drives = l
			list = l
		}

		c := a.daemon.GetClient()
		if c != client {
			change.Client = c
			client = c
		}

		if change.Drives == nil && change.Client == "" {
			continue
		}

		log.Info("daemon changes")

	Loop:
		for {
			select {
			case cl := <-a.longPollQueue:
				log.Info("notifying long poll client")
				cl <- change
			default:
				log.Info("all long poll clients notified")
				break Loop
			}
		}
	}

	log.Info("stopped watching for daemon changes")
}

//
func (a *api) list(w http.ResponseWriter, req *http.Request) {

	list := a.getCartridges()

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
func (a *api) getCartridges() []*Cartridge {

	ret := make([]*Cartridge, daemon.DriveCount)

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

		ret[drive-1] = c
	}

	return ret
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

	arg, err := getArg(req, "name")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}
	params := map[string]interface{}{"name": arg}
	cart, err := reader.Read(io.LimitReader(req.Body, 1048576), true,
		isFlagSet(req, "repair"), params)
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
	if handleError(
		writer.Write(cart, &out, nil), http.StatusInternalServerError, w) {
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

	start, err := getIntArg(req, "start")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	end, err := getIntArg(req, "end")
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
func (a *api) resync(w http.ResponseWriter, req *http.Request) {

	arg, err := getArg(req, "client")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	var cl client.Client = client.UNKNOWN

	if arg != "" {
		if cl = client.GetClient(arg); cl == client.UNKNOWN {
			handleError(fmt.Errorf("unknown client type: %s", arg),
				http.StatusUnprocessableEntity, w)
			return
		}
	}

	reset := isFlagSet(req, "reset")
	if handleError(
		a.daemon.Resync(cl, reset), http.StatusUnprocessableEntity, w) {
		return
	}

	msg := "re-syncing with adapter"
	if reset {
		msg = "resetting adapter"
	}
	sendReply([]byte(msg), http.StatusOK, w)
}

//
func (a *api) config(w http.ResponseWriter, req *http.Request) {

	item, err := getArg(req, "item")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	arg1, err := getIntArg(req, "arg1")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return
	}

	arg2, err := getIntArg(req, "arg2")
	if err != nil {
		arg2 = 0
	}

	if handleError(
		a.daemon.Configure(item, byte(arg1), byte(arg2)),
		http.StatusUnprocessableEntity, w) {
		return
	}

	sendReply([]byte("configuring"), http.StatusOK, w)
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
	arg, err := getArg(req, "type")
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return nil
	}
	ret, err := format.NewFormat(arg)
	if handleError(err, http.StatusUnprocessableEntity, w) {
		return nil
	}
	return ret
}

//
func isFlagSet(req *http.Request, flag string) bool {
	arg, _ := getArg(req, flag)
	return arg == "true"
}

//
func getArg(req *http.Request, arg string) (string, error) {
	ret := req.URL.Query().Get(arg)
	if ret != "" {
		return url.QueryUnescape(ret)
	}
	return ret, nil
}

//
func getIntArg(req *http.Request, arg string) (int, error) {
	if val, err := getArg(req, arg); err != nil {
		return -1, err
	} else {
		if ret, err := strconv.Atoi(val); err != nil {
			return -1, err
		} else {
			return ret, nil
		}
	}
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
