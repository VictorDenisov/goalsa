package main

import (
	"math"
)

type SignalDetector struct {
	signal []float64
	noise  []float64
}

func (sd *SignalDetector) isSignal(sample []float64) bool {
	return distSq(sample, sd.signal) < distSq(sample, sd.noise)
}

func classifySegments(segments [][]float64) (sd *SignalDetector) {

	m := len(segments[0])
	n := len(segments)

	labels := make([]bool, n)

	mx := make([]float64, n)

	for i := 0; i < n; i++ {
		mx[i], _ = segmentMax(segments[i])
	}

	allMin, _ := segmentMin(mx)
	allMax, _ := segmentMax(mx)

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

func expectationMaximizationClassifySegments(segments [][]float64) (sd *SignalDetector) {

	sq2pi := math.Sqrt(2 * math.Pi)
	m := len(segments[0])
	n := len(segments)

	r := make([][]float64, n)
	for i := 0; i < n; i++ {
		r[i] = make([]float64, 2)
	}

	mx := make([]float64, n) // Maximum in every segment

	for i := 0; i < n; i++ {
		mx[i], _ = segmentMax(segments[i])
	}

	_, minId := segmentMin(mx)
	_, maxId := segmentMax(mx)

	const nCentroids = 2
	centroids := make([][]float64, nCentroids)
	for i := 0; i < nCentroids; i++ {
		centroids[i] = make([]float64, m)
	}
	copy(centroids[0], segments[minId])
	copy(centroids[1], segments[maxId])
	sigma := make([]float64, nCentroids)
	for i := 0; i < nCentroids; i++ {
		sigma[i] = 0.5
	}
	pi := make([]float64, nCentroids)
	for i := 0; i < n; i++ {
		sumR := float64(0)
		for k := 0; k < nCentroids; k++ {
			r[i][k] = pi[k] * math.Exp(-dist(centroids[k], segments[i])/sigma[k]/sigma[k]) / math.Pow(sq2pi*sigma[k], float64(m))
			sumR = r[i][k]
		}

		for k := 0; k < nCentroids; k++ {
			r[i][k] /= sumR
		}
	}
	R := make([]float64, nCentroids)
	totalR := float64(0)
	for k := 0; k < nCentroids; k++ {
		var s []float64
		for j := 0; j < n; j++ {
			s = plus(s, mult(r[j][k], segments[j]))
			R[k] += r[j][k]
		}
		centroids[k] = divN(s, R[k])
		sigma[k] = 0
		for j := 0; j < n; j++ {
			sigma[k] += r[j][k] * dist(segments[k], centroids[k])
		}
		sigma[k] /= float64(m) * R[k]
		sigma[k] = math.Sqrt(sigma[k])
		totalR += R[k]
	}
	for k := 0; k < nCentroids; k++ {
		pi[k] = R[k] / totalR
	}
	return nil
}

func plus(a, b []float64) (r []float64) {
	r = make([]float64, len(b))
	for i := 0; i < len(b); i++ {
		r[i] = a[i] + b[i]
	}
	return
}

func divN(a []float64, b float64) (r []float64) {
	r = make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = a[i] / b
	}
	return
}

func mult(a float64, b []float64) (r []float64) {
	r = make([]float64, len(b))
	for i := 0; i < len(b); i++ {
		r[i] = b[i] * a
	}
	return
}

func dist(a, b []float64) (r float64) {
	for i := 0; i < len(a); i++ {
		r += (a[i] - b[i]) * (a[i] - b[i])
	}
	return r
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

func segmentMin(seg []float64) (mx float64, id int) {
	mx = seg[0]
	id = 0
	for i := 1; i < len(seg); i++ {
		if seg[i] < mx {
			mx = seg[i]
			id = i
		}
	}
	return
}

func segmentMax(seg []float64) (mx float64, id int) {
	mx = seg[0]
	id = 0
	for i := 1; i < len(seg); i++ {
		if seg[i] > mx {
			mx = seg[i]
		}
	}
	return mx, id
}
