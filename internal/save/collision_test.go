package save

import (
	"strings"
	"testing"
)

// farmFixtureForCollision is a minimal save with a Standard farm and one Silo.
const farmFixtureForCollision = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <player><UniqueMultiplayerID>1</UniqueMultiplayerID></player>
  <whichFarm>0</whichFarm>
  <locations>
    <GameLocation xsi:type="Farm">
      <buildings>
        <Building>
          <id>aaa-111</id>
          <tileX>10</tileX>
          <tileY>10</tileY>
          <tilesWide>3</tilesWide>
          <tilesHigh>3</tilesHigh>
          <maxOccupants>0</maxOccupants>
          <currentOccupants>0</currentOccupants>
          <daysOfConstructionLeft>0</daysOfConstructionLeft>
          <daysUntilUpgrade>0</daysUntilUpgrade>
          <buildingType>Silo</buildingType>
          <hayCapacity>240</hayCapacity>
          <buildingChests/>
          <humanDoor><X>1</X><Y>2</Y></humanDoor>
          <animalDoor><X>-1</X><Y>-1</Y></animalDoor>
          <animalDoorOpen>false</animalDoorOpen>
          <animalDoorOpenAmount>0</animalDoorOpenAmount>
          <magical>false</magical>
          <fadeWhenPlayerIsBehind>true</fadeWhenPlayerIsBehind>
          <owner>0</owner>
          <isMoving>false</isMoving>
        </Building>
      </buildings>
      <animals/>
    </GameLocation>
  </locations>
</SaveGame>`

func TestValidatePlacement_Valid(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	warnings := ValidatePlacement(root, "Barn", 20, 20)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %+v", warnings)
	}
}

func TestValidatePlacement_OutOfBounds(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	// tileX=75, Barn is 7 wide → 75+7=82 > 77
	warnings := ValidatePlacement(root, "Barn", 75, 20)
	hasBounds := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "boundary") {
			hasBounds = true
		}
	}
	if !hasBounds {
		t.Errorf("expected boundary warning, got %+v", warnings)
	}
}

func TestValidatePlacement_BuildingOverlap(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	// Existing Silo at (10,10) size 3×3; place Barn at (11,11)
	warnings := ValidatePlacement(root, "Barn", 11, 11)
	hasConflict := false
	for _, w := range warnings {
		if w.Severity == "conflict" {
			hasConflict = true
		}
	}
	if !hasConflict {
		t.Errorf("expected conflict warning, got %+v", warnings)
	}
}

func TestValidatePlacement_NoOverlapAdjacentBuilding(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	// Silo at (10,10) size 3×3; place Barn at (13,10) — just right of Silo
	warnings := ValidatePlacement(root, "Barn", 13, 10)
	for _, w := range warnings {
		if w.Severity == "conflict" {
			t.Errorf("unexpected conflict for adjacent (non-overlapping) placement: %+v", warnings)
		}
	}
}

func TestValidatePlacement_TerrainObstacle(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	// Standard farm obstacle at (66,3,77,38); place Barn at (68,10)
	warnings := ValidatePlacement(root, "Barn", 68, 10)
	hasTerrain := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "impassable terrain") {
			hasTerrain = true
		}
	}
	if !hasTerrain {
		t.Errorf("expected terrain obstacle warning, got %+v", warnings)
	}
}

func TestValidatePlacement_UnknownType(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	warnings := ValidatePlacement(root, "Castle", 10, 10)
	if len(warnings) == 0 {
		t.Error("expected warning for unknown type")
	}
}

func TestSuggestPlacement_FindsValidSpot(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	x, y, found := SuggestPlacement(root, "Barn")
	if !found {
		t.Fatal("expected to find a valid placement")
	}
	// The suggested placement should itself have no warnings.
	warnings := ValidatePlacement(root, "Barn", x, y)
	if len(warnings) != 0 {
		t.Errorf("suggested placement (%d,%d) has warnings: %+v", x, y, warnings)
	}
}

func TestSuggestPlacement_UnknownType(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	_, _, found := SuggestPlacement(root, "Castle")
	if found {
		t.Error("expected not found for unknown type")
	}
}

// farmFixtureWithObstacles has a tree at (20,20), a stone object at (21,20),
// and a 2×2 resource clump at (22,20).
const farmFixtureWithObstacles = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <player><UniqueMultiplayerID>1</UniqueMultiplayerID></player>
  <whichFarm>0</whichFarm>
  <locations>
    <GameLocation xsi:type="Farm">
      <buildings/>
      <animals/>
      <terrainFeatures>
        <SerializableDictionaryOfVector2TerrainFeature>
          <item>
            <key><Vector2><X>20</X><Y>20</Y></Vector2></key>
            <value><TerrainFeature xsi:type="Tree"><growthStage>5</growthStage></TerrainFeature></value>
          </item>
        </SerializableDictionaryOfVector2TerrainFeature>
      </terrainFeatures>
      <objects>
        <SerializableDictionaryOfVector2Object>
          <item>
            <key><Vector2><X>21</X><Y>20</Y></Vector2></key>
            <value><Object><name>Stone</name></Object></value>
          </item>
        </SerializableDictionaryOfVector2Object>
      </objects>
      <resourceClumps>
        <ResourceClump>
          <width>2</width><height>2</height>
          <tile><X>22</X><Y>20</Y></tile>
        </ResourceClump>
      </resourceClumps>
    </GameLocation>
  </locations>
</SaveGame>`

func TestValidatePlacement_TerrainFeatureBlocks(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureWithObstacles))
	// Barn at (19,19) size 7×4 covers tiles 19–25, 19–22 — hits tree at (20,20)
	warnings := ValidatePlacement(root, "Barn", 19, 19)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "clearable obstacles") && strings.Contains(w.Message, "Tree") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected tree obstacle warning, got %+v", warnings)
	}
}

func TestValidatePlacement_ObjectBlocks(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureWithObstacles))
	warnings := ValidatePlacement(root, "Barn", 19, 19)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "clearable obstacles") && strings.Contains(w.Message, "Stone") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Stone obstacle warning, got %+v", warnings)
	}
}

func TestValidatePlacement_ResourceClumpBlocks(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureWithObstacles))
	warnings := ValidatePlacement(root, "Barn", 19, 19)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "boulder") || strings.Contains(w.Message, "stump") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected resource clump warning, got %+v", warnings)
	}
}

func TestValidatePlacement_ClearOfObstacles(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureWithObstacles))
	// Place far away from the obstacles at (20-24, 20-21)
	warnings := ValidatePlacement(root, "Silo", 40, 40)
	for _, w := range warnings {
		if strings.Contains(w.Message, "clearable") || strings.Contains(w.Message, "boulder") {
			t.Errorf("unexpected obstacle warning for clear area: %+v", w)
		}
	}
}

func TestGetFarmType(t *testing.T) {
	root, _ := Parse(strings.NewReader(farmFixtureForCollision))
	ft := GetFarmType(root)
	if ft != FarmStandard {
		t.Errorf("expected FarmStandard (0), got %d", ft)
	}
}

func TestFarmLayoutFor_UnknownDefaultsToStandard(t *testing.T) {
	layout := farmLayoutFor(999)
	if layout.Name != "Standard" {
		t.Errorf("expected Standard layout for unknown type, got %q", layout.Name)
	}
}
