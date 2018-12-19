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

package factory

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	copts "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var log = logging.NewLogger("fabric-client-factory")

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

// OneOfSelectionProvider implements a selection provider that randomly chooses
// a peer from the list of available peers
type OneOfSelectionProvider struct {
	fab.ChannelProvider
	IsPeerFilterMandatory bool
}

// Initialize sets the provider context
func (cp *OneOfSelectionProvider) Initialize(providers context.Providers) error {
	if init, ok := cp.ChannelProvider.(initializer); ok {
		init.Initialize(providers)
	}
	return nil
}

// ChannelService creates a ChannelService
func (cp *OneOfSelectionProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	chService, err := cp.ChannelProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, err
	}
	discovery, err := chService.Discovery()
	if err != nil {
		return nil, err
	}

	return &staticSelectionChannelService{
		ChannelService:        chService,
		IsPeerFilterMandatory: cp.IsPeerFilterMandatory,
		discoveryService:      discovery,
	}, nil
}

type staticSelectionChannelService struct {
	fab.ChannelService
	IsPeerFilterMandatory bool
	discoveryService      fab.DiscoveryService
}

// CreateSelectionService creates a static selection service
func (p *staticSelectionChannelService) Selection() (fab.SelectionService, error) {
	return &service{
		discoveryService:      p.discoveryService,
		isPeerFilterMandatory: p.IsPeerFilterMandatory,
	}, nil
}

/*
 * Static Selection Service
 */

// service implements static selection service
type service struct {
	discoveryService      fab.DiscoveryService
	isPeerFilterMandatory bool
}

// GetEndorsersForChaincode returns a random peer from the list of peers
func (s *service) GetEndorsersForChaincode(chaincodes []*fab.ChaincodeCall, opts ...copts.Opt) ([]fab.Peer, error) {
	if s == nil {
		return nil, fmt.Errorf("service: s is nil")
	}
	if s.discoveryService == nil {
		return nil, fmt.Errorf("service: discoveryService is nil")
	}
	channelPeers, err := s.discoveryService.GetPeers()
	if err != nil {
		log.Errorf("Error retrieving peers from discovery service: %s", err)
		return nil, nil
	}

	var peers []fab.Peer

	// For delegated transaction, we filter out peers that are not part of our org
	// This functionality is now passed in as a peer filter in opts
	params := options.NewParams(opts)
	// Apply peer filter if provided
	if params.PeerFilter != nil {
		for _, peer := range channelPeers {
			if params.PeerFilter(peer) {
				peers = append(peers, peer)
			}
		}
		if len(peers) == 0 {
			if s.isPeerFilterMandatory {
				log.Errorf("No available peers meeting peer filter requirement and peer filter is mandatory")
				return nil, nil
			}
			log.Errorf("No available peers meeting peer filter requirement but peer filter not mandatory, skipping filter...")
			peers = channelPeers
		}
	} else {
		peers = channelPeers
	}

	if len(peers) == 0 {
		log.Errorf("No available peers")
		return nil, nil
	}

	// Choose a random peer
	index := rand.Intn(len(peers))

	log.Debugf("Choosing peer %s out of %d peer(s)", peers[index].URL(), len(peers))
	return []fab.Peer{peers[index]}, nil
}
