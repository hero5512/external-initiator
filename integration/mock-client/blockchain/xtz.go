package blockchain

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type XtzMonitorResponse struct {
	Hash           string   `json:"hash"`
	Level          int      `json:"level"`
	Proto          int      `json:"proto"`
	Predecessor    string   `json:"predecessor"`
	TimeStamp      string   `json:"timestamp"`
	ValidationPass int      `json:"validation_pass"`
	OperationsHash string   `json:"operations_hash"`
	Fitness        []string `json:"fitness"`
	Context        string   `json:"context"`
	ProtocolData   string   `json:"protocol_data"`
}

type XtzTransaction struct {
	Protocol string                  `json:"protocol"`
	ChainID  string                  `json:"chain_id"`
	Hash     string                  `json:"hash"`
	Branch   string                  `json:"branch"`
	Contents []XtzTransactionContent `json:"contents"`
}

type XtzTransactionContent struct {
	Kind         string                        `json:"kind"`
	Source       string                        `json:"source"`
	Fee          string                        `json:"fee"`
	Counter      string                        `json:"counter"`
	GasLimit     string                        `json:"gas_limit"`
	StorageLimit string                        `json:"storage_limit"`
	Amount       string                        `json:"amount"`
	Destination  string                        `json:"destination"`
	Parameters   interface{}                   `json:"parameters"`
	Metadata     XtzTransactionContentMetadata `json:"metadata"`
}

type XtzTransactionContentMetadata struct {
	BalanceUpdates           []XtzNullValue                `json:"balance_updates"`
	OperationResult          XtzNullValue                  `json:"operation_result"`
	InternalOperationResults *[]XtzInternalOperationResult `json:"internal_operation_results"`
}

type XtzInternalOperationResult struct {
	Kind        string      `json:"kind"`
	Source      string      `json:"source"`
	Nonce       int         `json:"nonce"`
	Amount      string      `json:"amount"`
	Destination string      `json:"destination"`
	Parameters  interface{} `json:"parameters"`
	Result      interface{} `json:"result"`
}

type XtzNullValue struct {
}

func HandleXtzMonitorRequest(chainId string) (XtzMonitorResponse, error) {
	return XtzMonitorResponse{
		Hash:           "8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D",
		Level:          0,
		Proto:          0,
		Predecessor:    "BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3",
		ValidationPass: 0,
		OperationsHash: "00000000000000000000000000000000000000000000000000000000",
		Context:        "00000000000000000000000000000000000000000000000000000000",
		ProtocolData:   "",
	}, nil
}

func HandleXtzOperationsRequest(blockId string) ([][]XtzTransaction, error) {
	addressFile, err := os.Open("/run/secrets/subscribed_address")
	if err != nil {
		log.Println("Couldn't open subscription address file: ", err)
		return nil, err
	}
	defer addressFile.Close()

	subscriptionAddress, err := ioutil.ReadAll(addressFile)
	if err != nil {
		log.Println("Couldn't read subscription address from file: ", err)
		return nil, err
	}

	transactionContents := []XtzTransactionContent{
		{
			Kind:         "transaction",
			Source:       "BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3BaDc0Ff3",
			Fee:          "666",
			Counter:      "666",
			GasLimit:     "6666",
			StorageLimit: "42",
			Amount:       "66666",
			Destination:  strings.TrimSpace(string(subscriptionAddress)),
			Parameters:   XtzNullValue{},
			Metadata: XtzTransactionContentMetadata{
				BalanceUpdates:           []XtzNullValue{},
				OperationResult:          XtzNullValue{},
				InternalOperationResults: &[]XtzInternalOperationResult{},
			},
		},
	}
	transactions := [][]XtzTransaction{
		{},
		{},
		{},
		{
			{
				Protocol: "nonsense",
				ChainID:  "nonsense",
				Hash:     "8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D8BADF00D",
				Branch:   "8BADF00D",
				Contents: transactionContents,
			},
		},
	}

	return transactions, nil
}
