package strata

import "math/bits"

// BinomialCoefficient calculates the binomial coefficient,
// aka. n choose k.
//
// If the result overflows uint64 ok is false.
func BinomialCoefficient(n, k int) (r uint64, ok bool) {
	// https://en.wikipedia.org/wiki/Binomial_coefficient
	// https://stackoverflow.com/questions/15301885/best-way-of-calculating-n-choose-k

	if k == 0 {
		return 1, true
	}

	r, ok = BinomialCoefficient(n-1, k-1)
	if !ok {
		return r, ok
	}

	hi, lo := bits.Mul64(uint64(n), r)
	if uint64(k) <= hi {
		// overflow
		return ^uint64(0), false
	}

	// remainder is always zero
	quo, _ := bits.Div64(hi, lo, uint64(k))
	return quo, true
}
