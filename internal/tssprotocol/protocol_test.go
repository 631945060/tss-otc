package tssprotocol

import (
	"crypto/ecdsa"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bnb-chain/tss-lib/common"
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/ecdsa/signing"
	tsstest "github.com/bnb-chain/tss-lib/test"
	"github.com/bnb-chain/tss-lib/tss"
)

const parties, threshold = 3, 1

// TestKeygenThenTwoPartySign is an integration test of tss-lib v1.5.0. It
// starts three DKG state machines, then signs with two of the resulting key
// shares and verifies the final ECDSA signature under the DKG public key.
func TestKeygenThenTwoPartySign(t *testing.T) {
	fixtures, _, err := keygen.LoadKeygenTestFixtures(parties)
	if err != nil {
		t.Fatal(err)
	}
	ids := tss.GenerateTestPartyIDs(parties)
	ctx := tss.NewPeerContext(ids)
	out := make(chan tss.Message, 128)
	done := make(chan keygen.LocalPartySaveData, parties)
	errs := make(chan *tss.Error, 16)
	local := make([]tss.Party, parties)
	for i := range ids {
		p := tss.NewParameters(tss.S256(), ctx, ids[i], parties, threshold)
		local[i] = keygen.NewLocalParty(p, out, done, fixtures[i].LocalPreParams)
		go func(party tss.Party) {
			if e := party.Start(); e != nil {
				errs <- e
			}
		}(local[i])
	}
	keys := make([]keygen.LocalPartySaveData, parties)
	completed := 0
	deadline := time.After(90 * time.Second)
	for completed < parties {
		select {
		case err := <-errs:
			t.Fatal(err)
		case msg := <-out:
			if to := msg.GetTo(); to == nil {
				for _, p := range local {
					if p.PartyID().Index != msg.GetFrom().Index {
						go tsstest.SharedPartyUpdater(p, msg, errs)
					}
				}
			} else {
				go tsstest.SharedPartyUpdater(local[to[0].Index], msg, errs)
			}
		case save := <-done:
			idx, e := save.OriginalIndex()
			if e != nil {
				t.Fatal(e)
			}
			keys[idx] = save
			completed++
		case <-deadline:
			t.Fatal("DKG timed out")
		}
	}
	if keys[0].ECDSAPub.X() == nil || keys[0].ECDSAPub.Y() == nil {
		t.Fatal("DKG did not return public key")
	}

	selectedIDs := tss.SortedPartyIDs{ids[0], ids[1]}
	signCtx := tss.NewPeerContext(selectedIDs)
	signOut := make(chan tss.Message, 128)
	signDone := make(chan common.SignatureData, 2)
	signErrs := make(chan *tss.Error, 16)
	message := big.NewInt(20260916)
	signers := make([]tss.Party, 2)
	for i := 0; i < 2; i++ {
		p := tss.NewParameters(tss.S256(), signCtx, selectedIDs[i], 2, threshold)
		signers[i] = signing.NewLocalParty(message, p, keys[i], signOut, signDone)
		go func(party tss.Party) {
			if e := party.Start(); e != nil {
				signErrs <- e
			}
		}(signers[i])
	}
	var count int32
	var signature common.SignatureData
	deadline = time.After(90 * time.Second)
	for atomic.LoadInt32(&count) < 2 {
		select {
		case err := <-signErrs:
			t.Fatal(err)
		case msg := <-signOut:
			if to := msg.GetTo(); to == nil {
				for _, p := range signers {
					if p.PartyID().Index != msg.GetFrom().Index {
						go tsstest.SharedPartyUpdater(p, msg, signErrs)
					}
				}
			} else {
				go tsstest.SharedPartyUpdater(signers[to[0].Index], msg, signErrs)
			}
		case sig := <-signDone:
			signature = sig
			atomic.AddInt32(&count, 1)
		case <-deadline:
			t.Fatal("signing timed out")
		}
	}
	pub := ecdsa.PublicKey{Curve: tss.S256(), X: keys[0].ECDSAPub.X(), Y: keys[0].ECDSAPub.Y()}
	if !ecdsa.Verify(&pub, message.Bytes(), new(big.Int).SetBytes(signature.R), new(big.Int).SetBytes(signature.S)) {
		t.Fatal("ECDSA verification failed")
	}
}
