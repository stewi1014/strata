package strata

import (
	"fmt"
	"testing"
)

func TestBinomialCoefficient(t *testing.T) {
	tests := []struct {
		n      int
		k      int
		wantR  uint64
		wantOk bool
	}{
		{
			n:      3,
			k:      1,
			wantR:  3,
			wantOk: true,
		},
		{
			n:      52,
			k:      5,
			wantR:  2598960,
			wantOk: true,
		},
		{
			n:      52,
			k:      7,
			wantR:  133784560,
			wantOk: true,
		},
		{
			n:      200,
			k:      20,
			wantR:  ^uint64(0),
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v choose %v", tt.n, tt.k), func(t *testing.T) {
			gotR, gotOk := BinomialCoefficient(tt.n, tt.k)
			if gotR != tt.wantR {
				t.Errorf("BinomialCoefficient() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotOk != tt.wantOk {
				t.Errorf("BinomialCoefficient() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
