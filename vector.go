package main

func sumV(dst, src []float64) {
	if len(src) != len(dst) {
		panic("Adding vectors of different size")
	}
	for i := 0; i < len(src); i++ {
		dst[i] += src[i]
	}
}

func divVS(dst []float64, s float64) {
	for i := 0; i < len(dst); i++ {
		dst[i] /= s
	}
}

func divNVS(dst []float64, s float64) (res []float64) {
	res = make([]float64, len(dst))
	for i := 0; i < len(dst); i++ {
		res[i] = dst[i] / s
	}
	return res
}

func maxV(v []float64) float64 {
}
