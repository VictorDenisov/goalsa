package main

import (
	"fmt"
	"math"
)

// -------------- KMeans Signal Detector ---------------

type KMeansSignalDetector struct {
	signal []float64
	noise  []float64
}

func (sd *KMeansSignalDetector) isSignal(sample []float64) bool {
	return distSq(sample, sd.signal) < distSq(sample, sd.noise)
}

func classifySegments(segments [][]float64) (sd *KMeansSignalDetector) {

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

	sd = &KMeansSignalDetector{signalMean, noiseMean}
	return sd
}

// ---------------- EM Signal Detector ----------------

type EMSignalDetector struct {
	r         [][]float64
	centroids [][]float64
	sigma     []float64
	pi        []float64
}

func (sd *EMSignalDetector) isSignal(sample []float64) bool {
	return distSq(sample, sd.centroids[0]) < distSq(sample, sd.centroids[1])
}

func expectationMaximizationClassifySegments(segments [][]float64) (sd *EMSignalDetector) {

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
	drawChart("signalMean0.html", centroids[0])
	drawChart("noiseMean0.html", centroids[1])
	sigma := make([]float64, nCentroids)
	for i := 0; i < nCentroids; i++ {
		sigma[i] = 100
	}
	pi := make([]float64, nCentroids)
	for i := 0; i < nCentroids; i++ {
		pi[i] = 1
	}
	for i := 0; i < n; i++ {
		//sumR := float64(0)
		//fmt.Printf("dist: ")
		for k := 0; k < nCentroids; k++ {
			//v := pi[k] * math.Exp(-dist(centroids[k], segments[i])/sigma[k]/sigma[k])
			//v = float64(m) //sq2pi * sigma[k]
			//fmt.Printf("%f ", v)
			r[i][k] = pi[k] * math.Exp(-dist(centroids[k], segments[i])/sigma[k]/sigma[k]) / math.Pow(sq2pi*sigma[k], float64(m))
			//sumR = r[i][k]
		}
		//fmt.Printf("\n")

		for k := 0; k < nCentroids; k++ {
			//r[i][k] /= sumR
			//fmt.Printf("%f ", r[i][k])
		}
		//fmt.Printf("\n")
	}
	R := make([]float64, nCentroids)
	totalR := float64(0)
	for k := 0; k < nCentroids; k++ {
		s := make([]float64, m)
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
	return &EMSignalDetector{r, centroids, sigma, pi}
}

// ------------- Single Frequency Detector -------------

type SingleFrequencyDetector struct {
	signal float64
	noise  float64
}

func (sd *SingleFrequencyDetector) isSignal(sample float64) bool {
	return math.Abs(sample-sd.signal) < math.Abs(sample-sd.noise)
}

func classifyFromSingleFrequency(signals []float64) (sd *SingleFrequencyDetector) {
	n := len(signals)

	labels := make([]bool, n)

	allMin, _ := segmentMin(signals)
	allMax, _ := segmentMax(signals)

	middle := (allMin + allMax) / 2

	for i := 0; i < n; i++ {
		labels[i] = signals[i] > middle
	}

	var signalMean float64
	var noiseMean float64
	for {
		signalMean = 0
		noiseMean = 0

		signalNum := 0
		noiseNum := 0
		for i := 0; i < n; i++ {
			if labels[i] {
				signalMean += signals[i]
				signalNum++
			} else {
				noiseMean += signals[i]
				noiseNum++
			}
		}

		signalMean /= float64(signalNum)
		noiseMean /= float64(noiseNum)

		labelChanged := false
		for i := 0; i < n; i++ {
			newLabel := math.Abs(signalMean-signals[i]) < math.Abs(noiseMean-signals[i])
			labelChanged = labelChanged || (labels[i] != newLabel)
			labels[i] = newLabel
		}
		if !labelChanged {
			break
		}
	}

	sd = &SingleFrequencyDetector{signalMean, noiseMean}
	return sd
}

// ----------- EM Single Frequency Detector ------------

type EMSingleFrequencyDetector struct {
	m     []float64
	sigma []float64
	pi    []float64
}

func (sd *EMSingleFrequencyDetector) isSignal(sample float64) bool {
	rk := assignPoint(sample, sd.m, sd.sigma, sd.pi)
	return rk[0] > rk[1]
}

func classifyEMFromSingleFrequency(signals []float64) (sd *EMSingleFrequencyDetector) {
	n := len(signals)

	r := make([][]float64, 2)

	m := make([]float64, 2)
	oldM := make([]float64, 2)

	sigma := make([]float64, 2)
	oldSigma := make([]float64, 2)

	pi := make([]float64, 2)
	oldPi := make([]float64, 2)

	for i := 0; i < 2; i++ {
		r[i] = make([]float64, n)
	}

	allMin, _ := segmentMin(signals)
	allMax, _ := segmentMax(signals)

	middle := (allMin + allMax) / 2

	m[0] = allMax
	m[1] = allMin

	aboveMiddle := 0
	sigma[0] = 0
	for i := 0; i < n; i++ {
		if signals[i] > middle {
			sigma[0] += (m[0] - signals[i]) * (m[0] - signals[i])
			aboveMiddle++
		}
	}
	sigma[0] /= float64(aboveMiddle)

	belowMiddle := 0
	sigma[1] = 0
	for i := 0; i < n; i++ {
		if signals[i] < middle {
			sigma[1] += (m[1] - signals[i]) * (m[1] - signals[i])
			belowMiddle++
		}
	}
	sigma[1] /= float64(belowMiddle)

	pi[0] = 0.5
	pi[1] = 0.5

	/*
		for i := 0; i < n; i++ {
			if signals[i] > middle {
				r[0][i] = 1
			} else {
				r[1][i] = 1
			}
		}
	*/
	//updateStep(n, r, signals, m, sigma, pi)
	assignmentStep(n, r, signals, m, sigma, pi)
	fmt.Printf("Initial assignment\n")
	fmt.Printf("r: %v\n", r)
	fmt.Printf("m: %v\n", m)
	fmt.Printf("sigma: %v\n", sigma)
	fmt.Printf("pi: %v\n", pi)
	stepCount := 0
	for {
		copy(oldM, m)
		copy(oldSigma, sigma)
		copy(oldPi, pi)
		assignmentStep(n, r, signals, m, sigma, pi)

		/*
			fmt.Printf("step count: %v\n", stepCount)
			fmt.Printf("r: %v\n", r)
			fmt.Printf("m: %v\n", m)
			fmt.Printf("sigma: %v\n", sigma)
			fmt.Printf("pi: %v\n", pi)
		*/

		updateStep(n, r, signals, m, sigma, pi)

		/*
			//changed := false
			for k := 0; k < 2; k++ {
				if math.Abs(m[k]-oldM[k]) > 0.000001 {
					changed = true
				}
				if math.Abs(sigma[k]-oldSigma[k]) > 0.000001 {
					changed = true
				}
				if math.Abs(pi[k]-oldPi[k]) > 0.000001 {
					changed = true
				}
			}
		*/
		if stepCount > 100 {
			break
		}
		stepCount++
	}
	return &EMSingleFrequencyDetector{m, sigma, pi}

}

func assignmentStep(n int, r [][]float64, x []float64, m []float64, sigma []float64, pi []float64) {
	for i := 0; i < n; i++ {
		rk := assignPoint(x[i], m, sigma, pi)
		for k := 0; k < 2; k++ {
			r[k][i] = rk[k]
		}
	}
}

func assignPoint(x float64, m []float64, sigma []float64, pi []float64) (r []float64) {
	var rkt float64
	r = make([]float64, 2)
	//fmt.Printf("Assign point: %v, %v, %v, %v\n", x, m, sigma, pi)
	for k := 0; k < 2; k++ {
		//fmt.Printf("%v\n", math.Abs(m[k]-x))
		//fmt.Printf("%v\n", math.Abs(m[k]-x)/sigma[k])
		//fmt.Printf("%v\n", math.Exp(-math.Abs(m[k]-x)/sigma[k]))
		rkt += pi[k] * math.Exp(-(m[k]-x)*(m[k]-x)/sigma[k]) / math.Sqrt(2*math.Pi*sigma[k])
	}
	//fmt.Printf("rkt: %v\n", rkt)
	for k := 0; k < 2; k++ {
		r[k] = pi[k] * math.Exp(-(m[k]-x)*(m[k]-x)/sigma[k]) / math.Sqrt(2*math.Pi*sigma[k]) / rkt
		//fmt.Printf("k: %v, rk: %v\n", k, r[k])
	}
	return r
}

func updateStep(n int, r [][]float64, x []float64, m []float64, sigma []float64, pi []float64) {
	R := make([]float64, 2)
	var Rsum float64
	for k := 0; k < 2; k++ {
		for i := 0; i < n; i++ {
			R[k] += r[k][i]
		}
		Rsum += R[k]
	}
	for k := 0; k < 2; k++ {
		m[k] = 0
		for i := 0; i < n; i++ {
			m[k] += r[k][i] * x[i]
		}
		m[k] /= R[k]

		sigma[k] = 0
		for i := 0; i < n; i++ {
			sigma[k] += r[k][i] * (x[i] - m[k]) * (x[i] - m[k])
		}
		sigma[k] /= R[k]

		pi[k] = R[k] / Rsum
	}

}

// ----------------------- Misc Functions -----------------------

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
	return math.Sqrt(r)
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
