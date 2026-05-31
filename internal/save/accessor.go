package save

import (
	"fmt"
	"strconv"
	"strings"
)

// --- Player ---

type PlayerStats struct {
	Name             string `json:"name"`
	FarmName         string `json:"farmName"`
	Gender           string `json:"gender"`
	Money            int    `json:"money"`
	MaxHealth        int    `json:"maxHealth"`
	MaxStamina       int    `json:"maxStamina"`
	DeepestMineLevel int    `json:"deepestMineLevel"`
	// ExperiencePoints: [Farming, Fishing, Foraging, Mining, Combat, Luck]
	ExperiencePoints [6]int `json:"experiencePoints"`
	// SkillLevels: same index order
	SkillLevels [6]int `json:"skillLevels"`
}

func GetPlayerStats(root *Node) (PlayerStats, error) {
	p := root.Child("player")
	if p == nil {
		return PlayerStats{}, fmt.Errorf("no player node")
	}
	s := PlayerStats{
		Name:             textOf(p, "name"),
		FarmName:         textOf(root, "player/farmName"),
		Gender:           textOf(p, "Gender"),
		Money:            intOf(p, "money"),
		MaxHealth:        intOf(p, "maxHealth"),
		MaxStamina:       intOf(p, "maxStamina"),
		DeepestMineLevel: intOf(p, "deepestMineLevel"),
	}
	expNode := p.Child("experiencePoints")
	if expNode != nil {
		ints := expNode.ChildrenNamed("int")
		for i := 0; i < 6 && i < len(ints); i++ {
			s.ExperiencePoints[i], _ = strconv.Atoi(ints[i].Text)
		}
	}
	skillNames := []string{"farmingLevel", "fishingLevel", "foragingLevel", "miningLevel", "combatLevel", "luckLevel"}
	for i, sn := range skillNames {
		s.SkillLevels[i] = intOf(p, sn)
	}
	return s, nil
}

func SetPlayerStats(root *Node, s PlayerStats) error {
	p := root.Child("player")
	if p == nil {
		return fmt.Errorf("no player node")
	}
	setLeaf(p, "name", s.Name)
	setLeaf(p, "farmName", s.FarmName)
	setLeaf(p, "Gender", s.Gender)
	setLeaf(p, "money", strconv.Itoa(s.Money))
	setLeaf(p, "maxHealth", strconv.Itoa(s.MaxHealth))
	setLeaf(p, "maxStamina", strconv.Itoa(s.MaxStamina))
	setLeaf(p, "deepestMineLevel", strconv.Itoa(s.DeepestMineLevel))
	expNode := p.Child("experiencePoints")
	if expNode != nil {
		ints := expNode.ChildrenNamed("int")
		for i := 0; i < 6 && i < len(ints); i++ {
			ints[i].Text = strconv.Itoa(s.ExperiencePoints[i])
		}
	}
	skillNames := []string{"farmingLevel", "fishingLevel", "foragingLevel", "miningLevel", "combatLevel", "luckLevel"}
	for i, sn := range skillNames {
		setLeaf(p, sn, strconv.Itoa(s.SkillLevels[i]))
	}
	return nil
}

// --- Friendships ---

type FriendshipEntry struct {
	NPC              string `json:"npc"`
	Points           int    `json:"points"`
	GiftsThisWeek    int    `json:"giftsThisWeek"`
	GiftsToday       int    `json:"giftsToday"`
	TalkedToToday    bool   `json:"talkedToToday"`
	ProposalRejected bool   `json:"proposalRejected"`
	Status           string `json:"status"`
	RoommateMarriage bool   `json:"roommateMarriage"`
}

func GetFriendships(root *Node) []FriendshipEntry {
	fd := root.Get("player/friendshipData")
	if fd == nil {
		return []FriendshipEntry{}
	}
	out := []FriendshipEntry{}
	for _, item := range fd.ChildrenNamed("item") {
		npc := item.Get("key/string")
		f := item.Get("value/Friendship")
		if npc == nil || f == nil {
			continue
		}
		out = append(out, FriendshipEntry{
			NPC:              npc.Text,
			Points:           intOf(f, "Points"),
			GiftsThisWeek:    intOf(f, "GiftsThisWeek"),
			GiftsToday:       intOf(f, "GiftsToday"),
			TalkedToToday:    boolOf(f, "TalkedToToday"),
			ProposalRejected: boolOf(f, "ProposalRejected"),
			Status:           textOf(f, "Status"),
			RoommateMarriage: boolOf(f, "RoommateMarriage"),
		})
	}
	return out
}

