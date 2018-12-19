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
