package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var ErrGameOver = fmt.Errorf("Game Over")

const foodPoints = 100

const cellSize = 32

type cellType int

const (
	cellTypeEmpty cellType = iota
	cellTypeFood
	cellTypeHead
	cellTypeBody
	cellTypeTail
)

type collisionType int

const (
	collisionTypeNone collisionType = iota
	collisionTypeWall
	collisionTypeSnake
	collisionTypeFood
)

type direction int

const (
	directionNone direction = iota
	directionLeft
	directionUp
	directionRight
	directionDown
)

type cell struct {
	current       cellType
	direction     direction
	nextDirection direction
}

type Game struct {
	screenWidth, screenHeight int

	snake []int
	food  int

	grid       []*cell
	rows, cols int

	fox, head, body, bodyTurnBig, bodyTurnLittle, tail *ebiten.Image

	direction direction

	score int
}

func NewGame(screenWidth, screenHeight int) (*Game, error) {

	// load assets
	head, _, err := ebitenutil.NewImageFromFile("./assets/head.png", ebiten.FilterDefault)
	if err != nil {
		return nil, errors.WithMessage(err, "head")
	}
	body, _, err := ebitenutil.NewImageFromFile("./assets/body.png", ebiten.FilterDefault)
	if err != nil {
		return nil, errors.WithMessage(err, "body")
	}
	bodyTurnBig, _, err := ebitenutil.NewImageFromFile("./assets/body-turn-big.png", ebiten.FilterDefault)
	if err != nil {
		return nil, errors.WithMessage(err, "body")
	}
	bodyTurnLittle, _, err := ebitenutil.NewImageFromFile("./assets/body-turn-little.png", ebiten.FilterDefault)
	if err != nil {
		return nil, errors.WithMessage(err, "body")
	}
	tail, _, err := ebitenutil.NewImageFromFile("./assets/tail.png", ebiten.FilterDefault)
	if err != nil {
		return nil, errors.WithMessage(err, "tail")
	}
	food, _, err := ebitenutil.NewImageFromFile("./assets/food.png", ebiten.FilterDefault)
	if err != nil {
		return nil, errors.WithMessage(err, "food")
	}

	// create our grid
	rows := screenHeight / cellSize
	cols := screenWidth / cellSize

	grid := make([]*cell, rows*cols)
	for i := 0; i < len(grid); i++ {
		grid[i] = &cell{}
	}

	// create our snake
	snake := make([]int, 3)

	g := &Game{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,

		snake: snake,

		grid: grid,
		rows: rows,
		cols: cols,

		head: head,

		body:           body,
		bodyTurnBig:    bodyTurnBig,
		bodyTurnLittle: bodyTurnLittle,

		tail: tail,
		fox:  food,
	}

	g.initSnake()
	if err := g.initFood(); err != nil {
		return nil, err
	}

	ebiten.SetMaxTPS(10)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Fake")

	return g, nil
}

