package game

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"strings"
	"time"
)

//go:embed content/starter/rooms.json
var starterContent embed.FS

type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Encounter   *Encounter        `json:"encounter,omitempty"`
	Recruitable *Recruitable      `json:"recruitable,omitempty"`
}

type Encounter struct {
	Name string `json:"name"`
	DC   int    `json:"dc"`
	Win  string `json:"win"`
	Lose string `json:"lose"`
}

type Recruitable struct {
	Name    string `json:"name"`
	DC      int    `json:"dc"`
	Success string `json:"success"`
	Failure string `json:"failure"`
}

type World struct {
	rooms map[string]Room
}

type Session struct {
	world  World
	roomID string
	party  map[string]bool
	roll   func(sides int) int
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
	world, err := LoadWorld(starterContent, "content/starter/rooms.json")
	if err != nil {
		panic(fmt.Sprintf("load starter world: %v", err))
	}
	return world
}

func LoadWorld(files fs.FS, path string) (World, error) {
	payload, err := fs.ReadFile(files, path)
	if err != nil {
		return World{}, err
	}

	var document struct {
		Rooms []Room `json:"rooms"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		return World{}, err
	}

	return NewWorld(document.Rooms)
}

func NewWorld(rooms []Room) (World, error) {
	if len(rooms) == 0 {
		return World{}, errors.New("world must include at least one room")
	}

	byID := make(map[string]Room, len(rooms))
	for _, room := range rooms {
		if strings.TrimSpace(room.ID) == "" {
			return World{}, errors.New("room id is required")
		}
		if strings.TrimSpace(room.Name) == "" {
			return World{}, fmt.Errorf("room %q name is required", room.ID)
		}
		if strings.TrimSpace(room.Description) == "" {
			return World{}, fmt.Errorf("room %q description is required", room.ID)
		}
		if _, exists := byID[room.ID]; exists {
			return World{}, fmt.Errorf("duplicate room id %q", room.ID)
		}
		if room.Exits == nil {
			room.Exits = map[string]string{}
		}
		byID[room.ID] = room
	}

	for _, room := range byID {
		for direction, targetID := range room.Exits {
			if strings.TrimSpace(direction) == "" {
				return World{}, fmt.Errorf("room %q has empty exit direction", room.ID)
			}
			if _, ok := byID[targetID]; !ok {
				return World{}, fmt.Errorf("room %q exit %q points to unknown room %q", room.ID, direction, targetID)
			}
		}
	}

	return World{rooms: byID}, nil
}

func NewSession(world World) Session {
	return Session{
		world:  world,
		roomID: "lantern-yard",
		party:  map[string]bool{},
		roll:   defaultRoller(),
	}
}

func NewSessionWithRoller(world World, roller func(sides int) int) Session {
	session := NewSession(world)
	session.roll = roller
	return session
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
	case "fight":
		return []Event{s.fight()}
	case "recruit":
		return []Event{s.recruit()}
	case "party":
		return []Event{s.partyStatus()}
	case "exits":
		room := s.currentRoom()
		return []Event{{Type: "system", Text: fmt.Sprintf("Exits: %s", strings.Join(sortedExitNames(room.Exits), ", "))}}
	default:
		return []Event{{Type: "error", Text: fmt.Sprintf("Unknown command `%s`. Try `look`, `go north`, `go east`, `fight`, `recruit`, `party`, `exits`, or `say hello`.", verb)}}
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

func (s *Session) fight() Event {
	room := s.currentRoom()
	if room.Encounter == nil {
		return Event{Type: "system", Text: "There is nothing here that wants a fight yet."}
	}

	roll := s.roll(20)
	total := roll + 2
	if total >= room.Encounter.DC {
		return Event{
			Type: "fight",
			Text: fmt.Sprintf("Rolled %d + 2 = %d against DC %d. %s", roll, total, room.Encounter.DC, room.Encounter.Win),
		}
	}

	return Event{
		Type: "fight",
		Text: fmt.Sprintf("Rolled %d + 2 = %d against DC %d. %s", roll, total, room.Encounter.DC, room.Encounter.Lose),
	}
}

func (s *Session) recruit() Event {
	room := s.currentRoom()
	if room.Recruitable == nil {
		return Event{Type: "system", Text: "No one here is ready to join you."}
	}

	name := room.Recruitable.Name
	if s.party[name] {
		return Event{Type: "party", Text: fmt.Sprintf("%s is already with you.", name)}
	}

	roll := s.roll(20)
	total := roll + 1
	if total >= room.Recruitable.DC {
		s.party[name] = true
		return Event{
			Type: "party",
			Text: fmt.Sprintf("Rolled %d + 1 = %d against DC %d. %s", roll, total, room.Recruitable.DC, room.Recruitable.Success),
		}
	}

	return Event{
		Type: "party",
		Text: fmt.Sprintf("Rolled %d + 1 = %d against DC %d. %s", roll, total, room.Recruitable.DC, room.Recruitable.Failure),
	}
}

func (s *Session) partyStatus() Event {
	if len(s.party) == 0 {
		return Event{Type: "party", Text: "Party: no companions yet."}
	}

	names := make([]string, 0, len(s.party))
	for name := range s.party {
		names = append(names, name)
	}
	sortStrings(names)
	return Event{Type: "party", Text: fmt.Sprintf("Party: %s.", strings.Join(names, ", "))}
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
	sortStrings(names)
	if len(names) == 0 {
		return []string{"none"}
	}
	return names
}

func sortStrings(values []string) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

func defaultRoller() func(sides int) int {
	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(sides int) int {
		if sides <= 1 {
			return 1
		}
		return source.Intn(sides) + 1
	}
}