func SetFriendships(root *Node, entries []FriendshipEntry) {
	fd := root.Get("player/friendshipData")
	if fd == nil {
		return
	}
	// build lookup by NPC name for fast patching
	idx := map[string]*Node{}
	for _, item := range fd.ChildrenNamed("item") {
		npc := item.Get("key/string")
		f := item.Get("value/Friendship")
		if npc != nil && f != nil {
			idx[npc.Text] = f
		}
	}
	for _, e := range entries {
		f, ok := idx[e.NPC]
		if !ok {
			continue
		}
		setLeaf(f, "Points", strconv.Itoa(e.Points))
		setLeaf(f, "GiftsThisWeek", strconv.Itoa(e.GiftsThisWeek))
		setLeaf(f, "GiftsToday", strconv.Itoa(e.GiftsToday))
		setLeaf(f, "TalkedToToday", strconv.FormatBool(e.TalkedToToday))
		setLeaf(f, "ProposalRejected", strconv.FormatBool(e.ProposalRejected))
		setLeaf(f, "Status", e.Status)
		setLeaf(f, "RoommateMarriage", strconv.FormatBool(e.RoommateMarriage))
	}
}

// --- World State ---

type WorldState struct {
	Season              string  `json:"season"`
	DayOfMonth          int     `json:"dayOfMonth"`
	Year                int     `json:"year"`
	DailyLuck           float64 `json:"dailyLuck"`
	WeatherForTomorrow  string  `json:"weatherForTomorrow"`
	MineLowestLevel     int     `json:"mineLowestLevel"`
	MineLowestForOrder  int     `json:"mineLowestForOrder"`
}

func GetWorldState(root *Node) WorldState {
	luck, _ := strconv.ParseFloat(textOf(root, "dailyLuck"), 64)
	return WorldState{
		Season:             textOf(root, "currentSeason"),
		DayOfMonth:         intOf(root, "dayOfMonth"),
		Year:               intOf(root, "year"),
		DailyLuck:          luck,
		WeatherForTomorrow: textOf(root, "weatherForTomorrow"),
		MineLowestLevel:    intOf(root, "mine_lowestLevelReached"),
		MineLowestForOrder: intOf(root, "mine_lowestLevelReachedForOrder"),
	}
}

func SetWorldState(root *Node, ws WorldState) {
	setLeaf(root, "currentSeason", ws.Season)
	setLeaf(root, "dayOfMonth", strconv.Itoa(ws.DayOfMonth))
	setLeaf(root, "year", strconv.Itoa(ws.Year))
	setLeaf(root, "dailyLuck", strconv.FormatFloat(ws.DailyLuck, 'f', 3, 64))
	setLeaf(root, "weatherForTomorrow", ws.WeatherForTomorrow)
	setLeaf(root, "mine_lowestLevelReached", strconv.Itoa(ws.MineLowestLevel))
	setLeaf(root, "mine_lowestLevelReachedForOrder", strconv.Itoa(ws.MineLowestForOrder))
}

// --- Buildings ---

type BuildingPaintColor struct {
	ColorName   string `json:"colorName"`
	H1          int    `json:"h1"`
	S1          int    `json:"s1"`
	L1          int    `json:"l1"`
	Default1    bool   `json:"default1"`
	H2          int    `json:"h2"`
	S2          int    `json:"s2"`
	L2          int    `json:"l2"`
	Default2    bool   `json:"default2"`
	H3          int    `json:"h3"`
	S3          int    `json:"s3"`
	L3          int    `json:"l3"`
	Default3    bool   `json:"default3"`
}

type BuildingEntry struct {
	ID                   string             `json:"id"`
	BuildingType         string             `json:"buildingType"`
	XsiType              string             `json:"xsiType"`
	TileX                int                `json:"tileX"`
	TileY                int                `json:"tileY"`
	TilesWide            int                `json:"tilesWide"`
	TilesHigh            int                `json:"tilesHigh"`
	DaysOfConstruction   int                `json:"daysOfConstruction"`
	DaysUntilUpgrade     int                `json:"daysUntilUpgrade"`
	Paint                BuildingPaintColor `json:"paint"`
	HayCapacity          int                `json:"hayCapacity"`
	MaxOccupants         int                `json:"maxOccupants"`
	CurrentOccupants     int                `json:"currentOccupants"`
	Animals              []AnimalEntry      `json:"animals"`
}

