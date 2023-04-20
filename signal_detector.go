package main

type SignalDetector struct {
	signal []float64
	noise  []float64
}

func classifySegments(segments [][]float64) (sd *SignalDetector) {

	m := len(segments[0])
	n := len(segments)

	labels := make([]bool, n)

	mx := make([]float64, n)

	for i := 0; i < n; i++ {
		mx[i] = segmentMax(segments[i])
	}

	allMin := segmentMin(mx)
	allMax := segmentMax(mx)

	middle := (allMin + allMax) / 2

	for i := 0; i < n; i++ {
		labels[i] = mx[i] > middle
	}

	signalMean := make([]float64, m)
	noiseMean := make([]float64, m)
	for {
		signalMean = make([]float64, m)
		noiseMean = make([]float64, m)

		signalNum := 0
		noiseNum := 0
		for i := 0; i < n; i++ {
			if labels[i] {
				add(signalMean, segments[i])
				signalNum++
			} else {
				add(noiseMean, segments[i])
				noiseNum++
			}
		}

		div(signalMean, float64(signalNum))
		div(noiseMean, float64(noiseNum))

		labelChanged := false
		for i := 0; i < n; i++ {
			newLabel := distSq(signalMean, segments[i]) < distSq(noiseMean, segments[i])
			labelChanged = labelChanged || (labels[i] != newLabel)
			labels[i] = newLabel
		}
		if !labelChanged {
			break
		}
	}

	sd = &SignalDetector{signalMean, noiseMean}
	return sd
}

func distSq(x, y []float64) float64 {
	res := float64(0)
	for i := 0; i < len(x); i++ {
		res += (y[i] - x[i]) * (y[i] - x[i])
	}
	return res
}

func add(x, y []float64) {
	n := len(x)
	for i := 0; i < n; i++ {
		x[i] += y[i]
	}
}

func div(x []float64, v float64) {
	n := len(x)
	for i := 0; i < n; i++ {
		x[i] /= v
	}
}

func segmentMin(seg []float64) float64 {
	mx := seg[0]
	for i := 1; i < len(seg); i++ {
		if seg[i] < mx {
			mx = seg[i]
		}
	}
	return mx
}

func segmentMax(seg []float64) float64 {
	mx := seg[0]
	for i := 1; i < len(seg); i++ {
		if seg[i] > mx {
			mx = seg[i]
		}
	}
	return mx
}
