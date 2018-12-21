/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/securekey/marbles-perf/fabric-client/factory"
	"github.com/securekey/marbles-perf/fabric-client/peerfilter"

	"github.com/spf13/viper"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	fabapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	sdkcfg "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/modlog"
	gologging "github.com/op/go-logging"
)

// Client ...
/**
 * The Client object captures settings for a participating member as a fabric client to the hyperledger network
 *
 */
type Client interface {

	// InvokeCC invokes a chancode on the specified channel
	InvokeCC(channelID string, chainCodeID string, args []string, transientData map[string][]byte) (data *CCResponse, err error)

	// QueryCC queries a chaincode
	QueryCC(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte) (data *CCResponse, err error)

	// QueryCCAtPeer query a chaincode from specified peer
	QueryCCAtPeer(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte, peerURL string) (data *CCResponse, err error)

	// QueryCCAtMSP query a chaincode from specified MSP
	QueryCCAtMSP(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte, MSPID string) (data *CCResponse, err error)

	// QueryCCAtOwnOrg query a chaincode from within one's own organziation
	QueryCCAtOwnOrg(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte) (data *CCResponse, err error)

	// ChannelClient returns a channel client for the given channel id
	ChannelClient(channelID string) (*channel.Client, error)

	// ConsortiumChannelID return channel ID of consortium channel
	ConsortiumChannelID() string

	// OrgChannelID return channel ID of own org
	OrgChannelID() string

	// CloseChannelClient (logically) closes a channel client obtained from a previous ChannelClient() call
	CloseChannelClient(client *channel.Client)

	// EventTimeoutSeconds returns the configured or default timeout value in seconds for chaincode/transaction event
	EventTimeoutSeconds() int

	// NewChannelProvider returns a channel provider based on fabricSDK field of client
	NewChannelProvider(channelID string) context.ChannelProvider

	// Close closes this client
	Close()
}

// CCResponse ...
type CCResponse struct {
	Payload     []byte
	FabricTxnID string
}

type fabClient struct {
	userID              string
	fabricSDK           *fabsdk.FabricSDK
	fabricSDKQuery      *fabsdk.FabricSDK
	queryMaxAttempts    int
	invokeMaxAttempts   int
	eventTimeoutSeconds int
	orgChannelID        string
	shareChannelClient  bool

	chclientMutex      *sync.RWMutex
	chClients          map[string]*channel.Client
	chclientqueryMutex *sync.RWMutex
	chClientsQuery     map[string]*channel.Client
	queryRetryOpts     retry.Opts
	invokeRetryOpts    retry.Opts

	orgConfig *fabapi.OrganizationConfig
	orgName   string
}

const (
	configTreeNameFabricSDK = "fabric_sdk"
	// ConfigUserID ..
	ConfigUserID                   = configTreeNameFabricSDK + ".user.id"
	configUserEnrollmentSecret     = configTreeNameFabricSDK + ".user.enroll_secret"
	configEventTimeoutSeconds      = configTreeNameFabricSDK + ".events.timeout_seconds"
	configUseTxnDelegation         = configTreeNameFabricSDK + ".txn.use_delegation"
	configQueryMaxAttempts         = configTreeNameFabricSDK + ".query.max_attempts"
	configInvokeMaxAttempts        = configTreeNameFabricSDK + ".invoke.max_attempts"
	configDelegationTimeoutSeconds = configTreeNameFabricSDK + ".delegation.timeout_seconds"

	configClientConfiguration = configTreeNameFabricSDK + ".client_configuration"
	configClientConfFile      = configTreeNameFabricSDK + ".client_conf_file"
	configOwnChannel          = "FABRIC_CHANNEL"

	consortiumChannelID = "consortium"
)

var logger = gologging.MustGetLogger("sdkclient")

// NewClient ...
/**
 * Create a new instance of Client
 */
func NewClient() (Client, error) {
	client := fabClient{}

	if err := client.init(); err != nil {
		return nil, err
	}
	return &client, nil
}

