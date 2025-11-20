package main

import (
	"fmt"
	"math"
	"path/filepath"
)

type MusicVisualizerOpts struct {
	barWidth   int
	barPadding int
	barSpacing int
}

type MusicVisualizerOptFunc func(*MusicVisualizerOpts)

type MusicVisualizer struct {
	MusicVisualizerOpts
	Canvas
	Rows      int
	Cols      int
	Bars      []float64
	musicName string
}

func NewMusicVisualizer(
	canvas Canvas,
	rows int,
	cols int,
	bars []float64,
	file string,
	opts ...MusicVisualizerOptFunc,
) *MusicVisualizer {
	config := MusicVisualizerOpts{
		barPadding: 3,
		barSpacing: 2,
		barWidth:   1,
	}
	for _, fn := range opts {
		fn(&config)
	}
	musicNameWithExt := filepath.Base(file)
	extension := filepath.Ext(musicNameWithExt)
	musicName := musicNameWithExt[0 : len(musicNameWithExt)-len(extension)]
	return &MusicVisualizer{
		MusicVisualizerOpts: config,
		Canvas:              canvas,
		Rows:                rows,
		Cols:                cols,
		Bars:                bars,
		musicName:           musicName,
	}
}

func (m *MusicVisualizer) Draw() {
	m.Reset()
	m.DrawRect(Point{0, 0}, Point{m.Rows - 1, m.Cols - 1})
	m.DrawText(Point{1, 1}, fmt.Sprintf("Now playing: %s", m.musicName))
	for i, bar := range m.Bars {
		height := int(math.Ceil(float64(m.Rows-3) * bar))
		row := m.Rows - height
		col := m.barPadding + i*(m.barWidth+m.barSpacing)
		m.DrawFilledRect(Point{row, col}, Point{m.Rows - 3, col + m.barWidth})
	}
	m.Display()
}
