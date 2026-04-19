package card

import (
	"fmt"
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

// formatTick renders a float tick label, abbreviating thousands / millions /
// billions so every possible y-axis label fits within ≤4 characters. The
// leftPad gutter of every chart card is sized for ≤4 chars at 10 px font,
// so anything wider would overflow past the card frame for busy profiles
// (1000+ monthly commits, 10k+ yearly contributions, etc).
//
// Examples:
//
//	999       -> "999"
//	1_000     -> "1k"
//	1_500     -> "1.5k"
//	12_345    -> "12k"
//	1_234_567 -> "1.2M"
func formatTick(v float64) string {
	if v == 0 {
		return "0"
	}
	abs := math.Abs(v)
	if abs < 1000 {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	var div float64
	var suffix string
	switch {
	case abs < 1_000_000:
		div, suffix = 1000, "k"
	case abs < 1_000_000_000:
		div, suffix = 1_000_000, "M"
	default:
		div, suffix = 1_000_000_000, "B"
	}
	n := v / div
	// One decimal place only when it matters. 1.5k stays "1.5k", 10k stays
	// "10k" not "10.0k", 500k stays "500k".
	if math.Abs(n) >= 10 || n == math.Trunc(n) {
		return strconv.FormatFloat(n, 'f', 0, 64) + suffix
	}
	return fmt.Sprintf("%.1f%s", n, suffix)
}