func GetBuildings(root *Node) []BuildingEntry {
	farm := farmNode(root)
	if farm == nil {
		return []BuildingEntry{}
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return []BuildingEntry{}
	}
	out := []BuildingEntry{}
	for _, b := range bldgs.ChildrenNamed("Building") {
		entry := buildingFromNode(b)
		out = append(out, entry)
	}
	return out
}

func buildingFromNode(b *Node) BuildingEntry {
	paint := BuildingPaintColor{}
	pc := b.Child("buildingPaintColor")
	if pc != nil {
		paint.Default1 = boolOf(pc, "Color1Default/boolean")
		paint.H1 = intOf(pc, "Color1Hue/int")
		paint.S1 = intOf(pc, "Color1Saturation/int")
		paint.L1 = intOf(pc, "Color1Lightness/int")
		paint.Default2 = boolOf(pc, "Color2Default/boolean")
		paint.H2 = intOf(pc, "Color2Hue/int")
		paint.S2 = intOf(pc, "Color2Saturation/int")
		paint.L2 = intOf(pc, "Color2Lightness/int")
		paint.Default3 = boolOf(pc, "Color3Default/boolean")
		paint.H3 = intOf(pc, "Color3Hue/int")
		paint.S3 = intOf(pc, "Color3Saturation/int")
		paint.L3 = intOf(pc, "Color3Lightness/int")
	}
	entry := BuildingEntry{
		ID:                 textOf(b, "id"),
		BuildingType:       textOf(b, "buildingType"),
		XsiType:            b.Attr("type"),
		TileX:              intOf(b, "tileX"),
		TileY:              intOf(b, "tileY"),
		TilesWide:          intOf(b, "tilesWide"),
		TilesHigh:          intOf(b, "tilesHigh"),
		DaysOfConstruction: intOf(b, "daysOfConstructionLeft"),
		DaysUntilUpgrade:   intOf(b, "daysUntilUpgrade"),
		HayCapacity:        intOf(b, "hayCapacity"),
		MaxOccupants:       intOf(b, "maxOccupants"),
		CurrentOccupants:   intOf(b, "currentOccupants"),
		Paint:              paint,
	}
	// animals inside indoors
	entry.Animals = animalsFromBuilding(b)
	return entry
}

// SetBuildingField patches a single field on a building identified by ID.
func SetBuildingField(root *Node, id, field, value string) error {
	farm := farmNode(root)
	if farm == nil {
		return fmt.Errorf("farm not found")
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return fmt.Errorf("buildings not found")
	}
	for _, b := range bldgs.ChildrenNamed("Building") {
		if textOf(b, "id") == id {
			return setLeaf(b, field, value)
		}
	}
	return fmt.Errorf("building %q not found", id)
}

// --- Animals ---

type AnimalEntry struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	Age                  int    `json:"age"`
	DaysOwned            int    `json:"daysOwned"`
	Friendship           int    `json:"friendship"`
	Happiness            int    `json:"happiness"`
	Fullness             int    `json:"fullness"`
	WasPet               bool   `json:"wasPet"`
	ProduceQuality       int    `json:"produceQuality"`
	BuildingID           string `json:"buildingId"`
}

// GetAnimals returns all farm animals found inside building indoors.
func GetAnimals(root *Node) []AnimalEntry {
	farm := farmNode(root)
	if farm == nil {
		return []AnimalEntry{}
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return []AnimalEntry{}
	}
	out := []AnimalEntry{}
	for _, b := range bldgs.ChildrenNamed("Building") {
		bid := textOf(b, "id")
		for _, a := range animalsFromBuilding(b) {
			a.BuildingID = bid
			out = append(out, a)
		}
	}
	return out
}

