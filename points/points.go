package points

import (
	"math"
	"math/rand"
)

type point struct {
	x, y                 float64
	xVelocity, yVelocity float64
	alpha                float64
	volume               float64
	fadeIn               bool
}

type murmuration struct {
	volumePoints  []point
	maxPoints     int
	width, height float64 // Dimensions of the simulation space
}

func (m *murmuration) updatePoints() {
	neighborRadius := 50.0
	alignmentFactor := 0.05
	cohesionFactor := 0.01
	separationFactor := 0.1
	maxSpeed := 5.0

	for i := range m.volumePoints {
		// Variables for calculating alignment, cohesion, and separation
		alignmentX, alignmentY := 0.0, 0.0
		cohesionX, cohesionY := 0.0, 0.0
		separationX, separationY := 0.0, 0.0
		neighborCount := 0

		for j := range m.volumePoints {
			if i == j {
				continue
			}
			dx := m.volumePoints[j].x - m.volumePoints[i].x
			dy := m.volumePoints[j].y - m.volumePoints[i].y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < neighborRadius {
				neighborCount++
				// Alignment: Match velocity with neighbors
				alignmentX += m.volumePoints[j].xVelocity
				alignmentY += m.volumePoints[j].yVelocity
				// Cohesion: Move toward neighbors' average position
				cohesionX += m.volumePoints[j].x
				cohesionY += m.volumePoints[j].y
				// Separation: Avoid crowding
				if dist > 0 {
					separationX -= dx / dist
					separationY -= dy / dist
				}
			}
		}

		if neighborCount > 0 {
			// Calculate average alignment
			alignmentX /= float64(neighborCount)
			alignmentY /= float64(neighborCount)

			// Calculate average cohesion
			cohesionX = (cohesionX / float64(neighborCount)) - m.volumePoints[i].x
			cohesionY = (cohesionY / float64(neighborCount)) - m.volumePoints[i].y

			// Apply alignment, cohesion, and separation adjustments
			m.volumePoints[i].xVelocity += alignmentX*alignmentFactor + cohesionX*cohesionFactor + separationX*separationFactor
			m.volumePoints[i].yVelocity += alignmentY*alignmentFactor + cohesionY*cohesionFactor + separationY*separationFactor
		}

		// Limit speed to maxSpeed
		speed := math.Sqrt(m.volumePoints[i].xVelocity*m.volumePoints[i].xVelocity + m.volumePoints[i].yVelocity*m.volumePoints[i].yVelocity)
		if speed > maxSpeed {
			m.volumePoints[i].xVelocity = (m.volumePoints[i].xVelocity / speed) * maxSpeed
			m.volumePoints[i].yVelocity = (m.volumePoints[i].yVelocity / speed) * maxSpeed
		}

		// Update position based on velocity
		m.volumePoints[i].x += m.volumePoints[i].xVelocity
		m.volumePoints[i].y += m.volumePoints[i].yVelocity

		// Wrap around screen boundaries
		if m.volumePoints[i].x < 0 {
			m.volumePoints[i].x += m.width
		} else if m.volumePoints[i].x > m.width {
			m.volumePoints[i].x -= m.width
		}
		if m.volumePoints[i].y < 0 {
			m.volumePoints[i].y += m.height
		} else if m.volumePoints[i].y > m.height {
			m.volumePoints[i].y -= m.height
		}
	}
}

func (m *murmuration) initializePoints() {
	for len(m.volumePoints) < m.maxPoints {
		m.volumePoints = append(m.volumePoints, point{
			x:         rand.Float64() * m.width,
			y:         rand.Float64() * m.height,
			xVelocity: rand.Float64()*2 - 1,
			yVelocity: rand.Float64()*2 - 1,
			alpha:     0.0,
			volume:    1.0,
			fadeIn:    true,
		})
	}
}
