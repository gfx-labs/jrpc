package main

import (
	"fmt"
	"os"
)

func main() {
	ctx := NewCLI()
	if err := ctx.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
