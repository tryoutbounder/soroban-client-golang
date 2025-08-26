package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/stellar/go/xdr"
	"github.com/tryOutbounder/soroban-client-golang/pkg/executor"
	soroban "github.com/tryOutbounder/soroban-client-golang/pkg/rpc"
	"github.com/tryOutbounder/soroban-client-golang/pkg/rpc/protocol"
)

func main() {
	sorobanClient := soroban.NewClient("https://soroban-rpc.creit.tech", http.DefaultClient)

	// poolAddr, _ := helpers.ContractAddressToScAddress("CAJJZSGMMM3PD7N33TAPHGBUGTB43OC73HVIK2L2G6BNGGGYOSSYBXBD")

	// backstopAddr, _ := helpers.ContractAddressToScAddress("CAQQR5SWBXKIGZKPBZDH3KM5GQ5GUTPKB7JAFCINLZBC5WXPJKRG3IM7")

	// userAddr, _ := helpers.StellarAddressToScAddress("GDW72XXDCZ2T2YZ5I5EEA55RYGUW3RF7S4VAMFUKMUHAHC3AOWERGUCT")
	// fmt.Println(executor.ContractCall(sorobanClient, poolAddr, &txnbuild.SimpleAccount{
	// 	AccountID: "GDW72XXDCZ2T2YZ5I5EEA55RYGUW3RF7S4VAMFUKMUHAHC3AOWERGUCT",
	// }, []xdr.ScVal{}, xdr.ScSymbol("get_reserve_list")))

	// resp, _ := (executor.LedgerEntryCall(sorobanClient, backstopAddr, []xdr.LedgerKey{UserBalanceLedgerKey(poolAddr, userAddr, backstopAddr)}))
	// var res xdr.LedgerEntryData
	// for _, d := range resp {
	// 	res = d
	// }
	// formattedResp, _ := json.MarshalIndent(res.ContractData.Val.Map, "", "  ")
	// fmt.Println(string(formattedResp))

	ledgers, _ := sorobanClient.GetHealth(context.TODO())
	newAuction := xdr.ScString("fill_auction")
	resp, _, err := executor.EventCall(
		sorobanClient,
		ledgers.OldestLedger,
		ledgers.LatestLedger,
		[]protocol.EventFilter{
			protocol.EventFilter{
				ContractIDs: []string{"CAJJZSGMMM3PD7N33TAPHGBUGTB43OC73HVIK2L2G6BNGGGYOSSYBXBD"},
				Topics: []protocol.TopicFilter{
					protocol.TopicFilter{
						protocol.SegmentFilter{
							ScVal: &xdr.ScVal{
								Type: xdr.ScValTypeScvSymbol,
								Str:  &newAuction,
							},
						},
					},
				},
			},
		},
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	events := resp["CAJJZSGMMM3PD7N33TAPHGBUGTB43OC73HVIK2L2G6BNGGGYOSSYBXBD"]

	lastEvent := events[len(events)-1]
	fmt.Println(lastEvent.Body)

}

func UserBalanceLedgerKey(
	poolAddr xdr.ScAddress,
	userAddr xdr.ScAddress,
	backstopAddr xdr.ScAddress,
) xdr.LedgerKey {
	scVec := &xdr.ScVec{}
	scMap := &xdr.ScMap{}

	userBalanceSymbol := xdr.ScSymbol("UserBalance")
	poolSymbol := xdr.ScSymbol("pool")
	userSymbol := xdr.ScSymbol("user")

	*scMap = append(
		*scMap,
		xdr.ScMapEntry{
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvSymbol,
				Sym:  &poolSymbol,
			},
			Val: xdr.ScVal{
				Type:    xdr.ScValTypeScvAddress,
				Address: &poolAddr,
			},
		},
		xdr.ScMapEntry{
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvSymbol,
				Sym:  &userSymbol,
			},
			Val: xdr.ScVal{
				Type:    xdr.ScValTypeScvAddress,
				Address: &userAddr,
			},
		},
	)

	*scVec = append(
		*scVec,
		xdr.ScVal{
			Type: xdr.ScValTypeScvSymbol,
			Sym:  &userBalanceSymbol,
		},
		xdr.ScVal{
			Type: xdr.ScValTypeScvMap,
			Map:  &scMap,
		},
	)

	ledgerKey := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.LedgerKeyContractData{
			Contract: backstopAddr,
			Key: xdr.ScVal{
				Type: xdr.ScValTypeScvVec,
				Vec:  &scVec,
			},
			Durability: xdr.ContractDataDurabilityPersistent,
		},
	}

	return ledgerKey
}
