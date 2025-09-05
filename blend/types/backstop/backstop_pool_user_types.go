package backstop

import (
	"fmt"
	"time"

	"github.com/stellar/go/xdr"
	"github.com/tryOutbounder/soroban-client-golang/blend/types"
	"github.com/tryOutbounder/soroban-client-golang/pkg/executor"
	"github.com/tryOutbounder/soroban-client-golang/pkg/helpers"
	soroban "github.com/tryOutbounder/soroban-client-golang/pkg/rpc"
)

type Q4W struct {
	Amount     float64
	Expiration time.Time
}

type BackstopUserBalance struct {
	Shares      float64
	Q4W         []Q4W
	UnlockedQ4W float64
	TotalQ4W    float64
}

type BackstopUserEmissions struct {
	Index   int64
	Accrued float64
}
type BackstopPoolUser struct {
	Balance   *BackstopUserBalance
	Emissions *BackstopUserEmissions
}

func LoadBackstopUser(
	rpc *soroban.RpcClient,
	backstopContract string,
	poolContract string,
	userAddress string,

) (*BackstopPoolUser, error) {
	backstopAddress, err := helpers.ContractAddressToScAddress(backstopContract)
	if err != nil {
		return nil, err
	}

	poolAddress, err := helpers.ContractAddressToScAddress(poolContract)
	if err != nil {
		return nil, err
	}

	userScAddress, err := helpers.StellarAddressToScAddress(userAddress)
	if err != nil {
		return nil, err
	}

	// Create UserBalance ledger key
	userBalanceSymbol := xdr.ScSymbol("UserBalance")
	poolSymbol := xdr.ScSymbol("pool")
	userSymbol := xdr.ScSymbol("user")

	userBalanceMap := &xdr.ScMap{
		xdr.ScMapEntry{
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvSymbol,
				Sym:  &poolSymbol,
			},
			Val: xdr.ScVal{
				Type:    xdr.ScValTypeScvAddress,
				Address: &poolAddress,
			},
		},
		xdr.ScMapEntry{
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvSymbol,
				Sym:  &userSymbol,
			},
			Val: xdr.ScVal{
				Type:    xdr.ScValTypeScvAddress,
				Address: &userScAddress,
			},
		},
	}

	userBalanceKeyVec := &xdr.ScVec{
		xdr.ScVal{
			Type: xdr.ScValTypeScvSymbol,
			Sym:  &userBalanceSymbol,
		},
		xdr.ScVal{
			Type: xdr.ScValTypeScvMap,
			Map:  &userBalanceMap,
		},
	}

	userBalanceKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract:   backstopAddress,
			Key:        xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &userBalanceKeyVec},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	// Create UEmisData ledger key
	uEmisDataSymbol := xdr.ScSymbol("UEmisData")

	uEmisDataMap := &xdr.ScMap{
		xdr.ScMapEntry{
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvSymbol,
				Sym:  &poolSymbol,
			},
			Val: xdr.ScVal{
				Type:    xdr.ScValTypeScvAddress,
				Address: &poolAddress,
			},
		},
		xdr.ScMapEntry{
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvSymbol,
				Sym:  &userSymbol,
			},
			Val: xdr.ScVal{
				Type:    xdr.ScValTypeScvAddress,
				Address: &userScAddress,
			},
		},
	}

	uEmisDataKeyVec := &xdr.ScVec{
		xdr.ScVal{
			Type: xdr.ScValTypeScvSymbol,
			Sym:  &uEmisDataSymbol,
		},
		xdr.ScVal{
			Type: xdr.ScValTypeScvMap,
			Map:  &uEmisDataMap,
		},
	}

	uEmisDataKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract:   backstopAddress,
			Key:        xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &uEmisDataKeyVec},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	ledgerKeys := []xdr.LedgerKey{userBalanceKey, uEmisDataKey}

	entries, err := executor.LedgerEntryCall(rpc, backstopAddress, ledgerKeys)
	if err != nil {
		return nil, err
	}

	backstopUser := &BackstopPoolUser{}

	for ledgerKey, entry := range entries {
		if entry.ContractData == nil {
			return nil, fmt.Errorf("contract data is nil for ledger entry")
		}

		data, ok := entry.ContractData.Val.GetMap()
		if !ok {
			return nil, fmt.Errorf("contract data val is not a map")
		}

		if data == nil {
			return nil, fmt.Errorf("contract data map is nil")
		}

		if ledgerKey.Equals(userBalanceKey) {
			balance, err := extractUserBalance(*data)
			if err != nil {
				return nil, err
			}

			// Calculate total and unlocked Q4W
			totalQ4W := 0.0
			unlockedQ4W := 0.0
			currentTime := time.Now()

			for _, q4w := range balance.Q4W {
				totalQ4W += q4w.Amount
				if currentTime.After(q4w.Expiration) {
					unlockedQ4W += q4w.Amount
				}
			}

			balance.TotalQ4W = totalQ4W
			balance.UnlockedQ4W = unlockedQ4W

			backstopUser.Balance = balance

		}

		if ledgerKey.Equals(uEmisDataKey) {
			emissions, err := extractUserEmissions(*data)
			if err != nil {
				return nil, err
			}

			backstopUser.Emissions = emissions

		}

	}

	return backstopUser, nil
}

