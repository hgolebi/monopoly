package neatnetwork

import cfg "monopoly/pkg/config"

// Price searches ask the network the same question at different prices and assume its answer is monotonic:
// a buyer that accepts a price accepts every lower one, a seller that accepts a price accepts every higher one.
//
// The search is exponential (galloping): it starts at lo and probes lo+1, lo+2, lo+4, ... steps until the answer
// changes, then finishes with a binary search between the last two probes. It needs ~2*log2(d) activations, where
// d is the distance from lo to the answer in steps, instead of d with a step-by-step search. Short distances are
// common (bids close to the current price), so this is cheaper than a plain binary search over the whole range.
// For a network that is not monotonic the result is some price at which the answer changes, not necessarily the first one.
//
// The searched prices are lo, lo+PRICE_STEP, lo+2*PRICE_STEP, ... and hi itself as the last price.

// priceGridSize returns the index of the last price on the grid between lo and hi.
func priceGridSize(lo, hi int) int {
	return (hi - lo + cfg.PRICE_STEP - 1) / cfg.PRICE_STEP
}

func priceAt(lo, hi, k int) int {
	return min(lo+k*cfg.PRICE_STEP, hi)
}

// highestAcceptedPrice returns the highest price between lo and hi accepted by a buyer.
// found is false when even lo is not accepted (or hi < lo).
func highestAcceptedPrice(lo, hi int, accept func(price int) bool) (price int, found bool) {
	if hi < lo || !accept(lo) {
		return 0, false
	}
	last := priceGridSize(lo, hi)
	low, high := 0, last // invariant: the price at low is accepted, prices above high are not
	for step := 1; low < high; step *= 2 {
		probe := min(low+step, high)
		if !accept(priceAt(lo, hi, probe)) {
			high = probe - 1
			break
		}
		low = probe
	}
	for low < high {
		mid := (low + high + 1) / 2
		if accept(priceAt(lo, hi, mid)) {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return priceAt(lo, hi, low), true
}

// lowestAcceptedPrice returns the lowest price between lo and hi accepted by a seller.
// found is false when even hi is not accepted (or hi < lo).
func lowestAcceptedPrice(lo, hi int, accept func(price int) bool) (price int, found bool) {
	if hi < lo {
		return 0, false
	}
	last := priceGridSize(lo, hi)
	rejected, high := -1, -1 // invariant: prices up to rejected are not accepted, the price at high is
	for step := 1; high == -1; step *= 2 {
		probe := min(rejected+step, last)
		if accept(priceAt(lo, hi, probe)) {
			high = probe
		} else if probe == last {
			return 0, false
		} else {
			rejected = probe
		}
	}
	low := rejected + 1
	for low < high {
		mid := (low + high) / 2
		if accept(priceAt(lo, hi, mid)) {
			high = mid
		} else {
			low = mid + 1
		}
	}
	return priceAt(lo, hi, low), true
}
