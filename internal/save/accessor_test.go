package save

import (
	"bytes"
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

const playerHealthFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame>
  <player>
    <name>Trent</name>
    <health>50</health>
    <maxHealth>100</maxHealth>
    <stamina>120</stamina>
    <maxStamina>270</maxStamina>
  </player>
</SaveGame>`

func TestGetPlayerStats_CurrentHealthStamina(t *testing.T) {
	root, _ := Parse(strings.NewReader(playerHealthFixture))
	s, err := GetPlayerStats(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.Health != 50 {
		t.Errorf("Health = %d, want 50", s.Health)
	}
	if s.Stamina != 120 {
		t.Errorf("Stamina = %d, want 120", s.Stamina)
	}
}

func TestSetPlayerStats_CurrentHealthStamina_RoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(playerHealthFixture))
	s, _ := GetPlayerStats(root)
	s.Health = 100
	s.Stamina = 270
	SetPlayerStats(root, s)

	s2, _ := GetPlayerStats(root)
	if s2.Health != 100 {
		t.Errorf("Health = %d after patch, want 100", s2.Health)
	}
	if s2.Stamina != 270 {
		t.Errorf("Stamina = %d after patch, want 270", s2.Stamina)
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

func TestAddFriendship_AppendsNewNPC(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))

	if err := AddFriendship(root, "Abigail"); err != nil {
		t.Fatalf("AddFriendship: %v", err)
	}

	entries := GetFriendships(root)
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}
	var abby *FriendshipEntry
	for i := range entries {
		if entries[i].NPC == "Abigail" {
			abby = &entries[i]
		}
	}
	if abby == nil {
		t.Fatal("Abigail not present after AddFriendship")
	}
	if abby.Points != 0 {
		t.Errorf("new NPC Points = %d, want 0", abby.Points)
	}
}

func TestAddFriendship_IdempotentPreservesExisting(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))

	if err := AddFriendship(root, "Lewis"); err != nil {
		t.Fatalf("AddFriendship: %v", err)
	}

	entries := GetFriendships(root)
	count := 0
	for _, e := range entries {
		if e.NPC == "Lewis" {
			count++
			if e.Points != 16 {
				t.Errorf("existing Lewis Points = %d, want 16 (untouched)", e.Points)
			}
		}
	}
	if count != 1 {
		t.Errorf("Lewis appears %d times, want 1 (no duplicate)", count)
	}
}

func TestAddFriendship_ThenSetRoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	if err := AddFriendship(root, "Abigail"); err != nil {
		t.Fatal(err)
	}

	entries := GetFriendships(root)
	for i := range entries {
		if entries[i].NPC == "Abigail" {
			entries[i].Points = 2500
			entries[i].Status = "Dating"
		}
	}
	SetFriendships(root, entries)

	for _, e := range GetFriendships(root) {
		if e.NPC == "Abigail" {
			if e.Points != 2500 {
				t.Errorf("Abigail Points = %d after set, want 2500", e.Points)
			}
			if e.Status != "Dating" {
				t.Errorf("Abigail Status = %q after set, want Dating", e.Status)
			}
		}
	}
}

func TestAddFriendship_RoundTripsThroughXML(t *testing.T) {
	root, _ := Parse(strings.NewReader(friendshipFixture))
	if err := AddFriendship(root, "Sebastian"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	root2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	found := false
	for _, e := range GetFriendships(root2) {
		if e.NPC == "Sebastian" {
			found = true
		}
	}
	if !found {
		t.Error("Sebastian not durable through round-trip")
	}
}

func TestKnownNPCs_NonEmpty(t *testing.T) {
	npcs := KnownNPCs()
	if len(npcs) == 0 {
		t.Fatal("KnownNPCs returned empty list")
	}
	// a few canonical villagers must be present
	want := map[string]bool{"Abigail": false, "Sebastian": false, "Lewis": false}
	for _, n := range npcs {
		if _, ok := want[n]; ok {
			want[n] = true
		}
	}
	for n, present := range want {
		if !present {
			t.Errorf("expected %q in KnownNPCs", n)
		}
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

const petFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <locations>
    <GameLocation xsi:type="Farm">
      <characters>
        <NPC xsi:type="Pet">
          <name>Rex</name>
          <petType>Dog</petType>
          <whichBreed>0</whichBreed>
          <friendshipTowardFarmer>100</friendshipTowardFarmer>
          <timesPet>5</timesPet>
          <Gender>Male</Gender>
        </NPC>
      </characters>
    </GameLocation>
  </locations>
</SaveGame>`

func TestSetPet_TypeAndGender_RoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(petFixture))
	pet := GetPet(root)
	if pet == nil {
		t.Fatal("no pet in fixture")
	}

	updated := *pet
	updated.PetType = "Cat"
	updated.Gender = "Female"
	if err := SetPet(root, updated); err != nil {
		t.Fatalf("SetPet: %v", err)
	}

	got := GetPet(root)
	if got.PetType != "Cat" {
		t.Errorf("PetType = %q, want Cat", got.PetType)
	}
	if got.Gender != "Female" {
		t.Errorf("Gender = %q, want Female", got.Gender)
	}
}

const noPetFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <locations>
    <GameLocation xsi:type="Farm">
      <characters/>
    </GameLocation>
  </locations>
</SaveGame>`

func TestAddPet_CreatesPetWhenNoneExists(t *testing.T) {
	root, _ := Parse(strings.NewReader(noPetFixture))
	if GetPet(root) != nil {
		t.Fatal("fixture should start with no pet")
	}

	if err := AddPet(root, "Cat", "Whiskers", 1); err != nil {
		t.Fatalf("AddPet: %v", err)
	}

	pet := GetPet(root)
	if pet == nil {
		t.Fatal("no pet after AddPet")
	}
	if pet.Name != "Whiskers" {
		t.Errorf("Name = %q, want Whiskers", pet.Name)
	}
	if pet.PetType != "Cat" {
		t.Errorf("PetType = %q, want Cat", pet.PetType)
	}
	if pet.Breed != 1 {
		t.Errorf("Breed = %d, want 1", pet.Breed)
	}
}

func TestAddPet_RefusesDuplicate(t *testing.T) {
	root, _ := Parse(strings.NewReader(petFixture)) // already has Rex
	if err := AddPet(root, "Dog", "Spot", 0); err == nil {
		t.Fatal("expected error adding a second pet")
	}
	if GetPet(root).Name != "Rex" {
		t.Error("existing pet disturbed by duplicate AddPet")
	}
}

func TestAddPet_RoundTripsThroughXML(t *testing.T) {
	root, _ := Parse(strings.NewReader(noPetFixture))
	if err := AddPet(root, "Dog", "Rover", 2); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	root2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	pet := GetPet(root2)
	if pet == nil || pet.Name != "Rover" || pet.PetType != "Dog" || pet.Breed != 2 {
		t.Errorf("pet not durable through round-trip: %+v", pet)
	}
}

func TestAddPet_ThenSetRoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(noPetFixture))
	if err := AddPet(root, "Cat", "Whiskers", 0); err != nil {
		t.Fatal(err)
	}
	pet := GetPet(root)
	pet.Friendship = 750
	pet.TimesPet = 30
	if err := SetPet(root, *pet); err != nil {
		t.Fatalf("SetPet on added pet: %v", err)
	}
	got := GetPet(root)
	if got.Friendship != 750 {
		t.Errorf("Friendship = %d after set, want 750", got.Friendship)
	}
	if got.TimesPet != 30 {
		t.Errorf("TimesPet = %d after set, want 30", got.TimesPet)
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

const inventoryFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <player>
    <name>Trent</name>
    <items>
      <Item xsi:nil="true" />
      <Item xsi:type="Object">
        <name>Stone</name><itemId>390</itemId><stack>50</stack><quality>0</quality>
      </Item>
      <Item xsi:nil="true" />
    </items>
  </player>
</SaveGame>`

func TestAddInventoryItem_FillsEmptySlot(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))

	if err := AddInventoryItem(root, 0, "388", "", 99, 2); err != nil {
		t.Fatalf("AddInventoryItem: %v", err)
	}

	items := GetInventory(root)
	if len(items) != 3 {
		t.Fatalf("inventory slot count changed: got %d, want 3", len(items))
	}
	if items[0].IsNil {
		t.Error("slot 0 still nil after add")
	}
	if items[0].ItemID != "388" {
		t.Errorf("ItemID = %q, want 388", items[0].ItemID)
	}
	if items[0].Name != "Wood" {
		t.Errorf("Name = %q, want catalog default Wood", items[0].Name)
	}
	if items[0].Stack != 99 {
		t.Errorf("Stack = %d, want 99", items[0].Stack)
	}
	if items[0].Quality != 2 {
		t.Errorf("Quality = %d, want 2", items[0].Quality)
	}
	if items[0].XsiType != "Object" {
		t.Errorf("XsiType = %q, want Object", items[0].XsiType)
	}
	// existing item untouched
	if items[1].ItemID != "390" || items[1].Stack != 50 {
		t.Errorf("slot 1 disturbed: %+v", items[1])
	}
	if !items[2].IsNil {
		t.Error("slot 2 should still be nil")
	}
}

