package waveforms

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	waveSpeed       = 0.001 // Speed of wave oscillation
	smoothingFactor = 0.02
	amplitudeFactor = 0.01 // Reduce sensitivity of amplitude changes
	centerMoveSpeed = 0.2
)

func SmoothWaveform(samples []float64, screenWidth, screenHeight int, offset float64) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, 0, len(samples)*2)
	centerY := float64(screenHeight) / 2
	stepX := float64(screenWidth) / float64(len(samples))

	for i, sample := range samples {
		x := float32(float64(i) * stepX)
		y := float32(centerY + math.Sin(float64(i)/float64(len(samples))*2*math.Pi)*sample*centerY/2)
		vertices = append(vertices, ebiten.Vertex{
			DstX: x, DstY: y,
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 0.1,
		})
	}

	return vertices
}

func FerroliquidWaveform(samples []float64, screenWidth, screenHeight int, offset float64) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, 0, len(samples)*2)
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight) / 2
	maxRadius := math.Min(float64(screenWidth), float64(screenHeight)) / 3

	// Smoothing for radii
	smoothedRadius := make([]float64, len(samples))
	for i := 0; i < len(samples); i++ {
		if i == 0 {
			smoothedRadius[i] = maxRadius * (1 + samples[i]*0.5)
		} else {
			smoothedRadius[i] = smoothingFactor*smoothedRadius[i-1] + (1-smoothingFactor)*maxRadius*(1+samples[i]*0.5)
		}
	}

	for i := 0; i < len(samples); i++ {
		angle := (float64(i) + offset) / float64(len(samples)) * 2 * math.Pi
		radius := smoothedRadius[i] * (1 + amplitudeFactor*math.Sin(offset+float64(i)/10))
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)
		vertices = append(vertices, ebiten.Vertex{
			DstX: float32(x), DstY: float32(y),
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		})
	}

	return vertices
}

func FoldingWaveform(samples []float64, screenWidth, screenHeight int, offset float64, attribute float64) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, 0, len(samples)*2)
	centerX := float64(screenWidth)/2 + math.Sin(offset*centerMoveSpeed)*100  // Dynamic horizontal movement
	centerY := float64(screenHeight)/2 + math.Cos(offset*centerMoveSpeed)*100 // Dynamic vertical movement
	maxRadius := attribute                                                    //math.Min(float64(screenWidth), float64(screenHeight)) / 3

	for i := 0; i < len(samples); i++ {
		angle := (float64(i) + offset) / float64(len(samples)) * 4 * math.Pi                             // Wrap around for folding
		radius := maxRadius * (1 + samples[i]*amplitudeFactor) * (1 + 0.2*math.Sin(offset+float64(i)/5)) // More pronounced folds

		// Add folding effect by modulating the radius
		foldEffect := 0.5 + 0.7*math.Sin(float64(i)/10+offset)
		radius *= foldEffect

		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)
		vertices = append(vertices, ebiten.Vertex{
			DstX: float32(x), DstY: float32(y),
			ColorR: float32(0.5 + 0.5*math.Sin(float64(i)/20+offset)), // Color modulation
			ColorG: float32(0.8), ColorB: float32(1), ColorA: 1,
		})
	}

	return vertices
}

func BezierWaveform(samples []float64, screenWidth, screenHeight int, offset float64) []ebiten.Vertex {
	vertices := make([]ebiten.Vertex, 0, len(samples)*2)
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight) / 2
	radius := math.Min(float64(screenWidth), float64(screenHeight)) / 3

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
