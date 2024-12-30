package utils

import "math"

func CalculateDominantFrequency(samples []float64, sampleRate int) float64 {
	n := len(samples)
	fftReal := make([]float64, n)
	fftImag := make([]float64, n)

	for k := 0; k < n; k++ {
		for t := 0; t < n; t++ {
			angle := 2 * math.Pi * float64(k) * float64(t) / float64(n)
			fftReal[k] += samples[t] * math.Cos(angle)
			fftImag[k] -= samples[t] * math.Sin(angle)
		}
	}

	highestFrequency := 0.0
	maxAmplitude := 0.0

	for k := 1; k < n/2; k++ {
		amplitude := math.Sqrt(fftReal[k]*fftReal[k] + fftImag[k]*fftImag[k])
		if amplitude > maxAmplitude {
			maxAmplitude = amplitude
			highestFrequency = float64(k) * float64(sampleRate) / float64(n)
		}
	}

	return highestFrequency
}
