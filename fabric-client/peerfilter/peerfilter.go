/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

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