func TestAddInventoryItem_RejectsOccupied(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	err := AddInventoryItem(root, 1, "388", "", 1, 0)
	if err == nil {
		t.Fatal("expected error when slot is occupied")
	}
}

func TestAddInventoryItem_OutOfRange(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := AddInventoryItem(root, 99, "388", "", 1, 0); err == nil {
		t.Fatal("expected error for out-of-range slot")
	}
}

func TestAddInventoryItem_CustomNameOverridesCatalog(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := AddInventoryItem(root, 0, "390", "Pet Rock", 1, 0); err != nil {
		t.Fatal(err)
	}
	items := GetInventory(root)
	if items[0].Name != "Pet Rock" {
		t.Errorf("Name = %q, want Pet Rock", items[0].Name)
	}
}

func TestAddInventoryItem_RoundTripsThroughXML(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := AddInventoryItem(root, 0, "388", "", 99, 2); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `<Item xsi:type="Object">`) {
		t.Errorf("serialized output missing Object Item; got: %s", out)
	}
	if strings.Count(out, `<Item xsi:nil="true"`) != 1 {
		t.Errorf("expected exactly one nil Item left; got: %s", out)
	}
	// re-parse to confirm shape is durable
	root2, err := Parse(strings.NewReader(out))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	items := GetInventory(root2)
	if len(items) != 3 || items[0].IsNil || items[0].Stack != 99 {
		t.Errorf("after round-trip, got %+v", items)
	}
}

func TestReplaceInventoryItem_ChangesType(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))

	// slot 1 holds Stone (390); replace with Iridium Ore (386)
	if err := ReplaceInventoryItem(root, 1, "386", "", 5, 2); err != nil {
		t.Fatalf("ReplaceInventoryItem: %v", err)
	}

	items := GetInventory(root)
	if len(items) != 3 {
		t.Fatalf("slot count changed: got %d, want 3", len(items))
	}
	if items[1].ItemID != "386" {
		t.Errorf("ItemID = %q, want 386", items[1].ItemID)
	}
	if items[1].Name != "Iridium Ore" {
		t.Errorf("Name = %q, want catalog default Iridium Ore", items[1].Name)
	}
	if items[1].Stack != 5 {
		t.Errorf("Stack = %d, want 5", items[1].Stack)
	}
	if items[1].Quality != 2 {
		t.Errorf("Quality = %d, want 2", items[1].Quality)
	}
	if items[1].XsiType != "Object" {
		t.Errorf("XsiType = %q, want Object", items[1].XsiType)
	}
}

func TestReplaceInventoryItem_RejectsEmptySlot(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := ReplaceInventoryItem(root, 0, "386", "", 1, 0); err == nil {
		t.Fatal("expected error replacing an empty slot")
	}
}

func TestReplaceInventoryItem_OutOfRange(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := ReplaceInventoryItem(root, 99, "386", "", 1, 0); err == nil {
		t.Fatal("expected error for out-of-range slot")
	}
}

func TestReplaceInventoryItem_RoundTripsThroughXML(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := ReplaceInventoryItem(root, 1, "386", "Custom Ore", 7, 1); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	root2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	items := GetInventory(root2)
	if items[1].ItemID != "386" || items[1].Name != "Custom Ore" || items[1].Stack != 7 {
		t.Errorf("after round-trip, slot 1 = %+v", items[1])
	}
}

func TestClearInventorySlot_EmptiesPopulatedSlot(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))

	if err := ClearInventorySlot(root, 1); err != nil {
		t.Fatalf("ClearInventorySlot: %v", err)
	}

	items := GetInventory(root)
	if len(items) != 3 {
		t.Fatalf("inventory slot count changed: got %d, want 3", len(items))
	}
	if !items[1].IsNil {
		t.Errorf("slot 1 still populated after clear: %+v", items[1])
	}
	if items[1].ItemID != "" {
		t.Errorf("cleared slot retains itemId %q", items[1].ItemID)
	}
}

