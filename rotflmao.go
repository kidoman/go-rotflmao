package main

import (
	"encoding/binary"
	"fmt"
	"github.com/jsgoecke/attspeech"
	"github.com/vova616/go-openal/openal"
	"io/ioutil"
	"math"
	"os"
	"time"
)

type Format struct {
	FormatTag     int16
	Channels      int16
	Samples       int32
	AvgBytes      int32
	BlockAlign    int16
	BitsPerSample int16
}

type Format2 struct {
	Format
	SizeOfExtension int16
}

type Format3 struct {
	Format2
	ValidBitsPerSample int16
	ChannelMask        int32
	SubFormat          [16]byte
}

func ReadWavFile(path string) (*Format, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	var buff [4]byte
	f.Read(buff[:4])

	if string(buff[:4]) != "RIFF" {
		return nil, nil, fmt.Errorf("Not a WAV file.\n")
	}

	var size int32
	binary.Read(f, binary.LittleEndian, &size)

	f.Read(buff[:4])

	if string(buff[:4]) != "WAVE" {
		return nil, nil, fmt.Errorf("Not a WAV file.\n")
	}

	f.Read(buff[:4])

	if string(buff[:4]) != "fmt " {
		return nil, nil, fmt.Errorf("Not a WAV file.\n")
	}

	binary.Read(f, binary.LittleEndian, &size)

	var format Format

	switch size {
	case 16:
		binary.Read(f, binary.LittleEndian, &format)
	case 18:
		var f2 Format2
		binary.Read(f, binary.LittleEndian, &f2)
		format = f2.Format
	case 40:
		var f3 Format3
		binary.Read(f, binary.LittleEndian, &f3)
		format = f3.Format
	}

	f.Read(buff[:4])

	if string(buff[:4]) != "data" {
		return nil, nil, fmt.Errorf("Not supported WAV file.\n")
	}

	binary.Read(f, binary.LittleEndian, &size)

	wavData := make([]byte, size)
	n, e := f.Read(wavData)
	if e != nil {
		return nil, nil, fmt.Errorf("Cannot read WAV data.\n")
	}
	if int32(n) != size {
		return nil, nil, fmt.Errorf("WAV data size doesnt match.\n")
	}

	return &format, wavData, nil
}

func Period(freq int, samples int) float64 {
	return float64(freq) * 2 * math.Pi * (1 / float64(samples))
}

func TimeToData(t time.Duration, samples int, channels int) int {
	return int((float64(samples)/(1/t.Seconds()))+0.5) * channels
}

func main() {
	expression := "ROTFLMAO!"
	holdyahorses := "WAIT FOR IT..."
	
	fmt.Println(expression)
	fmt.Println(holdyahorses)

	//TTS
	client := attspeech.New("fcs3ffhfpavzvc1iimvmtgxinnbb5agt", "cwdvrcfz7xm8dxw4bkstfenphwb8funw", "")
	client.SetAuthTokens()
	apiRequest := client.NewAPIRequest(attspeech.TTSResource)
	apiRequest.ContentType = "text/plain"
	apiRequest.Accept = "audio/x-wav"
	apiRequest.Text = expression
	data, err := client.TextToSpeech(apiRequest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile("./buffer.wav", data, 0444)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//OPEN-AL
	device := openal.OpenDevice("")
	context := device.CreateContext()
	context.Activate()

	source := openal.NewSource()
	source.SetPitch(1)
	source.SetGain(1)
	source.SetPosition(0, 0, 0)
	source.SetVelocity(0, 0, 0)
	source.SetLooping(false)

	buffer := openal.NewBuffer()
	format, data, err := ReadWavFile("./buffer.wav")
	if err != nil {
		panic(err)
	}

	switch format.Channels {
	case 1:
		buffer.SetData(openal.FormatMono16, data[:], int32(format.Samples))
	case 2:
		buffer.SetData(openal.FormatStereo16, data[:], int32(format.Samples))
	}

	source.SetBuffer(buffer)
	source.Play()
	
	for source.State() == openal.Playing {
		//loop long enough to let the wave file finish
	}
	
	source.Pause()
	source.Stop()
	return
	context.Destroy()
	time.Sleep(time.Second)
}
