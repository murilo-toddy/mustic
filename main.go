package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"math"
	"math/cmplx"
	"os"
	"os/exec"

	"github.com/gordonklaus/portaudio"
	"golang.org/x/term"
)

const (
    logFileName = "log.txt"
)

type MusicVisualizer struct {
    TopLeft Point
    Rows int
    Cols int
    Bars []float64
    barWidth int
    barSpacing int
    barPadding int
    canvas *Canvas
}

func NewMusicVisualizer(canvas *Canvas, topLeft Point, rows, cols int, bars []float64) *MusicVisualizer {
    barPadding := 3
    barSpacing := 2
    barWidth := 1
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
        height := int(math.Ceil(float64(m.Rows - 3) * m.Bars[index]))
        if index >= len(m.Bars) || row < height {
            return false
        }
        return true
    }

    return false
}

func (m *MusicVisualizer) Draw() error {
    m.canvas.DrawRect(m.TopLeft, Point{m.TopLeft.X + m.Rows - 1, m.TopLeft.Y + m.Cols - 1})

    for i, bar := range m.Bars {
        height := int(math.Ceil(float64(m.Rows - 3) * bar))
        row := m.TopLeft.Y + m.Rows - height
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

func fft(signal []complex128) []complex128 {
    n := len(signal)
    if n == 1 {
        return signal
    }

    odd := make([]complex128, n / 2)
    even := make([]complex128, n / 2)
    for i := range n / 2 {
        odd[i] = signal[2 * i]
        even[i] = signal[2 * i + 1]
    }

    oddFft := fft(odd)
    evenFft := fft(even)
    out := make([]complex128, n)

    for i := range n / 2 {
        component := -2.0 * math.Pi * float64(i) / float64(n)
        oddFactor := complex(math.Cos(component), math.Sin(component)) * oddFft[i]

        out[i] = evenFft[i] + oddFactor
        out[i + n / 2] = evenFft[i] - oddFactor
    }

    return out
}

func toComplexArray(arr []float64) []complex128 {
    out := make([]complex128, len(arr))
    for i, v := range arr {
        out[i] = complex(v, 0)
    }
    return out
}

func main() {
    logFile, err := os.Create(logFileName)
    defer func() {
        check(logFile.Close())
    }()

    w := bufio.NewWriter(logFile)
    log.SetOutput(w)
    check(err)


    width, height, err := term.GetSize(0)
    check(err)
    rows := height - 1
    cols := width

    numBars := 60
    bars := make([]float64, numBars)
    canvas := NewCanvas(rows, cols)
    musicVisualizer := NewMusicVisualizer(canvas, Point{0, 0}, rows, cols, bars)

    filename := getAudioFileArg()    

    output := createFfmpegPipe(filename)

    bufSize := 1 << 11
    samples := make([]float64, bufSize)

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
        bars = make([]float64, numBars)
        canvas = NewCanvas(rows, cols)
        musicVisualizer = NewMusicVisualizer(canvas, Point{0, 0}, rows, cols, bars)

        check(stream.Write())
        for i, sample := range audiobuf {
            samples[i] = float64(sample) / float64(math.Pow(2, 32))
        }
        frequencies := fft(toComplexArray(samples))
        amplitudes := make([]float64, len(frequencies))
        for i, freq := range frequencies {
            amplitudes[i] = math.Log(cmplx.Abs(freq))
        }

        step := 1.08
        lowf := 1.0
        m := 0
        maxAmp := 1.0
        outLog := make([]float64, 0)
        for f := lowf; f < float64(bufSize) / 2; f = math.Ceil(f * step) {
            f1 := math.Ceil(f*step)
            a := 0.0
            for q := int(f); q < bufSize / 2 && q < int(f1); q++ {
                b := amplitudes[q]
                if b > a {
                    a = b
                }
            }
            if maxAmp < a {
                maxAmp = a
            }
            outLog = append(outLog, a)
            m++
        }

        for i, v := range outLog {
            outLog[i] = v / maxAmp
        }

        for i := range bars {
            bars[i] = 0
        }

        for i := range bars {
            bars[i] = outLog[i]
        }

        canvas.Reset()
        musicVisualizer.Draw()
        canvas.Display()

        check(err)
        w.Flush()
    }

    check(err) 
}
