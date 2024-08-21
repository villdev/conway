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

type State string

const (
	Start   State = "Start"
	Paused  State = "Paused"
	Running State = "Running"
)

func main() {
	var cells Cells
	currentState := Start
	showGrid := true
	screen := newGameScreen()
	generation := 0
	frameDelay := 100 * time.Millisecond
	idleFrameDelay := 200 * time.Millisecond

	renderUI(screen, cells, currentState, generation, showGrid)

	for {
		switch currentState {
		case Start:
			time.Sleep(idleFrameDelay)
			var defaultCells Cells
			cells = defaultCells
		case Running:
			time.Sleep(frameDelay)
			if generation == 0 {
				cells = initCells()
			} else {
				cells = getNextGeneration(cells)
			}
			generation++
		case Paused:
			time.Sleep(idleFrameDelay)
		default:
			time.Sleep(idleFrameDelay)
		}

		renderUI(screen, cells, currentState, generation, showGrid)

		if screen.HasPendingEvent() {
			handleKeyInputs(screen, &currentState, &showGrid, &generation)
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

func renderUI(s tcell.Screen, cells Cells, currentState State, generation int, showGrid bool) {
	s.Clear()

	offsetX, offsetY := 1, 1

	liveCount := renderCellGrid(cells, offsetX, offsetY, s, showGrid)
	renderState(cells, offsetY, generation, liveCount, s, currentState)
	renderControls(cells, offsetY, s, currentState, showGrid)

	s.Show()
}

func renderCellGrid(cells Cells, offsetX, offsetY int, s tcell.Screen, showGrid bool) int {
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

			if showGrid {
				if j < len(cells[0]) && i <= len(cells) {
					s.SetContent(x+offsetX, y, tcell.RuneVLine, nil, gridStyle)
				}

				if i < len(cells) && j <= len(cells[0]) {
					s.SetContent(x, y+offsetY, tcell.RuneHLine, nil, gridStyle)
				}
			}
		}
	}
	return liveCount
}

func renderState(cells Cells, offsetY, generation, liveCount int, s tcell.Screen, currentState State) {
	infoStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	infoX := 20
	infoY := (len(cells) + 1) * (offsetY + 1)
	infoText := fmt.Sprintf("Generation: %d, Live Cells: %d | %s", generation, liveCount, currentState)
	for _, rune := range infoText {
		s.SetContent(infoX, infoY, rune, nil, infoStyle)
		infoX++
	}
}

func renderControls(cells Cells, offsetY int, s tcell.Screen, currentState State, showGrid bool) {
	infoStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	infoX := 18
	infoY := (len(cells)+1)*(offsetY+1) + 2
	var infoText string
	if currentState == Running {
		infoText = "Space -> Pause"
	} else {
		infoText = "Space -> Play"
	}
	if showGrid {
		infoText += " | G -> Hide Grid"
	} else {
		infoText += " | G -> Show Grid"
	}
	infoText += " | R -> Reset"
	for _, rune := range infoText {
		s.SetContent(infoX, infoY, rune, nil, infoStyle)
		infoX++
	}

	infoText = "Press ESC to exit..."
	infoX = 30
	infoY = (len(cells)+1)*(offsetY+1) + 4
	for _, rune := range infoText {
		s.SetContent(infoX, infoY, rune, nil, infoStyle)
		infoX++
	}
}

func handleKeyInputs(s tcell.Screen, currentState *State, showGrid *bool, generation *int) {
	event := s.PollEvent()

	switch event := event.(type) {
	case *tcell.EventResize:
		s.Sync()
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			s.Fini()
			os.Exit(0)
		case tcell.KeyRune:
			if event.Rune() == ' ' {
				switch *currentState {
				case Start, Paused:
					*currentState = Running
				case Running:
					*currentState = Paused
				}
			} else if event.Rune() == 'G' || event.Rune() == 'g' {
				*showGrid = !*showGrid
			} else if event.Rune() == 'R' || event.Rune() == 'r' {
				*currentState = Start
				*generation = 0
			}
		}
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
