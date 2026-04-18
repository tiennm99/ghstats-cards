package card

import (
	"math"
	"strconv"
)

// niceTicks returns evenly-spaced tick values in [0, max] such that the step
// is a 1/2/5/10 × 10^k number and the tick count is roughly targetTicks.
//
// Mirrors d3.scaleLinear().nice() / d3.axisLeft().ticks(n) so charts built
// on top look visually consistent with the d3 reference.
func niceTicks(max float64, targetTicks int) []float64 {
	if max <= 0 || targetTicks <= 0 {
		return []float64{0}
	}
	rough := max / float64(targetTicks)
	exp := math.Pow(10, math.Floor(math.Log10(rough)))
	frac := rough / exp
	var step float64
	switch {
	case frac < 1.5:
		step = 1 * exp
	case frac < 3:
		step = 2 * exp
	case frac < 7:
		step = 5 * exp
	default:
		step = 10 * exp
	}

	out := []float64{}
	for v := 0.0; v <= max+step/1e9; v += step {
		out = append(out, v)
	}
	return out
}

// formatTick renders a float tick label. Integer-valued ticks drop decimals.
func formatTick(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
