package visualiser

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/idroz/mezmer/utils"
	"github.com/idroz/mezmer/waveforms"
	"golang.org/x/image/font/basicfont"
)

const (
	chunkSize       = 512   // Number of samples per chunk
	pointSize       = 3     // Size of each point
	radiateSpeed    = 2     // Speed at which points radiate outward
	radiateVariance = 0.1   // Maximum variance in radiate speed
	waveSpeed       = 0.001 // Speed of wave oscillation
	smoothingFactor = 0.02
	amplitudeFactor = 0.01 // Reduce sensitivity of amplitude changes
	centerMoveSpeed = 0.2
)

type point struct {
	x, y      float64
	xVelocity float64
	yVelocity float64
	alpha     float64
	volume    float64
	fadeIn    bool
	red       int
	green     int
	blue      int
}

type colorSceme struct {
	red   int
	green int
	blue  int
}

// AudioVisualizer represents the visualization logic.
type audioVisualizer struct {
	samples         []float64
	currentChunk    []float64
	chunkSamples    int
	volumePoints    []point // Stores the radiating points with volume and alpha
	maxPoints       int     // Maximum number of radiating points
	screenWidth     int
	screenHeight    int
	showText        bool
	volume          float64
	frequency       float64
	spacePressed    bool
	waveForm        string
	pointType       string
	waveOffset      float64
	connectedDevice string
	colorScheme     colorSceme
}

func newAudioVisualizer(chunkSize, screenWidth, screenHeight int) *audioVisualizer {
	return &audioVisualizer{
		samples:         make([]float64, chunkSize),
		currentChunk:    make([]float64, chunkSize),
		chunkSamples:    chunkSize,
		volumePoints:    make([]point, 0),
		maxPoints:       1000, // Initial maximum points
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		showText:        true,
		volume:          0,
		frequency:       0,
		spacePressed:    false,
		waveForm:        "smooth",
		pointType:       "radial",
		waveOffset:      0,
		connectedDevice: "No Device",
		colorScheme:     colorSceme{red: 255, green: 0, blue: 255},
	}
}

