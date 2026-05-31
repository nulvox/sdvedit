package save

import (
	"strings"
	"testing"
)

const emptyFarmFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <player><UniqueMultiplayerID>99</UniqueMultiplayerID></player>
  <locations>
    <GameLocation xsi:type="Farm">
      <buildings/>
      <animals/>
    </GameLocation>
  </locations>
</SaveGame>`

func TestAddBuilding_Barn(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))

	if err := AddBuilding(root, "Barn", 10, 20); err != nil {
		t.Fatalf("AddBuilding: %v", err)
	}

	bldgs := GetBuildings(root)
	if len(bldgs) != 1 {
		t.Fatalf("expected 1 building, got %d", len(bldgs))
	}
	b := bldgs[0]
	if b.BuildingType != "Barn" {
		t.Errorf("BuildingType = %q", b.BuildingType)
	}
	if b.TileX != 10 || b.TileY != 20 {
		t.Errorf("position = (%d,%d), want (10,20)", b.TileX, b.TileY)
	}
	if b.TilesWide != 7 || b.TilesHigh != 4 {
		t.Errorf("size = (%d,%d), want (7,4)", b.TilesWide, b.TilesHigh)
	}
	if b.HayCapacity != 240 {
		t.Errorf("HayCapacity = %d, want 240", b.HayCapacity)
	}
	if b.ID == "" {
		t.Error("ID is empty")
	}
}

func TestAddBuilding_Coop(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	if err := AddBuilding(root, "Coop", 5, 5); err != nil {
		t.Fatalf("AddBuilding: %v", err)
	}
	bldgs := GetBuildings(root)
	if bldgs[0].BuildingType != "Coop" {
		t.Errorf("BuildingType = %q", bldgs[0].BuildingType)
	}
}

func TestAddBuilding_HasIndoors(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Barn", 10, 10)

	farm := farmNode(root)
	bldg := farm.Child("buildings").Child("Building")
	if bldg == nil {
		t.Fatal("no building node")
	}
	indoors := bldg.Child("indoors")
	if indoors == nil {
		t.Fatal("Barn has no indoors node")
	}
	animals := indoors.Get("Animals/SerializableDictionaryOfInt64FarmAnimal")
	if animals == nil {
		t.Error("indoors missing Animals/SerializableDictionaryOfInt64FarmAnimal")
	}
}

func TestAddBuilding_SiloNoIndoors(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Silo", 3, 3)

	farm := farmNode(root)
	bldg := farm.Child("buildings").Child("Building")
	if bldg.Child("indoors") != nil {
		t.Error("Silo should not have an indoors node")
	}
}

func TestAddBuilding_UnknownType(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	if err := AddBuilding(root, "Castle", 0, 0); err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestAddBuilding_IDsUnique(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	for i := 0; i < 5; i++ {
		AddBuilding(root, "Silo", i*5, 0)
	}
	bldgs := GetBuildings(root)
	if len(bldgs) != 5 {
		t.Fatalf("expected 5 buildings, got %d", len(bldgs))
	}
	seen := map[string]bool{}
	for _, b := range bldgs {
		if seen[b.ID] {
			t.Errorf("duplicate ID %q", b.ID)
		}
		seen[b.ID] = true
	}
}

func TestAddBuilding_RoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Barn", 15, 15)

	var buf strings.Builder
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	root2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	bldgs := GetBuildings(root2)
	if len(bldgs) != 1 || bldgs[0].BuildingType != "Barn" {
		t.Errorf("round-trip failed: %+v", bldgs)
	}
}

func TestAddBuilding_ThenAddAnimal(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	if err := AddBuilding(root, "Barn", 10, 10); err != nil {
		t.Fatal(err)
	}

	bldgs := GetBuildings(root)
	if len(bldgs) == 0 {
		t.Fatal("no buildings")
	}
	barnID := bldgs[0].ID

	if err := AddAnimal(root, barnID, "Cow", "Bessie"); err != nil {
		t.Fatalf("AddAnimal after AddBuilding: %v", err)
	}
	animals := GetAnimals(root)
	if len(animals) != 1 || animals[0].Name != "Bessie" {
		t.Errorf("animal not found after add: %+v", animals)
	}
}
