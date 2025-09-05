package executor

import (
	"context"
	"fmt"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
	soroban "github.com/tryoutbounder/soroban-client-golang/pkg/rpc"
	"github.com/tryoutbounder/soroban-client-golang/pkg/rpc/protocol"
)

func SimulateContractCall(
	rpc *soroban.RpcClient,
	contractAddress xdr.ScAddress,
	sourceAccount txnbuild.Account,
	args []xdr.ScVal,
	functionName xdr.ScSymbol,
) (*xdr.ScVal, error) {

	transactionXdr, err := buildContractTx(contractAddress, sourceAccount, args, functionName)
	if err != nil {
		return nil, err
	}

	transactionBase64, err := transactionXdr.Base64()
	if err != nil {
		return nil, err
	}

	response, err := rpc.SimulateTransaction(
		context.TODO(),
		protocol.SimulateTransactionRequest{
			Transaction: transactionBase64,
		},
	)

	if err != nil {
		return nil, err
	}

	var responseScVal xdr.ScVal

	if len(response.Results) != 1 {
		return nil, fmt.Errorf("unexpected number of simulation results: %d", len(response.Results))
	}

	err = xdr.SafeUnmarshalBase64(
		*response.Results[0].ReturnValueXDR,
		&responseScVal,
	)

	if err != nil {
		return nil, err
	}

	return &responseScVal, err
}

// Untested
func SubmitContractCall(
	rpc *soroban.RpcClient,
	contractAddress xdr.ScAddress,
	sourceAccount txnbuild.Account,
	args []xdr.ScVal,
	functionName xdr.ScSymbol,
	networkPassphrase string,
	signingKeypairs []*keypair.Full,
) (string, error) {
	transactionXdr, err := buildContractTx(contractAddress, sourceAccount, args, functionName)
	if err != nil {
		return "", err
	}

	for _, keypair := range signingKeypairs {
		transactionXdr, err = transactionXdr.Sign(
			networkPassphrase,
			keypair,
		)
		if err != nil {
			return "", err
		}
	}

	transactionBase64, err := transactionXdr.Base64()
	if err != nil {
		return "", err
	}

	response, err := rpc.SendTransaction(
		context.TODO(),
		protocol.SendTransactionRequest{
			Transaction: transactionBase64,
		},
	)
	if err != nil {
		return "", err
	}

	if response.ErrorResultXDR != "" {
		var xdrErr xdr.ScError

		err := xdr.SafeUnmarshalBase64(response.ErrorResultXDR, &xdrErr)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("contract error: %+v", xdrErr)
	}
	return response.Hash, nil
}