func (t *fabClient) init() error {

	t.chClients = make(map[string]*channel.Client)
	t.chClientsQuery = make(map[string]*channel.Client)
	t.chclientMutex = &sync.RWMutex{}
	t.chclientqueryMutex = &sync.RWMutex{}

	if t.userID = viper.GetString(ConfigUserID); len(t.userID) == 0 {
		return fmt.Errorf("configuration error, %s not set", ConfigUserID)
	}
	if t.queryMaxAttempts = viper.GetInt(configQueryMaxAttempts); t.queryMaxAttempts == 0 {
		t.queryMaxAttempts = 5
	}
	if t.invokeMaxAttempts = viper.GetInt(configInvokeMaxAttempts); t.invokeMaxAttempts == 0 {
		t.invokeMaxAttempts = 4
	}
	if t.eventTimeoutSeconds = viper.GetInt(configEventTimeoutSeconds); t.eventTimeoutSeconds == 0 {
		t.eventTimeoutSeconds = 60
	}

	t.orgChannelID = viper.GetString(configOwnChannel)
	t.shareChannelClient = true // new sdk does not support closing of channel client anymore, so no use for NOT sharing

	t.queryRetryOpts = retry.DefaultOpts
	t.queryRetryOpts.RetryableCodes[status.ChaincodeStatus] = []status.Code{
		status.Code(common.Status_BAD_REQUEST),
	}
	t.queryRetryOpts.Attempts = t.queryMaxAttempts
	AddExtraRetryableCodes(t.queryRetryOpts.RetryableCodes)

	t.invokeRetryOpts = t.queryRetryOpts
	t.invokeRetryOpts.Attempts = t.invokeMaxAttempts

	invokeRetryOptsJSON, _ := json.Marshal(t.invokeRetryOpts)
	logger.Infof("effective invoke retry options: %s", string(invokeRetryOptsJSON))

	// make fabric-sdk use our logger for logging for consistency
	sdkLoggerProvider := &sdkLoggerProvider{logger}
	modlog.InitLogger(sdkLoggerProvider)

	sdkConfData, err := SetupClientConfFile()
	if err != nil {
		return err
	}

	apiConfig := sdkcfg.FromRaw(sdkConfData, "yaml")
	t.fabricSDK, err = fabsdk.New(
		apiConfig,
		fabsdk.WithServicePkg(factory.NewMPerfServiceProviderFactory(t.userID)))
	if err != nil {
		return fmt.Errorf("failed to create new SDK: %v", err)
	}

	// the peer filter used in query calls is not mandatory
	// this is to allow peer-specific peer filter to query other peers in the rare event that the peer specified is shut down
	isQueryPeerFilterMandatory := false

	apiConfigQuery := sdkcfg.FromRaw(sdkConfData, "yaml")
	t.fabricSDKQuery, err = fabsdk.New(
		apiConfigQuery,
		//fabsdk.WithCorePkg(factory.NewMPerfCoreProviderFactory()),
		fabsdk.WithServicePkg(factory.NewMPerfQueryServiceProviderFactory(isQueryPeerFilterMandatory)))

	if err != nil {
		return fmt.Errorf("failed to create new SDK for chaincode query: %v", err)
	}

	t.orgName, t.orgConfig, err = GetDefaultOrganizationConfig(t.fabricSDK)
	if err != nil || t.orgConfig == nil {
		return fmt.Errorf("failed to get default organization config: %v", err)
	}

	if err := EnrollMemberIfNecessary(t.fabricSDK, t.userID); err != nil {
		return fmt.Errorf("failed to enroll member: %v", err)
	}

	return nil
}

// AddExtraRetryableCodes Add extra retryable codes that are not part of SDK's default, but are needed for our application
//
func AddExtraRetryableCodes(codes map[status.Group][]status.Code) {
	clientStatusGroup, hasClientStatusGroup := codes[status.ClientStatus]
	if !hasClientStatusGroup {
		clientStatusGroup = []status.Code{}
	}

	// retry for chaincode errors is set in sdk config now
	codes[status.ClientStatus] = append(clientStatusGroup, status.NoPeersFound)
	codes[status.ChaincodeStatus] = append(codes[status.ChaincodeStatus], []status.Code{status.Code(412), status.Code(500), status.Code(501), status.Code(502)}...)
}

