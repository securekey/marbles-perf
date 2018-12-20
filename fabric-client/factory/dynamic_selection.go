/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package factory

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type dynamicSelectionChannelProvider struct {
	fab.ChannelProvider
	services map[string]*dynamicselection.SelectionService
	lock     sync.RWMutex
	config   fab.EndpointConfig
}

// ChannelService creates a ChannelService
func (cp *dynamicSelectionChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	chService, err := cp.ChannelProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, err
	}
	cp.lock.RLock()
	selection, ok := cp.services[channelID]
	cp.lock.RUnlock()
	if !ok {
		discovery, err := chService.Discovery()
		if err != nil {
			return nil, err
		}
		selection, err = dynamicselection.NewService(ctx, channelID, discovery)
		if err != nil {
			return nil, err
		}
		cp.lock.Lock()
		cp.services[channelID] = selection
		cp.lock.Unlock()
	}

	return &dynamicSelectionChannelService{
		ChannelService: chService,
		selection:      selection,
	}, nil
}

type dynamicSelectionChannelService struct {
	fab.ChannelService
	selection fab.SelectionService
}

func (cs *dynamicSelectionChannelService) Selection() (fab.SelectionService, error) {
	return cs.selection, nil
}

type initializer interface {
	Initialize(providers context.Providers) error
}

// Initialize sets the provider context
func (cp *dynamicSelectionChannelProvider) Initialize(providers context.Providers) error {
	if init, ok := cp.ChannelProvider.(initializer); ok {
		init.Initialize(providers)
	}
	return nil
}

type closable interface {
	Close()
}

// Close frees resources and caches.
func (cp *dynamicSelectionChannelProvider) Close() {
	if c, ok := cp.ChannelProvider.(closable); ok {
		c.Close()
	}

	cp.lock.RLock()
	for _, service := range cp.services {
		service.Close()
	}
	cp.lock.RUnlock()
}