func extractUserBalance(data xdr.ScMap) (*BackstopUserBalance, error) {
	balance := &BackstopUserBalance{}
	for _, scVal := range data {
		key, ok := scVal.Key.GetSym()
		if !ok {
			return nil, fmt.Errorf("failed to get symbol from key")
		}
		val := scVal.Val

		switch key {
		case "shares":
			i128, ok := val.GetI128()
			if !ok {
				return nil, fmt.Errorf("shares val is not an i128")
			}

			balance.Shares = helpers.I128ToFloat64(i128, types.SCALAR_7)

		case "q4w":
			vec, ok := val.GetVec()
			if !ok {
				return nil, fmt.Errorf("q4w val is not a vector")
			}

			if vec == nil {
				return nil, fmt.Errorf("q4w vec is nil")
			}

			queuedWithdrawals := make([]Q4W, len(*vec))
			for i, queued := range *vec {
				q4wMap, ok := queued.GetMap()
				if !ok {
					return nil, fmt.Errorf("queued item is not a map")
				}

				q4w, err := extractQ4wData(*q4wMap)
				if err != nil {
					return nil, err
				}

				queuedWithdrawals[i] = *q4w

			}
			balance.Q4W = queuedWithdrawals

		}

	}

	return balance, nil

}

func extractUserEmissions(data xdr.ScMap) (*BackstopUserEmissions, error) {
	emissions := &BackstopUserEmissions{}
	for _, scVal := range data {
		key, ok := scVal.Key.GetSym()
		if !ok {
			return nil, fmt.Errorf("failed to get symbol from key")
		}

		val := scVal.Val

		switch key {
		case "accrued":
			i128, ok := val.GetI128()
			if !ok {
				return nil, fmt.Errorf("accrued val is not an i128")
			}

			emissions.Accrued = helpers.I128ToFloat64(i128, types.SCALAR_7)

		case "index":
			i128, ok := val.GetI128()
			if !ok {
				return nil, fmt.Errorf("index val is not an i128")
			}

			emissions.Index = helpers.I128ToInt64(i128)
		}

	}

	return emissions, nil

}

func extractQ4wData(data xdr.ScMap) (*Q4W, error) {
	q4w := Q4W{}
	for _, scVal := range data {
		key, ok := scVal.Key.GetSym()
		if !ok {
			return nil, fmt.Errorf("failed to get symbol from key")
		}

		val := scVal.Val

		switch key {
		case "amount":
			i128, ok := val.GetI128()
			if !ok {
				return nil, fmt.Errorf("amount val is not an i128")
			}

			q4w.Amount = helpers.I128ToFloat64(i128, types.SCALAR_7)

		case "exp":
			u64, ok := val.GetU64()
			if !ok {
				return nil, fmt.Errorf("exp val is not an i128")
			}

			q4w.Expiration = time.Unix(int64(u64), 0)

		}

	}

	return &q4w, nil

}