func animalsFromBuilding(b *Node) []AnimalEntry {
	// Animals are in indoors/Animals/SerializableDictionaryOfInt64FarmAnimal/item/value/FarmAnimal
	indoors := b.Child("indoors")
	if indoors == nil {
		return []AnimalEntry{}
	}
	animals := indoors.Child("Animals")
	if animals == nil {
		return []AnimalEntry{}
	}
	dict := animals.Child("SerializableDictionaryOfInt64FarmAnimal")
	if dict == nil {
		return []AnimalEntry{}
	}
	out := []AnimalEntry{}
	for _, item := range dict.ChildrenNamed("item") {
		keyNode := item.Get("key/long")
		valNode := item.Get("value/FarmAnimal")
		if valNode == nil {
			continue
		}
		id := ""
		if keyNode != nil {
			id = keyNode.Text
		}
		out = append(out, AnimalEntry{
			ID:             id,
			Name:           textOf(valNode, "name"),
			Type:           textOf(valNode, "type"),
			Age:            intOf(valNode, "age"),
			DaysOwned:      intOf(valNode, "daysOwned"),
			Friendship:     intOf(valNode, "friendshipTowardFarmer"),
			Happiness:      intOf(valNode, "happiness"),
			Fullness:       intOf(valNode, "fullness"),
			WasPet:         boolOf(valNode, "wasPet"),
			ProduceQuality: intOf(valNode, "produceQuality"),
		})
	}
	return out
}

// SetAnimalField patches a field on a specific animal by ID (long key).
func SetAnimalField(root *Node, animalID, field, value string) error {
	farm := farmNode(root)
	if farm == nil {
		return fmt.Errorf("farm not found")
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return fmt.Errorf("buildings not found")
	}
	for _, b := range bldgs.ChildrenNamed("Building") {
		indoors := b.Child("indoors")
		if indoors == nil {
			continue
		}
		animals := indoors.Child("Animals")
		if animals == nil {
			continue
		}
		dict := animals.Child("SerializableDictionaryOfInt64FarmAnimal")
		if dict == nil {
			continue
		}
		for _, item := range dict.ChildrenNamed("item") {
			key := item.Get("key/long")
			if key == nil || key.Text != animalID {
				continue
			}
			valNode := item.Get("value/FarmAnimal")
			if valNode == nil {
				continue
			}
			return setLeaf(valNode, field, value)
		}
	}
	return fmt.Errorf("animal %q not found", animalID)
}

// --- Pet ---

type PetEntry struct {
	Name               string `json:"name"`
	PetType            string `json:"petType"`
	Breed              int    `json:"breed"`
	Friendship         int    `json:"friendship"`
	TimesPet           int    `json:"timesPet"`
	Gender             string `json:"gender"`
}

func GetPet(root *Node) *PetEntry {
	farm := farmNode(root)
	if farm == nil {
		return nil
	}
	for _, ch := range farm.ChildrenNamed("characters") {
		for _, npc := range ch.ChildrenNamed("NPC") {
			if npc.Attr("type") == "Pet" {
				return &PetEntry{
					Name:       textOf(npc, "name"),
					PetType:    textOf(npc, "petType"),
					Breed:      intOf(npc, "whichBreed"),
					Friendship: intOf(npc, "friendshipTowardFarmer"),
					TimesPet:   intOf(npc, "timesPet"),
					Gender:     textOf(npc, "Gender"),
				}
			}
		}
	}
	return nil
}

func SetPet(root *Node, e PetEntry) error {
	farm := farmNode(root)
	if farm == nil {
		return fmt.Errorf("farm not found")
	}
	for _, ch := range farm.ChildrenNamed("characters") {
		for _, npc := range ch.ChildrenNamed("NPC") {
			if npc.Attr("type") != "Pet" {
				continue
			}
			setLeaf(npc, "name", e.Name)
			setLeaf(npc, "friendshipTowardFarmer", strconv.Itoa(e.Friendship))
			setLeaf(npc, "timesPet", strconv.Itoa(e.TimesPet))
			setLeaf(npc, "whichBreed", strconv.Itoa(e.Breed))
			return nil
		}
	}
	return fmt.Errorf("pet not found")
}

// --- Recipes ---

type RecipeEntry struct {
	Name        string `json:"name"`
	TimesMade   int    `json:"timesMade"`
}

