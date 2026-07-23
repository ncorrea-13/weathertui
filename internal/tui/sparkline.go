package tui

var sparkBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// sparkline renders a compact block-chart trend line for a slice of values.
func sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	out := make([]rune, len(values))
	span := max - min
	for i, v := range values {
		if span == 0 {
			out[i] = sparkBlocks[0]
			continue
		}
		idx := int((v - min) / span * float64(len(sparkBlocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkBlocks) {
			idx = len(sparkBlocks) - 1
		}
		out[i] = sparkBlocks[idx]
	}
	return string(out)
}
