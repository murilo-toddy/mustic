package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"log"
	"math"
	"os"
	"os/exec"

	"github.com/gordonklaus/portaudio"
	"golang.org/x/term"
)

const (
	logFileName = "log.txt"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func createFfmpegPipe(filepath string) (output io.ReadCloser) {
	cmd := exec.Command("ffmpeg", "-i", filepath, "-f", "s32le", "-")
	output, err := cmd.StdoutPipe()
	check(err)
	err = cmd.Start()
	check(err)
	return
}

func toComplexArray(arr []float64) []complex128 {
	out := make([]complex128, len(arr))
	for i, v := range arr {
		out[i] = complex(v, 0)
	}
	return out
}

func redirectLogs() (logFileWriter *bufio.Writer) {
	// Redirect log messages to log file
	logFile, err := os.Create(logFileName)
	check(err)
	defer func() {
		check(logFile.Close())
	}()
	logFileWriter = bufio.NewWriter(logFile)
	log.SetOutput(logFileWriter)
	return
}

func startPortaudioStream(audiobuf []int32, bufSize int) (stream *portaudio.Stream) {
	inputChannels := 0
	outputChannels := 2
	sampleRate := 44100.0

	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, sampleRate, bufSize, &audiobuf)
	check(err)
	return
}

func main() {
	file := flag.String("filepath", "", "Name of the file to play music from")
	numBars := flag.Int("num_bars", 60, "Number of frequency bars to draw")
	flag.Parse()
	if file == nil || *file == "" {
		log.Fatal("Failed to parse flag \"filepath\"")
	}

	if _, err := os.Stat(*file); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("File \"%s\" not found", *file)
	}
	output := createFfmpegPipe(*file)

	width, height, err := term.GetSize(0)
	check(err)
	rows := height - 1
	cols := width

	bars := make([]float64, *numBars)
	canvas := NewCanvas(rows, cols)

	musicVisualizer := NewMusicVisualizer(*canvas, rows, cols, bars, *file)

	bufSize := 1 << 11
	samples := make([]float64, bufSize)
	fft := NewFFTCalculator(WithBufSize(bufSize))

	portaudio.Initialize()
	defer portaudio.Terminate()

	audiobuf := make([]int32, bufSize)
	stream := startPortaudioStream(audiobuf, bufSize)
	defer stream.Close()

	check(stream.Start())
	defer stream.Stop()

	logFileWriter := redirectLogs()
	for err = binary.Read(output, binary.LittleEndian, &audiobuf); err == nil; err = binary.Read(output, binary.LittleEndian, &audiobuf) {
		check(err)
		check(stream.Write())

		// normalize input sample
		for i, sample := range audiobuf {
			samples[i] = float64(sample) / float64(math.Pow(2, 32))
		}

		// calculate sample FFT
		outLog := fft.Calculate(samples)

		// update bars
		for i := range bars {
			bars[i] = outLog[i]
		}

		musicVisualizer.Draw()

		logFileWriter.Flush()
	}
	check(err)
}