func getRecipes(root *Node, section string) []RecipeEntry {
	node := root.Get("player/" + section)
	if node == nil {
		return []RecipeEntry{}
	}
	out := []RecipeEntry{}
	for _, item := range node.ChildrenNamed("item") {
		name := item.Get("key/string")
		val := item.Get("value/int")
		if name == nil {
			continue
		}
		times := 0
		if val != nil {
			times, _ = strconv.Atoi(val.Text)
		}
		out = append(out, RecipeEntry{Name: name.Text, TimesMade: times})
	}
	return out
}

func GetCookingRecipes(root *Node) []RecipeEntry { return getRecipes(root, "cookingRecipes") }
func GetCraftingRecipes(root *Node) []RecipeEntry { return getRecipes(root, "craftingRecipes") }

func setRecipes(root *Node, section string, entries []RecipeEntry) {
	node := root.Get("player/" + section)
	if node == nil {
		return
	}
	idx := map[string]*Node{}
	for _, item := range node.ChildrenNamed("item") {
		name := item.Get("key/string")
		val := item.Get("value/int")
		if name != nil && val != nil {
			idx[name.Text] = val
		}
	}
	for _, e := range entries {
		if v, ok := idx[e.Name]; ok {
			v.Text = strconv.Itoa(e.TimesMade)
		}
	}
}

func SetCookingRecipes(root *Node, entries []RecipeEntry)  { setRecipes(root, "cookingRecipes", entries) }
func SetCraftingRecipes(root *Node, entries []RecipeEntry) { setRecipes(root, "craftingRecipes", entries) }

// --- Mail ---

func GetMailReceived(root *Node) []string {
	node := root.Get("player/mailReceived")
	if node == nil {
		return []string{}
	}
	out := []string{}
	for _, s := range node.ChildrenNamed("string") {
		out = append(out, s.Text)
	}
	return out
}

func AddMailFlag(root *Node, flag string) {
	node := root.Get("player/mailReceived")
	if node == nil {
		return
	}
	// check not already present
	for _, s := range node.ChildrenNamed("string") {
		if s.Text == flag {
			return
		}
	}
	node.Children = append(node.Children, &Node{Name: "string", Text: flag})
}

func RemoveMailFlag(root *Node, flag string) {
	node := root.Get("player/mailReceived")
	if node == nil {
		return
	}
	filtered := node.Children[:0]
	for _, c := range node.Children {
		if !(c.Name == "string" && c.Text == flag) {
			filtered = append(filtered, c)
		}
	}
	node.Children = filtered
}

// --- Quests ---

type QuestEntry struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	QuestType   string `json:"questType"`
	Completed   bool   `json:"completed"`
	DailyQuest  bool   `json:"dailyQuest"`
	Destroy     bool   `json:"destroy"`
	MoneyReward int    `json:"moneyReward"`
	DaysLeft    int    `json:"daysLeft"`
}

func GetQuests(root *Node) []QuestEntry {
	ql := root.Get("player/questLog")
	if ql == nil {
		return []QuestEntry{}
	}
	out := []QuestEntry{}
	for _, q := range ql.ChildrenNamed("Quest") {
		out = append(out, QuestEntry{
			ID:          intOf(q, "id"),
			Title:       textOf(q, "questTitle"),
			Description: textOf(q, "_questDescription"),
			QuestType:   q.Attr("type"),
			Completed:   boolOf(q, "completed"),
			DailyQuest:  boolOf(q, "dailyQuest"),
			Destroy:     boolOf(q, "destroy"),
			MoneyReward: intOf(q, "moneyReward"),
			DaysLeft:    intOf(q, "daysLeft"),
		})
	}
	return out
}

// --- Inventory ---

type ItemEntry struct {
	Slot     int    `json:"slot"`
	ItemID   string `json:"itemId"`
	Name     string `json:"name"`
	Stack    int    `json:"stack"`
	Quality  int    `json:"quality"`
	IsNil    bool   `json:"isNil"`
	Category int    `json:"category"`
	XsiType  string `json:"xsiType"`
}

func GetInventory(root *Node) []ItemEntry {
	items := root.Get("player/items")
	if items == nil {
		return []ItemEntry{}
	}
	out := []ItemEntry{}
	for i, item := range items.ChildrenNamed("Item") {
		if item.Attr("nil") == "true" {
			out = append(out, ItemEntry{Slot: i, IsNil: true})
			continue
		}
		out = append(out, ItemEntry{
			Slot:     i,
			ItemID:   textOf(item, "itemId"),
			Name:     textOf(item, "name"),
			Stack:    intOf(item, "stack"),
			Quality:  intOf(item, "quality"),
			Category: intOf(item, "category"),
			XsiType:  item.Attr("type"),
		})
	}
	return out
}

