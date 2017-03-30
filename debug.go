package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/julienschmidt/httprouter"
	"github.com/zonesan/clog"
)

func debugIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	pprof.Index(w, r)
}

func debugProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	pprof.Profile(w, r)
}

func debugSymbol(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clog.Info("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	pprof.Symbol(w, r)
}
