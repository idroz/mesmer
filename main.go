package main

import (
	"log"

	"github.com/idroz/mezmer/visualiser"
)

func main() {
	err := visualiser.RunMezmer()
	if err != nil {
		log.Fatalf("Failed to start Mezmer: %v", err)
	}
}
