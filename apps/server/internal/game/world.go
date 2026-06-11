package game

import (
	"fmt"
	"strings"
)

type Room struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string
}

type World struct {
	rooms map[string]Room
}

type Session struct {
	world  World
	roomID string
}

type Event struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Room *View  `json:"room,omitempty"`
}

type View struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

func NewStarterWorld() World {
	return World{
		rooms: map[string]Room{
			"lantern-yard": {
				ID:          "lantern-yard",
				Name:        "Lantern Yard",
				Description: "The rain over Camelot tastes faintly of iron. Squires trade rumors beside a shrine of old river-stone.",
				Exits: map[string]string{
					"north": "old-bridge",
					"east":  "market-under",
				},
			},
			"old-bridge": {
				ID:          "old-bridge",
				Name:        "Old Bridge",
				Description: "A bridge older than the crown leans over black water. Someone has scratched a Greek oath into one stone.",
				Exits: map[string]string{
					"south": "lantern-yard",
				},
			},
			"market-under": {
				ID:          "market-under",
				Name:        "Market Under",
				Description: "Below the respectable stalls, charm-sellers auction debts, curses, and maps that disagree with themselves.",
				Exits: map[string]string{
					"west": "lantern-yard",
				},
			},
		},
	}
}

func NewSession(world World) Session {
	return Session{
		world:  world,
		roomID: "lantern-yard",
	}
}

func (s *Session) Welcome() []Event {
	return []Event{
		{Type: "system", Text: "Connected to Eldermere."},
		s.look(),
	}
}

func (s *Session) Handle(input string) []Event {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return []Event{{Type: "error", Text: "Enter a command."}}
	}

	parts := strings.Fields(trimmed)
	verb := strings.ToLower(parts[0])
	args := parts[1:]

	switch verb {
	case "look", "l":
		return []Event{s.look()}
	case "go", "move", "walk":
		if len(args) == 0 {
			return []Event{{Type: "error", Text: "Go where? Try `go north`."}}
		}
		return s.goDirection(strings.ToLower(args[0]))
	case "say":
		if len(args) == 0 {
			return []Event{{Type: "error", Text: "Say what?"}}
		}
		return []Event{{Type: "say", Text: fmt.Sprintf("You say, %q", strings.Join(args, " "))}}
	case "exits":
		room := s.currentRoom()
		return []Event{{Type: "system", Text: fmt.Sprintf("Exits: %s", strings.Join(sortedExitNames(room.Exits), ", "))}}
	default:
		return []Event{{Type: "error", Text: fmt.Sprintf("Unknown command `%s`. Try `look`, `go north`, `go east`, `exits`, or `say hello`.", verb)}}
	}
}

func (s *Session) goDirection(direction string) []Event {
	room := s.currentRoom()
	nextID, ok := room.Exits[direction]
	if !ok {
		return []Event{{Type: "error", Text: fmt.Sprintf("No exit %s from %s.", direction, room.Name)}}
	}

	s.roomID = nextID
	return []Event{
		{Type: "move", Text: fmt.Sprintf("You go %s.", direction)},
		s.look(),
	}
}

func (s *Session) look() Event {
	room := s.currentRoom()
	return Event{
		Type: "room",
		Text: fmt.Sprintf("%s: %s Exits: %s.", room.Name, room.Description, strings.Join(sortedExitNames(room.Exits), ", ")),
		Room: &View{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			Exits:       room.Exits,
		},
	}
}

func (s *Session) currentRoom() Room {
	room, ok := s.world.rooms[s.roomID]
	if !ok {
		return s.world.rooms["lantern-yard"]
	}
	return room
}

func sortedExitNames(exits map[string]string) []string {
	names := make([]string, 0, len(exits))
	for name := range exits {
		names = append(names, name)
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	if len(names) == 0 {
		return []string{"none"}
	}
	return names
}
