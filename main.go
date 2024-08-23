package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Cell struct {
	alive   bool
	updated bool
}

type Cells = [20][40]Cell

type State string

const (
	Start   State = "Start"
	Paused  State = "Paused"
	Running State = "Running"
)

const (
	Random int = iota
	Block
	Beehive
	Blinker
	Toad
	Glider
	Spaceship
	Spaceship2
	Diehard
	Acorn
)

func main() {
	var cells Cells
	currentState := Start
	showGrid := true
	previousShowGrid := !showGrid
	patternIndex := 0
	patterns := []string{
		"Random",
		"Block",
		"Beehive",
		"Blinker",
		"Toad",
		"Glider",
		"Spaceship",
		"Spaceship2",
		"Diehard",
		"Acorn",
	}
	screen := newGameScreen()
	generation := 0
	frameDelay := 100 * time.Millisecond
	idleFrameDelay := 200 * time.Millisecond

	resetUI := func() {
		screen.Clear()
		for i := range cells {
			for j := range cells[i] {
				cells[i][j].updated = true
			}
		}
		previousShowGrid = !showGrid
		renderUI(screen, cells, currentState, generation, showGrid, previousShowGrid, patterns[patternIndex])
	}

	for {
		switch currentState {
		case Start:
			time.Sleep(idleFrameDelay)
			cells = newCells()
		case Running:
			time.Sleep(frameDelay)
			if generation == 0 {
				cells = initCells(patternIndex)
			} else {
				cells = getNextGeneration(cells)
			}
			generation++
		case Paused:
			time.Sleep(idleFrameDelay)
		default:
			time.Sleep(idleFrameDelay)
		}

		renderUI(screen, cells, currentState, generation, showGrid, previousShowGrid, patterns[patternIndex])
		previousShowGrid = showGrid

		if screen.HasPendingEvent() {
			handleKeyInputs(screen, &currentState, &showGrid, &generation, patterns, &patternIndex, resetUI)
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

func renderUI(s tcell.Screen, cells Cells, currentState State, generation int, showGrid bool, previousShowGrid bool, pattern string) {
	_, gridHeight, startX, startY := getScreenDimensions(s, cells)

	liveCount := renderCellGrid(cells, s, showGrid, previousShowGrid, startX, startY)
	renderState(cells, generation, liveCount, s, currentState, pattern, startX, startY+gridHeight+1)
	renderControls(cells, s, currentState, showGrid, startX, startY+gridHeight+3)
	renderStaticUI(s, cells, startX, startY+gridHeight+5)
	s.Show()
}

func renderCellGrid(cells Cells, s tcell.Screen, showGrid bool, previousShowGrid bool, startX int, startY int) int {
	liveCount := 0
	blockStyle := tcell.StyleDefault.Background(tcell.ColorDarkViolet).Foreground(tcell.ColorDarkViolet)
	gridStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorGrey)
	emptyStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	for i := 0; i <= len(cells); i++ {
		for j := 0; j <= len(cells[0]); j++ {
			x, y := startX+j*2, startY+i*2

			if i < len(cells) && j < len(cells[0]) {
				if cells[i][j].updated {
					if cells[i][j].alive {
						liveCount++
						s.SetContent(x, y, '█', nil, blockStyle)
					} else {
						s.SetContent(x, y, ' ', nil, emptyStyle)
					}
				}
			}

			if showGrid != previousShowGrid {
				gridChar := ' '
				style := emptyStyle
				if showGrid {
					style = gridStyle
				}

				if j < len(cells[0]) && i <= len(cells) {
					if showGrid {
						gridChar = tcell.RuneVLine
					}
					s.SetContent(x+1, y, gridChar, nil, style)
				}
				if i < len(cells) && j <= len(cells[0]) {
					if showGrid {
						gridChar = tcell.RuneHLine
					}
					s.SetContent(x, y+1, gridChar, nil, style)
				}
			}
		}
	}
	return liveCount
}

func renderState(cells Cells, generation, liveCount int, s tcell.Screen, currentState State, pattern string, startX int, startY int) {
	var infoText string
	if currentState == Start {
		infoText = fmt.Sprintf("Pattern: < %s >, ", pattern)
	} else {
		infoText = fmt.Sprintf("Pattern: %s, ", pattern)
	}
	infoText += fmt.Sprintf("Generation: %d, Live Cells: %d | %s", generation, liveCount, currentState)

	maxLen := len(cells[0]) * 2
	drawString(infoText, startY, startX, maxLen, s)
}

func renderControls(cells Cells, s tcell.Screen, currentState State, showGrid bool, startX int, startY int) {
	var infoText string
	switch currentState {
	case Start:
		infoText = "Right/Left -> Change Pattern | Space -> Play"
	case Running:
		infoText = "Space -> Pause | R -> Reset"
	case Paused:
		infoText = "Space -> Play | R -> Reset"
	}

	if showGrid {
		infoText += " | G -> Hide Grid"
	} else {
		infoText += " | G -> Show Grid"
	}

	maxLen := len(cells[0]) * 2
	drawString(infoText, startY, startX, maxLen, s)
}

func renderStaticUI(s tcell.Screen, cells Cells, startX int, startY int) {
	infoText := "Press ESC to exit..."
	maxLen := len(cells[0]) * 2
	drawString(infoText, startY, startX, maxLen, s)
}

func drawString(text string, row int, startX int, maxLen int, s tcell.Screen) {
	infoStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	if len(text) < maxLen {
		padding := (maxLen - len(text)) / 2
		text = fmt.Sprintf("%s%s%s", strings.Repeat(" ", padding), text, strings.Repeat(" ", maxLen-len(text)-padding))
	}

	for x, rune := range text {
		s.SetContent(startX+x, row, rune, nil, infoStyle)
	}
}

func getScreenDimensions(s tcell.Screen, cells Cells) (int, int, int, int) {
	width, height := s.Size()

	gridWidth := len(cells[0])*2 + 1
	gridHeight := len(cells)*2 + 1
	startX := (width - gridWidth) / 2
	startY := (height - gridHeight - 3) / 2

	return gridWidth, gridHeight, startX, startY
}

func handleKeyInputs(s tcell.Screen, currentState *State, showGrid *bool, generation *int, patterns []string, patternIndex *int, resetUI func()) {
	event := s.PollEvent()

	switch event := event.(type) {
	case *tcell.EventResize:
		s.Sync()
		resetUI()
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
				*patternIndex = 0
			}
		case tcell.KeyLeft:
			if *currentState == Start {
				if *patternIndex == 0 {
					*patternIndex = len(patterns) - 1
				} else {
					*patternIndex--
				}
			}
		case tcell.KeyRight:
			if *currentState == Start {
				*patternIndex = (*patternIndex + 1) % len(patterns)
			}
		}
	}
}