// Update reads new audio data into the visualizer and updates the points.
func (v *audioVisualizer) Update() error {
	// Handle toggling text visibility on space key press
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if !v.spacePressed {
			v.showText = !v.showText
			v.spacePressed = true
		}
	} else {
		v.spacePressed = false
	}

	if ebiten.IsKeyPressed(ebiten.KeyDigit0) {
		v.waveForm = ""
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit1) {
		v.waveForm = "smooth"
	}

	if ebiten.IsKeyPressed(ebiten.KeyDigit5) {
		v.pointType = "radial"
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit6) {
		v.pointType = "spiral"
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit7) {
		v.pointType = "slinky"
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit8) {
		v.pointType = "spikes"
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) && ebiten.IsKeyPressed(ebiten.KeyR) {
		v.colorScheme.red = int(math.Min(255, float64(v.colorScheme.red+1)))
	}
	if ebiten.IsKeyPressed(ebiten.KeyR) && (!ebiten.IsKeyPressed(ebiten.KeyShift)) {
		v.colorScheme.red = int(math.Max(0, float64(v.colorScheme.red-1)))
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) && ebiten.IsKeyPressed(ebiten.KeyG) {
		v.colorScheme.green = int(math.Min(255, float64(v.colorScheme.green+1)))
	}
	if ebiten.IsKeyPressed(ebiten.KeyG) && (!ebiten.IsKeyPressed(ebiten.KeyShift)) {
		v.colorScheme.green = int(math.Max(0, float64(v.colorScheme.green-1)))
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) && ebiten.IsKeyPressed(ebiten.KeyB) {
		v.colorScheme.blue = int(math.Min(255, float64(v.colorScheme.blue+1)))
	}
	if ebiten.IsKeyPressed(ebiten.KeyB) && (!ebiten.IsKeyPressed(ebiten.KeyShift)) {
		v.colorScheme.blue = int(math.Max(0, float64(v.colorScheme.blue-1)))
	}

	// Copy the latest audio data into the visualizer's current chunk.
	copy(v.samples, v.currentChunk)

	// Compute the RMS (volume) of the current chunk
	sumSquares := 0.0
	for _, sample := range v.currentChunk {
		sumSquares += sample * sample
	}
	volume := math.Sqrt(sumSquares/float64(len(v.currentChunk))) * 15

	// Normalize volume to a range for better control
	normalizedVolume := volume
	v.volume = normalizedVolume

	if v.volume > normalizedVolume {
		v.volume -= (normalizedVolume - v.volume) * 0.5
	}

	// Adjust maxPoints based on normalized volume
	v.maxPoints = int(normalizedVolume * 500) // Scale normalized volume to a reasonable number of points
	if v.maxPoints > 1000 {
		v.maxPoints = 1000
	} else if v.maxPoints < 3 {
		v.maxPoints = 0
	}

	randX := float64(v.screenWidth / 2)
	randY := float64(v.screenHeight / 2)

	if v.volume >= 4 {
		randX = rand.Float64() * float64(v.screenWidth)
		randY = rand.Float64() * float64(v.screenHeight)
	}

	if v.pointType == "radial" {
		// Add new points radiating from the center of the screen based on the volume
		for len(v.volumePoints) < v.maxPoints {
			angle := rand.Float64() * 2 * math.Pi // Random angle
			speed := normalizedVolume + rand.Float64()*radiateVariance
			v.volumePoints = append(v.volumePoints, point{
				x:         randX,
				y:         randY,
				xVelocity: math.Cos(angle) * speed,
				yVelocity: math.Sin(angle) * speed,
				alpha:     0.0, // Start with alpha 0 for fade-in effect
				volume:    normalizedVolume,
				fadeIn:    true,
			})
		}
	} else if v.pointType == "spiral" {
		for len(v.volumePoints) < v.maxPoints {
			// Spiral parameters
			angle := float64(len(v.volumePoints)) * 2.2 // Increment angle for spiral
			radius := 0.1 * angle                       // Archimedean spiral: radius increases linearly with angle

			// Compute initial position and velocity for the Archimedean spiral
			speed := normalizedVolume*5 + rand.Float64()*radiateVariance
			xVelocity := math.Cos(angle) * speed
			yVelocity := math.Sin(angle) * speed

			// Generate the point with initial properties
			v.volumePoints = append(v.volumePoints, point{
				x:         randX + math.Cos(angle)*radius, // Start at spiral position
				y:         randY + math.Sin(angle)*radius, // Start at spiral position
				xVelocity: xVelocity,
				yVelocity: yVelocity,
				alpha:     0.0, // Start with alpha 0 for fade-in effect
				volume:    normalizedVolume,
				fadeIn:    true,
			})
		}
	} else if v.pointType == "slinky" {
		// Add new points radiating in a spiral pattern based on the volume
		for len(v.volumePoints) < v.maxPoints {
			angle := rand.Float64() * 2 * math.Pi   // Random initial angle
			spiralRadius := normalizedVolume * 10.0 // Spiral radius influenced by volume
			angleIncrement := 0.1                   // Controls the spacing of the spiral

			// Generate wave points
			for i := 0; i < v.maxPoints; i++ {
				angle += angleIncrement

				//xOffset := ((rand.Float64() * 2) - 1) * (float64(i) * 10.0) // Linear spacing in x
				xOffset := float64(i) * angleIncrement
				yOffset := math.Sin(float64(i)*angleIncrement) * spiralRadius

				v.volumePoints = append(v.volumePoints, point{
					x:         randX + xOffset,
					y:         randY + yOffset,
					xVelocity: math.Cos(angle) * (normalizedVolume + rand.Float64()*radiateVariance),
					yVelocity: math.Sin(angle) * (normalizedVolume + rand.Float64()*radiateVariance),
					alpha:     0.0, // Start with alpha 0 for fade-in effect
					volume:    normalizedVolume,
					fadeIn:    true,
				})
			}
		}
	} else if v.pointType == "spikes" {
		for len(v.volumePoints) < v.maxPoints {

			branchCount := int(math.Max(1, normalizedVolume*5))

			// Randomly select one of the branches
			branchIndex := rand.Intn(branchCount)
			baseAngle := (2 * math.Pi / float64(branchCount)) * float64(branchIndex)

			// Add random deviation from the branch angle
			angleDeviation := (rand.Float64() - 0.5) * math.Pi / 12 // Small deviation for natural look
			angle := baseAngle + angleDeviation

			// Generate random distance from the center
			distance := rand.Float64() * normalizedVolume * 10.0

			// Calculate point coordinates
			randX := randX + math.Cos(angle)*distance
			randY := randY + math.Sin(angle)*distance

			// Speed calculation (same as before)
			speed := normalizedVolume + rand.Float64()*radiateVariance

			// Add the point to the list
			v.volumePoints = append(v.volumePoints, point{
				x:         randX,
				y:         randY,
				xVelocity: math.Cos(angle) * speed,
				yVelocity: math.Sin(angle) * speed,
				alpha:     0.0, // Start with alpha 0 for fade-in effect
				volume:    normalizedVolume,
				fadeIn:    true,
			})
		}
	}

	// Update existing points (radiate, fade in, fade out, and remove if off screen or alpha <= 0)
	for i := 0; i < len(v.volumePoints); i++ {
		v.volumePoints[i].x += v.volumePoints[i].xVelocity // Move in x direction
		v.volumePoints[i].y += v.volumePoints[i].yVelocity // Move in y direction

		if v.volumePoints[i].fadeIn {
			v.volumePoints[i].alpha += 0.05 // Gradually fade in
			if v.volumePoints[i].alpha >= 1.0 {
				v.volumePoints[i].alpha = 1.0
				v.volumePoints[i].fadeIn = false // Switch to fade-out mode
			}
		}

		// Remove the point if it is off the screen or fully transparent
		if v.volumePoints[i].x < 0 || v.volumePoints[i].x > float64(v.screenWidth) ||
			v.volumePoints[i].y < 0 || v.volumePoints[i].y > float64(v.screenHeight) ||
			v.volumePoints[i].alpha <= 0 {
			v.volumePoints = append(v.volumePoints[:i], v.volumePoints[i+1:]...)
			i--
		}
	}

	return nil
}

