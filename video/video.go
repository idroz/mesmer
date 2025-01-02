package video

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/user"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type Recording struct {
	width          int
	height         int
	frameRate      int
	frameCounter   int
	frameDuration  time.Duration
	audioChunkSize int
	sampleRate     int
	outputVideo    string
	OutputAudio    string
	AudioData      []byte
	tempFramesDir  string
	lastFrameTime  time.Time
	//MicDevice      *malgo.Device
}

// NewRecording initializes theencoder.
func NewRecording(width int, height int, frameRate int) (*Recording, error) {
	recordingDir, err := os.MkdirTemp("", "mezmer")
	if err != nil {
		return nil, err
	}
	//recordingDir := "/Users/ignat/go/src/github.com/idroz/mezmer/png"

	// Create a directory for temporary frames
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return nil, err
	}

	myself, error := user.Current()
	if error != nil {
		panic(error)
	}
	homedir := myself.HomeDir
	desktop := homedir + "/Desktop/"
	now := time.Now()

	return &Recording{
		width:         width,
		height:        height,
		frameRate:     frameRate,
		frameCounter:  0,
		lastFrameTime: time.Now(),
		AudioData:     []byte{},
		//MicDevice:     nil,
		tempFramesDir: recordingDir,
		frameDuration: time.Second / time.Duration(frameRate),
		outputVideo:   desktop + "/mezmer_" + now.Format("2006-01-02 15:04:05") + ".mp4",
		OutputAudio:   desktop + "/mezmer_" + now.Format("2006-01-02 15:04:05") + ".wav",
	}, nil
}

// Close cleans up resources.
func (r *Recording) Close() {
	os.RemoveAll(r.tempFramesDir)
}

func (r *Recording) NextFrame() {
	r.frameCounter += 1
}

// saveAudio writes the captured audio to a WAV file.
func (r *Recording) SaveAudio() error {
	file, err := os.Create(r.OutputAudio)
	if err != nil {
		log.Fatalf("Failed to create audio file: %v", err)
	}
	defer file.Close()

	// Write WAV header
	//dataLength := len(r.AudioInput) * 2 // Each int16 sample is 2 bytes
	sampleRate := uint32(44100) // 44.1kHz
	channels := uint16(2)       // Stereo
	bitsPerSample := uint16(16) // 16-bit PCM
	byteRate := sampleRate * uint32(channels) * uint32(bitsPerSample/8)
	blockAlign := channels * bitsPerSample / 8

	// WAV file header
	var wavHeader = struct {
		ChunkID       [4]byte
		ChunkSize     uint32
		Format        [4]byte
		Subchunk1ID   [4]byte
		Subchunk1Size uint32
		AudioFormat   uint16
		NumChannels   uint16
		SampleRate    uint32
		ByteRate      uint32
		BlockAlign    uint16
		BitsPerSample uint16
		Subchunk2ID   [4]byte
		Subchunk2Size uint32
	}{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     uint32(36 + len(r.AudioData)),
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16, // PCM header size
		AudioFormat:   1,  // PCM
		NumChannels:   channels,
		SampleRate:    sampleRate,
		ByteRate:      byteRate,
		BlockAlign:    blockAlign,
		BitsPerSample: bitsPerSample,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: uint32(len(r.AudioData)),
	}

	// Create the WAV file
	file, err = os.Create(r.OutputAudio)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the header
	err = binary.Write(file, binary.LittleEndian, wavHeader)
	if err != nil {
		return err
	}

	// Write the audio data
	_, err = file.Write(r.AudioData)
	if err != nil {
		return err
	}

	return nil
}

// Update advances the recording logic.
func (r *Recording) Update() error {
	// Limit frame rate
	now := time.Now()
	if now.Sub(r.lastFrameTime) < r.frameDuration {
		return nil
	}
	r.lastFrameTime = now
	r.frameCounter++

	return nil
}

func (r *Recording) SaveFrame(screen *ebiten.Image) error {
	img := screen.SubImage(screen.Bounds()).(*ebiten.Image)
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		return nil
	}

	decodedImg, err := png.Decode(buf)
	if err != nil {
		return nil
	}

	rgba := image.NewRGBA(decodedImg.Bounds())
	draw.Draw(rgba, decodedImg.Bounds(), decodedImg, image.Point{}, draw.Src)

	// Save the frame to the temporary directory
	framePath := fmt.Sprintf("%s/frame_%05d.png", r.tempFramesDir, r.frameCounter)
	file, err := os.Create(framePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, rgba); err != nil {
		return err
	}

	return nil
}

func (r *Recording) Layout(outsideWidth, outsideHeight int) (int, int) {
	return r.width, r.height
}

func (r *Recording) CreateVideoFromFrames() error {
	// Use ffmpeg-go to combine frames into a video
	err := ffmpeg_go.
		Input(fmt.Sprintf("%s/frame_%%05d.png", r.tempFramesDir), ffmpeg_go.KwArgs{"framerate": r.frameRate}).
		Output(r.outputVideo, ffmpeg_go.KwArgs{"pix_fmt": "yuv420p"}).
		OverWriteOutput().
		Run()

	if err != nil {
		return err
	}
	return nil
}
