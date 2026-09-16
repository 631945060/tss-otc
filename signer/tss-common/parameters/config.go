// Package tssconfig records the tss-lib configuration used by this demo.
package tssconfig

import (
	"fmt"

	"github.com/bnb-chain/tss-lib/tss"
)

const (
	PartyCount = 3
	// LibraryThreshold is the polynomial degree. A degree of one requires two
	// participants, therefore it represents the 2-of-3 policy used here.
	LibraryThreshold = 1
)

func New2of3Parameters() ([]*tss.Parameters, error) {
	ids := tss.GenerateTestPartyIDs(PartyCount)
	ctx := tss.NewPeerContext(ids)
	params := make([]*tss.Parameters, 0, PartyCount)
	for _, id := range ids {
		params = append(params, tss.NewParameters(tss.S256(), ctx, id, PartyCount, LibraryThreshold))
	}
	if params[0].Threshold()+1 != 2 {
		return nil, fmt.Errorf("unexpected threshold configuration")
	}
	return params, nil
}
