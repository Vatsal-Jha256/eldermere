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
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Exits       map[string]string    `json:"exits"`
	GatedExits  map[string]GatedExit `json:"gated_exits,omitempty"`
	Encounter   *Encounter           `json:"encounter,omitempty"`
	Recruitable *Recruitable         `json:"recruitable,omitempty"`
	Item        *Item                `json:"item,omitempty"`
	Quest       *QuestMarker         `json:"quest,omitempty"`
	Atmosphere  Atmosphere           `json:"atmosphere,omitempty"`
}

type Encounter struct {
	Name           string         `json:"name"`
	DC             int            `json:"dc"`
	Modifier       int            `json:"modifier,omitempty"`
	RollMode       string         `json:"roll_mode,omitempty"`
	Win            string         `json:"win"`
	Lose           string         `json:"lose"`
	CritWin        string         `json:"crit_win,omitempty"`
	CritLose       string         `json:"crit_lose,omitempty"`
	FactionEffects map[string]int `json:"faction_effects,omitempty"`
}

type Recruitable struct {
	Name     string `json:"name"`
	DC       int    `json:"dc"`
	Modifier int    `json:"modifier,omitempty"`
	RollMode string `json:"roll_mode,omitempty"`
	Success  string `json:"success"`
	Failure  string `json:"failure"`
	CritWin  string `json:"crit_win,omitempty"`
	CritLose string `json:"crit_lose,omitempty"`
}

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type QuestMarker struct {
	Start         string   `json:"start,omitempty"`
	StartVariants []string `json:"start_variants,omitempty"`
	Incomplete    string   `json:"incomplete,omitempty"`
	Complete      string   `json:"complete,omitempty"`
}

type GatedExit struct {
	Target       string `json:"target"`
	RequiresItem string `json:"requires_item"`
	LockedText   string `json:"locked_text"`
}

type Atmosphere struct {
	Palette   string   `json:"palette,omitempty"`
	Weather   string   `json:"weather,omitempty"`
	MythLayer string   `json:"myth_layer,omitempty"`
	Motifs    []string `json:"motifs,omitempty"`
}

type World struct {
	rooms         map[string]Room
	stories       map[string]StoryArc
	storySeedTags []string
	packEntries   map[string]string
}

type Session struct {
	world    World
	roomID   string
	party    map[string]bool
	items    map[string]Item
	quest    QuestState
	story    StoryState
	factions map[string]int
	roll     func(sides int) int
}

type QuestState struct {
	Started   bool
	Completed bool
	Variant   string
}

type PersistentState struct {
	RoomID   string         `json:"room_id"`
	Party    []string       `json:"party"`
	Items    []Item         `json:"items"`
	Quest    QuestState     `json:"quest"`
	Story    StoryState     `json:"story"`
	Factions map[string]int `json:"factions"`
}

type StoryState struct {
	ActiveArcID     string   `json:"active_arc_id"`
	StepIndex       int      `json:"step_index"`
	CompletedArcIDs []string `json:"completed_arc_ids"`
	Tags            []string `json:"tags"`
	VariantTag      string   `json:"variant_tag"`
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
	Atmosphere  Atmosphere        `json:"atmosphere"`
}

func NewStarterWorld() World {
	world, err := LoadWorld(starterContent, "content/starter/rooms.json")
	if err != nil {
		panic(fmt.Sprintf("load starter world: %v", err))
	}
	return world
}

func LoadWorld(files fs.FS, path string) (World, error) {
	rooms, err := LoadRooms(files, path)
	if err != nil {
		return World{}, err
	}
	return NewWorld(rooms)
}

