package factory

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
)

// VmeQueryServiceProviderFactory A fabric-sdk service provider with a static Selection Provider,
// in order to query one specific peer
type VmeQueryServiceProviderFactory struct {
	defsvc.ProviderFactory
	IsPeerFilterMandatory bool
}

// VmeServiceProviderFactory A fabric-sdk service provider factory customized to Verified.Me's needs.
// In particular, the selection provider created by this factory is a Dynamic Selection Provider
type VmeServiceProviderFactory struct {
	defsvc.ProviderFactory
	userID string
	org    string
}

// NewVmeQueryServiceProviderFactory create a new instance of VmeQueryServiceProviderFactory
func NewVmeQueryServiceProviderFactory(isPeerFilterMandatory bool) *VmeQueryServiceProviderFactory {
	return &VmeQueryServiceProviderFactory{IsPeerFilterMandatory: isPeerFilterMandatory}
}

// NewVmeServiceProviderFactory create a new instance of VmeServiceProviderFactory
func NewVmeServiceProviderFactory(userID string) *VmeServiceProviderFactory {
	return &VmeServiceProviderFactory{userID: userID}
}

// CreateChannelProvider return a new implementation of OneOfSelectionProvider
func (f *VmeQueryServiceProviderFactory) CreateChannelProvider(config fab.EndpointConfig) (fab.ChannelProvider, error) {
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
func (f *VmeServiceProviderFactory) CreateChannelProvider(config fab.EndpointConfig) (fab.ChannelProvider, error) {
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
