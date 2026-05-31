package save

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const savePath = "/home/this/.config/StardewValley/Saves/trentshire_409715412/trentshire_409715412"

// loadReal opens and parses the actual save file.
// Tests that call this are skipped if the file is absent.
func loadReal(t *testing.T) *Node {
	t.Helper()
	f, err := os.Open(savePath)
	if err != nil {
		t.Skipf("save file not found: %v", err)
	}
	defer f.Close()
	root, err := Parse(f)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return root
}

func TestGetPlayerStats_Fixture(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	s, err := GetPlayerStats(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "Trent" {
		t.Errorf("Name = %q, want Trent", s.Name)
	}
	if s.Money != 500 {
		t.Errorf("Money = %d, want 500", s.Money)
	}
	if s.DeepestMineLevel != 5 {
		t.Errorf("DeepestMineLevel = %d, want 5", s.DeepestMineLevel)
	}
}

func TestSetPlayerStats_RoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	s, _ := GetPlayerStats(root)
	s.Money = 99999
	s.DeepestMineLevel = 120
	s.Name = "Patched"
	SetPlayerStats(root, s)

	s2, _ := GetPlayerStats(root)
	if s2.Money != 99999 {
		t.Errorf("Money = %d after patch", s2.Money)
	}
	if s2.DeepestMineLevel != 120 {
		t.Errorf("DeepestMineLevel = %d after patch", s2.DeepestMineLevel)
	}
	if s2.Name != "Patched" {
		t.Errorf("Name = %q after patch", s2.Name)
	}
}

func TestGetFriendships_Fixture(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	entries := GetFriendships(root)
	if len(entries) != 2 {
		t.Fatalf("len = %d, want 2", len(entries))
	}
	lewis := entries[0]
	if lewis.NPC != "Lewis" {
		t.Errorf("NPC = %q, want Lewis", lewis.NPC)
	}
	if lewis.Points != 16 {
		t.Errorf("Points = %d, want 16", lewis.Points)
	}
	if lewis.Status != "Friendly" {
		t.Errorf("Status = %q, want Friendly", lewis.Status)
	}
}

func TestSetFriendships_RoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	entries := GetFriendships(root)
	entries[0].Points = 2500 // 10 hearts
	entries[0].Status = "Dating"
	SetFriendships(root, entries)

	entries2 := GetFriendships(root)
	if entries2[0].Points != 2500 {
		t.Errorf("Points = %d after patch", entries2[0].Points)
	}
	if entries2[0].Status != "Dating" {
		t.Errorf("Status = %q after patch", entries2[0].Status)
	}
}

func TestGetWorldState_Fixture(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	ws := GetWorldState(root)
	if ws.Season != "spring" {
		t.Errorf("Season = %q, want spring", ws.Season)
	}
	if ws.DayOfMonth != 16 {
		t.Errorf("Day = %d, want 16", ws.DayOfMonth)
	}
	if ws.Year != 1 {
		t.Errorf("Year = %d, want 1", ws.Year)
	}
	if ws.WeatherForTomorrow != "Sun" {
		t.Errorf("Weather = %q, want Sun", ws.WeatherForTomorrow)
	}
	if ws.MineLowestLevel != 5 {
		t.Errorf("MineLevel = %d, want 5", ws.MineLowestLevel)
	}
}

func TestSetWorldState_RoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	ws := GetWorldState(root)
	ws.Season = "winter"
	ws.DayOfMonth = 1
	ws.Year = 3
	ws.WeatherForTomorrow = "Rain"
	ws.MineLowestLevel = 80
	SetWorldState(root, ws)

	ws2 := GetWorldState(root)
	if ws2.Season != "winter" {
		t.Errorf("Season = %q after patch", ws2.Season)
	}
	if ws2.DayOfMonth != 1 {
		t.Errorf("Day = %d after patch", ws2.DayOfMonth)
	}
	if ws2.WeatherForTomorrow != "Rain" {
		t.Errorf("Weather = %q after patch", ws2.WeatherForTomorrow)
	}
	if ws2.MineLowestLevel != 80 {
		t.Errorf("MineLevel = %d after patch", ws2.MineLowestLevel)
	}
}

func TestMailFlags(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	// fixture has no mailReceived, skip if absent
	if root.Get("player/mailReceived") == nil {
		t.Skip("no mailReceived in fixture")
	}
	AddMailFlag(root, "testFlag1")
	AddMailFlag(root, "testFlag2")
	AddMailFlag(root, "testFlag1") // duplicate, should be ignored

	flags := GetMailReceived(root)
	if len(flags) != 2 {
		t.Errorf("flags = %v, want 2 unique", flags)
	}

	RemoveMailFlag(root, "testFlag1")
	flags = GetMailReceived(root)
	if len(flags) != 1 || flags[0] != "testFlag2" {
		t.Errorf("after remove: flags = %v", flags)
	}
}