func LoadRooms(files fs.FS, path string) ([]Room, error) {
	payload, err := fs.ReadFile(files, path)
	if err != nil {
		return nil, err
	}

	var document struct {
		Rooms []Room `json:"rooms"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		return nil, err
	}

	return document.Rooms, nil
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
		if room.GatedExits == nil {
			room.GatedExits = map[string]GatedExit{}
		}
		if room.Encounter != nil {
			if err := validateChallenge("encounter", room.ID, room.Encounter.DC, room.Encounter.RollMode); err != nil {
				return World{}, err
			}
		}
		if room.Recruitable != nil {
			if err := validateChallenge("recruitable", room.ID, room.Recruitable.DC, room.Recruitable.RollMode); err != nil {
				return World{}, err
			}
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
		for direction, exit := range room.GatedExits {
			if strings.TrimSpace(direction) == "" {
				return World{}, fmt.Errorf("room %q has empty gated exit direction", room.ID)
			}
			if strings.TrimSpace(exit.Target) == "" {
				return World{}, fmt.Errorf("room %q gated exit %q target is required", room.ID, direction)
			}
			if _, ok := byID[exit.Target]; !ok {
				return World{}, fmt.Errorf("room %q gated exit %q points to unknown room %q", room.ID, direction, exit.Target)
			}
			if strings.TrimSpace(exit.RequiresItem) == "" {
				return World{}, fmt.Errorf("room %q gated exit %q requires_item is required", room.ID, direction)
			}
		}
	}

	return World{rooms: byID, stories: map[string]StoryArc{}, packEntries: map[string]string{}}, nil
}

func validateChallenge(kind string, roomID string, dc int, mode string) error {
	if dc < 2 || dc > 30 {
		return fmt.Errorf("room %q %s dc must be between 2 and 30", roomID, kind)
	}
	if !validRollMode(mode) {
		return fmt.Errorf("room %q %s roll_mode must be empty, normal, advantage, or disadvantage", roomID, kind)
	}
	return nil
}

func (w World) WithRooms(rooms []Room) (World, error) {
	if len(rooms) == 0 {
		return w, nil
	}

	combined := make([]Room, 0, len(w.rooms)+len(rooms))
	for _, room := range w.rooms {
		combined = append(combined, room)
	}
	combined = append(combined, rooms...)

	next, err := NewWorld(combined)
	if err != nil {
		return World{}, err
	}
	next.stories = w.stories
	next.storySeedTags = w.storySeedTags
	next.packEntries = copyStringMap(w.packEntries)
	return next, nil
}

func (w World) WithStoryArcs(arcs []StoryArc) (World, error) {
	if len(arcs) == 0 {
		return w, nil
	}
	if err := ValidateStoryDocument(StoryDocument{Arcs: arcs}); err != nil {
		return World{}, err
	}

	stories := make(map[string]StoryArc, len(w.stories)+len(arcs))
	for id, arc := range w.stories {
		stories[id] = arc
	}
	for _, arc := range arcs {
		stories[arc.ID] = arc
	}
	w.stories = stories
	return w, nil
}

func (w World) WithStoryContent(content StoryContent) (World, error) {
	withStories, err := w.WithStoryArcs(content.Arcs)
	if err != nil {
		return World{}, err
	}
	withStories.storySeedTags = appendStoryTags(withStories.storySeedTags, content.Tags...)
	return withStories, nil
}

func (w World) WithPackRuntimeContent(content PackRuntimeContent) (World, error) {
	withRooms, err := w.WithRooms(content.Rooms)
	if err != nil {
		return World{}, err
	}
	withStories, err := withRooms.WithStoryContent(content.Stories)
	if err != nil {
		return World{}, err
	}
	if withStories.packEntries == nil {
		withStories.packEntries = map[string]string{}
	}
	for packID, roomID := range content.Entries {
		withStories.packEntries[packID] = roomID
	}
	return withStories, nil
}

func NewSession(world World) Session {
	return Session{
		world:  world,
		roomID: "lantern-yard",
		party:  map[string]bool{},
		items:  map[string]Item{},
		quest:  QuestState{},
		story: StoryState{
			Tags: appendStoryTags(nil, world.storySeedTags...),
		},
		factions: map[string]int{},
		roll:     defaultRoller(),
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
	session.story = state.Story
	session.story.Tags = appendStoryTags(session.story.Tags, world.storySeedTags...)
	for name, value := range state.Factions {
		if strings.TrimSpace(name) != "" {
			session.factions[name] = value
		}
	}
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
		RoomID:   s.roomID,
		Party:    party,
		Items:    items,
		Quest:    s.quest,
		Story:    normalizeStoryState(s.story),
		Factions: copyFactions(s.factions),
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
	case "help", "commands", "?":
		return []Event{s.helpStatus(args)}
	case "look", "l":
		return []Event{s.look()}
	case "go", "move", "walk":
		if len(args) == 0 {
			return []Event{{Type: "error", Text: "Go where? Try `go north`."}}
		}
		return s.goDirection(strings.ToLower(args[0]))
	case "travel":
		if len(args) == 0 {
			return []Event{{Type: "error", Text: "Travel where? Try `travel arthurian-core`."}}
		}
		return s.travelToPack(strings.ToLower(args[0]))
	case "say", "talk":
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
	case "factions", "reputation", "rep":
		return []Event{s.factionStatus()}
	case "map":
		return []Event{s.mapStatus()}
	case "inventory", "inv", "i":
		return []Event{s.inventoryStatus()}
	case "quest":
		return []Event{s.questStatus()}
	case "story", "stories", "arcs":
		return []Event{s.storyStatus(args)}
	case "take", "get":
		return []Event{s.takeItem()}
	case "exits":
		room := s.currentRoom()
		return []Event{{Type: "system", Text: fmt.Sprintf("Exits: %s", strings.Join(s.visibleExitNames(room), ", "))}}
	default:
		return []Event{{Type: "error", Text: fmt.Sprintf("Unknown command `%s`. Try `help`.", verb)}}
	}
}

func (s *Session) helpStatus(args []string) Event {
	if len(args) == 0 {
		return Event{
			Type: "help",
			Text: "Help topics: movement, story, combat, inventory, social, world. Commands: look, exits, go <direction>, travel <pack-id>, quest, story, fight, recruit, take, inventory, party, factions, map, say <text>. Try `help story`.",
		}
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "movement", "move", "go", "travel":
		return Event{Type: "help", Text: "Movement: `look` shows the current room, `exits` lists visible exits, `go north` moves through exits, and `travel arthurian-core` enters a loaded content pack entry room."}
	case "story", "quest", "plot":
		return Event{Type: "help", Text: "Story: `quest` runs the starter quest. `story` lists loaded arcs, `story eligible` shows playable arcs, `story locked` explains blocked arcs, `story <id>` inspects one, `story start <id>` begins it, `story status` shows the current room/objective, `story next` advances when requirements are met, and `story tags` lists earned branch tags."}
	case "combat", "fight", "recruit":
		return Event{Type: "help", Text: "Combat and companions: `fight` rolls against the current room encounter, party members add a small bonus, `recruit` rolls to add a room companion, and `party` lists recruited allies."}
	case "inventory", "items", "take":
		return Event{Type: "help", Text: "Inventory: `take` collects the current room item, `inventory` lists carried relics/items, and some items unlock hidden map routes."}
	case "social", "say", "talk", "multiplayer":
		return Event{Type: "help", Text: "Social: `say hello` or `talk hello` sends local speech to players in the same room. Room presence and recent room events are shared over the WebSocket session."}
	case "world", "factions", "map":
		return Event{Type: "help", Text: "World state: `factions` shows reputation changes from encounters and story steps. `map` shows hidden or gated exits from the current room."}
	default:
		return Event{Type: "help", Text: fmt.Sprintf("Unknown help topic `%s`. Try `help`, `help story`, `help movement`, `help combat`, `help inventory`, `help social`, or `help world`.", args[0])}
	}
}

func (s *Session) goDirection(direction string) []Event {
	room := s.currentRoom()
	nextID, ok := room.Exits[direction]
	if !ok {
		gated, gatedOK := room.GatedExits[direction]
		if !gatedOK {
			return []Event{{Type: "error", Text: fmt.Sprintf("No exit %s from %s.", direction, room.Name)}}
		}
		if !s.hasItem(gated.RequiresItem) {
			text := gated.LockedText
			if text == "" {
				text = fmt.Sprintf("You need %s to go %s.", gated.RequiresItem, direction)
			}
			return []Event{{Type: "error", Text: text}}
		}
		nextID = gated.Target
	}

	s.roomID = nextID
	return []Event{
		{Type: "move", Text: fmt.Sprintf("You go %s.", direction)},
		s.look(),
	}
}

func (s *Session) travelToPack(packID string) []Event {
	roomID, ok := s.world.packEntries[packID]
	if !ok {
		packs := make([]string, 0, len(s.world.packEntries))
		for id := range s.world.packEntries {
			packs = append(packs, id)
		}
		sortStrings(packs)
		if len(packs) == 0 {
			return []Event{{Type: "error", Text: "No travel destinations are loaded yet."}}
		}
		return []Event{{Type: "error", Text: fmt.Sprintf("Unknown travel destination `%s`. Known packs: %s.", packID, strings.Join(packs, ", "))}}
	}
	if _, exists := s.world.rooms[roomID]; !exists {
		return []Event{{Type: "error", Text: fmt.Sprintf("Travel destination `%s` points to missing room `%s`.", packID, roomID)}}
	}
	s.roomID = roomID
	return []Event{
		{Type: "move", Text: fmt.Sprintf("You travel to %s.", packID)},
		s.look(),
	}
}

func (s *Session) look() Event {
	room := s.currentRoom()
	text := fmt.Sprintf("%s: %s Exits: %s.", room.Name, room.Description, strings.Join(s.visibleExitNames(room), ", "))
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
			Exits:       s.visibleExits(room),
			Atmosphere:  room.Atmosphere,
		},
	}
}

func (s *Session) fight() Event {
	room := s.currentRoom()
	if room.Encounter == nil {
		return Event{Type: "system", Text: "There is nothing here that wants a fight yet."}
	}

	partyBonus := s.partyBonus()
	modifier := room.Encounter.Modifier + 2 + partyBonus
	result := resolveCheck(s.roll, room.Encounter.DC, modifier, room.Encounter.RollMode)
	if result.Success {
		for faction, delta := range room.Encounter.FactionEffects {
			s.factions[faction] += delta
		}
		text := room.Encounter.Win
		if result.Critical == "critical success" && room.Encounter.CritWin != "" {
			text = room.Encounter.CritWin
		}
		return Event{
			Type: "fight",
			Text: fmt.Sprintf("%s. %s", formatCheck(result), text),
		}
	}

	text := room.Encounter.Lose
	if result.Critical == "critical failure" && room.Encounter.CritLose != "" {
		text = room.Encounter.CritLose
	}
	return Event{
		Type: "fight",
		Text: fmt.Sprintf("%s. %s", formatCheck(result), text),
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

	result := resolveCheck(s.roll, room.Recruitable.DC, room.Recruitable.Modifier+1, room.Recruitable.RollMode)
	if result.Success {
		s.party[name] = true
		text := room.Recruitable.Success
		if result.Critical == "critical success" && room.Recruitable.CritWin != "" {
			text = room.Recruitable.CritWin
		}
		return Event{
			Type: "party",
			Text: fmt.Sprintf("%s. %s", formatCheck(result), text),
		}
	}

	text := room.Recruitable.Failure
	if result.Critical == "critical failure" && room.Recruitable.CritLose != "" {
		text = room.Recruitable.CritLose
	}
	return Event{
		Type: "party",
		Text: fmt.Sprintf("%s. %s", formatCheck(result), text),
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

func (s *Session) factionStatus() Event {
	if len(s.factions) == 0 {
		return Event{Type: "factions", Text: "Factions: no reputation changes yet."}
	}

	names := make([]string, 0, len(s.factions))
	for name := range s.factions {
		names = append(names, name)
	}
	sortStrings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s %+d", name, s.factions[name]))
	}
	return Event{Type: "factions", Text: fmt.Sprintf("Factions: %s.", strings.Join(parts, ", "))}
}

func (s *Session) mapStatus() Event {
	room := s.currentRoom()
	gated := make([]string, 0, len(room.GatedExits))
	for direction, exit := range room.GatedExits {
		if s.hasItem(exit.RequiresItem) {
			gated = append(gated, fmt.Sprintf("%s -> %s", direction, exit.Target))
		} else {
			gated = append(gated, fmt.Sprintf("%s locked by %s", direction, exit.RequiresItem))
		}
	}
	sortStrings(gated)
	if len(gated) == 0 {
		return Event{Type: "map", Text: "Map: no hidden or gated routes from here."}
	}
	return Event{Type: "map", Text: fmt.Sprintf("Map: %s.", strings.Join(gated, ", "))}
}

func (s *Session) questStatus() Event {
	room := s.currentRoom()
	if room.Quest != nil {
		if room.Quest.Start != "" && !s.quest.Started {
			s.quest.Started = true
			s.quest.Variant = s.chooseQuestVariant(room.Quest)
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
		if s.quest.Variant != "" {
			return Event{Type: "quest", Text: fmt.Sprintf("Quest active: %s", s.quest.Variant)}
		}
		return Event{Type: "quest", Text: "Quest active: find the stolen Excalibur fragment in the under-market route."}
	}
	return Event{Type: "quest", Text: "No quest is active. Try asking around in Lantern Yard."}
}

func (s *Session) storyStatus(args []string) Event {
	if len(s.world.stories) == 0 {
		return Event{Type: "story", Text: "No story arcs are loaded yet."}
	}

	if len(args) > 0 {
		action := strings.ToLower(strings.TrimSpace(args[0]))
		switch action {
		case "start":
			if len(args) < 2 {
				return Event{Type: "story", Text: "Start which story arc? Try `story start sword-test`."}
			}
			return s.startStoryArc(strings.ToLower(strings.TrimSpace(args[1])))
		case "next", "advance":
			return s.advanceStoryArc()
		case "status", "active":
			return s.activeStoryStatus()
		case "tags":
			return s.storyTagsStatus()
		case "eligible", "available", "open":
			return s.storyEligibilityStatus(false)
		case "locked", "locks":
			return s.storyEligibilityStatus(true)
		}

		id := strings.ToLower(strings.TrimSpace(args[0]))
		arc, ok := s.world.stories[id]
		if !ok {
			return Event{Type: "story", Text: fmt.Sprintf("Story arc `%s` is not loaded. Try `story` to list known arcs or `story start sword-test` to begin one.", id)}
		}
		stepTitles := make([]string, 0, len(arc.Steps))
		for _, step := range arc.Steps {
			stepTitles = append(stepTitles, step.Title)
		}
		return Event{
			Type: "story",
			Text: fmt.Sprintf("%s [%s]: %s Sources: %s. Steps: %s.", arc.Title, arc.Kind, arc.Summary, strings.Join(arc.SourceIDs, ", "), strings.Join(stepTitles, " -> ")),
		}
	}

	mainIDs := make([]string, 0, len(s.world.stories))
	sideIDs := make([]string, 0, len(s.world.stories))
	for id, arc := range s.world.stories {
		if arc.Kind == "main" {
			mainIDs = append(mainIDs, id)
			continue
		}
		sideIDs = append(sideIDs, id)
	}
	sortStrings(mainIDs)
	sortStrings(sideIDs)

	return Event{
		Type: "story",
		Text: fmt.Sprintf("Story arcs loaded. Main: %s. Side: %s. Try `story eligible`, `story locked`, `story sword-test`, or `story start sword-test`.", strings.Join(mainIDs, ", "), strings.Join(sideIDs, ", ")),
	}
}

func (s *Session) startStoryArc(id string) Event {
	arc, ok := s.world.stories[id]
	if !ok {
		return Event{Type: "story", Text: fmt.Sprintf("Story arc `%s` is not loaded. Try `story` to list known arcs.", id)}
	}
	if storyContains(s.story.CompletedArcIDs, id) {
		return Event{Type: "story", Text: fmt.Sprintf("Story arc `%s` is already complete. You can inspect it with `story %s`.", id, id)}
	}
	missing := missingStoryTags(s.story.Tags, arc.RequiredTags)
	missingFactions := s.missingFactionRequirements(arc.RequiredFactions)
	if len(missing) > 0 || len(missingFactions) > 0 {
		return Event{Type: "story", Text: fmt.Sprintf("Story arc `%s` is locked. %s", id, formatStoryLockReason(missing, missingFactions))}
	}

	var variant string
	if len(arc.VariationTags) > 0 {
		index := s.roll(len(arc.VariationTags)) - 1
		if index < 0 || index >= len(arc.VariationTags) {
			index = 0
		}
		variant = arc.VariationTags[index]
	}

	s.story.ActiveArcID = id
	s.story.StepIndex = 0
	s.story.VariantTag = variant
	step := arc.Steps[0]
	text := fmt.Sprintf("Story started: %s. Step 1/%d - %s", arc.Title, len(arc.Steps), formatStoryStep(step))
	if variant != "" {
		text = fmt.Sprintf("%s Variant: %s.", text, variant)
	}
	return Event{Type: "story", Text: text}
}

func (s *Session) advanceStoryArc() Event {
	if s.story.ActiveArcID == "" {
		return Event{Type: "story", Text: "No story arc is active. Try `story start sword-test`."}
	}
	arc, ok := s.world.stories[s.story.ActiveArcID]
	if !ok {
		s.story.ActiveArcID = ""
		s.story.StepIndex = 0
		s.story.VariantTag = ""
		return Event{Type: "story", Text: "The active story arc is no longer loaded, so progress was cleared."}
	}

	if s.story.StepIndex < 0 {
		s.story.StepIndex = 0
	}
	if s.story.StepIndex >= len(arc.Steps) {
		return s.completeStoryArc(arc)
	}

	step := arc.Steps[s.story.StepIndex]
	if strings.TrimSpace(step.RoomHint) != "" && step.RoomHint != s.roomID {
		return Event{
			Type: "story",
			Text: fmt.Sprintf("Story step needs room `%s`; you are in `%s`. Travel or move there before advancing.", step.RoomHint, s.roomID),
		}
	}
	s.story.Tags = appendStoryTags(s.story.Tags, step.OutcomeTags...)
	for faction, delta := range step.FactionEffects {
		if strings.TrimSpace(faction) != "" {
			s.factions[faction] += delta
		}
	}
	if s.story.StepIndex < len(arc.Steps)-1 {
		s.story.StepIndex++
		next := arc.Steps[s.story.StepIndex]
		return Event{
			Type: "story",
			Text: fmt.Sprintf("Story advanced: %s. Step %d/%d - %s", arc.Title, s.story.StepIndex+1, len(arc.Steps), formatStoryStep(next)),
		}
	}

	return s.completeStoryArc(arc)
}

func (s *Session) completeStoryArc(arc StoryArc) Event {
	s.story.Tags = appendStoryTags(s.story.Tags, arc.AddsTags...)
	if s.story.VariantTag != "" {
		s.story.Tags = appendStoryTags(s.story.Tags, s.story.VariantTag)
	}
	s.story.CompletedArcIDs = appendStoryTags(s.story.CompletedArcIDs, arc.ID)
	s.story.ActiveArcID = ""
	s.story.StepIndex = 0
	s.story.VariantTag = ""
	return Event{Type: "story", Text: fmt.Sprintf("Story complete: %s. Tags gained: %s.", arc.Title, strings.Join(s.story.Tags, ", "))}
}

func (s *Session) activeStoryStatus() Event {
	if s.story.ActiveArcID == "" {
		completed := append([]string{}, s.story.CompletedArcIDs...)
		sortStrings(completed)
		if len(completed) == 0 {
			return Event{Type: "story", Text: "No story arc is active and no story arcs are complete."}
		}
		return Event{Type: "story", Text: fmt.Sprintf("No story arc is active. Completed: %s.", strings.Join(completed, ", "))}
	}

	arc, ok := s.world.stories[s.story.ActiveArcID]
	if !ok {
		return Event{Type: "story", Text: fmt.Sprintf("Active story arc `%s` is not loaded.", s.story.ActiveArcID)}
	}
	index := s.story.StepIndex
	if index < 0 {
		index = 0
	}
	if index >= len(arc.Steps) {
		index = len(arc.Steps) - 1
	}
	step := arc.Steps[index]
	text := fmt.Sprintf("Story active: %s. Step %d/%d - %s", arc.Title, index+1, len(arc.Steps), formatStoryStep(step))
	if s.story.VariantTag != "" {
		text = fmt.Sprintf("%s Variant: %s.", text, s.story.VariantTag)
	}
	return Event{Type: "story", Text: text}
}

func (s *Session) storyTagsStatus() Event {
	tags := appendStoryTags(nil, s.story.Tags...)
	if len(tags) == 0 {
		return Event{Type: "story", Text: "Story tags: none."}
	}
	return Event{Type: "story", Text: fmt.Sprintf("Story tags: %s.", strings.Join(tags, ", "))}
}

func (s *Session) storyEligibilityStatus(lockedOnly bool) Event {
	ids := make([]string, 0, len(s.world.stories))
	for id := range s.world.stories {
		ids = append(ids, id)
	}
	sortStrings(ids)

	parts := []string{}
	for _, id := range ids {
		arc := s.world.stories[id]
		if storyContains(s.story.CompletedArcIDs, id) {
			continue
		}
		missingTags := missingStoryTags(s.story.Tags, arc.RequiredTags)
		missingFactions := s.missingFactionRequirements(arc.RequiredFactions)
		locked := len(missingTags) > 0 || len(missingFactions) > 0
		if lockedOnly {
			if locked {
				parts = append(parts, fmt.Sprintf("%s (%s)", id, formatStoryLockReason(missingTags, missingFactions)))
			}
			continue
		}
		if !locked {
			parts = append(parts, id)
		}
	}

	if lockedOnly {
		if len(parts) == 0 {
			return Event{Type: "story", Text: "Locked story arcs: none."}
		}
		return Event{Type: "story", Text: fmt.Sprintf("Locked story arcs: %s.", strings.Join(parts, "; "))}
	}
	if len(parts) == 0 {
		return Event{Type: "story", Text: "Eligible story arcs: none."}
	}
	return Event{Type: "story", Text: fmt.Sprintf("Eligible story arcs: %s.", strings.Join(parts, ", "))}
}

func formatStoryStep(step StoryStep) string {
	parts := []string{fmt.Sprintf("%s: %s", step.Title, step.Objective)}
	if strings.TrimSpace(step.RoomHint) != "" {
		parts = append(parts, fmt.Sprintf("Room: %s", step.RoomHint))
	}
	if len(step.Commands) > 0 {
		parts = append(parts, fmt.Sprintf("Try: %s", strings.Join(step.Commands, ", ")))
	}
	return strings.Join(parts, ". ")
}

func (s *Session) missingFactionRequirements(required map[string]int) []string {
	if len(required) == 0 {
		return nil
	}
	missing := []string{}
	for faction, requiredValue := range required {
		faction = strings.TrimSpace(faction)
		if faction == "" {
			continue
		}
		current := s.factions[faction]
		if current < requiredValue {
			missing = append(missing, fmt.Sprintf("%s %+d (current %+d)", faction, requiredValue, current))
		}
	}
	sortStrings(missing)
	return missing
}

func formatStoryLockReason(missingTags []string, missingFactions []string) string {
	parts := []string{}
	if len(missingTags) > 0 {
		parts = append(parts, fmt.Sprintf("Missing tags: %s", strings.Join(missingTags, ", ")))
	}
	if len(missingFactions) > 0 {
		parts = append(parts, fmt.Sprintf("Missing factions: %s", strings.Join(missingFactions, ", ")))
	}
	return fmt.Sprintf("%s.", strings.Join(parts, ". "))
}

func (s *Session) hasItem(id string) bool {
	_, ok := s.items[id]
	return ok
}

func (s *Session) chooseQuestVariant(marker *QuestMarker) string {
	if len(marker.StartVariants) == 0 {
		return ""
	}
	index := s.roll(len(marker.StartVariants)) - 1
	if index < 0 || index >= len(marker.StartVariants) {
		index = 0
	}
	return marker.StartVariants[index]
}

func (s *Session) partyBonus() int {
	if len(s.party) > 3 {
		return 3
	}
	return len(s.party)
}

func (s *Session) visibleExitNames(room Room) []string {
	exits := s.visibleExits(room)
	names := make([]string, 0, len(exits))
	for direction := range exits {
		names = append(names, direction)
	}
	sortStrings(names)
	if len(names) == 0 {
		return []string{"none"}
	}
	return names
}

func (s *Session) visibleExits(room Room) map[string]string {
	exits := make(map[string]string, len(room.Exits)+len(room.GatedExits))
	for direction, target := range room.Exits {
		exits[direction] = target
	}
	for direction, exit := range room.GatedExits {
		if s.hasItem(exit.RequiresItem) {
			exits[direction] = exit.Target
		}
	}
	return exits
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

func copyFactions(values map[string]int) map[string]int {
	copied := make(map[string]int, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func copyStringMap(values map[string]string) map[string]string {
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func normalizeStoryState(state StoryState) StoryState {
	state.CompletedArcIDs = appendStoryTags(nil, state.CompletedArcIDs...)
	state.Tags = appendStoryTags(nil, state.Tags...)
	return state
}

func appendStoryTags(existing []string, incoming ...string) []string {
	seen := map[string]bool{}
	tags := make([]string, 0, len(existing)+len(incoming))
	for _, tag := range append(existing, incoming...) {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	sortStrings(tags)
	return tags
}

func storyContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func missingStoryTags(existing []string, required []string) []string {
	missing := []string{}
	for _, tag := range required {
		tag = strings.TrimSpace(tag)
		if tag == "" || storyContains(existing, tag) {
			continue
		}
		missing = append(missing, tag)
	}
	sortStrings(missing)
	return missing
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
