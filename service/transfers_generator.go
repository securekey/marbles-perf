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
	"math"
	"sync"
	"time"

	"encoding/hex"
	"fmt"
	"math/rand"

	"encoding/json"

	"github.com/securekey/marbles-perf/api"
)

const (
	createMarbleMaxAttempts = 3000

	ledgerKeyBatchResults = "BRR"

	statusSuccess          = "success"
	statusFailOwnerCreate  = "owner_create_failed"
	statusFailMarbleCreate = "marble_create_failed"
)

var colorArray = []string{"red", "orange", "yellow", "green", "blue", "indigo", "violet"}

type WorkerPerfData struct {
	transferTimes []time.Duration
	successes     int
	failures      int
	status        string
}

type MarbleWorker struct {
	id       int
	tg       *TransfersGenerator
	perfData *WorkerPerfData
	wg       *sync.WaitGroup
}

type TransfersGenerator struct {
	batchRunID string
	request    api.InitBatchRequest
	owners     map[string]*api.Owner
	ownerArray []string
}

func NewTransfersGenerator(id string, req api.InitBatchRequest) *TransfersGenerator {
	return &TransfersGenerator{
		batchRunID: id,
		request:    req,
	}
}

func (tg *TransfersGenerator) run() {
	// Create array of perf data objects (one per worker)
	perfData := make([]WorkerPerfData, tg.request.Concurrency)

	logger.Infof("concurrency=%d, iterations=%d, extraDataLength=%d\n", tg.request.Concurrency, tg.request.Iterations, tg.request.ExtraDataLength)
	if err := tg.initializeState(); err != nil {
		logger.Errorf("failed to initialize state for batch run: %s", err)
		tg.abortBatchRun(statusFailOwnerCreate)
		return
	}

	var wg sync.WaitGroup

	for workerId := 1; workerId <= tg.request.Concurrency; workerId++ {
		perfData[workerId-1].transferTimes = make([]time.Duration, tg.request.Iterations)
		worker := &MarbleWorker{
			id:       workerId,
			tg:       tg,
			perfData: &perfData[workerId-1],
			wg:       &wg,
		}

		wg.Add(1)
		go worker.startWorker()
	}

	wg.Wait()

	tg.processPerfData(perfData)
}

func (tg *TransfersGenerator) initializeState() error {
	tg.populateUsers()
	if err := tg.createOwners(); err != nil {
		return fmt.Errorf("failed to createOwners: %s", err)
	}

	if tg.request.ClearMarbles {
		resp, err := doClearMarbles()
		if err != nil {
			logger.Warningf("failed to clear marbles from ledger: %s", err)
		} else {
			logger.Infof("removed existng marbles - found: %d, removed: %d", resp.Found, resp.TxId)
		}
	}
	return nil
}

func (tg *TransfersGenerator) populateUsers() {
	tg.owners = map[string]*api.Owner{}
	tg.owners["o1"] = &api.Owner{Id: "o1", Username: "user1", Company: "United Marbles"}
	tg.owners["o2"] = &api.Owner{Id: "o2", Username: "user2", Company: "Spherical Arts"}
	tg.owners["o3"] = &api.Owner{Id: "o3", Username: "user3", Company: "Round Rollers"}
	tg.owners["o4"] = &api.Owner{Id: "o4", Username: "user4", Company: "Alley Baba."}
	tg.owners["o5"] = &api.Owner{Id: "o5", Username: "user5", Company: "ACME Inc."}

	tg.ownerArray = make([]string, len(tg.owners))
	index := 0
	for o := range tg.owners {
		tg.ownerArray[index] = o
		index++
	}

}

func (tg *TransfersGenerator) createOwners() error {

	for _, o := range tg.owners {
		// See if owner exists
		owner, err := doGetOwner(o.Id)
		if err == nil && owner != nil {
			// User already exists
			tg.owners[o.Id] = owner
			continue
		}

		// create new owner
		if _, err := doCreateOwner(*o); err != nil {
			return err
		}
	}
	return nil
}

func (tg *TransfersGenerator) abortBatchRun(code string) {
	logger.Errorf("aborting batch run %s: %s", tg.batchRunID, code)
	tg.storeBatchRunResults(api.BatchResult{
		Status:  code,
		Request: tg.request,
	})
}

func (tg *TransfersGenerator) writeLedger(key, value string) {
	key = tg.batchRunID + key
	if _, err := fc.InvokeCC(ConsortiumChannelID, MarblesCC, []string{"write", key, value}, nil); err != nil {
		logger.Errorf("failed to write to ledger: %s: %s - %s", key, value, err)
	}
}

func (tg *TransfersGenerator) readLedger(key string) string {
	key = tg.batchRunID + key
	if resp, err := fc.InvokeCC(ConsortiumChannelID, MarblesCC, []string{"read", key}, nil); err != nil {
		logger.Errorf("failed to read from ledger: %s: %s", key, err)
	} else {
		return string(resp.Payload)
	}
	return ""
}

