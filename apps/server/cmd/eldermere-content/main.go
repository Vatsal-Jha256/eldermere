package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: eldermere-content validate <rooms.json|pack-directory> | validate-all <content-packs-directory> [source-manifest]")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "validate":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: eldermere-content validate <rooms.json|pack-directory>")
			os.Exit(2)
		}
		validatePath(os.Args[2])
	case "validate-all":
		if len(os.Args) > 4 {
			fmt.Fprintln(os.Stderr, "usage: eldermere-content validate-all <content-packs-directory> [source-manifest]")
			os.Exit(2)
		}
		sourceManifest := ""
		if len(os.Args) == 4 {
			sourceManifest = os.Args[3]
		}
		validateAll(os.Args[2], sourceManifest)
	default:
		fmt.Fprintln(os.Stderr, "usage: eldermere-content validate <rooms.json|pack-directory> | validate-all <content-packs-directory> [source-manifest]")
		os.Exit(2)
	}
}

func validatePath(path string) {
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

func validateAll(contentPacksRoot string, sourceManifest string) {
	content, err := game.LoadPackRuntimeContentFromContentPacks(contentPacksRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid content packs: %v\n", err)
		os.Exit(1)
	}

	sourceIDs := map[string]bool(nil)
	if sourceManifest != "" {
		sourceIDs, err = loadSourceIDs(sourceManifest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid source manifest: %v\n", err)
			os.Exit(1)
		}
	}

	if err := game.ValidatePackRuntimeReferences(content, game.NewStarterWorld(), sourceIDs); err != nil {
		fmt.Fprintf(os.Stderr, "invalid runtime content references: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("valid runtime content: %s\n", contentPacksRoot)
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

func loadSourceIDs(path string) (map[string]bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ids := map[string]bool{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "| `") {
			continue
		}
		start := strings.Index(line, "`")
		if start < 0 {
			continue
		}
		rest := line[start+1:]
		end := strings.Index(rest, "`")
		if end <= 0 {
			continue
		}
		ids[rest[:end]] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no source ids found in %s", path)
	}
	return ids, nil
}
