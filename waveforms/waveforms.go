package waveforms

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	waveSpeed       = 0.001 // Speed of wave oscillation
	smoothingFactor = 0.02
	amplitudeFactor = 0.01 // Reduce sensitivity of amplitude changes
	centerMoveSpeed = 0.2
)

func SmoothWaveform(samples []float64, screenWidth, screenHeight int, offset float64) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, 0, len(samples))
	centerY := float64(screenHeight) / 2
	//stepX := float64(screenWidth) / float64(len(samples)+1)
	stepX := float64(screenWidth) / 128
	scaleY := float64(screenHeight) / 2

	for i, sample := range samples {
		x := float32(float64(i) * stepX)
		y := float32(centerY + sample*scaleY)
		vertices = append(vertices, ebiten.Vertex{
			DstX: x, DstY: y,
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 0.1,
		})
	}

	return vertices
}

func FerroliquidWaveform(samples []float64, screenWidth, screenHeight int, offset float64, radiusControl float64, smoothingFactor float64, amplitudeFactor float64) []ebiten.Vertex {
	var previousVertices []ebiten.Vertex
	vertices := make([]ebiten.Vertex, 0, len(samples)*2)
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight) / 2
	maxRadius := radiusControl * rand.Float64()

	// Smoothing for radii
	smoothedRadius := make([]float64, len(samples))
	for i := 0; i < len(samples); i++ {
		if i == 0 {
			smoothedRadius[i] = maxRadius * (1 + samples[i]*0.5)
		} else {
			smoothedRadius[i] = smoothingFactor*smoothedRadius[i-1] + (1-smoothingFactor)*maxRadius*(1+samples[i]*0.5)
		}
	}

	// Generate current vertices
	currentVertices := make([]ebiten.Vertex, len(samples))
	for i := 0; i < len(samples); i++ {
		angle := (float64(i) + offset) / float64(len(samples)) * 2 * math.Pi
		radius := smoothedRadius[i] * (1 + amplitudeFactor*math.Sin(offset+float64(i)/10))
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)
		currentVertices[i] = ebiten.Vertex{
			DstX: float32(x), DstY: float32(y),
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}
	}

	// Interpolate between previous and current vertices
	interpolationFactor := 0.1 // Adjust this factor for smoother transitions
	if len(previousVertices) == len(currentVertices) {
		for i := 0; i < len(samples); i++ {
			x := float32((1-interpolationFactor)*float64(previousVertices[i].DstX) + interpolationFactor*float64(currentVertices[i].DstX))
			y := float32((1-interpolationFactor)*float64(previousVertices[i].DstY) + interpolationFactor*float64(currentVertices[i].DstY))
			vertices = append(vertices, ebiten.Vertex{
				DstX: x, DstY: y,
				ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
			})
		}
	} else {
		vertices = currentVertices
	}

	// Store current vertices as the previous for the next frame
	previousVertices = currentVertices
	return vertices
}

func BezierWaveform(samples []float64, screenWidth, screenHeight int, offset float64, radiusControl float64) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, 0, len(samples)*2)
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight) / 2
	radius := radiusControl * rand.Float64()

	for i := 0; i < len(samples)-2; i += 3 {
		angle1 := (float64(i) + offset) / float64(len(samples)) * 2 * math.Pi
		angle2 := (float64(i+1) + offset) / float64(len(samples)) * 2 * math.Pi
		angle3 := (float64(i+2) + offset) / float64(len(samples)) * 2 * math.Pi

		x1 := centerX + radius*(1+samples[i])*math.Cos(angle1)
		y1 := centerY + radius*(1+samples[i])*math.Sin(angle1)
		x2 := centerX + radius*(1+samples[(i+1)%len(samples)])*math.Cos(angle2)
		y2 := centerY + radius*(1+samples[(i+1)%len(samples)])*math.Sin(angle2)
		x3 := centerX + radius*(1+samples[(i+2)%len(samples)])*math.Cos(angle3)
		y3 := centerY + radius*(1+samples[(i+2)%len(samples)])*math.Sin(angle3)

		vertices = append(vertices, ebiten.Vertex{
			DstX: float32(x1), DstY: float32(y1),
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}, ebiten.Vertex{
			DstX: float32(x2), DstY: float32(y2),
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}, ebiten.Vertex{
			DstX: float32(x3), DstY: float32(y3),
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		})
	}

	return vertices
}
