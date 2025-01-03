package audio

import (
	"fmt"
	"log"

	"github.com/gen2brain/malgo"
)

const (
	chunkSize = 512
)

type AudioDevice struct {
	Device     *malgo.Device
	DeviceName string
	AudioInput []float64
}

func newAudioDevice() *AudioDevice {
	return &AudioDevice{
		Device:     nil,
		DeviceName: "",
		AudioInput: make([]float64, chunkSize),
	}

}

func MinAudio() (*AudioDevice, error) {
	audioDevice := newAudioDevice()

	// Initialize MiniAudio context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("MiniAudio Log: %s\n", message)
	})
	if err != nil {
		return nil, err
	}
	defer ctx.Uninit()
	defer ctx.Free()

	// List available capture devices
	captureDevices, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, err
	}

	fmt.Println("Capture Devices:")
	var targetDevice *malgo.DeviceInfo
	for _, device := range captureDevices {
		log.Printf("  %s\n", device.Name())
		if device.Name() == "OP-XY" {
			targetDevice = &device
			break
		}
	}

	if targetDevice == nil {
		return nil, err
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
		return nil, err
	}

	audioDevice.Device = device
	audioDevice.DeviceName = targetDevice.Name()
	audioDevice.AudioInput = audioInput

	return audioDevice, nil
}
