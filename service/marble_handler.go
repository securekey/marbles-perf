package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"fmt"

	"github.com/gorilla/mux"
	"github.com/securekey/marbles-perf/api"
	"github.com/securekey/marbles-perf/fabric-client"
	"github.com/securekey/marbles-perf/utils"
)

// getOwner retrieves an existing owner
//
func getOwner(w http.ResponseWriter, r *http.Request) {
	var owner api.Owner
	getEntity(w, r, &owner)
}

// getMarble retrieves an existing marble
//
func getMarble(w http.ResponseWriter, r *http.Request) {
	var marble api.Marble
	getEntity(w, r, &marble)
}

// createOwner creates a new owner
//
func createOwner(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read request body: %s", err)
		return
	}

	var owner api.Owner
	if err := json.Unmarshal(payload, &owner); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to parse payload json: %s", err)
		return
	}

	response, err := doCreateOwner(owner)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func doCreateOwner(owner api.Owner) (resp api.Response, err error) {
	id := owner.Id
	if id == "" {
		id, err = utils.GenerateRandomAlphaNumericString(31)
		if err != nil {
			err = fmt.Errorf("failed to generate random string for id: %s", err)
			return
		}
		id = "o" + id
	}

	args := []string{
		"init_owner",
		id,
		owner.Username,
		owner.Company,
	}

	var data *fabricclient.CCResponse
	data, err = fc.InvokeCC(ConsortiumChannelID, MarblesCC, args, nil)
	if err != nil {
		err = fmt.Errorf("cc invoke failed: %s: %v", err, args)
		return
	}

	resp = api.Response{
		Id:   id,
		TxId: data.FabricTxnID,
	}
	return
}

// createMarble creates a new marble
//
func createMarble(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read request body: %s", err)
		return
	}

	var marble api.Marble
	if err := json.Unmarshal(payload, &marble); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to parse payload json: %s", err)
		return
	}

	response, err := doCreateMarble(marble)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func doCreateMarble(marble api.Marble) (resp api.Response, err error) {
	id := marble.Id
	if id == "" {
		id, err = utils.GenerateRandomAlphaNumericString(31)
		if err != nil {
			err = fmt.Errorf("failed to generate random string for id: %s", err)
			return
		}
		id = "m" + id
	}

	args := []string{
		"init_marble",
		id,
		marble.Color,
		strconv.Itoa(marble.Size),
		marble.Owner.Id,
		marble.Owner.Company,
	}

	// optional additonal data
	if marble.AdditionalData != "" {
		args = append(args, marble.AdditionalData)
	}

	data, ccErr := fc.InvokeCC(ConsortiumChannelID, MarblesCC, args, nil)
	if ccErr != nil {
		err = fmt.Errorf("cc invoke failed: %s: %v", err, args)
		return
	}

	resp = api.Response{
		Id:   id,
		TxId: data.FabricTxnID,
	}
	return
}

// transfer transfers marble ownership
//
func transfer(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to read request body: %s", err)
		return
	}

	var transfer api.Transfer
	if err := json.Unmarshal(payload, &transfer); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "failed to parse payload json: %s", err)
		return
	}

	args := []string{
		"set_owner",
		transfer.MarbleId,
		transfer.ToOwnerId,
		transfer.AuthCompany,
	}

	data, err := fc.InvokeCC(ConsortiumChannelID, MarblesCC, args, nil)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "cc invoke failed: %s: %v", err, args)
		return
	}
	response := api.Response{
		Id:   transfer.MarbleId,
		TxId: data.FabricTxnID,
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func doTransfer(transfer api.Transfer) (resp api.Response, err error) {
	args := []string{
		"set_owner",
		transfer.MarbleId,
		transfer.ToOwnerId,
		transfer.AuthCompany,
	}

	data, err := fc.InvokeCC(ConsortiumChannelID, MarblesCC, args, nil)
	if err != nil {
		err = fmt.Errorf("cc invoke failed: %s: %v", err, args)
		return
	}
	resp = api.Response{
		Id:   transfer.MarbleId,
		TxId: data.FabricTxnID,
	}
	return
}

// clearMarbles remove all marbles from ledger
//
func clearMarbles(w http.ResponseWriter, r *http.Request) {
	response, err := doClearMarbles()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
	}
	writeJSONResponse(w, http.StatusOK, response)
}

func doClearMarbles() (response api.ClearMarblesResponse, err error) {
	args := []string{"clear_marbles"}
	data, ccErr := fc.InvokeCC(ConsortiumChannelID, MarblesCC, args, nil)
	if ccErr != nil {
		err = fmt.Errorf("cc invoke failed: %s: %v", ccErr, args)
		return
	}

	if err = json.Unmarshal(data.Payload, &response); err != nil {
		err = fmt.Errorf("failed to JSON unmarshal cc response: %s: %v: %s", err, args, data.Payload)
		return
	}
	response.TxId = data.FabricTxnID
	return
}

// getEntity retrieves an existing entity
//
func getEntity(w http.ResponseWriter, r *http.Request, entity interface{}) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		writeErrorResponse(w, http.StatusBadRequest, "id not provided")
		return
	}

	data, err := doGetEntity(id, entity)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(data) == 0 {
		writeErrorResponse(w, http.StatusNotFound, "id not found")
		return
	}

	writeJSONResponse(w, http.StatusOK, entity)
}

func doGetEntity(id string, entity interface{}) ([]byte, error) {
	args := []string{
		"read",
		id,
	}

	data, err := fc.QueryCC(0, ConsortiumChannelID, MarblesCC, args, nil)
	if err != nil {
		return nil, fmt.Errorf("cc invoke failed: %s", err)
	}

	payloadJSON := data.Payload

	if len(payloadJSON) > 0 && entity != nil {
		if err := json.Unmarshal([]byte(payloadJSON), entity); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cc response payload: %s: %s", err, payloadJSON)
		}
	}
	return payloadJSON, nil
}

func doGetOwner(id string) (*api.Owner, error) {
	var owner api.Owner
	if data, err := doGetEntity(id, &owner); err != nil {
		return nil, err
	} else if len(data) == 0 {
		return nil, nil
	}
	return &owner, nil
}

func doGetMarble(id string) (*api.Marble, error) {
	var marble api.Marble
	if data, err := doGetEntity(id, &marble); err != nil {
		return nil, err
	} else if len(data) == 0 {
		return nil, nil
	}
	return &marble, nil
}
