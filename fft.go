package main

import (
	"math"
	"math/cmplx"
)

type FFTOpts struct {
	step    float64
	lowf    float64
	maxAmp  float64
	bufSize int
}

type FFTOptsFunc func(*FFTOpts)

func WithBufSize(n int) FFTOptsFunc {
	return func(opts *FFTOpts) {
		opts.bufSize = n
	}
}

type FFTCalculator struct {
	FFTOpts
}

func NewFFTCalculator(opts ...FFTOptsFunc) *FFTCalculator {
	config := FFTOpts{
		step:    1.08,
		lowf:    1.0,
		maxAmp:  1.0,
		bufSize: 10,
	}
	for _, fn := range opts {
		fn(&config)
	}
	return &FFTCalculator{
		FFTOpts: config,
	}
}

func (c *FFTCalculator) fft(signal []complex128) []complex128 {
	n := len(signal)
	if n == 1 {
		return signal
	}

	odd := make([]complex128, n/2)
	even := make([]complex128, n/2)
	for i := range n / 2 {
		odd[i] = signal[2*i]
		even[i] = signal[2*i+1]
	}

	oddFft := c.fft(odd)
	evenFft := c.fft(even)
	out := make([]complex128, n)

	for i := range n / 2 {
		component := -2.0 * math.Pi * float64(i) / float64(n)
		oddFactor := complex(math.Cos(component), math.Sin(component)) * oddFft[i]

		out[i] = evenFft[i] + oddFactor
		out[i+n/2] = evenFft[i] - oddFactor
	}

	return out
}

func (c *FFTCalculator) Calculate(samples []float64) []float64 {
	frequencies := c.fft(toComplexArray(samples))
	amplitudes := make([]float64, len(frequencies))
	for i, freq := range frequencies {
		amplitudes[i] = math.Log(cmplx.Abs(freq))
	}

	var outLog []float64
	// TODO: simplify and improve variable names
	for f := c.lowf; f < float64(c.bufSize)/2; f = math.Ceil(f * c.step) {
		f1 := math.Ceil(f * c.step)
		a := 0.0
		for q := int(f); q < c.bufSize/2 && q < int(f1); q++ {
			b := amplitudes[q]
			if b > a {
				a = b
			}
		}
		if c.maxAmp < a {
			c.maxAmp = a
		}
		outLog = append(outLog, a)
	}

	for i, v := range outLog {
		outLog[i] = v / c.maxAmp
	}

	return outLog
}
