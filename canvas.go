package main

import (
    "fmt"
    "strings"
    "time"
)

const (
    topLeftChar = "┌"
    topRightChar = "┐"
    bottomLeftChar = "└"
    bottomRightChar = "┘"
    horizontalBarChar = "─"
    verticalBarChar = "│"
    filledChar = "█"
)

type Point struct {
    X int
    Y int
}

func (p *Point) unwrap() (int, int) {
    return p.X, p.Y
}

type Canvas struct {
    Rows int
    Cols int
    canvas [][]string
}

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
        Rows: rows,
        Cols: cols,
        canvas: canvas,
    }
}

func (c *Canvas) DrawCell(x, y int, value string) {
    c.canvas[x][y] = value
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

func (c *Canvas) Display() {
    rows := make([]string, 0)
    for _, row := range c.canvas {
        rows = append(rows, strings.Join(row, ""))
    }
    fmt.Print(strings.Join(rows, "\n"))
    fmt.Println()

    time.Sleep(300 * time.Millisecond)

    fmt.Printf("\033[%dA", c.Rows)
    fmt.Printf("\033[%dD", c.Cols)
}
