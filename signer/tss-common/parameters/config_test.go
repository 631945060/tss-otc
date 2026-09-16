package tssconfig

import "testing"

func TestTwoOfThreeParameterMapping(t *testing.T) {
	params, err := New2of3Parameters()
	if err != nil {
		t.Fatal(err)
	}
	if len(params) != 3 || params[0].Threshold() != 1 || params[0].Threshold()+1 != 2 {
		t.Fatalf("unexpected 2-of-3 mapping: parties=%d threshold=%d", len(params), params[0].Threshold())
	}
}