func newCells() Cells {
	var defaultCells Cells
	for i := 0; i < len(defaultCells); i++ {
		for j := 0; j < len(defaultCells[i]); j++ {
			defaultCells[i][j].updated = true
		}
	}
	return defaultCells
}

func initCells(patternIndex int) Cells {
	var cells Cells
	switch patternIndex {
	case Block:
		cells[1][1] = Cell{alive: true, updated: true}
		cells[1][2] = Cell{alive: true, updated: true}
		cells[2][1] = Cell{alive: true, updated: true}
		cells[2][2] = Cell{alive: true, updated: true}
	case Beehive:
		cells[1][2] = Cell{alive: true, updated: true}
		cells[1][3] = Cell{alive: true, updated: true}
		cells[2][1] = Cell{alive: true, updated: true}
		cells[2][4] = Cell{alive: true, updated: true}
		cells[3][2] = Cell{alive: true, updated: true}
		cells[3][3] = Cell{alive: true, updated: true}
	case Blinker:
		cells[10][20] = Cell{alive: true, updated: true}
		cells[10][21] = Cell{alive: true, updated: true}
		cells[10][22] = Cell{alive: true, updated: true}
	case Toad:
		cells[10][20] = Cell{alive: true, updated: true}
		cells[10][21] = Cell{alive: true, updated: true}
		cells[10][22] = Cell{alive: true, updated: true}
		cells[11][19] = Cell{alive: true, updated: true}
		cells[11][20] = Cell{alive: true, updated: true}
		cells[11][21] = Cell{alive: true, updated: true}
	case Glider:
		cells[1][2] = Cell{alive: true, updated: true}
		cells[2][3] = Cell{alive: true, updated: true}
		cells[3][1] = Cell{alive: true, updated: true}
		cells[3][2] = Cell{alive: true, updated: true}
		cells[3][3] = Cell{alive: true, updated: true}
	case Spaceship:
		cells[1][2] = Cell{alive: true, updated: true}
		cells[1][3] = Cell{alive: true, updated: true}
		cells[1][4] = Cell{alive: true, updated: true}
		cells[1][5] = Cell{alive: true, updated: true}
		cells[2][1] = Cell{alive: true, updated: true}
		cells[2][5] = Cell{alive: true, updated: true}
		cells[3][5] = Cell{alive: true, updated: true}
		cells[4][1] = Cell{alive: true, updated: true}
		cells[4][4] = Cell{alive: true, updated: true}
	case Spaceship2:
		cells[10][1] = Cell{alive: true, updated: true}
		cells[10][2] = Cell{alive: true, updated: true}
		cells[10][3] = Cell{alive: true, updated: true}
		cells[10][4] = Cell{alive: true, updated: true}
		cells[10][5] = Cell{alive: true, updated: true}
		cells[11][0] = Cell{alive: true, updated: true}
		cells[11][5] = Cell{alive: true, updated: true}
		cells[12][5] = Cell{alive: true, updated: true}
		cells[13][4] = Cell{alive: true, updated: true}
		cells[13][0] = Cell{alive: true, updated: true}
	case Diehard:
		cells[10][25] = Cell{alive: true, updated: true}
		cells[11][19] = Cell{alive: true, updated: true}
		cells[11][20] = Cell{alive: true, updated: true}
		cells[12][20] = Cell{alive: true, updated: true}
		cells[12][24] = Cell{alive: true, updated: true}
		cells[12][25] = Cell{alive: true, updated: true}
		cells[12][26] = Cell{alive: true, updated: true}
	case Acorn:
		cells[10][20] = Cell{alive: true, updated: true}
		cells[11][22] = Cell{alive: true, updated: true}
		cells[12][19] = Cell{alive: true, updated: true}
		cells[12][20] = Cell{alive: true, updated: true}
		cells[12][23] = Cell{alive: true, updated: true}
		cells[12][24] = Cell{alive: true, updated: true}
		cells[12][25] = Cell{alive: true, updated: true}
	default:
		for i := 0; i < len(cells); i++ {
			for j := 0; j < len(cells[i]); j++ {
				if rand.Intn(10) == 0 {
					cells[i][j] = Cell{alive: true, updated: true}
				}
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
			alive := cells[i][j].alive
			if alive && (neighbourCount == 2 || neighbourCount == 3) {
				newCells[i][j] = Cell{alive: true, updated: false}
			} else if !alive && neighbourCount == 3 {
				newCells[i][j] = Cell{alive: true, updated: true}
			} else {
				if alive {
					newCells[i][j] = Cell{alive: false, updated: true}
				} else {
					newCells[i][j] = Cell{alive: false, updated: false}
				}
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
			if cells[i][j].alive && !isMainCell {
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
