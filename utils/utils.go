package utils

import (
	"math"

	"github.com/mjibson/go-dsp/fft"
)

func ComputeFFT(waveform []float64) []float64 {
	complexData := make([]complex128, len(waveform))
	for i, sample := range waveform {
		complexData[i] = complex(sample, 0)
	}
	fftResult := fft.FFT(complexData)
	frequencies := make([]float64, len(fftResult)/2)
	for i := range frequencies {
		frequencies[i] = math.Abs(real(fftResult[i]))
	}
	return frequencies
}

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
