package backstop

import (
	"fmt"
	"strings"

	"github.com/stellar/go/xdr"
	"github.com/tryOutbounder/soroban-client-golang/pkg/executor"
	"github.com/tryOutbounder/soroban-client-golang/pkg/helpers"
	soroban "github.com/tryOutbounder/soroban-client-golang/pkg/rpc"
)

type BackstopConfig struct {
	PublicEmitter string
	BlndTkn       string
	UsdcTkn       string
	BackstopTkn   string
	PoolFactory   string
	RewardZone    []string
}

func LoadConfig(
	rpc *soroban.RpcClient,
	backstopContract string,
) (*BackstopConfig, error) {
	backstopAddress, err := helpers.ContractAddressToScAddress(backstopContract)
	if err != nil {
		return nil, err
	}

	contractDataLedgerKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract: backstopAddress,
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvLedgerKeyContractInstance,
			},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	rewardZoneSymbol := xdr.ScSymbol("RZ")
	rewardZoneLedgerKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract: backstopAddress,
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvSymbol,
				Sym:  &rewardZoneSymbol,
			},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	ledgerKeys := []xdr.LedgerKey{contractDataLedgerKey, rewardZoneLedgerKey}

	entries, err := executor.LedgerEntryCall(rpc, backstopAddress, ledgerKeys)

	if err != nil {
		return nil, err
	}

	contractDataFound := false
	rewardZoneFound := false

	backstopConfig := &BackstopConfig{}
	for ledgerKey, entry := range entries {
		if ledgerKey.Equals(contractDataLedgerKey) {
			contractDataFound = true
			contractData := entry.ContractData
			if contractData == nil {
				return nil, fmt.Errorf("contract data is nil")
			}
			configData := contractData.Val.Instance.Storage
			backstopConfigData, err := extractBackstopConfigData(configData)
			if err != nil {
				return nil, err
			}

			backstopConfig.BlndTkn = backstopConfigData.BlndTkn
			backstopConfig.UsdcTkn = backstopConfigData.UsdcTkn
			backstopConfig.BackstopTkn = backstopConfigData.BackstopTkn
			backstopConfig.PoolFactory = backstopConfigData.PoolFactory
			backstopConfig.PublicEmitter = backstopConfigData.PublicEmitter

		}

		if ledgerKey.Equals(rewardZoneLedgerKey) {
			rewardZoneFound = true
			rewardZoneData := entry.ContractData

			rewardZoneAddresses, err := extractRewardZoneData(rewardZoneData)
			if err != nil {
				return nil, err
			}

			backstopConfig.RewardZone = rewardZoneAddresses
		}

	}

	if !contractDataFound || !rewardZoneFound {
		msg := "missing data:"
		missingItems := []string{}
		if !contractDataFound {
			missingItems = append(missingItems, "contract data")
		}
		if !rewardZoneFound {
			missingItems = append(missingItems, "reward zone")
		}
		return nil, fmt.Errorf("%s %s", msg, strings.Join(missingItems, ", "))
	}

	return backstopConfig, nil
}

func extractRewardZoneData(rewardZoneData *xdr.ContractDataEntry) ([]string, error) {
	if rewardZoneData == nil {
		return nil, fmt.Errorf("reward zone data is nil")
	}
	vec, ok := rewardZoneData.Val.GetVec()

	if !ok {
		return nil, fmt.Errorf("reward zone list is nil")
	}

	rewardZoneList := *vec

	rewardZoneAddresses := make([]string, len(rewardZoneList))

	for i, addr := range rewardZoneList {
		address, ok := addr.GetAddress()
		if !ok {
			return nil, fmt.Errorf("address.ContractId is nil")
		}
		contractId, err := helpers.EncodeContractAddress(*address.ContractId)
		if err != nil {
			return nil, err
		}

		rewardZoneAddresses[i] = contractId

	}

	return rewardZoneAddresses, nil

}

type backstopConfigData struct {
	PublicEmitter string
	BlndTkn       string
	UsdcTkn       string
	BackstopTkn   string
	PoolFactory   string
}

func extractBackstopConfigData(configData *xdr.ScMap) (*backstopConfigData, error) {
	mapData := *configData
	if mapData == nil {
		return nil, fmt.Errorf("config map is nil")
	}

	data := backstopConfigData{}

	for _, v := range mapData {
		key := v.Key
		val := v.Val

		if key.Type != xdr.ScValTypeScvSymbol {
			return nil, fmt.Errorf("config key is not a symbol type")
		}

		symbol := *key.Sym
		switch symbol {
		case "BLNDTkn":
			contractId, err := helpers.EncodeContractAddress(*val.Address.ContractId)
			if err != nil {
				return nil, err
			}
			data.BlndTkn = contractId
		case "BToken":
			contractId, err := helpers.EncodeContractAddress(*val.Address.ContractId)
			if err != nil {
				return nil, err
			}
			data.BackstopTkn = contractId
		case "USDCTkn":
			contractId, err := helpers.EncodeContractAddress(*val.Address.ContractId)
			if err != nil {
				return nil, err
			}
			data.UsdcTkn = contractId
		case "PoolFact":
			contractId, err := helpers.EncodeContractAddress(*val.Address.ContractId)
			if err != nil {
				return nil, err
			}
			data.PoolFactory = contractId
		case "Emitter":
			contractId, err := helpers.EncodeContractAddress(*val.Address.ContractId)
			if err != nil {
				return nil, err
			}
			data.PublicEmitter = contractId
		case "IsInit":
			// do nothing
		default:
			return nil, fmt.Errorf("invalid backstop instance storage key: should not contain %s", symbol)
		}

	}

	if data.BlndTkn == "" || data.UsdcTkn == "" || data.BackstopTkn == "" || data.PoolFactory == "" || data.PublicEmitter == "" {
		return nil, fmt.Errorf("incomplete backstop configuration: missing one or more required fields")
	}

	return &data, nil
}
