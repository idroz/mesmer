package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/idroz/mezmer/visualiser"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Exiting Visualizer loop...")
		os.Exit(1) // Terminate the program gracefully
	}()

	err := visualiser.RunMezmer()
	if err != nil {
		log.Fatalf("Failed to start Mezmer: %v", err)
	}
}
