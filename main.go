package main

import (
	"os"
	"os/signal"
)

func main() {
	commandArguments := NewCommandArguments()
	blastInstance := commandArguments.Parse()

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	go func() {
		<-interruptChannel
		blastInstance.Stop()
	}()

	blastInstance.WaitForCompletion()
}
