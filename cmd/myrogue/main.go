package main

import (
	"log"

	"github.com/gentaman/myrogue/internal/platform/ebiten"
)

func main() {
	if err := ebiten.Run(); err != nil {
		log.Fatal(err)
	}
}
