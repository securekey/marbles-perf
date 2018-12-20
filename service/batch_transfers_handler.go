/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

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
		writeErrorResponse(w, http.StatusNotFound, "Batch run status not yet available (not complete)")
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
