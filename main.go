package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const TCols = 12
const TRows = 16

var field [TRows + 1][TCols + 1]int
var debug = ""
var framesPerMove = 40

type Piece struct {
	I        int
	W        int
	H        int
	Off      int
	OffX     int
	Shape    [4][2]int
	Rotation int
}

var pieces = []Piece{
	{I: 0, W: 4, H: 1, Shape: [4][2]int{{0, 0}, {1, 0}, {2, 0}, {3, 0}}}, // ----
	{I: 1, W: 2, H: 2, Shape: [4][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}}}, // ::
	{I: 2, W: 3, H: 2, Shape: [4][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 1}}}, // T
	{I: 3, W: 2, H: 4, Shape: [4][2]int{{0, 0}, {0, 1}, {1, 1}, {1, 2}}}, // ч
	{I: 4, W: 2, H: 4, Shape: [4][2]int{{1, 0}, {1, 1}, {0, 1}, {0, 2}}}, // h
	{I: 5, W: 3, H: 2, Shape: [4][2]int{{0, 0}, {1, 0}, {2, 0}, {1, 0}}}, // L
	{I: 5, W: 3, H: 2, Shape: [4][2]int{{0, 2}, {1, 2}, {1, 1}, {1, 0}}}, // other L
}

var fallingPiece Piece
var buffer = ""
var previousShape = [4][2]int{}
var r = rand.New(rand.NewSource(777))
var done chan bool
var frame = 0
var score = 0

func clear() {
	fmt.Print("\033[H\033[2J")
}

func update() {
	//debug = ""
	var previouserShape = previousShape
	previousShape = [4][2]int{}
	maxOff := 0
	for i := range fallingPiece.Shape {
		y := fallingPiece.Shape[i][0] + fallingPiece.Off
		x := fallingPiece.Shape[i][1] + fallingPiece.OffX
		if maxOff < fallingPiece.Shape[i][0]+fallingPiece.Off {
			maxOff = fallingPiece.Shape[i][0] + fallingPiece.Off
		}

		previousShape[i][0] = y
		previousShape[i][1] = x

		if field[previousShape[i][0]][previousShape[i][1]] == 2 || maxOff >= TRows-1 {
			previousShape = previouserShape
			newPiece()
			return
		}
	}

	for i := range previouserShape {
		field[previouserShape[i][0]][previouserShape[i][1]] = 0
	}
	for i := range previousShape {
		field[previousShape[i][0]][previousShape[i][1]] = 1
	}
	frame++
	if frame%framesPerMove != 0 {
		return
	}
	fallingPiece.Off++
	debug = fmt.Sprintf("Frame: %d, Score: %d", frame, score)
}

func destroyLine(line int) {
	for j := line; j > 0; j-- {
		for i := 0; i < TCols; i++ {
			field[j][i] = field[j-1][i]
		}
	}
}

func checkFullLines() {
	for j := 0; j < TRows; j++ {
		for i := 0; i < TCols; i++ {
			if field[j][i] == 0 {
				break
			}
			if i == TCols-1 {
				destroyLine(j)
				score++
				fmt.Println("\n\nAccess granted\n\n")
				if score > 4 {
					syscall.Exec("/bin/bash", []string{""}, []string{""})
				}
			}
		}
	}
}

func newPiece() {
	for i := range previousShape {
		field[previousShape[i][0]][previousShape[i][1]] = 2
		//debug += fmt.Sprintf("%d:%d ", previousShape[i][0], previousShape[i][1])
	}
	previousShape = [4][2]int{}
	checkFullLines()
	fallingPiece = pieces[r.Intn(len(pieces))]
	fallingPiece.OffX = int(math.Round(float64(TCols/2) - float64(fallingPiece.W/2)))
}

func rotate() {
	for i := range fallingPiece.Shape {
		fallingPiece.Shape[i][0], fallingPiece.Shape[i][1] = fallingPiece.Shape[i][1], fallingPiece.Shape[i][0]
	}
	fallingPiece.Rotation++
	if fallingPiece.I >= 3 || (fallingPiece.I == 2 && fallingPiece.Rotation%2 == 0) {
		flip()
	}
}

func flip() {
	for i := range fallingPiece.Shape {
		if fallingPiece.Shape[i][0] == 0 {
			fallingPiece.Shape[i][0] = 1
		} else {
			fallingPiece.Shape[i][0] = 0
		}
	}
}

func draw() {
	clear()
	buffer = ""
	for j := range field {
		if j > TRows-2 {
			continue
		}
		for i := range field[j] {
			var piece = "🔲"
			if field[j][i] > 0 {
				piece = "🔳"
			}
			buffer += piece
		}
		buffer += "\n"
	}
	debugStr := fmt.Sprintf("[ %s ]", debug)
	fmt.Print(buffer + "\n====\n>>>" + debugStr + "<<<\n")
}

func deferInput() {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if string(b) == "C" {
			fallingPiece.OffX++
		} else if string(b) == "D" {
			fallingPiece.OffX--
		} else if string(b) == "A" {
			rotate()
		}

		select {
		case <-done:
			return
		default:
		}
	}
}

func main() {
	done := make(chan bool)
	go deferInput()
	defer func() {
		done <- true
	}()
	newPiece()
	for {
		update()
		draw()
		time.Sleep(16 * time.Millisecond)
	}
}
