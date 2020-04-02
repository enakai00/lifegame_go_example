package main

import (
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

var mu sync.Mutex

type environ struct {
	sizeX    int
	sizeY    int
	field    [][]bool
	cursorX  int
	cursorY  int
	pause    bool
	duration int
}

func drawLine(x, y int, str string) {
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		termbox.SetCell(x+i, y, runes[i],
			termbox.ColorDefault, termbox.ColorDefault)
	}
}

func (env *environ) show(pause bool) {
	mu.Lock()
	for x := 0; x < env.sizeX+2; x++ {
		termbox.SetCell(x*2, 0, '＃',
			termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(x*2, env.sizeY+1, '＃',
			termbox.ColorDefault, termbox.ColorDefault)
	}
	for y := 0; y < env.sizeY; y++ {
		termbox.SetCell(0, y+1, '＃',
			termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(env.sizeX*2+2, y+1, '＃',
			termbox.ColorDefault, termbox.ColorDefault)
		for x := 0; x < env.sizeX; x++ {
			fgColor := termbox.ColorDefault
			bgColor := termbox.ColorDefault
			if pause && x == env.cursorX && y == env.cursorY {
				fgColor = termbox.ColorWhite
				bgColor = termbox.ColorMagenta
			}
			char := '　'
			if env.field[y][x] {
				char = '＋'
			}
			termbox.SetCell(x*2+2, y+1, char, fgColor, bgColor)
		}
	}
	message1 := "Move Cursor: [←][↓][↑][→] (or [h][j][k][l]), Flip Cell State: [SPACE]"
	message2 := "Pause/Run: [ESC], Quit: [Ctrl]+[C]"
	drawLine(1, env.sizeY+2, message1)
	drawLine(1, env.sizeY+3, message2)
	termbox.Flush()
	mu.Unlock()
}

func (env *environ) moveCursor(dx, dy int) {
	env.cursorX += dx
	env.cursorY += dy
	if env.cursorX < 0 {
		env.cursorX = 0
	}
	if env.cursorX > env.sizeX-1 {
		env.cursorX = env.sizeX - 1
	}
	if env.cursorY < 0 {
		env.cursorY = 0
	}
	if env.cursorY > env.sizeY-1 {
		env.cursorY = env.sizeY - 1
	}
}

func (env *environ) neighbors(x, y int) int {
	nb := 0
	for dy := -1; dy < 2; dy++ {
		for dx := -1; dx < 2; dx++ {
			if dx == 0 && dy == 0 ||
				x+dx < 0 || x+dx >= env.sizeX ||
				y+dy < 0 || y+dy >= env.sizeY {
				continue
			}
			if env.field[y+dy][x+dx] {
				nb += 1
			}
		}
	}
	return nb
}

func (env *environ) evolve() {
	newField := make([][]bool, env.sizeY)
	for y := 0; y < env.sizeY; y++ {
		newField[y] = make([]bool, env.sizeX)
		for x := 0; x < env.sizeX; x++ {
			nb := env.neighbors(x, y)
			if env.field[y][x] && (nb == 2 || nb == 3) {
				newField[y][x] = true
			}
			if !env.field[y][x] && nb == 3 {
				newField[y][x] = true
			}
		}
	}
	env.field = newField
}

func newEnviron() *environ {
	env := new(environ)
	env.sizeX, env.sizeY = 38, 20
	env.cursorX, env.cursorY = 0, 0
	env.duration = 100
	env.pause = true

	env.field = make([][]bool, env.sizeY)
	for y := 0; y < env.sizeY; y++ {
		env.field[y] = make([]bool, env.sizeX)
	}

	return env
}

func getKey() (termbox.Key, rune) {
	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			return ev.Key, ev.Ch
		}
	}
}

func evolve(env *environ, ch <-chan bool) {
	tick := time.Tick(time.Duration(env.duration) * time.Millisecond)
	for {
		select {
		case <-tick:
			if !env.pause {
				env.evolve()
				env.show(env.pause)
			}
		case pause := <-ch:
			env.pause = pause
		}
	}
}

func play() {
	env := newEnviron()
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	pauseCh := make(chan bool)
	go evolve(env, pauseCh)

	for {
		env.show(env.pause)
		key, ch := getKey()
		switch {
		case key == termbox.KeyEsc:
			env.pause = !env.pause
			pauseCh <- env.pause
		case key == termbox.KeyCtrlC:
			return
		}
		if !env.pause {
			continue
		}
		switch {
		case key == termbox.KeyArrowLeft || ch == 'h': // Left
			env.moveCursor(-1, 0)
		case key == termbox.KeyArrowDown || ch == 'j': // Down
			env.moveCursor(0, 1)
		case key == termbox.KeyArrowUp || ch == 'k': // Up
			env.moveCursor(0, -1)
		case key == termbox.KeyArrowRight || ch == 'l': // Right
			env.moveCursor(1, 0)
		case key == termbox.KeySpace:
			env.field[env.cursorY][env.cursorX] =
				!env.field[env.cursorY][env.cursorX]
		}
	}
}

func main() {
	play()
}