// Draw renders both visualizations: waveform and radiating points.
func (v *audioVisualizer) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black) // Clear the screen

	// Calculate dominant frequency
	dominantFrequency := utils.CalculateDominantFrequency(v.samples, 44100)
	v.frequency = dominantFrequency
	clr := color.RGBA{R: uint8(math.Min(0, float64(v.colorScheme.red)-255*dominantFrequency/10)),
		G: uint8(v.colorScheme.green),
		B: uint8(math.Min(float64(v.colorScheme.blue), float64(v.colorScheme.blue)*dominantFrequency/10)), A: uint8(255 * v.volume)}

	if v.waveForm == "smooth" {
		vertices := waveforms.SmoothWaveform(v.samples, v.screenWidth, v.screenHeight, v.waveOffset)
		if len(vertices) > 1 {
			for i := 0; i < len(vertices)-1; i++ {
				x1, y1 := vertices[i].DstX, vertices[i].DstY
				x2, y2 := vertices[i+1].DstX, vertices[i+1].DstY
				vector.StrokeLine(screen, x1, y1, x2, y2, 3, clr, false)
			}
		}
	} else if v.waveForm == "ferroliquid" {

		// Draw Ferroliquid waveform visualizer
		vertices := waveforms.FerroliquidWaveform(v.samples, v.screenWidth, v.screenHeight, v.waveOffset, v.volume*100, 0.5, 0.005)
		if len(vertices) > 1 {
			for i := 0; i < len(vertices)-1; i++ {
				x1, y1 := vertices[i].DstX, vertices[i].DstY
				x2, y2 := vertices[i+1].DstX, vertices[i+1].DstY
				vector.StrokeLine(screen, x1, y1, x2, y2, 2, clr, false)
			}
		}
		// Connect last and first points for a closed loop
		if len(vertices) > 2 {
			x1, y1 := vertices[len(vertices)-1].DstX, vertices[len(vertices)-1].DstY
			x2, y2 := vertices[0].DstX, vertices[0].DstY
			vector.StrokeLine(screen, x1, y1, x2, y2, 2, clr, false)
		}
	} else if v.waveForm == "bezier" {
		vertices := waveforms.BezierWaveform(v.samples, v.screenWidth, v.screenHeight, v.waveOffset, v.volume*100)
		if len(vertices) > 1 {
			for i := 0; i < len(vertices)-2; i += 3 {
				x1, y1 := vertices[i].DstX, vertices[i].DstY
				x2, y2 := vertices[i+1].DstX, vertices[i+1].DstY
				x3, y3 := vertices[i+2].DstX, vertices[i+2].DstY
				vector.StrokeLine(screen, x1, y1, x2, y2, 2, clr, false)
				vector.StrokeLine(screen, x2, y2, x3, y3, 2, clr, false)
			}
		}
	} else {
		fmt.Println("No Waveform")
	}

	// Draw radiating points visualizer
	for _, p := range v.volumePoints {
		clr := color.RGBA{
			R: uint8(float64(v.colorScheme.red) * p.volume),
			G: uint8(v.colorScheme.green),
			B: uint8(float64(v.colorScheme.blue) * (1 - p.volume)),
			A: uint8(255 * p.volume * p.alpha),
		}
		for dx := -pointSize / 2; dx <= pointSize/2; dx++ {
			for dy := -pointSize / 2; dy <= pointSize/2; dy++ {
				screen.Set(int(p.x)+dx, int(p.y)+dy, clr)
			}
		}
	}

	// Draw text overlay
	if v.showText {
		textFace := basicfont.Face7x13
		text.Draw(screen, fmt.Sprintf("Connected Device: %s", v.connectedDevice), textFace, 10, 20, color.RGBA{R: 128, G: 128, B: 128, A: 10})
		text.Draw(screen, fmt.Sprintf("Volume: %.2f", float64(v.maxPoints)), textFace, 10, 50, color.RGBA{R: 128, G: 128, B: 128, A: 10})
		text.Draw(screen, fmt.Sprintf("Frequency: %.2f", float64(v.frequency)), textFace, 10, 70, color.RGBA{R: 128, G: 128, B: 128, A: 10})

		text.Draw(screen, fmt.Sprintf("R: %d", v.colorScheme.red), textFace, 10, 100, color.RGBA{R: 128, G: 128, B: 128, A: 10})
		text.Draw(screen, fmt.Sprintf("G: %d", v.colorScheme.green), textFace, 10, 120, color.RGBA{R: 128, G: 128, B: 128, A: 10})
		text.Draw(screen, fmt.Sprintf("B: %d", v.colorScheme.blue), textFace, 10, 140, color.RGBA{R: 128, G: 128, B: 128, A: 10})

		text.Draw(screen, fmt.Sprint("Waveforms: 0 (None),   1 (Smooth)"), textFace, 10, v.screenHeight-30, color.RGBA{R: 128, G: 128, B: 128, A: 10})
		text.Draw(screen, fmt.Sprint("Patterns:  5 (Radial), 6 (Spiral), 7 (Slinky) 8 (Spikes)"), textFace, 10, v.screenHeight-10, color.RGBA{R: 128, G: 128, B: 128, A: 10})
	}
}

