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
package cc

import (
	"time"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fc "github.com/securekey/marbles-perf/fabric-client"
)

var log = logging.NewLogger("fcClient_cc")

// EventListener is a chaincode event handler
type EventListener interface {
	// Register registers for a chaincode event
	Register(eventID string) error
	// BlockWait Block and wait for a chaincode event matching the previously registered eventID
	// Return cc event URL if event is received before timing out
	BlockWait(timeoutSeconds int, numMessages int) string
	// Close closes the listener and releases any assoicated resources
	Close()
}

// EventListenerImpl is the default implementation of CCEventHandler
type EventListenerImpl struct {
	ChannelID   string
	ChaincodeID string

	// an optional extra delay to apply before unblocking
	// upon receipt of event notification
	// a small time gap before returning notification so that
	// data will be available at more peers already by the time notitfication is received
	NotificationDelayMilliseconds int

	fabClient    fc.Client
	chClient     *channel.Client
	registration fab.Registration
	notifier     <-chan *fab.CCEvent
	eventID      string
}

// Register registers for a chaincode event
func (t *EventListenerImpl) Register(eventID string) error {
	var err error
	if t.chClient == nil {
		t.chClient, err = t.fabClient.ChannelClient(t.ChannelID)
		if err != nil {
			return err
		}
	}

	t.notifier = make(chan *fab.CCEvent)
	t.registration, t.notifier, err = t.chClient.RegisterChaincodeEvent(t.ChaincodeID, eventID)
	if err != nil {
		return fmt.Errorf("failed to register chaincode event: cc: %s, event: %s, error: %v", t.ChaincodeID, eventID, err)
	}
	t.eventID = eventID

	return nil
}

// BlockWait Blocks and wait for a chaincode event
func (t *EventListenerImpl) BlockWait(timeoutSeconds int, numMessages int) (ccEventURL string) {
	// blocking wait
	defer func() {
		t.chClient.UnregisterChaincodeEvent(t.registration)
		t.registration = nil
	}()
	if timeoutSeconds == 0 {
		timeoutSeconds = t.fabClient.EventTimeoutSeconds()
	}
	numMessagesReceived := 0
	log.Debugf("====== blocking for up to %d seconds to wait for chaincode notification %s", timeoutSeconds, t.eventID)
	waitStart := time.Now().Unix()
	var ccEvent *fab.CCEvent
	timeout := time.After(time.Second * time.Duration(timeoutSeconds))
	for numMessagesReceived < numMessages {
		select {
		case ccEvent = <-t.notifier:
			numMessagesReceived++
		case <-timeout:
			timeTaken := time.Now().Unix() - waitStart
			log.Warnf("timeout waiting for chaincode notification to arrive event ID %s, total wait time %d seconds", t.eventID, timeTaken)
			return
		}
	}
	if numMessagesReceived == numMessages {
		if numMessages != 1 {
			// only care about notification delay when there's multiple messages,
			// else we return the peer URL and any queries can start immediately
			if t.NotificationDelayMilliseconds > 0 {
				// optional, configurable brief pause to delay the event notification
				// this give a chance for more peers to receive the new data
				// before further query on the expected data to avoid unnecessary (and likely) retries
				//
				time.Sleep(time.Millisecond * time.Duration(t.NotificationDelayMilliseconds))
			}
		} else {
			// returning ccEventURL only when there's only one message
			ccEventURL = ccEvent.SourceURL
		}
		timeTaken := time.Now().Unix() - waitStart
		log.Debugf("got event notification for event ID %s - total wait time %d seconds", t.eventID, timeTaken)
	}
	return
}

// Close closes the listener and releases any assoicated resources
func (t *EventListenerImpl) Close() {
	if t.chClient != nil {
		if t.registration != nil {
			t.chClient.UnregisterChaincodeEvent(t.registration)
			t.registration = nil
		}
		t.fabClient.CloseChannelClient(t.chClient)
		t.chClient = nil
	}
}