func TestParseBundle(t *testing.T) {
	b := parseBundle("Pantry/0", "Spring Crops/O 465 20/24 1 0 188 1 0 190 1 0 192 1 0/0///Spring Crops")
	if b.Room != "Pantry" {
		t.Errorf("Room = %q, want Pantry", b.Room)
	}
	if b.BundleID != 0 {
		t.Errorf("BundleID = %d, want 0", b.BundleID)
	}
	if b.Name != "Spring Crops" {
		t.Errorf("Name = %q, want Spring Crops", b.Name)
	}
	if len(b.Items) != 4 {
		t.Errorf("Items = %d, want 4", len(b.Items))
	}
	if b.Items[0].ItemID != "24" {
		t.Errorf("Item[0].ItemID = %q, want 24", b.Items[0].ItemID)
	}
	if b.Items[0].Quantity != 1 {
		t.Errorf("Item[0].Quantity = %d, want 1", b.Items[0].Quantity)
	}
	if b.ItemsNeeded != 0 {
		t.Errorf("ItemsNeeded = %d, want 0", b.ItemsNeeded)
	}
}

// Integration tests against the real save file.

func TestRealSave_PlayerStats(t *testing.T) {
	root := loadReal(t)
	s, err := GetPlayerStats(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name == "" {
		t.Error("player name is empty")
	}
	t.Logf("Player: %q money=%d mine=%d", s.Name, s.Money, s.DeepestMineLevel)
}

func TestRealSave_Friendships(t *testing.T) {
	root := loadReal(t)
	entries := GetFriendships(root)
	if len(entries) == 0 {
		t.Error("no friendship entries found")
	}
	t.Logf("Friendships: %d NPCs", len(entries))
	for _, e := range entries {
		t.Logf("  %s: %d pts (%s)", e.NPC, e.Points, e.Status)
	}
}

func TestRealSave_WorldState(t *testing.T) {
	root := loadReal(t)
	ws := GetWorldState(root)
	if ws.Season == "" {
		t.Error("season is empty")
	}
	t.Logf("World: season=%s day=%d year=%d weather=%s mine=%d",
		ws.Season, ws.DayOfMonth, ws.Year, ws.WeatherForTomorrow, ws.MineLowestLevel)
}

func TestRealSave_Buildings(t *testing.T) {
	root := loadReal(t)
	buildings := GetBuildings(root)
	t.Logf("Buildings: %d", len(buildings))
	for _, b := range buildings {
		t.Logf("  [%s] type=%q at (%d,%d) animals=%d",
			b.ID[:8], b.BuildingType, b.TileX, b.TileY, len(b.Animals))
	}
}

func TestRealSave_Animals(t *testing.T) {
	root := loadReal(t)
	animals := GetAnimals(root)
	t.Logf("Animals: %d", len(animals))
	for _, a := range animals {
		t.Logf("  %s (%s) friend=%d happy=%d", a.Name, a.Type, a.Friendship, a.Happiness)
	}
}

func TestGetAnimals_EmptyIsArray(t *testing.T) {
	// nil slice marshals to JSON null; initialized slice marshals to [].
	// The JS layer does animals.length which throws on null.
	root, _ := Parse(strings.NewReader(friendshipFixture))
	animals := GetAnimals(root)

	b, err := json.Marshal(animals)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "null" {
		t.Error("GetAnimals marshaled to null; JS .length will throw — must be []")
	}
	if string(b) != "[]" {
		t.Errorf("expected [], got %s", b)
	}
}

func TestRealSave_Pet(t *testing.T) {
	root := loadReal(t)
	pet := GetPet(root)
	if pet == nil {
		t.Log("no pet found")
		return
	}
	t.Logf("Pet: %s (%s breed=%d) friendship=%d", pet.Name, pet.PetType, pet.Breed, pet.Friendship)
}

func TestSetPet_RoundTrip(t *testing.T) {
	root := loadReal(t)
	pet := GetPet(root)
	if pet == nil {
		t.Skip("no pet in save")
	}

	updated := *pet
	updated.Name = "Biscuit"
	updated.Friendship = 999
	updated.TimesPet = 42

	if err := SetPet(root, updated); err != nil {
		t.Fatalf("SetPet: %v", err)
	}

	got := GetPet(root)
	if got == nil {
		t.Fatal("pet gone after SetPet")
	}
	if got.Name != "Biscuit" {
		t.Errorf("Name = %q, want Biscuit", got.Name)
	}
	if got.Friendship != 999 {
		t.Errorf("Friendship = %d, want 999", got.Friendship)
	}
	if got.TimesPet != 42 {
		t.Errorf("TimesPet = %d, want 42", got.TimesPet)
	}
}

func TestRealSave_Bundles(t *testing.T) {
	root := loadReal(t)
	bundles := GetBundles(root)
	if len(bundles) == 0 {
		t.Error("no bundles found")
	}
	t.Logf("Bundles: %d", len(bundles))
}

func TestRealSave_RoundTrip(t *testing.T) {
	f, err := os.Open(savePath)
	if err != nil {
		t.Skip("save file not found")
	}
	defer f.Close()

	root, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	// Mutate and re-parse; check value survives
	orig, _ := GetPlayerStats(root)
	modified := orig
	modified.Money = orig.Money + 1
	SetPlayerStats(root, modified)

	s2, _ := GetPlayerStats(root)
	if s2.Money != orig.Money+1 {
		t.Errorf("money not patched: got %d, want %d", s2.Money, orig.Money+1)
	}
}
