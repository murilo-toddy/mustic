package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

const (
	topLeftChar       = "┌"
	topRightChar      = "┐"
	bottomLeftChar    = "└"
	bottomRightChar   = "┘"
	horizontalBarChar = "─"
	verticalBarChar   = "│"
	filledChar        = "█"
)

type Point struct {
	X int
	Y int
}

func (p *Point) unwrap() (int, int) {
	return p.X, p.Y
}

type Canvas struct {
	Rows   int
	Cols   int
	canvas [][]string
}

// TODO(perf): canvas operations should be buffered until next draw to avoid multiple changes to same cell
func NewCanvas(rows, cols int) *Canvas {
	canvas := make([][]string, rows)
	for i := range rows {
		canvas[i] = make([]string, cols)
	}
	for row := range rows {
		for col := range cols {
			canvas[row][col] = " "
		}
	}
	return &Canvas{
		Rows:   rows,
		Cols:   cols,
		canvas: canvas,
	}
}

func (c *Canvas) DrawCell(x, y int, value string) error {
	if x < 0 || x >= c.Rows || y < 0 || y >= c.Cols {
		invalidDrawMessage := fmt.Sprintf("WARN: attempting to draw on invalid cell (%d, %d), ignoring...", x, y)
		log.Printf(invalidDrawMessage)
		return errors.New(invalidDrawMessage)
	}

	c.canvas[x][y] = value
	return nil
}

func (c *Canvas) DrawPoint(point Point, value string) {
	c.DrawCell(point.X, point.Y, value)
}

func (c *Canvas) DrawFilledRect(topLeft, bottomRight Point) {
	rowStart, colStart := topLeft.unwrap()
	rowEnd, colEnd := bottomRight.unwrap()
	for row := rowStart; row <= rowEnd; row++ {
		for col := colStart; col <= colEnd; col++ {
			c.DrawCell(row, col, filledChar)
		}
	}
}

func (c *Canvas) DrawRect(topLeft, bottomRight Point) {
	rowStart, colStart := topLeft.unwrap()
	rowEnd, colEnd := bottomRight.unwrap()
	for row := rowStart + 1; row < rowEnd; row++ {
		c.DrawCell(row, colStart, verticalBarChar)
		c.DrawCell(row, colEnd, verticalBarChar)
	}
	for col := colStart + 1; col < colEnd; col++ {
		c.DrawCell(rowStart, col, horizontalBarChar)
		c.DrawCell(rowEnd, col, horizontalBarChar)
	}

	c.DrawCell(rowStart, colStart, topLeftChar)
	c.DrawCell(rowStart, colEnd, topRightChar)
	c.DrawCell(rowEnd, colStart, bottomLeftChar)
	c.DrawCell(rowEnd, colEnd, bottomRightChar)
}

func (c *Canvas) DrawText(startingPoint Point, value string) {
	for i := range value {
		c.DrawCell(startingPoint.X, startingPoint.Y+i, string(value[i]))
	}
}

func (c *Canvas) Reset() {
	for row := range c.Rows {
		for col := range c.Cols {
			c.canvas[row][col] = " "
		}
	}
}

func (c *Canvas) Display() {
	rows := make([]string, 0)
	for _, row := range c.canvas {
		rows = append(rows, strings.Join(row, ""))
	}
	fmt.Print(strings.Join(rows, "\n"))
	fmt.Println()

	fmt.Printf("\033[%dA", c.Rows)
	fmt.Printf("\033[%dD", c.Cols)
}