// SetupClientConfFile ..
func SetupClientConfFile() ([]byte, error) {
	if !viper.InConfig(configTreeNameFabricSDK) {
		return nil, fmt.Errorf("configuration section missing %s", configTreeNameFabricSDK)
	}

	clientConfFile := viper.GetString(configClientConfFile)
	clientConfContent := viper.GetString(configClientConfiguration)

	if clientConfFile == "" && clientConfContent == "" {
		return nil, fmt.Errorf("fabric SDK client configuration not specified, must provide a file (%s) or content (%s)", configClientConfFile, configClientConfiguration)
	} else if clientConfFile != "" && clientConfContent != "" {
		return nil, fmt.Errorf("only one of %s or %s can be specified", configClientConfFile, configClientConfiguration)
	}

	var err error
	rawConfData := []byte(clientConfContent)
	if clientConfContent == "" {
		rawConfData, err = ioutil.ReadFile(clientConfFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read Fabric SDK client conf file %s: %v", clientConfFile, err)
		}
	}
	confData := os.ExpandEnv(string(rawConfData))
	return []byte(confData), nil
}

// EnrollMemberIfNecessary ..
func EnrollMemberIfNecessary(fabricSDK *fabsdk.FabricSDK, userID string) error {

	logger.Infof("In enrollMemberIfNecessary ...")

	mspClient, err := msp.New(fabricSDK.Context())
	if err != nil {
		return fmt.Errorf("failed to create mspClient: %v", err)
	}

	user, err := mspClient.GetSigningIdentity(userID)
	if user != nil && err == nil {
		logger.Infof("successfully loaded fabric user credential: %s", userID)
		return nil
	}

	enrollSecret := viper.GetString(configUserEnrollmentSecret)
	if enrollSecret == "" {
		return fmt.Errorf("failed to load user credentials (%v) and no user enrollment secret configured", err)
	}

	logger.Infof("will attempt to enroll user now...")
	if err := mspClient.Enroll(userID, msp.WithSecret(enrollSecret)); err != nil {

		return fmt.Errorf("user enrollment failed, enroll call returned error: %v", err)
	}
	logger.Infof("enrollment success for %s", userID)

	return nil
}

// GetDefaultOrganizationConfig ..
func GetDefaultOrganizationConfig(fabricSDK *fabsdk.FabricSDK) (string, *fabapi.OrganizationConfig, error) {
	configBackend, err := fabricSDK.Config()
	if err != nil {
		return "", nil, fmt.Errorf("fabric-sdk failed to get config backend %v", err)
	}
	config, err := fab.ConfigFromBackend(configBackend)
	if err != nil {
		return "", nil, fmt.Errorf("fabric-sdk failed to get config from backend %v", err)
	}
	networkCfg := config.NetworkConfig()
	context := fabricSDK.Context()
	client, err := context()
	if err != nil {
		return "", nil, fmt.Errorf("fabric-sdk failed to obtain client config %v", err)
	}

	org, orgExists := networkCfg.Organizations[client.IdentityConfig().Client().Organization]
	if !orgExists {
		// It seems the stored org keys are always in lowercase, go figure!
		org, orgExists = networkCfg.Organizations[strings.ToLower(client.IdentityConfig().Client().Organization)]
		if !orgExists {
			return "", nil, fmt.Errorf("fabric-sdk configuration error - default organization %s not found under organizations", client.IdentityConfig().Client().Organization)
		}
	}
	return client.IdentityConfig().Client().Organization, &org, nil
}

func (t *fabClient) buildTxnRequest(channelID string, chainCodeID string, args []string, transientData map[string][]byte) channel.Request {

	fcn := args[0]
	args = args[1:]

	argBytes := [][]byte{}
	for _, argStr := range args {
		argBytes = append(argBytes, []byte(argStr))
	}

	request := channel.Request{
		ChaincodeID:  chainCodeID,
		Fcn:          fcn,
		Args:         argBytes,
		TransientMap: transientData,
	}

	return request
}

func extractFuncNameFromArgs(args []string) string {
	return fmt.Sprintf("%v", args)
}