func (w *MarbleWorker) startWorker() {

	// Create a marble
	owner := w.tg.pickRandomOwner(nil)

	marble := api.Marble{
		Id:             "m" + generateRandomValue(16),
		Color:          pickRandomColor(),
		Size:           generateRandomSize(),
		Owner:          *owner,
		AdditionalData: generateRandomValue(w.tg.request.ExtraDataLength),
	}

	var marbleCreated bool
	var err error
	for i := 0; i < createMarbleMaxAttempts; i++ {
		if _, err = doCreateMarble(marble); err == nil {
			marbleCreated = true
			break
		}
		logger.Infof("Failed to create marble, attempt %d: %s", i, err)
	}
	if !marbleCreated {
		logger.Errorf("Error creating marble: Worker %d, Create marble %s for %s: %s", w.id, marble.Id, owner.Username, err)
		w.perfData.status = statusFailMarbleCreate
		w.wg.Done()
		return
	}

	logger.Infof("Worker %d, Marble %s created for %s", w.id, marble.Id, owner.Username)

	prevOwner := owner

	// Loop through each iteration of the test.
	for t := 1; t <= w.tg.request.Iterations; t++ {
		if w.tg.request.DelaySeconds > 0 {
			time.Sleep(time.Duration(w.tg.request.DelaySeconds) * time.Second)
		}
		newOwner := w.tg.pickRandomOwner(prevOwner)
		transfer := api.Transfer{
			MarbleId:    marble.Id,
			ToOwnerId:   newOwner.Id,
			AuthCompany: prevOwner.Company,
		}
		start := time.Now()
		resp, err := doTransfer(transfer)
		if err == nil {
			w.perfData.transferTimes[t-1] = time.Since(start)
			w.perfData.successes++
			logger.Debugf("Worker %d, Iteration %d: Marble %s transferred from %s to %s", w.id, t, marble.Id, prevOwner.Username, newOwner.Username)
			prevOwner = newOwner
		} else {
			w.perfData.failures++
			logger.Debugf("Error transferring marble: Worker %d, Iteration %d: Transfer marble %s from %s to %s: %s", w.id, t, marble.Id, prevOwner.Username, newOwner.Username, resp.Error)
		}
	}
	w.wg.Done()

}

// Process the collected data.
// Note that durations are only captured for successes so we'll
// ignore zero values as they are for errors.
func (tg *TransfersGenerator) processPerfData(perfDataArray []WorkerPerfData) {

	totalSuccesses := 0
	totalFailures := 0
	totalDuration := time.Duration(0)
	maxDuration := time.Duration(0)
	minDuration := time.Hour // Hopefully an hour is long enough to be a minimum

	successWorkerCount := 0 // number of workers that have at least 1 successful transfer
	var workerFailureStatus string

	for _, perfData := range perfDataArray {
		if perfData.successes > 0 {
			successWorkerCount += 1
		} else {
			workerFailureStatus = perfData.status
		}
		totalSuccesses += perfData.successes
		totalFailures += perfData.failures

		for _, duration := range perfData.transferTimes {
			if duration > 0 {
				if duration > maxDuration {
					maxDuration = duration
				}
				if duration < minDuration {
					minDuration = duration
				}
				totalDuration += duration
			}
		}
	}

	avgTrfSecs := totalDuration.Seconds() / float64(totalSuccesses)
	avgTrfSecs = math.Round(avgTrfSecs*1000) / 1000
	minTrfSecs := math.Round(minDuration.Seconds()*1000) / 1000
	maxTrfSecs := math.Round(maxDuration.Seconds()*1000) / 1000

	logger.Infof("batch run completed %s", tg.batchRunID)
	logger.Infof("concurrency=%d, iterations=%d, extraDataLength=%d", tg.request.Concurrency, tg.request.Iterations, tg.request.ExtraDataLength)
	logger.Infof("Total number of transfers:         %d", totalSuccesses)
	logger.Infof("Total number of failures :         %d", totalFailures)
	logger.Infof("Total seconds taken for successes: %d", int(totalDuration.Seconds()))
	logger.Infof("Average seconds per transfer:      %3.3f", avgTrfSecs)
	logger.Infof("Minimum seconds per transfer:      %3.3f", minTrfSecs)
	logger.Infof("Maximum seconds per transfer:      %3.3f", maxTrfSecs)

	runStatus := statusSuccess
	if successWorkerCount < len(perfDataArray) {
		// at least 1 worker didn't complete ANY transfers at all
		runStatus = workerFailureStatus
	}

	results := api.BatchResult{
		Request:                tg.request,
		Status:                 runStatus,
		TotalSuccesses:         totalSuccesses,
		TotalFailures:          totalFailures,
		TotalSuccessSeconds:    int(totalDuration.Seconds()),
		AverageTransferSeconds: avgTrfSecs,
		MinTransferSeconds:     minTrfSecs,
		MaxTransferSeconds:     maxTrfSecs,
	}

	tg.storeBatchRunResults(results)
}

func (tg *TransfersGenerator) storeBatchRunResults(results api.BatchResult) {
	resultsJSON, err := json.MarshalIndent(results, "", "   ")
	if err != nil {
		logger.Errorf("failed to JSON marshal batch run results: %s", err)
		return
	}
	tg.writeLedger(ledgerKeyBatchResults, string(resultsJSON))
}

func (tg *TransfersGenerator) pickRandomOwner(currOwner *api.Owner) *api.Owner {

	for {
		index := rand.Intn(len(tg.ownerArray))
		newOwner := tg.owners[tg.ownerArray[index]]
		if newOwner != currOwner {
			return newOwner
		}
	}
}

func pickRandomColor() string {
	index := rand.Intn(len(colorArray))
	return colorArray[index]
}

func generateRandomSize() int {
	size := rand.Intn(10) + 1
	return size
}

func generateRandomValue(length int) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}