func TestClearInventorySlot_RejectsEmptySlot(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := ClearInventorySlot(root, 0); err == nil {
		t.Fatal("expected error clearing an already-empty slot")
	}
}

func TestClearInventorySlot_OutOfRange(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := ClearInventorySlot(root, 99); err == nil {
		t.Fatal("expected error for out-of-range slot")
	}
}

func TestClearInventorySlot_RoundTripsThroughXML(t *testing.T) {
	root, _ := Parse(strings.NewReader(inventoryFixture))
	if err := ClearInventorySlot(root, 1); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "<name>Stone</name>") {
		t.Errorf("cleared item content still serialized: %s", out)
	}
	if strings.Count(out, `<Item xsi:nil="true"`) != 3 {
		t.Errorf("expected three nil Items after clear; got: %s", out)
	}
	root2, err := Parse(strings.NewReader(out))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	items := GetInventory(root2)
	if len(items) != 3 || !items[1].IsNil {
		t.Errorf("after round-trip, slot 1 not nil: %+v", items)
	}
}

const recipeFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame>
  <player>
    <cookingRecipes>
      <item><key><string>Fried Egg</string></key><value><int>3</int></value></item>
    </cookingRecipes>
    <craftingRecipes>
      <item><key><string>Chest</string></key><value><int>1</int></value></item>
    </craftingRecipes>
  </player>
</SaveGame>`

func TestAddRecipe_AppendsNewRecipe(t *testing.T) {
	root, _ := Parse(strings.NewReader(recipeFixture))

	if err := AddRecipe(root, "cookingRecipes", "Pizza"); err != nil {
		t.Fatalf("AddRecipe: %v", err)
	}

	recipes := GetCookingRecipes(root)
	var found *RecipeEntry
	for i := range recipes {
		if recipes[i].Name == "Pizza" {
			found = &recipes[i]
		}
	}
	if found == nil {
		t.Fatal("Pizza not present after AddRecipe")
	}
	if found.TimesMade != 0 {
		t.Errorf("new recipe TimesMade = %d, want 0", found.TimesMade)
	}
}

func TestAddRecipe_IdempotentPreservesCount(t *testing.T) {
	root, _ := Parse(strings.NewReader(recipeFixture))

	if err := AddRecipe(root, "cookingRecipes", "Fried Egg"); err != nil {
		t.Fatalf("AddRecipe: %v", err)
	}

	recipes := GetCookingRecipes(root)
	count := 0
	for _, r := range recipes {
		if r.Name == "Fried Egg" {
			count++
			if r.TimesMade != 3 {
				t.Errorf("existing recipe count changed to %d, want 3", r.TimesMade)
			}
		}
	}
	if count != 1 {
		t.Errorf("Fried Egg appears %d times, want 1 (no duplicate)", count)
	}
}

func TestAddRecipe_RoundTripsThroughXML(t *testing.T) {
	root, _ := Parse(strings.NewReader(recipeFixture))
	if err := AddRecipe(root, "craftingRecipes", "Furnace"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	root2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	recipes := GetCraftingRecipes(root2)
	found := false
	for _, r := range recipes {
		if r.Name == "Furnace" {
			found = true
		}
	}
	if !found {
		t.Errorf("Furnace not durable through round-trip: %+v", recipes)
	}
}

func TestLearnAllRecipes_UnlocksWholeCatalog(t *testing.T) {
	root, _ := Parse(strings.NewReader(recipeFixture))

	added := LearnAllRecipes(root, "cookingRecipes")
	if added <= 0 {
		t.Fatalf("LearnAllRecipes added %d, want > 0", added)
	}

	recipes := GetCookingRecipes(root)
	known := map[string]bool{}
	for _, r := range recipes {
		known[r.Name] = true
	}
	for _, name := range KnownCookingRecipes() {
		if !known[name] {
			t.Errorf("recipe %q not learned after LearnAllRecipes", name)
		}
	}
	// pre-existing recipe preserved exactly once
	if !known["Fried Egg"] {
		t.Error("pre-existing Fried Egg missing after LearnAll")
	}
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
