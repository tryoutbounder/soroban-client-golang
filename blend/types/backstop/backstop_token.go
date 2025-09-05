package backstop

import (
	"fmt"

	"github.com/stellar/go/xdr"
	"github.com/tryOutbounder/soroban-client-golang/blend/types"
	"github.com/tryOutbounder/soroban-client-golang/pkg/executor"
	"github.com/tryOutbounder/soroban-client-golang/pkg/helpers"
	soroban "github.com/tryOutbounder/soroban-client-golang/pkg/rpc"
)

type BackstopToken struct {
	ID             string
	BLND           float64
	USDC           float64
	Shares         float64
	BLNDPerLPToken float64
	USDCPerLPToken float64
	LPTokenPrice   float64
}

func LoadToken(
	rpc *soroban.RpcClient,
	cometContract string,
	blndTokenContract string,
	usdcTokenContract string,
) (*BackstopToken, error) {
	backstopTokenAddress, err := helpers.ContractAddressToScAddress(cometContract)
	if err != nil {
		return nil, err
	}
	blndTokenAddress, err := helpers.ContractAddressToScAddress(blndTokenContract)
	if err != nil {
		return nil, err
	}

	usdcTokenAddress, err := helpers.ContractAddressToScAddress(usdcTokenContract)
	if err != nil {
		return nil, err
	}

	recordSymbol := xdr.ScSymbol("AllRecordData")
	totalSharesSymbol := xdr.ScSymbol("TotalShares")

	recordKeyVec := &xdr.ScVec{
		{Type: xdr.ScValTypeScvSymbol, Sym: &recordSymbol},
	}

	totalSharesKeyVec := &xdr.ScVec{
		{Type: xdr.ScValTypeScvSymbol, Sym: &totalSharesSymbol},
	}

	recordDataKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract:   backstopTokenAddress,
			Key:        xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &recordKeyVec},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	totalSharesKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract:   backstopTokenAddress,
			Key:        xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: &totalSharesKeyVec},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	ledgerKeys := []xdr.LedgerKey{recordDataKey, totalSharesKey}

	entries, err := executor.LedgerEntryCall(rpc, backstopTokenAddress, ledgerKeys)

	if err != nil {
		return nil, err
	}

	recordDataFound := false
	totalSharesFound := false

	tokenData := &BackstopToken{
		ID: cometContract,
	}
	for ledgerKey, entry := range entries {
		if ledgerKey.Equals(recordDataKey) {
			recordDataFound = true
			recordData := entry.ContractData
			if recordData == nil {
				return nil, fmt.Errorf("record data is nil")
			}

			blndBalance, usdcBalance, err := extractTokenBalances(recordData, blndTokenAddress, usdcTokenAddress)
			if err != nil {
				return nil, err
			}

			tokenData.BLND = blndBalance
			tokenData.USDC = usdcBalance
		}

		if ledgerKey.Equals(totalSharesKey) {
			totalSharesFound = true
			totalSharesData := entry.ContractData

			shares, err := extractTotalShares(totalSharesData)
			if err != nil {
				return nil, err
			}

			tokenData.Shares = shares
			tokenData.BLNDPerLPToken = float64(tokenData.BLND) / float64(tokenData.Shares)
			tokenData.USDCPerLPToken = float64(tokenData.USDC) / float64(tokenData.Shares)
			tokenData.LPTokenPrice = (float64(tokenData.USDC) * 5) / float64(tokenData.Shares)
		}
	}

	if !recordDataFound || !totalSharesFound {
		return nil, fmt.Errorf("could not find all required ledger entries")
	}

	return tokenData, nil
}

func extractTokenBalances(recordData *xdr.ContractDataEntry, blndTokenAddress, usdcTokenAddress xdr.ScAddress) (float64, float64, error) {
	data, ok := recordData.Val.GetMap()
	if !ok {
		return 0, 0, fmt.Errorf("failed to get map from contract data value")
	}

	var blndBalance, usdcBalance float64

	for _, balance := range *data {
		balanceAddr, ok := balance.Key.GetAddress()
		if !ok {
			return 0, 0, fmt.Errorf("failed to get address from balance key")
		}

		if balanceAddr.Equals(blndTokenAddress) {
			balanceMap, ok := balance.Val.GetMap()
			if !ok {
				return 0, 0, fmt.Errorf("failed to get balance map for BLND")
			}
			for _, balanceEntry := range *balanceMap {
				if balanceSymbol, ok := balanceEntry.Key.GetSym(); ok && string(balanceSymbol) == "balance" {

					if blndBal, ok := balanceEntry.Val.GetI128(); ok {
						blndBalance = helpers.I128ToFloat64(blndBal, types.SCALAR_7)
					}
				}
			}
		}

		if balanceAddr.Equals(usdcTokenAddress) {
			balanceMap, ok := balance.Val.GetMap()
			if !ok {
				return 0, 0, fmt.Errorf("failed to get balance map for USDC")
			}
			for _, balanceEntry := range *balanceMap {
				if balanceSymbol, ok := balanceEntry.Key.GetSym(); ok && string(balanceSymbol) == "balance" {
					if usdcBal, ok := balanceEntry.Val.GetI128(); ok {
						usdcBalance = helpers.I128ToFloat64(usdcBal, types.SCALAR_7)
					}
				}
			}
		}
	}

	return blndBalance, usdcBalance, nil
}

func extractTotalShares(totalSharesData *xdr.ContractDataEntry) (float64, error) {
	if totalSharesData == nil {
		return 0, fmt.Errorf("total shares data is nil")
	}

	value, ok := totalSharesData.Val.GetI128()
	if !ok {
		return 0, fmt.Errorf("failed to get i128 from total shares value")
	}

	return helpers.I128ToFloat64(value, types.SCALAR_7), nil
}
