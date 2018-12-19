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
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/securekey/marbles-perf/api"
	"github.com/securekey/marbles-perf/utils"
)

func initBatchTransfers(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to read request body: %s", err)
		return
	}

	var batchRequest api.InitBatchRequest
	if err := json.Unmarshal(reqBody, &batchRequest); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to json unmarshal request content: %s", err)
		return
	}

	id, err := utils.GenerateRandomAlphaNumericString(24)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to generate random batch run id: %s", err)
		return
	}
	id = "b" + id
	resp := api.InitBatchResponse{
		BatchID: id,
	}
	writeJSONResponse(w, http.StatusOK, resp)

	go doBatchTransfers(id, batchRequest)

}

func fetchBatchResults(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		writeErrorResponse(w, http.StatusBadRequest, "missing batch ids")
		return
	}

	statusResp, err := fc.QueryCC(1, ConsortiumChannelID, MarblesCC, []string{"read", id + ledgerKeyBatchResults}, nil)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to fetch batch run status from ledger: %s", err)
		return
	}

	if len(statusResp.Payload) == 0 {
		writeErrorResponse(w, http.StatusNotFound, "Batch run status not yet avaialble (not complete)")
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(statusResp.Payload)
}

func doBatchTransfers(id string, batchReq api.InitBatchRequest) {
	tg := NewTransfersGenerator(id, batchReq)
	tg.run()
}
