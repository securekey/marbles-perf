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
package peerfilter

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MSPFilter accept peers with given MSP ID
type MSPFilter struct {
	MSPID string
}

// Accept ..
func (filter MSPFilter) Accept(peer fab.Peer) bool {
	return filter.MSPID == peer.MSPID()
}

// URLFilter accept peers with given URL
type URLFilter struct {
	PeerURL string
}

// Accept ..
func (filter URLFilter) Accept(peer fab.Peer) bool {
	return filter.PeerURL == peer.URL()
}
