package main

import (
	"os"

	"github.com/kheina/openconman/src/cmd"
)

func main() {
	os.Exit(cmd.Run(os.Args[1:]))
}