func (v *audioVisualizer) Layout(outsideWidth, outsideHeight int) (int, int) {
	v.screenWidth = outsideWidth
	v.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func RunMezmer() error {
	runtime.LockOSThread()

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctxAudio, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("Log: %s\n", message)
	})
	if err != nil {
		return err
	}
	defer ctxAudio.Uninit()
	defer ctxAudio.Free()

	audioInput := make([]float64, chunkSize)
	var (
		deviceMutex     sync.Mutex
		deviceAvailable bool
		targetDevice    *malgo.DeviceInfo
		activeDevice    *malgo.Device // Track the active audio device
	)

	// Initialize the visualizer
	initialWidth, initialHeight := 1280, 720
	visualizer := newAudioVisualizer(chunkSize, initialWidth, initialHeight)

	errChan := make(chan error)

	// Goroutine to check for the OP-XY/OP-Z device
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Device discovery goroutine shutting down...")
				return
			default:
				devices, err := ctxAudio.Devices(malgo.Capture)
				if err != nil {
					log.Printf("Error listing devices: %v", err)
					time.Sleep(1 * time.Second)
					continue
				}

				deviceMutex.Lock()
				foundDevice := false
				for _, device := range devices {
					if device.Name() == "OP-XY" || device.Name() == "OP-Z" {
						if targetDevice == nil || targetDevice.ID.Pointer() != device.ID.Pointer() {
							targetDevice = &device
							deviceAvailable = true
							fmt.Printf("Found device: %s\n", device.Name())
							visualizer.connectedDevice = device.Name()
						}
						foundDevice = true
						break
					}
				}

				if !foundDevice {
					if deviceAvailable {
						fmt.Println("Device disconnected.")
						deviceAvailable = false
						targetDevice = nil
						visualizer.connectedDevice = "No Device"

						if activeDevice != nil {
							fmt.Println("Stopping active audio device...")
							activeDevice.Stop()
							activeDevice.Uninit()
							activeDevice = nil
						}
					}
				}
				deviceMutex.Unlock()

				if !deviceAvailable {
					fmt.Println("Waiting for the OP-XY/Z device...")
				}
				time.Sleep(1 * time.Second)
			}
		}
	}(ctx)

	// Goroutine to handle audio capture
	go func(ctx context.Context) {
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Audio capture goroutine shutting down...")
				if activeDevice != nil {
					activeDevice.Stop()
					activeDevice.Uninit()
				}
				return
			default:
				deviceMutex.Lock()
				if deviceAvailable && targetDevice != nil && activeDevice == nil {
					deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
					deviceConfig.Capture.Format = malgo.FormatS16
					deviceConfig.Capture.Channels = 1
					deviceConfig.SampleRate = 44100
					deviceConfig.Capture.DeviceID = targetDevice.ID.Pointer()

					deviceCallbacks := malgo.DeviceCallbacks{
						Data: func(_, inputSamples []byte, frameCount uint32) {
							for i := 0; i < int(frameCount) && i < len(audioInput); i++ {
								sample := int16(inputSamples[2*i]) | int16(inputSamples[2*i+1])<<8
								audioInput[i] = float64(sample) / 32768.0 // Convert to float64
							}
						},
					}

					device, err := malgo.InitDevice(ctxAudio.Context, deviceConfig, deviceCallbacks)
					if err != nil {
						log.Printf("Failed to initialize device: %v", err)
						deviceMutex.Unlock()
						continue
					}

					if err := device.Start(); err != nil {
						log.Printf("Failed to start device: %v", err)
						device.Uninit()
						deviceMutex.Unlock()
						continue
					}

					fmt.Println("Audio device started.")
					activeDevice = device
				}

				if !deviceAvailable && activeDevice != nil {
					fmt.Println("Stopping active audio device due to disconnection...")
					activeDevice.Stop()
					activeDevice.Uninit()
					activeDevice = nil
				}
				deviceMutex.Unlock()

				if activeDevice != nil {
					copy(visualizer.currentChunk, audioInput)
				}
				time.Sleep(100 * time.Millisecond) // Wait for device to be available
			}
		}
	}(ctx)

	// Run the Ebiten visualizer
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(initialWidth, initialHeight)
	ebiten.SetWindowTitle("Mezmer")
	if err := ebiten.RunGame(visualizer); err != nil {
		cancel() // Cancel context to stop goroutines
		return err
	}

	return nil
}