func (g *Game) Update(screen *ebiten.Image) error {

	g.handleInput()

	if g.food == -1 {
		if err := g.initFood(); err != nil {
			return err
		}
	}

	if err := g.handleSnakeChanges(); err != nil {
		return err
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	var img *ebiten.Image

DRAW:
	for i := range g.grid {
		switch g.grid[i].current {
		case cellTypeFood:
			img = g.fox

		case cellTypeHead:
			img = g.head

		case cellTypeBody:
			img = g.body

			if g.grid[i].direction != g.grid[i].nextDirection {
				var clockwise bool

				d := g.grid[i].direction
				n := g.grid[i].nextDirection

				switch {
				case d == directionUp && n == directionRight:
					fallthrough
				case d == directionRight && n == directionDown:
					fallthrough
				case d == directionDown && n == directionLeft:
					fallthrough
				case d == directionLeft && n == directionUp:
					clockwise = true
				}

				// clockwise is a big turn
				if clockwise {
					img = g.bodyTurnLittle
				} else {
					// counter clockwise is a little turn
					img = g.bodyTurnBig
				}
			}

		case cellTypeTail:
			img = g.tail

		default:
			continue DRAW
		}

		x, y := g.getCellXY(i)

		op := &ebiten.DrawImageOptions{}

		// handle rotations
		switch g.grid[i].direction {
		case directionUp:
			w, h := img.Size()

			// move center of image to the origin 0,0
			op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			op.GeoM.Rotate(float64(90%360) * 2 * math.Pi / 360)
			// reset the center of the image to the top left of the image
			op.GeoM.Translate(float64(w)/2, float64(h)/2)

		case directionRight:
			// op.GeoM.Scale(-1, 1)
			w, h := img.Size()
			op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			op.GeoM.Rotate(float64(180%360) * 2 * math.Pi / 360)
			op.GeoM.Translate(float64(w)/2, float64(h)/2)

		case directionDown:
			w, h := img.Size()
			op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
			op.GeoM.Rotate(float64(270%360) * 2 * math.Pi / 360)
			op.GeoM.Translate(float64(w)/2, float64(h)/2)
		}

		op.GeoM.Translate(x, y)

		screen.DrawImage(img, op)
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return g.screenWidth, g.screenHeight
}

func (g *Game) getCellXY(idx int) (float64, float64) {
	row := idx / g.rows
	col := math.Floor(float64(idx % g.rows))

	return col * float64(cellSize), float64(row * cellSize)
}

func (g *Game) handleInput() {
	headCell := g.grid[g.snake[len(g.snake)-1]]

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if headCell.direction != directionDown {
			headCell.direction = directionUp
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
		if headCell.direction != directionUp {
			headCell.direction = directionDown
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if headCell.direction != directionRight {
			headCell.direction = directionLeft
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if headCell.direction != directionLeft {
			headCell.direction = directionRight
		}
	}
}

func (g *Game) handleSnakeChanges() error {
	// snake is backwards to avoid shuffling

	var lastCell cell

SNAKE:
	for i := len(g.snake) - 1; i >= 0; i-- {

		current := g.snake[i]
		direction := g.grid[current].direction

		next, colType := g.nextCell(g.snake[i], direction)

		switch colType {
		case collisionTypeWall:
			fallthrough
		case collisionTypeSnake:
			// game over
			return ErrGameOver

		case collisionTypeFood:
			g.score += foodPoints

			// food becomes the head
			g.snake = append(g.snake, g.food)
			g.grid[g.food].current = cellTypeHead
			g.grid[g.food].direction = direction
			g.grid[current].current = cellTypeBody

			// reset food
			g.food = -1

			// break to allow for a clean update
			break SNAKE

		case collisionTypeNone:
			break
		}

		nextDirection := direction
		if lastCell.direction != directionNone {
			direction = lastCell.direction
		}
		lastCell = *g.grid[current]

		// move segment
		g.grid[next].direction = direction
		g.grid[next].current = g.grid[current].current
		g.grid[next].nextDirection = nextDirection

		// remove previous
		g.grid[current].direction = directionNone
		g.grid[current].current = cellTypeEmpty

		// update snake
		g.snake[i] = next
	}

	return nil
}

func (g *Game) nextCell(idx int, direction direction) (int, collisionType) {
	row := idx / g.rows
	col := int(math.Floor(float64(idx % g.rows)))

	switch direction {
	case directionUp:
		row--
	case directionDown:
		row++
	case directionLeft:
		col--
	case directionRight:
		col++
	}

	next := (row * g.cols) + col

	if col < 0 || col >= g.cols || next < 0 || next >= len(g.grid) {
		return 0, collisionTypeWall
	}

	// dont allow to bump in to score line
	if next <= g.cols {
		return 0, collisionTypeWall
	}

	switch g.grid[next].current {
	case cellTypeFood:
		return next, collisionTypeFood
	case cellTypeHead:
	case cellTypeBody:
	case cellTypeTail:
		return 0, collisionTypeSnake
	}

	return next, collisionTypeNone
}

func (g *Game) initSnake() {
	// hardcoded for now
	g.snake[2] = 48
	g.snake[1] = 49
	g.snake[0] = 50

	// hardcoded for now
	g.grid[48].direction = directionLeft
	g.grid[48].current = cellTypeHead
	g.grid[49].direction = directionLeft
	g.grid[49].current = cellTypeBody
	g.grid[50].direction = directionLeft
	g.grid[50].current = cellTypeTail
}

func (g *Game) initFood() error {
	for i := 0; i < len(g.grid)*2; i++ {
		col := rand.Intn(len(g.grid)-g.cols-1) + g.cols
		if g.grid[col].current == cellTypeEmpty {
			g.food = col
			g.grid[col].current = cellTypeFood
			return nil
		}
	}
	return errors.New("No space left")
}
