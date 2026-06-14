package save

import (
	"bytes"
	"strings"
	"testing"
)

// equipmentFixture mirrors a real 1.6 save: shirtItem/pantsItem are always
// present (clothing XML copied from a real save), boots and leftRing are
// populated, while hat and rightRing are omitted entirely (the game drops
// unequipped slots rather than writing xsi:nil placeholders).
const equipmentFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <player>
    <name>Trent</name>
    <shirtItem><isLostItem>false</isLostItem><category>-100</category><hasBeenInInventory>true</hasBeenInInventory><name>Shirt</name><itemId>1109</itemId><specialItem>false</specialItem><isRecipe>false</isRecipe><quality>0</quality><stack>1</stack><SpecialVariable>0</SpecialVariable><price>50</price><indexInTileSheet>109</indexInTileSheet><indexInTileSheetFemale xsi:nil="true" /><clothesType>SHIRT</clothesType><dyeable>false</dyeable><clothesColor><B>255</B><G>255</G><R>255</R><A>255</A><PackedValue>4294967295</PackedValue></clothesColor><isPrismatic>false</isPrismatic><Price>50</Price></shirtItem>
    <pantsItem><isLostItem>false</isLostItem><category>-100</category><hasBeenInInventory>true</hasBeenInInventory><name>Farmer Pants</name><itemId>0</itemId><specialItem>false</specialItem><isRecipe>false</isRecipe><quality>0</quality><stack>1</stack><SpecialVariable>0</SpecialVariable><price>50</price><indexInTileSheet>0</indexInTileSheet><indexInTileSheetFemale xsi:nil="true" /><clothesType>PANTS</clothesType><dyeable>true</dyeable><clothesColor><B>183</B><G>85</G><R>46</R><A>255</A><PackedValue>4290204974</PackedValue></clothesColor><isPrismatic>false</isPrismatic><Price>50</Price></pantsItem>
    <boots><name>Leather Boots</name><itemId>504</itemId><indexInTileSheet>0</indexInTileSheet><defenseBonus>1</defenseBonus><immunityBonus>0</immunityBonus></boots>
    <leftRing><name>Glow Ring</name><itemId>517</itemId></leftRing>
  </player>
