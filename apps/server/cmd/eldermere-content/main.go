package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "validate" {
		fmt.Fprintln(os.Stderr, "usage: eldermere-content validate <rooms.json>")
		os.Exit(2)
	}

	path := os.Args[2]
	dir := filepath.Dir(path)
	file := filepath.Base(path)

	if _, err := game.LoadWorld(os.DirFS(dir), file); err != nil {
		fmt.Fprintf(os.Stderr, "invalid content pack: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("valid content pack: %s\n", path)
}
