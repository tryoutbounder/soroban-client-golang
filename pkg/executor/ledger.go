package executor

import (
	"context"
	"fmt"

	"github.com/stellar/go/xdr"
	soroban "github.com/tryOutbounder/soroban-client-golang/pkg/rpc"
	"github.com/tryOutbounder/soroban-client-golang/pkg/rpc/protocol"
)

func LedgerEntryCall(
	rpc *soroban.RpcClient,
	contractAddress xdr.ScAddress,
	ledgerKeys []xdr.LedgerKey,
) (map[xdr.LedgerKey]xdr.LedgerEntryData, error) {

	keys := make([]string, len(ledgerKeys))
	for idx, ledgerKey := range ledgerKeys {
		encodedKey, err := ledgerKey.MarshalBinaryBase64()
		if err != nil {
			return nil, fmt.Errorf("error encoding ledger key at index %d: %w", idx, err)
		}

		keys[idx] = encodedKey
	}

	resp, err := rpc.GetLedgerEntries(
		context.TODO(),
		protocol.GetLedgerEntriesRequest{
			Keys: keys,
		},
	)

	if err != nil {
		return nil, err
	}

	result := make(map[xdr.LedgerKey]xdr.LedgerEntryData)
	for idx, entry := range resp.Entries {
		var ledgerKeyXdr xdr.LedgerKey

		err := xdr.SafeUnmarshalBase64(entry.KeyXDR, &ledgerKeyXdr)
		if err != nil {
			return nil, err
		}

		if ledgerKeyXdr.Equals(ledgerKeys[idx]) {

			var bodyXdr xdr.LedgerEntryData

			err := xdr.SafeUnmarshalBase64(entry.DataXDR, &bodyXdr)

			if err != nil {
				return nil, fmt.Errorf("error unmarshaling entry data at index %d: %w", idx, err)
			}

			result[ledgerKeyXdr] = bodyXdr
		}

	}

	return result, nil

}
