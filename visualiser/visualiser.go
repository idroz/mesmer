package visualiser

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"runtime"

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
	waveOffset      float64
	connectedDevice string
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
		showText:        false,
		volume:          0,
		frequency:       0,
		spacePressed:    false,
		waveForm:        "smooth",
		waveOffset:      0,
		connectedDevice: "",
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

	// Increment wave offset to control gradual oscillation
	v.waveOffset += waveSpeed

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

	// Add new points radiating from the center of the screen based on the volume
	for len(v.volumePoints) < v.maxPoints {
		angle := rand.Float64() * 2 * math.Pi // Random angle
		speed := normalizedVolume + rand.Float64()*radiateVariance
		v.volumePoints = append(v.volumePoints, point{
			x:         float64(v.screenWidth / 2),
			y:         float64(v.screenHeight / 2),
			xVelocity: math.Cos(angle) * speed,
			yVelocity: math.Sin(angle) * speed,
			alpha:     0.0, // Start with alpha 0 for fade-in effect
			volume:    normalizedVolume,
			fadeIn:    true,
		})
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
	clr := color.RGBA{R: uint8(math.Min(0, 255-255*dominantFrequency/10)), G: 2, B: uint8(math.Min(255, 255*dominantFrequency/10)), A: uint8(255 * v.volume)}

	if v.waveForm == "folding" {
		vertices := waveforms.FoldingWaveform(v.samples, v.screenWidth, v.screenHeight, v.waveOffset, v.volume)
		if len(vertices) > 1 {
			for i := 0; i < len(vertices)-1; i++ {
				x1, y1 := vertices[i].DstX, vertices[i].DstY
				x2, y2 := vertices[i+1].DstX, vertices[i+1].DstY
				vector.StrokeLine(screen, x1, y1, x2, y2, 2, color.RGBA{R: 120, G: 200, B: 255, A: 255}, false)
			}
		}
		// Connect the last point to the first to close the loop
		if len(vertices) > 2 {
			x1, y1 := vertices[len(vertices)-1].DstX, vertices[len(vertices)-1].DstY
			x2, y2 := vertices[0].DstX, vertices[0].DstY
			vector.StrokeLine(screen, x1, y1, x2, y2, 2, color.RGBA{R: 120, G: 200, B: 255, A: 255}, false)
		}
	} else if v.waveForm == "ferroliquid" {

		// Draw Ferroliquid waveform visualizer
		vertices := waveforms.FerroliquidWaveform(v.samples, v.screenWidth, v.screenHeight, v.waveOffset)
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
		vertices := waveforms.BezierWaveform(v.samples, v.screenWidth, v.screenHeight, v.waveOffset)
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
		vertices := waveforms.SmoothWaveform(v.samples, v.screenWidth, v.screenHeight, v.waveOffset)
		if len(vertices) > 1 {
			for i := 0; i < len(vertices)-1; i++ {
				x1, y1 := vertices[i].DstX, vertices[i].DstY
				x2, y2 := vertices[i+1].DstX, vertices[i+1].DstY
				vector.StrokeLine(screen, x1, y1, x2, y2, 3, clr, false)
			}
		}
	}

	// Draw radiating points visualizer
	for _, p := range v.volumePoints {

		clr := color.RGBA{
			R: uint8(255 * p.volume),
			G: 0,
			B: uint8(255 * (1 - p.volume)),
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
		text.Draw(screen, fmt.Sprintf("Connected Device: %s", v.connectedDevice), textFace, 10, 20, color.RGBA{
			R: 128,
			G: 128,
			B: 128,
			A: 10})
		text.Draw(screen, fmt.Sprintf("Volume: %.2f", float64(v.maxPoints)), textFace, 10, 40, color.RGBA{
			R: 128,
			G: 128,
			B: 128,
			A: 10})
		text.Draw(screen, fmt.Sprintf("Frequency: %.2f", float64(v.frequency)), textFace, 10, 60, color.RGBA{
			R: 128,
			G: 128,
			B: 128,
			A: 10})
	}
}

func (v *audioVisualizer) Layout(outsideWidth, outsideHeight int) (int, int) {
	v.screenWidth = outsideWidth
	v.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func RunMezmer() error {
	runtime.LockOSThread()

	// Initialize MiniAudio context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("MiniAudio Log: %s\n", message)
	})
	if err != nil {
		return fmt.Errorf("Failed to initialize MiniAudio: %s", err.Error())
	}
	defer ctx.Uninit()
	defer ctx.Free()

	// List available capture devices
	captureDevices, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return fmt.Errorf("Failed to list capture devices: %s", err.Error())
	}

	fmt.Println("Capture Devices:")
	var targetDevice *malgo.DeviceInfo
	for _, device := range captureDevices {
		fmt.Printf("  %s\n", device.Name())
		if device.Name() == "OP-XY" {
			targetDevice = &device
			break
		}
	}

	if targetDevice == nil {
		return fmt.Errorf("Capture device 'OP-XY' not found.")
	}

	// Prepare buffer for audio input
	audioInput := make([]float64, chunkSize)

	// Configure audio capture
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = 44100
	deviceConfig.Alsa.NoMMap = 1
	deviceConfig.Capture.DeviceID = targetDevice.ID.Pointer()

	// Configure callbacks for capturing audio
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: func(_, inputSamples []byte, frameCount uint32) {
			for i := 0; i < int(frameCount) && i < len(audioInput); i++ {
				sample := int16(inputSamples[2*i]) | int16(inputSamples[2*i+1])<<8
				audioInput[i] = float64(sample) / 32768.0 // Convert to float64
			}
		},
	}

	// Initialize audio device
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return fmt.Errorf("Failed to initialize audio device: %v", err)
	}
	defer device.Uninit()

	// Start audio capture
	if err := device.Start(); err != nil {
		return fmt.Errorf("Failed to start audio device: %v", err)
	}

	// Initialize the visualizer
	initialWidth, initialHeight := 800, 400
	visualizer := newAudioVisualizer(chunkSize, initialWidth, initialHeight)

	visualizer.connectedDevice = targetDevice.Name()
	go func() {
		for {
			if visualizer.volume <= 600 {
				visualizer.waveForm = "smooth"
			} else if visualizer.volume > 600 {
				visualizer.waveForm = "smooth"
			}
		}
	}()

	// Capture audio data in a separate goroutine
	go func() {
		for {
			copy(visualizer.currentChunk, audioInput)
		}
	}()

	// Run the Ebiten game loop
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(initialWidth, initialHeight)
	ebiten.SetWindowTitle("Mezmer")
	if err := ebiten.RunGame(visualizer); err != nil {
		return fmt.Errorf("Failed to run visualizer: %v", err)
	}

	return nil
}

func StartMezmer() {
	err := RunMezmer()
	if err != nil {
		log.Println(err.Error())
	}
}
