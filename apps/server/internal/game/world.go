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
	Item        *Item             `json:"item,omitempty"`
	Quest       *QuestMarker      `json:"quest,omitempty"`
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

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type QuestMarker struct {
	Start      string `json:"start,omitempty"`
	Incomplete string `json:"incomplete,omitempty"`
	Complete   string `json:"complete,omitempty"`
}

type World struct {
	rooms map[string]Room
}

type Session struct {
	world  World
	roomID string
	party  map[string]bool
	items  map[string]Item
	quest  QuestState
	roll   func(sides int) int
}

type QuestState struct {
	Started   bool
	Completed bool
}

type PersistentState struct {
	RoomID string     `json:"room_id"`
	Party  []string   `json:"party"`
	Items  []Item     `json:"items"`
	Quest  QuestState `json:"quest"`
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
		items:  map[string]Item{},
		quest:  QuestState{},
		roll:   defaultRoller(),
	}
}

func NewSessionWithRoller(world World, roller func(sides int) int) Session {
	session := NewSession(world)
	session.roll = roller
	return session
}

func NewSessionFromState(world World, state PersistentState) Session {
	session := NewSession(world)
	if _, ok := world.rooms[state.RoomID]; ok {
		session.roomID = state.RoomID
	}
	for _, name := range state.Party {
		if strings.TrimSpace(name) != "" {
			session.party[name] = true
		}
	}
	for _, item := range state.Items {
		if strings.TrimSpace(item.ID) != "" {
			session.items[item.ID] = item
		}
	}
	session.quest = state.Quest
	return session
}

func (s *Session) PersistentState() PersistentState {
	party := make([]string, 0, len(s.party))
	for name := range s.party {
		party = append(party, name)
	}
	sortStrings(party)

	items := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	sortItems(items)

	return PersistentState{
		RoomID: s.roomID,
		Party:  party,
		Items:  items,
		Quest:  s.quest,
	}
}

func (s *Session) RoomID() string {
	return s.roomID
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
	case "inventory", "inv", "i":
		return []Event{s.inventoryStatus()}
	case "quest":
		return []Event{s.questStatus()}
	case "take", "get":
		return []Event{s.takeItem()}
	case "exits":
		room := s.currentRoom()
		return []Event{{Type: "system", Text: fmt.Sprintf("Exits: %s", strings.Join(sortedExitNames(room.Exits), ", "))}}
	default:
		return []Event{{Type: "error", Text: fmt.Sprintf("Unknown command `%s`. Try `look`, `quest`, `go north`, `fight`, `recruit`, `take`, `inventory`, `party`, `exits`, or `say hello`.", verb)}}
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
	text := fmt.Sprintf("%s: %s Exits: %s.", room.Name, room.Description, strings.Join(sortedExitNames(room.Exits), ", "))
	if room.Item != nil && !s.hasItem(room.Item.ID) {
		text = fmt.Sprintf("%s You notice %s.", text, room.Item.Description)
	}
	return Event{
		Type: "room",
		Text: text,
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

func (s *Session) takeItem() Event {
	room := s.currentRoom()
	if room.Item == nil {
		return Event{Type: "system", Text: "There is nothing obvious to take here."}
	}
	if s.hasItem(room.Item.ID) {
		return Event{Type: "inventory", Text: fmt.Sprintf("You already have %s.", room.Item.Name)}
	}
	s.items[room.Item.ID] = *room.Item
	return Event{Type: "inventory", Text: fmt.Sprintf("Taken: %s.", room.Item.Name)}
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

func (s *Session) inventoryStatus() Event {
	if len(s.items) == 0 {
		return Event{Type: "inventory", Text: "Inventory: empty."}
	}

	names := make([]string, 0, len(s.items))
	for _, item := range s.items {
		names = append(names, item.Name)
	}
	sortStrings(names)
	return Event{Type: "inventory", Text: fmt.Sprintf("Inventory: %s.", strings.Join(names, ", "))}
}

func (s *Session) questStatus() Event {
	room := s.currentRoom()
	if room.Quest != nil {
		if room.Quest.Start != "" && !s.quest.Started {
			s.quest.Started = true
			return Event{Type: "quest", Text: room.Quest.Start}
		}
		if room.Quest.Complete != "" && s.hasItem("excalibur-fragment") && !s.quest.Completed {
			s.quest.Completed = true
			return Event{Type: "quest", Text: room.Quest.Complete}
		}
		if room.Quest.Incomplete != "" && s.quest.Started && !s.quest.Completed {
			return Event{Type: "quest", Text: room.Quest.Incomplete}
		}
	}

	if s.quest.Completed {
		return Event{Type: "quest", Text: "Quest complete: the stolen Excalibur fragment is back under safer eyes."}
	}
	if s.quest.Started {
		return Event{Type: "quest", Text: "Quest active: find the stolen Excalibur fragment in the under-market route."}
	}
	return Event{Type: "quest", Text: "No quest is active. Try asking around in Lantern Yard."}
}

func (s *Session) hasItem(id string) bool {
	_, ok := s.items[id]
	return ok
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

func sortItems(values []Item) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j].ID < values[i].ID {
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
