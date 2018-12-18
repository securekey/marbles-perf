package api

// Marble data structure
//
type Marble struct {
	Id             string `json:"id"` //the fieldtags are needed to keep case from bouncing around
	Color          string `json:"color"`
	Size           int    `json:"size"` //size in mm of marble
	Owner          Owner  `json:"owner"`
	AdditionalData string `json:"additionalData,omitempty"`
}

// Owner (user) of a marble
//
type Owner struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Company  string `json:"company"`
}

// Transfer reprensents an ownership transfer request
//
type Transfer struct {
	MarbleId    string `json:"marbleId"`
	ToOwnerId   string `json:"toOwnerId"`
	AuthCompany string `json:"authCompany"` // should be fromOwner's company
}

// Response data structure for entity creation or transfer
//
type Response struct {
	Id    string `json:"id"`              // entity id (owner or marble)
	TxId  string `json:"txId"`            // fabric transaction id
	Error string `json:"error,omitempty"` // error message if any from chaincode
}

type ClearMarblesResponse struct {
	TxId    string `json:"txId"`            // fabric transaction id
	Error   string `json:"error,omitempty"` // error message if any from chaincode
	Found   int    `json:"found"`
	Deleted int    `json:"deleted"`
}

type InitBatchRequest struct {
	Concurrency     int  `json:"concurrency"`     // concurrency
	Iterations      int  `json:"iterations"`      //# iterations indicates the number of marbles transfers that each worker performs
	DelaySeconds    int  `json:"delaySeconds"`    // delay_seconds indicates the time the worker will wait between transfers
	ClearMarbles    bool `json:"clearMarbles"`    // clearMarbles indicates whether the client will delete all marbles from the ledger prior to the test
	ExtraDataLength int  `json:"extraDataLength"` // extraDataLength specifies the size of extra data attached to the marble to increase block size
}

type InitBatchResponse struct {
	BatchID string `json:"batchId"`
}

type BatchResult struct {
	Request                InitBatchRequest `json:"request"`
	Status                 string           `json:"status"`
	TotalSuccesses         int              `json:"totalSuccesses"`
	TotalFailures          int              `json:"totalFailures"`
	TotalSuccessSeconds    int              `json:"totalSuccessSeconds"`
	AverageTransferSeconds float64          `json:"averageTransferSeconds"`
	MinTransferSeconds     float64          `json:"minTransferSeconds"`
	MaxTransferSeconds     float64          `json:"maxTransferSeconds"`
}
