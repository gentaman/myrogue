package main

import (
	"log"

	"github.com/gentaman/myrogue/internal/game"
)

func main() {
	if err := game.Run(); err != nil {
		log.Fatal(err)
	}
}
