package main

import (
	"github.com/SarthakMakhija/blast-core/cmd"
	"os"
	"os/signal"
)

func main() {
	commandArguments := blast.NewCommandArguments()
	blastInstance := commandArguments.Parse("blast")

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	go func() {
		<-interruptChannel
		blastInstance.Stop()
	}()

	blastInstance.WaitForCompletion()
}
