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

func TestRemoveBuilding_DropsBuilding(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	if err := AddBuilding(root, "Barn", 10, 10); err != nil {
		t.Fatal(err)
	}
	if err := AddBuilding(root, "Silo", 20, 20); err != nil {
		t.Fatal(err)
	}
	var siloID string
	for _, b := range GetBuildings(root) {
		if b.BuildingType == "Silo" {
			siloID = b.ID
		}
	}

	if err := RemoveBuilding(root, siloID); err != nil {
		t.Fatalf("RemoveBuilding: %v", err)
	}

	bldgs := GetBuildings(root)
	if len(bldgs) != 1 {
		t.Fatalf("expected 1 building after remove, got %d", len(bldgs))
	}
	if bldgs[0].BuildingType != "Barn" {
		t.Errorf("surviving building = %q, want Barn", bldgs[0].BuildingType)
	}
}

func TestRemoveBuilding_RefusesWhenOccupied(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	if err := AddBuilding(root, "Barn", 10, 10); err != nil {
		t.Fatal(err)
	}
	barnID := GetBuildings(root)[0].ID
	if err := AddAnimal(root, barnID, "Cow", "Bessie"); err != nil {
		t.Fatal(err)
	}

	if err := RemoveBuilding(root, barnID); err == nil {
		t.Fatal("expected error removing occupied building")
	}
	if len(GetBuildings(root)) != 1 {
		t.Error("occupied building was removed despite error")
	}
}

func TestRemoveBuilding_UnknownID(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	if err := RemoveBuilding(root, "no-such-building"); err == nil {
		t.Fatal("expected error removing unknown building")
	}
}

func TestChangeBuildingType_RecomputesStructuralFields(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	if err := AddBuilding(root, "Barn", 10, 10); err != nil {
		t.Fatal(err)
	}
	id := GetBuildings(root)[0].ID

	if err := ChangeBuildingType(root, id, "Big Barn", true); err != nil {
		t.Fatalf("ChangeBuildingType: %v", err)
	}

	b := GetBuildings(root)[0]
	if b.BuildingType != "Big Barn" {
		t.Errorf("BuildingType = %q, want Big Barn", b.BuildingType)
	}
	if b.MaxOccupants != 8 {
		t.Errorf("MaxOccupants = %d, want 8 (recomputed for Big Barn)", b.MaxOccupants)
	}
	if b.HayCapacity != 240 {
		t.Errorf("HayCapacity = %d, want 240", b.HayCapacity)
	}
}

func TestChangeBuildingType_WithoutRecomputePreservesFields(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Barn", 10, 10)
	id := GetBuildings(root)[0].ID
	// hand-tune a structural value
	SetBuildingField(root, id, "maxOccupants", "99")

	if err := ChangeBuildingType(root, id, "Big Barn", false); err != nil {
		t.Fatalf("ChangeBuildingType: %v", err)
	}

	b := GetBuildings(root)[0]
	if b.BuildingType != "Big Barn" {
		t.Errorf("BuildingType = %q, want Big Barn", b.BuildingType)
	}
	if b.MaxOccupants != 99 {
		t.Errorf("MaxOccupants = %d, want 99 preserved (no recompute)", b.MaxOccupants)
	}
}

func TestChangeBuildingType_UnknownType(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Barn", 10, 10)
	id := GetBuildings(root)[0].ID
	if err := ChangeBuildingType(root, id, "Castle", true); err == nil {
		t.Fatal("expected error for unknown target type")
	}
}

func TestChangeBuildingType_RefusesCrossHabitatWhenOccupied(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Barn", 10, 10)
	id := GetBuildings(root)[0].ID
	if err := AddAnimal(root, id, "Cow", "Bessie"); err != nil {
		t.Fatal(err)
	}
	if err := ChangeBuildingType(root, id, "Coop", true); err == nil {
		t.Fatal("expected error switching an occupied barn to a coop")
	}
	if GetBuildings(root)[0].BuildingType != "Barn" {
		t.Error("type changed despite refusal")
	}
}

func TestChangeBuildingType_RefusesNonHousingWhenOccupied(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Barn", 10, 10)
	id := GetBuildings(root)[0].ID
	AddAnimal(root, id, "Cow", "Bessie")
	if err := ChangeBuildingType(root, id, "Silo", true); err == nil {
		t.Fatal("expected error switching an occupied barn to a silo")
	}
}

func TestChangeBuildingType_SameHabitatWhenOccupiedUpdatesAnimals(t *testing.T) {
	root, _ := Parse(strings.NewReader(emptyFarmFixture))
	AddBuilding(root, "Barn", 10, 10)
	id := GetBuildings(root)[0].ID
	AddAnimal(root, id, "Cow", "Bessie")

	if err := ChangeBuildingType(root, id, "Big Barn", true); err != nil {
		t.Fatalf("ChangeBuildingType barn upgrade while occupied: %v", err)
	}

	farm := farmNode(root)
	bldg := farm.Child("buildings").Child("Building")
	dict := animalDict(bldg)
	item := dict.ChildrenNamed("item")[0]
	if got := textOf(item, "value/FarmAnimal/buildingTypeILiveIn"); got != "Big Barn" {
		t.Errorf("buildingTypeILiveIn = %q, want Big Barn", got)
	}
}
