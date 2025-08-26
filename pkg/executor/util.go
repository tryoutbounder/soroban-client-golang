package executor

import (
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
)

func buildContractTx(
	contractAddress xdr.ScAddress,
	sourceAccount txnbuild.Account,
	args []xdr.ScVal,
	functionName xdr.ScSymbol,
) (*txnbuild.Transaction, error) {

	invokeHostOp := &txnbuild.InvokeHostFunction{
		HostFunction: xdr.HostFunction{
			Type: xdr.HostFunctionTypeHostFunctionTypeInvokeContract,
			InvokeContract: &xdr.InvokeContractArgs{
				ContractAddress: contractAddress,
				FunctionName:    functionName,
				Args:            args,
			},
		},
	}

	return txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount: sourceAccount,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewTimeout(30),
		},
		Operations: []txnbuild.Operation{invokeHostOp},
	})

}
