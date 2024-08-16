package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"

	"github.com/gordonklaus/portaudio"
)

type MusicVisualizer struct {
    TopLeft Point
    Rows int
    Cols int
    Bars []int
    barWidth int
    barSpacing int
    barPadding int
    canvas *Canvas
}

func NewMusicVisualizer(canvas *Canvas, topLeft Point, rows, cols int, bars []int) *MusicVisualizer {
    barPadding := 3
    barSpacing := 2
    barWidth := 10
    return &MusicVisualizer{
        barWidth: barWidth,
        barSpacing: barSpacing,
        barPadding: barPadding,
        TopLeft: topLeft,
        Rows: rows,
        Cols: cols,
        Bars: bars,
        canvas: canvas,
    }
}

func (m *MusicVisualizer) onHorizontalBar(row, col int) bool {
    if col < m.barPadding || col > m.Cols - m.barPadding - 1 {
        return false
    } 
    col = col - m.barPadding - 1

    if col % (m.barSpacing + m.barWidth) < m.barWidth {
        index := col / (m.barSpacing + m.barWidth)
        if index >= len(m.Bars) || row < m.Rows - 2 - m.Bars[index] {
            return false
        }
        return true
    }

    return false
}

func (m *MusicVisualizer) Draw() error {
    m.canvas.DrawRect(m.TopLeft, Point{m.TopLeft.X + m.Rows - 1, m.TopLeft.Y + m.Cols - 1})

    for i, bar := range m.Bars {
        row := m.TopLeft.Y + m.Rows - bar
        col := m.TopLeft.X + m.barPadding + i * (m.barWidth + m.barSpacing)
        m.canvas.DrawFilledRect(Point{row, col}, Point{m.Rows - 2, col + m.barWidth})
    }

    return nil
}


func getAudioFileArg() (filename string) {
    if len(os.Args) < 2 {
        log.Fatal("Missing argument: input file name")
    }
    filename = os.Args[1]
    return
}

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func createFfmpegPipe(filename string) (output io.ReadCloser) {
    cmd := exec.Command("ffmpeg", "-i", filename, "-f", "s32le", "-")
    output, err := cmd.StdoutPipe()
    check(err)

    err = cmd.Start()
    check(err)

    return
}

func save(frames []float64) {
    path := "./out.txt"
    fmt.Println("saving")
    file, err := os.Create(path)
    check(err)
    defer file.Close()

    w := bufio.NewWriter(file)
    fmt.Fprintln(w, frames)
    err = w.Flush()
    fmt.Println("saved")
    check(err)
}

func playAudioFile(filename string) {
    output := createFfmpegPipe(filename)
    samples := make([]float64, 0)

    bufSize := 2048
    audiobuf := make([]int32, bufSize)
    portaudio.Initialize()
    defer portaudio.Terminate()
    
    inputChannels := 0
    outputChannels := 2
    sampleRate := 44100.0

    stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, sampleRate, bufSize, &audiobuf)
    check(err)
    defer stream.Close()
    
    check(stream.Start())
    defer stream.Stop()

    for err = binary.Read(output, binary.LittleEndian, &audiobuf); err == nil; err = binary.Read(output, binary.LittleEndian, &audiobuf) {
        check(stream.Write())
        for _, sample := range audiobuf {
            samples = append(samples, float64(sample) / float64(math.Pow(2, 32)))
        }
        // fft
    }

    check(err) 
}


func main() {
    cols := 200
    rows := 46
    numBars := 15

    bars := make([]int, numBars)
    canvas := NewCanvas(rows, cols)
    musicVisualizer := NewMusicVisualizer(canvas, Point{0, 0}, rows, cols, bars)

    for {
        for i := range bars {
            bars[i] = bars[i] + 1
        }

        musicVisualizer.Draw()
        canvas.Display()
    }
//    filename := getAudioFileArg()    
//    playAudioFile(filename)
}
