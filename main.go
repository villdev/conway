package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Cells = [20][40]bool

func main() {
	screen := newGameScreen()
	cells := initCells()
	generation := 1
	frameDelay := 100 * time.Millisecond

	renderUI(screen, cells, generation)

	for {
		time.Sleep(frameDelay)
		cells = getNextGeneration(cells)
		generation++

		renderUI(screen, cells, generation)

		if screen.HasPendingEvent() {
			event := screen.PollEvent()

			switch event := event.(type) {
			case *tcell.EventResize:
				screen.Sync()
			case *tcell.EventKey:
				if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
					screen.Fini()
					os.Exit(0)
				}
			}
		}
	}
}

func newGameScreen() tcell.Screen {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s.SetStyle(defStyle)
	s.Clear()

	return s
}

func renderUI(s tcell.Screen, cells Cells, generation int) {
	s.Clear()

	offsetX, offsetY := 1, 1

	liveCount := renderCellGrid(cells, offsetX, offsetY, s)
	renderCounts(cells, offsetY, generation, liveCount, s)
	renderControls(cells, offsetY, s)

	s.Show()
}

func renderCellGrid(cells Cells, offsetX, offsetY int, s tcell.Screen) int {
	liveCount := 0
	blockStyle := tcell.StyleDefault.Background(tcell.ColorDarkViolet).Foreground(tcell.ColorDarkViolet)
	gridStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorGrey)
	for i := 0; i <= len(cells); i++ {
		for j := 0; j <= len(cells[0]); j++ {
			x, y := j*(offsetX+1), i*(offsetY+1)

			if i < len(cells) && j < len(cells[0]) {
				if cells[i][j] {
					liveCount++
					s.SetContent(x, y, '█', nil, blockStyle)
				} else {
					s.SetContent(x, y, ' ', nil, gridStyle)
				}
			}

			if j < len(cells[0]) && i <= len(cells) {
				s.SetContent(x+offsetX, y, tcell.RuneVLine, nil, gridStyle)
			}

			if i < len(cells) && j <= len(cells[0]) {
				s.SetContent(x, y+offsetY, tcell.RuneHLine, nil, gridStyle)
			}
		}
	}
	return liveCount
}

func renderCounts(cells Cells, offsetY, generation, liveCount int, s tcell.Screen) {
	infoStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	infoX := 20
	infoY := (len(cells) + 1) * (offsetY + 1)
	infoText := fmt.Sprintf("Generation: %d, Live Cells: %d", generation, liveCount)
	for _, rune := range infoText {
		s.SetContent(infoX, infoY, rune, nil, infoStyle)
		infoX++
	}
}

func renderControls(cells Cells, offsetY int, s tcell.Screen) {
	infoStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	infoX := 25
	infoY := (len(cells)+1)*(offsetY+1) + 2
	infoText := "Press ESC to exit..."
	for _, rune := range infoText {
		s.SetContent(infoX, infoY, rune, nil, infoStyle)
		infoX++
	}
}

func initCells() Cells {
	var cells Cells
	for i := 0; i < len(cells); i++ {
		for j := 0; j < len(cells[i]); j++ {
			if rand.Intn(10) == 0 {
				cells[i][j] = true
			}
		}
	}
	return cells
}

func getNextGeneration(cells Cells) Cells {
	var newCells Cells
	for i := 0; i < len(cells); i++ {
		for j := 0; j < len(cells[i]); j++ {
			neighbourCount := countNeighbours(i, j, cells)
			alive := cells[i][j]
			if alive && (neighbourCount == 2 || neighbourCount == 3) {
				newCells[i][j] = true
			} else if !alive && neighbourCount == 3 {
				newCells[i][j] = true
			} else {
				newCells[i][j] = false
			}
		}
	}
	return newCells
}

func countNeighbours(currentRow, currentCol int, cells Cells) int {
	count := 0
	for i := bigger(0, currentRow-1); i <= smaller(currentRow+1, len(cells)-1); i++ {
		for j := bigger(0, currentCol-1); j <= smaller(currentCol+1, len(cells[i])-1); j++ {
			isMainCell := i == currentRow && j == currentCol
			if cells[i][j] && !isMainCell {
				count++
			}
		}
	}
	return count
}

func bigger(lowest, b int) int {
	if lowest > b {
		return lowest
	}
	return b
}

func smaller(a, highest int) int {
	if a > highest {
		return highest
	}
	return a
}
