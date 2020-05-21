package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten"
	"github.com/jawr/fake/internal/game"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	g, err := game.NewGame(640, 640)
	if err != nil {
		return err
	}

	if err := ebiten.RunGame(g); err != nil {
		return err
	}

	return nil
}
