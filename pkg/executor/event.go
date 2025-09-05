package executor

import (
	"context"
	"fmt"

	"github.com/stellar/go/xdr"
	soroban "github.com/tryoutbounder/soroban-client-golang/pkg/rpc"
	"github.com/tryoutbounder/soroban-client-golang/pkg/rpc/protocol"
)

type Event struct {
	Topics []xdr.ScVal
	Body   xdr.ScVal
}

type PaginationInfo struct {
	OldestLedger uint32
	LatestLedger uint32
	Cursor       string
}

func EventCall(
	rpc *soroban.RpcClient,
	startLedger uint32,
	endLedger uint32,
	filters []protocol.EventFilter,
	paginationOptions *protocol.PaginationOptions,

) (
	map[string][]Event,
	*PaginationInfo,
	error,
) {

	if startLedger > endLedger {
		return nil, nil, fmt.Errorf("startLedger (%d) cannot be greater than endLedger (%d)", startLedger, endLedger)
	}

	events, err := rpc.GetEvents(
		context.TODO(),
		protocol.GetEventsRequest{
			StartLedger: startLedger,
			EndLedger:   endLedger,
			Filters:     filters,
			Pagination:  paginationOptions,
		},
	)

	if err != nil {
		return nil, nil, err
	}

	eventsResp := make(map[string][]Event)
	for _, event := range events.Events {

		topicsXdr := make([]xdr.ScVal, len(event.TopicXDR))

		for idx, topic := range event.TopicXDR {
			var topicXdr xdr.ScVal
			err := xdr.SafeUnmarshalBase64(topic, &topicXdr)
			if err != nil {
				return nil, nil, err
			}

			topicsXdr[idx] = topicXdr
		}

		var eventBodyXdr xdr.ScVal
		err := xdr.SafeUnmarshalBase64(event.ValueXDR, &eventBodyXdr)
		if err != nil {
			return nil, nil, err
		}

		formattedEvent := Event{
			Topics: topicsXdr,
			Body:   eventBodyXdr,
		}

		if _, ok := eventsResp[event.ContractID]; !ok {
			eventsResp[event.ContractID] = []Event{}
		}
		eventsResp[event.ContractID] = append(eventsResp[event.ContractID], formattedEvent)
	}

	paginationInfo := &PaginationInfo{
		OldestLedger: events.OldestLedger,
		LatestLedger: events.LatestLedger,
		Cursor:       events.Cursor,
	}
	return eventsResp, paginationInfo, nil
}

func WildcardSwitch(exactlyOne bool) *string {
	wildcard := "**"
	if exactlyOne {
		wildcard = "*"
	}

	return &wildcard
}