// InvokeCC invokes a chancode on the specified channel
//
func (t *fabClient) InvokeCC(channelID string, chainCodeID string, args []string, transientData map[string][]byte) (*CCResponse, error) {

	logger.Debugf("--> InvokeCC: %s %s %s", channelID, chainCodeID, extractFuncNameFromArgs(args))

	request := t.buildTxnRequest(channelID, chainCodeID, args, transientData)

	chClient, err := t.ChannelClient(channelID)
	if err != nil {
		return nil, err
	}
	defer t.CloseChannelClient(chClient)

	resp, err := chClient.Execute(request, channel.WithRetry(t.invokeRetryOpts))
	if err != nil {
		return nil, fmt.Errorf("fabClient invokeCC failed for %v: %v", args, err)
	}

	return t.extractCCResponse(&resp)
}

// extractCCResponse extracts chaincode response from TransactionProposalResponse
//
func (t *fabClient) extractCCResponse(txnResp *channel.Response) (*CCResponse, error) {

	/*
		if txnProposalResponse != nil && txnProposalResponse.Err != nil {
			return nil, txnProposalResponse.Err
		}
	*/
	txnProposalResponse := txnResp.Responses[0]
	if txnProposalResponse != nil && txnProposalResponse.Status != 200 {
		var errmsg string
		if txnProposalResponse.ProposalResponse != nil && txnProposalResponse.ProposalResponse.Response != nil {
			errmsg = txnProposalResponse.ProposalResponse.Response.Message
		}
		return nil, fmt.Errorf(errmsg)
	}
	ccResponse := CCResponse{
		FabricTxnID: string(txnResp.TransactionID),
	}
	//	txnProposalResponse.ProposalResponse.GetResponse().Payload
	if txnProposalResponse != nil && txnProposalResponse.ProposalResponse != nil {
		if resp := txnProposalResponse.ProposalResponse.GetResponse(); resp != nil {
			ccResponse.Payload = resp.Payload
			return &ccResponse, nil
		}
	}
	return &ccResponse, nil
}

// QueryCC queries a chancode on the specified channel
//
func (t *fabClient) QueryCC(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte) (*CCResponse, error) {
	logger.Debugf("--> QueryCC: %s %s %s", channelID, chainCodeID, extractFuncNameFromArgs(args))

	retryOpts := t.queryRetryOpts
	if maxAttempts > retryOpts.Attempts {
		retryOpts.Attempts = maxAttempts
	}

	chClient, err := t.ChannelClientQuery(channelID)
	if err != nil {
		return nil, err
	}
	defer t.CloseChannelClient(chClient)

	resp, err := chClient.Query(t.buildTxnRequest(channelID, chainCodeID, args, transientData), channel.WithRetry(retryOpts))
	if err != nil {
		return nil, err
	}

	return t.extractCCResponse(&resp)
}

func (t *fabClient) QueryCCAtPeer(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte, peerURL string) (*CCResponse, error) {
	logger.Debugf("--> QueryCCAtPeer: %s %s %s", channelID, chainCodeID, extractFuncNameFromArgs(args))

	retryOpts := t.queryRetryOpts
	if maxAttempts > retryOpts.Attempts {
		retryOpts.Attempts = maxAttempts
	}

	chClient, err := t.ChannelClientQuery(channelID)
	if err != nil {
		return nil, err
	}
	defer t.CloseChannelClient(chClient)

	resp, err := chClient.Query(t.buildTxnRequest(channelID, chainCodeID, args, transientData), channel.WithTargetFilter(peerfilter.URLFilter{PeerURL: peerURL}), channel.WithRetry(retryOpts))
	if err != nil {
		return nil, err
	}

	return t.extractCCResponse(&resp)
}

func (t *fabClient) QueryCCAtMSP(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte, MSPID string) (*CCResponse, error) {
	logger.Debugf("--> QueryCCAtMSP: %s %s %s", channelID, chainCodeID, extractFuncNameFromArgs(args))

	retryOpts := t.queryRetryOpts
	if maxAttempts > retryOpts.Attempts {
		retryOpts.Attempts = maxAttempts
	}

	chClient, err := t.ChannelClientQuery(channelID)
	if err != nil {
		return nil, err
	}
	defer t.CloseChannelClient(chClient)

	resp, err := chClient.Query(t.buildTxnRequest(channelID, chainCodeID, args, transientData), channel.WithTargetFilter(peerfilter.MSPFilter{MSPID: MSPID}), channel.WithRetry(retryOpts))
	if err != nil {
		return nil, err
	}

	return t.extractCCResponse(&resp)
}

