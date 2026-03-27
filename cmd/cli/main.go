package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("AWD Arena Platform CLI v0.1.0")
		fmt.Println("Usage: awd-cli <command>")
		fmt.Println("Commands: version, game list")
		return
	}

	switch os.Args[1] {
	case "version":
		fmt.Println("awd-arena v0.1.0")
	case "game":
		fmt.Println("No games found.")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
