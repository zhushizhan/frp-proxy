// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	httppkg "github.com/fatedier/frp/pkg/util/http"
	netpkg "github.com/fatedier/frp/pkg/util/net"
	"github.com/fatedier/frp/pkg/util/version"
	adminapi "github.com/fatedier/frp/server/http"
	adminapi_model "github.com/fatedier/frp/server/http/model"
	webuifrps "github.com/fatedier/frp/webui/frps"
)

func (svr *Service) registerRouteHandlers(helper *httppkg.RouterRegisterHelper) {
	helper.Router.HandleFunc("/healthz", healthz)
	subRouter := helper.Router.NewRoute().Subrouter()

	subRouter.Use(helper.AuthMiddleware)
	subRouter.Use(httppkg.NewRequestLogger)

	// metrics
	if svr.cfg.EnablePrometheus {
		subRouter.Handle("/metrics", promhttp.Handler())
	}

	apiController := adminapi.NewController(svr.cfg, svr.clientRegistry, svr.pxyManager, func() adminapi.ConfigManager {
		return svr.configManager
	}).WithKickFunc(svr.ctlManager.KickByRunID)

	// apis
	subRouter.HandleFunc("/api/serverinfo", httppkg.MakeHTTPHandlerFunc(apiController.APIServerInfo)).Methods("GET")
	subRouter.HandleFunc("/api/files/upload", httppkg.MakeHTTPHandlerFunc(apiController.UploadFile)).Methods("POST")
	subRouter.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		if svr.configManager == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(httppkg.GeneralResponse{Code: 500, Msg: "server config manager unavailable"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			settings, err := svr.configManager.GetSettings()
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(httppkg.GeneralResponse{Code: 400, Msg: err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(settings)
		case http.MethodPut:
			var payload adminapi_model.ServerSettings
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(httppkg.GeneralResponse{Code: 400, Msg: err.Error()})
				return
			}
			if err := svr.configManager.UpdateSettings(payload); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(httppkg.GeneralResponse{Code: 400, Msg: err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(httppkg.GeneralResponse{Code: 200, Msg: "saved and restarting"})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}).Methods("GET", "PUT")
	subRouter.HandleFunc("/api/proxy/{type}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyByType)).Methods("GET")
	subRouter.HandleFunc("/api/proxy/{type}/{name}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyByTypeAndName)).Methods("GET")
	subRouter.HandleFunc("/api/proxies/{name}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyByName)).Methods("GET")
	subRouter.HandleFunc("/api/traffic/{name}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyTraffic)).Methods("GET")
	subRouter.HandleFunc("/api/clients", httppkg.MakeHTTPHandlerFunc(apiController.APIClientList)).Methods("GET")
	subRouter.HandleFunc("/api/clients/{key}", httppkg.MakeHTTPHandlerFunc(apiController.APIClientDetail)).Methods("GET")
	subRouter.HandleFunc("/api/clients/{key}", httppkg.MakeHTTPHandlerFunc(apiController.KickClient)).Methods("DELETE")
	subRouter.HandleFunc("/api/proxies", httppkg.MakeHTTPHandlerFunc(apiController.DeleteProxies)).Methods("DELETE")

	// view
	subRouter.Handle("/favicon.ico", http.FileServer(helper.AssetsFS)).Methods("GET")
	subRouter.PathPrefix("/static/").Handler(
		netpkg.MakeHTTPGzipHandler(http.StripPrefix("/static/", http.FileServer(helper.AssetsFS))),
	).Methods("GET")
	if webuiFS, ok := webuifrps.HTTPFileSystem(); ok {
		registerWebUIRoutes(subRouter, webuiFS)
	}

	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/webui/", http.StatusMovedPermanently)
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"ok","version":%q}`, version.Full())
}

func registerWebUIRoutes(router *mux.Router, webuiFS http.FileSystem) {
	fileServer := http.FileServer(webuiFS)
	servePath := func(w http.ResponseWriter, r *http.Request, path string) {
		req := r.Clone(r.Context())
		req.URL.Path = path
		fileServer.ServeHTTP(w, req)
	}
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		relPath := strings.TrimPrefix(r.URL.Path, "/webui/")
		if hasWebUIAsset(webuiFS, relPath) {
			servePath(w, r, "/"+relPath)
			return
		}
		servePath(w, r, "/")
	})

	router.HandleFunc("/webui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/webui/", http.StatusMovedPermanently)
	}).Methods(http.MethodGet)
	router.PathPrefix("/webui/").Handler(netpkg.MakeHTTPGzipHandler(spaHandler)).Methods(http.MethodGet)
}

func hasWebUIAsset(webuiFS http.FileSystem, relPath string) bool {
	relPath = strings.TrimPrefix(relPath, "/")
	if relPath == "" {
		return false
	}

	file, err := webuiFS.Open(relPath)
	if err != nil {
		return false
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false
	}
	return !info.IsDir()
}
