/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package factory

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
)

// MPerfQueryServiceProviderFactory A fabric-sdk service provider with a static Selection Provider,
// in order to query one specific peer
type MPerfQueryServiceProviderFactory struct {
	defsvc.ProviderFactory
	IsPeerFilterMandatory bool
}

// MPerfServiceProviderFactory A fabric-sdk service provider factory customized to Marbles needs.
// In particular, the selection provider created by this factory is a Dynamic Selection Provider
type MPerfServiceProviderFactory struct {
	defsvc.ProviderFactory
	userID string
	org    string
}

// NewMPerfQueryServiceProviderFactory create a new instance of MPerfQueryServiceProviderFactory
func NewMPerfQueryServiceProviderFactory(isPeerFilterMandatory bool) *MPerfQueryServiceProviderFactory {
	return &MPerfQueryServiceProviderFactory{IsPeerFilterMandatory: isPeerFilterMandatory}
}

// NewMPerfServiceProviderFactory create a new instance of MPerfServiceProviderFactory
func NewMPerfServiceProviderFactory(userID string) *MPerfServiceProviderFactory {
	return &MPerfServiceProviderFactory{userID: userID}
}

// CreateChannelProvider return a new implementation of OneOfSelectionProvider
func (f *MPerfQueryServiceProviderFactory) CreateChannelProvider(config fab.EndpointConfig) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config)
	if err != nil {
		return nil, err
	}
	return &OneOfSelectionProvider{
		ChannelProvider:       chProvider,
		IsPeerFilterMandatory: f.IsPeerFilterMandatory,
	}, nil
}

// CreateChannelProvider returns a new implementation of dynamic selection channel provider
func (f *MPerfServiceProviderFactory) CreateChannelProvider(config fab.EndpointConfig) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config)
	if err != nil {
		return nil, err
	}
	return &dynamicSelectionChannelProvider{
		ChannelProvider: chProvider,
		services:        make(map[string]*dynamicselection.SelectionService),
		config:          config,
	}, nil
}