</SaveGame>`

func slotByName(slots []EquipmentSlot, name string) (EquipmentSlot, bool) {
	for _, s := range slots {
		if s.Slot == name {
			return s, true
		}
	}
	return EquipmentSlot{}, false
}

func TestGetEquipment_ReadsClothingAndPresence(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	slots := GetEquipment(root)

	if len(slots) != 6 {
		t.Fatalf("expected 6 slots, got %d", len(slots))
	}

	shirt, ok := slotByName(slots, "shirt")
	if !ok {
		t.Fatal("shirt slot missing")
	}
	if !shirt.Present {
		t.Error("shirt should be present")
	}
	if shirt.ItemID != "1109" {
		t.Errorf("shirt ItemID = %q, want 1109", shirt.ItemID)
	}
	if shirt.ColorR != 255 || shirt.ColorA != 255 {
		t.Errorf("shirt color = (%d,%d,%d,%d)", shirt.ColorR, shirt.ColorG, shirt.ColorB, shirt.ColorA)
	}

	hat, ok := slotByName(slots, "hat")
	if !ok {
		t.Fatal("hat slot missing from result")
	}
	if hat.Present {
		t.Error("hat is unequipped (absent) and should report not present")
	}
}

func TestGetEquipment_ReadsBootsAndRing(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	slots := GetEquipment(root)

	boots, _ := slotByName(slots, "boots")
	if !boots.Present || boots.ItemID != "504" {
		t.Errorf("boots = %+v", boots)
	}
	if boots.DefenseBonus != 1 {
		t.Errorf("boots DefenseBonus = %d, want 1", boots.DefenseBonus)
	}

	left, _ := slotByName(slots, "leftRing")
	if !left.Present || left.ItemID != "517" {
		t.Errorf("leftRing = %+v", left)
	}
	right, _ := slotByName(slots, "rightRing")
	if right.Present {
		t.Error("rightRing is absent and should report not present")
	}
}

func TestSetEquipmentColor_UpdatesComponentsAndPackedValue(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	// pantsItem real values: B=183 G=85 R=46 A=255 -> PackedValue 4290204974
	if err := SetEquipmentColor(root, "pants", 46, 85, 183, 255); err != nil {
		t.Fatalf("SetEquipmentColor: %v", err)
	}
	slots := GetEquipment(root)
	pants, _ := slotByName(slots, "pants")
	if pants.ColorR != 46 || pants.ColorG != 85 || pants.ColorB != 183 || pants.ColorA != 255 {
		t.Errorf("pants color = %+v", pants)
	}
	color := root.Get("player/pantsItem/clothesColor")
	if got := textOf(color, "PackedValue"); got != "4290204974" {
		t.Errorf("PackedValue = %q, want 4290204974", got)
	}
}

func TestSetEquipmentColor_RejectsNonClothing(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	if err := SetEquipmentColor(root, "boots", 1, 2, 3, 255); err == nil {
		t.Fatal("expected error setting color on boots")
	}
}

func TestSetEquipmentField_BootsBonuses(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	if err := SetEquipmentField(root, "boots", "immunityBonus", "4"); err != nil {
		t.Fatalf("SetEquipmentField: %v", err)
	}
	boots, _ := slotByName(GetEquipment(root), "boots")
	if boots.ImmunityBonus != 4 {
		t.Errorf("ImmunityBonus = %d, want 4", boots.ImmunityBonus)
	}
}

func TestSetEquipmentField_RejectsEmptySlot(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	if err := SetEquipmentField(root, "hat", "itemId", "5"); err == nil {
		t.Fatal("expected error editing an empty slot")
	}
}

func TestClearEquipmentSlot_RemovesElement(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	if err := ClearEquipmentSlot(root, "leftRing"); err != nil {
		t.Fatalf("ClearEquipmentSlot: %v", err)
	}
	left, _ := slotByName(GetEquipment(root), "leftRing")
	if left.Present {
		t.Error("leftRing still present after clear")
	}
	if root.Get("player/leftRing") != nil {
		t.Error("leftRing element should be removed entirely, not nil-placeheld")
	}
	if err := ClearEquipmentSlot(root, "rightRing"); err == nil {
		t.Fatal("expected error clearing an already-empty slot")
	}
}

func TestAddClothing_FillsEmptySlotAndRoundTrips(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	// clear shirt then re-add to exercise the add path
	if err := ClearEquipmentSlot(root, "shirt"); err != nil {
		t.Fatal(err)
	}
	if err := AddClothing(root, "shirt", "1000", "Test Shirt"); err != nil {
		t.Fatalf("AddClothing: %v", err)
	}

	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatal(err)
	}
	root2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	shirt, _ := slotByName(GetEquipment(root2), "shirt")
	if !shirt.Present || shirt.ItemID != "1000" || shirt.Name != "Test Shirt" {
		t.Errorf("after round-trip, shirt = %+v", shirt)
	}
	if shirt.ClothesType != "SHIRT" {
		t.Errorf("clothesType = %q, want SHIRT", shirt.ClothesType)
	}
}

func TestAddClothing_RejectsOccupied(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	if err := AddClothing(root, "shirt", "1000", "Test"); err == nil {
		t.Fatal("expected error adding to an occupied clothing slot")
	}
}

func TestAddClothing_RejectsNonClothingSlot(t *testing.T) {
	root, _ := Parse(strings.NewReader(equipmentFixture))
	if err := AddClothing(root, "leftRing", "517", "Ring"); err == nil {
		t.Fatal("expected error: ring creation is not supported via AddClothing")
	}
}