// QueryCCAtOwnOrg implementation of QueryCCAtOwnOrg of Client interface
func (t *fabClient) QueryCCAtOwnOrg(maxAttempts int, channelID string, chainCodeID string, args []string, transientData map[string][]byte) (*CCResponse, error) {
	return t.QueryCCAtMSP(maxAttempts, channelID, chainCodeID, args, transientData, t.orgConfig.MSPID)
}

// ChannelClient returns a channelClient for the specified channel ID
func (t *fabClient) ChannelClient(channelID string) (*channel.Client, error) {

	if !t.shareChannelClient {
		return t.newChannelClient(channelID)
	}

	t.chclientMutex.RLock()
	chClient, exists := t.chClients[channelID]
	t.chclientMutex.RUnlock()

	if !exists {
		var err error
		chClient, err = t.newChannelClient(channelID)
		if err != nil {
			return nil, err
		}
		t.chclientMutex.Lock()
		t.chClients[channelID] = chClient
		t.chclientMutex.Unlock()
	}
	return chClient, nil
}

// ConsortiumChannel return channel ID of consortium channel
func (t *fabClient) ConsortiumChannelID() string {
	return consortiumChannelID
}

// OrgChannel return channel ID of own org
func (t *fabClient) OrgChannelID() string {
	return t.orgChannelID
}

func (t *fabClient) newChannelClient(channelID string) (*channel.Client, error) {

	chProvider := t.fabricSDK.ChannelContext(channelID, fabsdk.WithUser(t.userID), fabsdk.WithOrg(t.orgName))
	chClient, err := channel.New(chProvider)
	if err != nil {
		return nil, fmt.Errorf("newChannelClient: failed to obtain ChannelClient for %s, channel: %v", channelID, err)
	}
	return chClient, nil
}

func (t *fabClient) NewChannelProvider(channelID string) context.ChannelProvider {

	return t.fabricSDK.ChannelContext(channelID, fabsdk.WithUser(t.userID), fabsdk.WithOrg(t.orgName))
}

// ChannelClient returns a channelClient for the specified channel ID
func (t *fabClient) ChannelClientQuery(channelID string) (*channel.Client, error) {

	if !t.shareChannelClient {
		return t.newChannelClientQuery(channelID)
	}

	t.chclientqueryMutex.RLock()
	chClient, exists := t.chClientsQuery[channelID]
	t.chclientqueryMutex.RUnlock()

	if !exists {
		var err error
		chClient, err = t.newChannelClientQuery(channelID)
		if err != nil {
			return nil, err
		}
		t.chclientqueryMutex.Lock()
		t.chClientsQuery[channelID] = chClient
		t.chclientqueryMutex.Unlock()
	}
	return chClient, nil
}

func (t *fabClient) newChannelClientQuery(channelID string) (*channel.Client, error) {

	var err error
	var chClient *channel.Client
	chProvider := t.fabricSDKQuery.ChannelContext(channelID, fabsdk.WithUser(t.userID), fabsdk.WithOrg(t.orgName))
	chClient, err = channel.New(chProvider)

	if err != nil {
		return nil, fmt.Errorf("newChannelClientQuery: failed to obtain ChannelClient for %s, channel: %v", channelID, err)
	}
	return chClient, nil
}

// CloseChannelClient (logically) closes a channel client obtained from a previous ChannelClient() call
func (t *fabClient) CloseChannelClient(client *channel.Client) {
	//if !t.shareChannelClient && client != nil {
	//	client.Close()
	//}
	// don't close if shared instance is used
}

// Close closes this client
func (t *fabClient) Close() {
	//if !t.shareChannelClient {
	//	return
	//}
	//
	//t.chclientMutex.Lock()
	//for _, chclient := range t.chClients {
	//	chclient.Close()
	//}
	//t.chclientMutex.Unlock()
	t.fabricSDK.Close()
	t.fabricSDKQuery.Close()
}

// EventTimeoutSeconds returns the configured or default timeout value in seconds for chaincode/transaction event
func (t *fabClient) EventTimeoutSeconds() int {
	return t.eventTimeoutSeconds
}