// SetInventoryItem updates stack and quality for the item at a given slot.
func SetInventoryItem(root *Node, slot int, stack, quality int) error {
	items := root.Get("player/items")
	if items == nil {
		return fmt.Errorf("items not found")
	}
	children := items.ChildrenNamed("Item")
	if slot < 0 || slot >= len(children) {
		return fmt.Errorf("slot %d out of range", slot)
	}
	item := children[slot]
	if item.Attr("nil") == "true" {
		return fmt.Errorf("slot %d is empty", slot)
	}
	setLeaf(item, "stack", strconv.Itoa(stack))
	setLeaf(item, "quality", strconv.Itoa(quality))
	return nil
}

// --- Bundles ---

type BundleItem struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
	Quality  int    `json:"quality"`
}

type BundleEntry struct {
	Key          string       `json:"key"`   // "Pantry/0"
	Room         string       `json:"room"`
	BundleID     int          `json:"bundleId"`
	Name         string       `json:"name"`
	RewardRaw    string       `json:"rewardRaw"`
	Items        []BundleItem `json:"items"`
	ItemsNeeded  int          `json:"itemsNeeded"`
}

func GetBundles(root *Node) []BundleEntry {
	bd := root.Child("bundleData")
	if bd == nil {
		return []BundleEntry{}
	}
	out := []BundleEntry{}
	for _, item := range bd.ChildrenNamed("item") {
		key := item.Get("key/string")
		val := item.Get("value/string")
		if key == nil || val == nil {
			continue
		}
		out = append(out, parseBundle(key.Text, val.Text))
	}
	return out
}

// parseBundle decodes the bundle value string format:
// "Name/Reward/item1 qty1 qual1 item2 qty2 qual2.../itemsNeeded/color/..."
func parseBundle(key, raw string) BundleEntry {
	parts := strings.Split(key, "/")
	room := parts[0]
	bid := 0
	if len(parts) > 1 {
		bid, _ = strconv.Atoi(parts[1])
	}

	segments := strings.Split(raw, "/")
	name := ""
	if len(segments) > 0 {
		name = segments[0]
	}
	rewardRaw := ""
	if len(segments) > 1 {
		rewardRaw = segments[1]
	}

	var items []BundleItem
	itemsNeeded := -1
	if len(segments) > 2 {
		itemStr := segments[2]
		fields := strings.Fields(itemStr)
		for i := 0; i+2 < len(fields); i += 3 {
			qty, _ := strconv.Atoi(fields[i+1])
			qual, _ := strconv.Atoi(fields[i+2])
			items = append(items, BundleItem{
				ItemID:   fields[i],
				Quantity: qty,
				Quality:  qual,
			})
		}
	}
	if len(segments) > 3 {
		itemsNeeded, _ = strconv.Atoi(segments[3])
	}

	return BundleEntry{
		Key:         key,
		Room:        room,
		BundleID:    bid,
		Name:        name,
		RewardRaw:   rewardRaw,
		Items:       items,
		ItemsNeeded: itemsNeeded,
	}
}

// --- helpers ---

func farmNode(root *Node) *Node {
	locs := root.Child("locations")
	if locs == nil {
		return nil
	}
	for _, gl := range locs.ChildrenNamed("GameLocation") {
		if gl.Attr("type") == "Farm" {
			return gl
		}
	}
	return nil
}

func textOf(n *Node, path string) string {
	child := n.Get(path)
	if child == nil {
		return ""
	}
	return child.Text
}

func intOf(n *Node, path string) int {
	v, _ := strconv.Atoi(textOf(n, path))
	return v
}

func boolOf(n *Node, path string) bool {
	return strings.EqualFold(textOf(n, path), "true")
}

func setLeaf(n *Node, path string, value string) error {
	child := n.Get(path)
	if child == nil {
		return fmt.Errorf("path %q not found", path)
	}
	if len(child.Children) > 0 {
		return fmt.Errorf("path %q has children", path)
	}
	child.Text = value
	return nil
}
