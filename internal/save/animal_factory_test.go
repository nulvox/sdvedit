package save

import (
	"strings"
	"testing"
)

// minimal save with a Barn building that has an empty Animals dict
const barnFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <player>
    <UniqueMultiplayerID>12345678</UniqueMultiplayerID>
  </player>
  <locations>
    <GameLocation xsi:type="Farm">
      <buildings>
        <Building>
          <id>barn-building-id</id>
          <buildingType>Barn</buildingType>
          <indoors>
            <Animals>
              <SerializableDictionaryOfInt64FarmAnimal/>
            </Animals>
          </indoors>
        </Building>
        <Building>
          <id>coop-building-id</id>
          <buildingType>Coop</buildingType>
          <indoors>
            <Animals>
              <SerializableDictionaryOfInt64FarmAnimal/>
            </Animals>
          </indoors>
        </Building>
      </buildings>
      <animals/>
    </GameLocation>
  </locations>
</SaveGame>`

func TestAddAnimal_Cow(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))

	if err := AddAnimal(root, "barn-building-id", "Cow", "Bessie"); err != nil {
		t.Fatalf("AddAnimal: %v", err)
	}

	animals := GetAnimals(root)
	if len(animals) != 1 {
		t.Fatalf("expected 1 animal, got %d", len(animals))
	}
	a := animals[0]
	if a.Name != "Bessie" {
		t.Errorf("Name = %q, want Bessie", a.Name)
	}
	if a.Type != "Cow" {
		t.Errorf("Type = %q, want Cow", a.Type)
	}
	if a.Happiness != 255 {
		t.Errorf("Happiness = %d, want 255", a.Happiness)
	}
	if a.Fullness != 255 {
		t.Errorf("Fullness = %d, want 255", a.Fullness)
	}
	if a.BuildingID != "barn-building-id" {
		t.Errorf("BuildingID = %q", a.BuildingID)
	}
}

func TestAddAnimal_Chicken(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))

	if err := AddAnimal(root, "coop-building-id", "Chicken", "Henny"); err != nil {
		t.Fatalf("AddAnimal: %v", err)
	}

	animals := GetAnimals(root)
	if len(animals) != 1 {
		t.Fatalf("expected 1 animal, got %d", len(animals))
	}
	if animals[0].Name != "Henny" {
		t.Errorf("Name = %q", animals[0].Name)
	}
}

func TestAddAnimal_MultipleAnimals(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))

	names := []string{"Bessie", "Daisy", "Milky"}
	for _, n := range names {
		if err := AddAnimal(root, "barn-building-id", "Cow", n); err != nil {
			t.Fatalf("AddAnimal %q: %v", n, err)
		}
	}

	animals := GetAnimals(root)
	if len(animals) != 3 {
		t.Fatalf("expected 3 animals, got %d", len(animals))
	}
	got := map[string]bool{}
	for _, a := range animals {
		got[a.Name] = true
	}
	for _, n := range names {
		if !got[n] {
			t.Errorf("missing animal %q", n)
		}
	}
}

func TestAddAnimal_WrongBuildingType(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))

	// Cow in a Coop should fail
	if err := AddAnimal(root, "coop-building-id", "Cow", "Bessie"); err == nil {
		t.Error("expected error placing barn animal in coop, got nil")
	}
	// Chicken in a Barn should fail
	if err := AddAnimal(root, "barn-building-id", "Chicken", "Henny"); err == nil {
		t.Error("expected error placing coop animal in barn, got nil")
	}
}

func TestAddAnimal_UnknownType(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))

	if err := AddAnimal(root, "barn-building-id", "Dragon", "Puff"); err == nil {
		t.Error("expected error for unknown animal type, got nil")
	}
}

func TestAddAnimal_IDsAreUnique(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))

	for i := 0; i < 10; i++ {
		AddAnimal(root, "barn-building-id", "Cow", "Animal")
	}

	animals := GetAnimals(root)
	if len(animals) != 10 {
		t.Fatalf("expected 10 animals, got %d", len(animals))
	}
	ids := map[string]bool{}
	for _, a := range animals {
		if ids[a.ID] {
			t.Errorf("duplicate animal ID %q", a.ID)
		}
		ids[a.ID] = true
	}
}

func TestAddAnimal_RoundTrip(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))
	AddAnimal(root, "barn-building-id", "Cow", "Bessie")

	// serialize and re-parse
	var buf strings.Builder
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	root2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatal(err)
	}
	animals := GetAnimals(root2)
	if len(animals) != 1 || animals[0].Name != "Bessie" {
		t.Errorf("round-trip failed: %+v", animals)
	}
}

func TestSetAnimalField_Rename(t *testing.T) {
	root, _ := Parse(strings.NewReader(barnFixture))
	if err := AddAnimal(root, "barn-building-id", "Cow", "Bessie"); err != nil {
		t.Fatal(err)
	}
	id := GetAnimals(root)[0].ID

	if err := SetAnimalField(root, id, "name", "Buttercup"); err != nil {
		t.Fatalf("SetAnimalField name: %v", err)
	}

	animals := GetAnimals(root)
	if len(animals) != 1 {
		t.Fatalf("expected 1 animal, got %d", len(animals))
	}
	if animals[0].Name != "Buttercup" {
		t.Errorf("Name = %q after rename, want Buttercup", animals[0].Name)
	}
}

func TestAddAnimal_RealSave(t *testing.T) {
	root := loadReal(t)

	// Find a barn or coop building
	buildings := GetBuildings(root)
	var barnID, coopID string
	for _, b := range buildings {
		if barnTypes[b.BuildingType] && barnID == "" {
			barnID = b.ID
		}
		if coopTypes[b.BuildingType] && coopID == "" {
			coopID = b.ID
		}
	}

	if barnID == "" && coopID == "" {
		t.Skip("no barn or coop in this save")
	}

	before := len(GetAnimals(root))

	if barnID != "" {
		if err := AddAnimal(root, barnID, "Cow", "TestCow"); err != nil {
			t.Fatalf("AddAnimal Cow: %v", err)
		}
	}
	if coopID != "" {
		if err := AddAnimal(root, coopID, "Chicken", "TestChicken"); err != nil {
			t.Fatalf("AddAnimal Chicken: %v", err)
		}
	}

	after := GetAnimals(root)
	expected := before
	if barnID != "" {
		expected++
	}
	if coopID != "" {
		expected++
	}
	if len(after) != expected {
		t.Errorf("animals: before=%d after=%d expected=%d", before, len(after), expected)
	}
}
