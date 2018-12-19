//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
package main

import (
	"log"
	"net/http"

	"time"

	"os"

	"encoding/json"
	"fmt"

	"math/rand"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	fabclient "github.com/securekey/marbles-perf/fabric-client"
	"github.com/securekey/marbles-perf/utils"
	"github.com/spf13/viper"
)

const (
	ConsortiumChannelID = "consortium"
	MarblesCC           = "marblescc"
)

var fc fabclient.Client
var logger = logging.MustGetLogger("marbles-service")

func main() {

	if len(os.Args) < 2 {
		log.Fatal("expecting configuration file as first argument")
	}
	cfgFile := os.Args[1]

	err := SetupViper(cfgFile)
	if err != nil {
		log.Fatalf("error setting up viper using config file and environmental variables: %v ", err)
	}

	utils.InitLogger()

	fc, err = fabclient.NewClient()
	if err != nil {
		log.Fatalf("failed to initialize fabric client: %s", err)
	}

	r := mux.NewRouter()
	// ping
	r.HandleFunc("/hello", handleHello)
	// CRUD
	r.HandleFunc("/marble", createMarble).Methods(http.MethodPost)
	r.HandleFunc("/marble/{id}", getMarble).Methods(http.MethodGet)
	r.HandleFunc("/owner", createOwner).Methods(http.MethodPost)
	r.HandleFunc("/owner/{id}", getOwner).Methods(http.MethodGet)
	r.HandleFunc("/transfer", transfer).Methods(http.MethodPost)
	r.HandleFunc("/clear_marbles", clearMarbles).Methods(http.MethodPost)

	// batch (random) transfers
	r.HandleFunc("/batch_run", initBatchTransfers).Methods(http.MethodPost)
	r.HandleFunc("/batch_run/{id}", fetchBatchResults).Methods(http.MethodGet)

	// Seed the random generator so we get different values each time
	rand.Seed(time.Now().UTC().UnixNano())

	srv := &http.Server{
		Handler:      r,
		Addr:         viper.GetString("http.server.address"),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World\n"))
}

func writeErrorResponse(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args)
	}
	w.Write([]byte(fmt.Sprintf(`{error: "%s"}`, msg)))
	logger.Infof("error: %s", msg)
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	jsonStr, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to JSON marshal response: %s", err)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonStr)
}
