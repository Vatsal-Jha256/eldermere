package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "validate" {
		fmt.Fprintln(os.Stderr, "usage: eldermere-content validate <rooms.json|pack-directory>")
		os.Exit(2)
	}

	path := os.Args[2]
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid path: %v\n", err)
		os.Exit(1)
	}
	if info.IsDir() {
		validatePack(path)
		return
	}

	validateRooms(path)
}

func validateRooms(path string) {
	dir := filepath.Dir(path)
	file := filepath.Base(path)

	if _, err := game.LoadWorld(os.DirFS(dir), file); err != nil {
		fmt.Fprintf(os.Stderr, "invalid room file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("valid room file: %s\n", path)
}

func validatePack(path string) {
	pack, err := game.LoadContentPack(os.DirFS(path), "pack.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid content pack manifest: %v\n", err)
		os.Exit(1)
	}

	if _, err := game.LoadWorld(os.DirFS(path), pack.RoomsFile); err != nil {
		fmt.Fprintf(os.Stderr, "invalid content pack rooms: %v\n", err)
		os.Exit(1)
	}
	if pack.EntryRoom != "" {
		rooms, err := game.LoadRooms(os.DirFS(path), pack.RoomsFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid content pack rooms: %v\n", err)
			os.Exit(1)
		}
		if !packRoomExists(rooms, pack.EntryRoom) {
			fmt.Fprintf(os.Stderr, "invalid content pack entry room: %q does not exist\n", pack.EntryRoom)
			os.Exit(1)
		}
	}
	if pack.StoryFile != "" {
		if _, err := game.LoadStoryDocument(os.DirFS(path), pack.StoryFile); err != nil {
			fmt.Fprintf(os.Stderr, "invalid content pack story: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("valid content pack: %s\n", path)
}

func packRoomExists(rooms []game.Room, id string) bool {
	for _, room := range rooms {
		if room.ID == id {
			return true
		}
	}
	return false
}
