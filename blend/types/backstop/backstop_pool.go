package backstop

import (
	"fmt"

	"github.com/stellar/go/xdr"
	"github.com/tryoutbounder/soroban-client-golang/blend/types"
	"github.com/tryoutbounder/soroban-client-golang/pkg/executor"
	"github.com/tryoutbounder/soroban-client-golang/pkg/helpers"
	soroban "github.com/tryoutbounder/soroban-client-golang/pkg/rpc"
)

type BackstopPoolBalance struct {
	Shares float64
	Tokens float64
	Q4w    float64
}

func LoadPoolBalance(
	rpc *soroban.RpcClient,
	backstopContract string,
	poolContract string,

) (*BackstopPoolBalance, error) {
	backstopAddress, err := helpers.ContractAddressToScAddress(backstopContract)
	if err != nil {
		return nil, err
	}

	poolAddress, err := helpers.ContractAddressToScAddress(poolContract)
	if err != nil {
		return nil, err
	}

	poolBalanceSymbol := xdr.ScSymbol("PoolBalance")

	poolBalanceKeyVec := &xdr.ScVec{
		xdr.ScVal{
			Type: xdr.ScValTypeScvSymbol,
			Sym:  &poolBalanceSymbol,
		},
		xdr.ScVal{
			Type:    xdr.ScValTypeScvAddress,
			Address: &poolAddress,
		},
	}

	poolBalanceKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract:   backstopAddress,
			Key:        xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &poolBalanceKeyVec},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	ledgerKeys := []xdr.LedgerKey{poolBalanceKey}

	entries, err := executor.LedgerEntryCall(rpc, backstopAddress, ledgerKeys)
	if err != nil {
		return nil, err
	}

	var entry xdr.LedgerEntryData
	found := false
	for k, v := range entries {
		if k.Equals(poolBalanceKey) {
			found = true
			entry = v
		}
	}

	if !found {
		return nil, fmt.Errorf("pool balance entry not found for pool %s", poolContract)
	}

	data, ok := entry.ContractData.Val.GetMap()
	if !ok {

		return nil, fmt.Errorf("contract data value is not a map for pool %s", poolContract)
	}

	poolBalance := &BackstopPoolBalance{}
	for _, scVal := range *data {
		key := scVal.Key
		val := scVal.Val

		sym, ok := key.GetSym()
		if !ok {
			return nil, fmt.Errorf("expected symbol key in pool balance data for pool %s", poolContract)
		}

		switch sym {
		case "shares":
			i128, ok := val.GetI128()
			if !ok {
				return nil, fmt.Errorf("expected i128 value for shares in pool %s", poolContract)
			}
			poolBalance.Shares = helpers.I128ToFloat64(i128, types.SCALAR_7)

		case "tokens":
			i128, ok := val.GetI128()
			if !ok {
				return nil, fmt.Errorf("expected i128 value for tokens in pool %s", poolContract)
			}
			poolBalance.Tokens = helpers.I128ToFloat64(i128, types.SCALAR_7)

		case "q4w":
			i128, ok := val.GetI128()
			if !ok {
				return nil, fmt.Errorf("expected i128 value for q4w in pool %s", poolContract)
			}
			poolBalance.Q4w = helpers.I128ToFloat64(i128, types.SCALAR_7)

		}

	}
	return poolBalance, nil

}
