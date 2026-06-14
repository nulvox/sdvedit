package save

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// equipmentSlots defines the six worn-gear slots and how each maps onto the
// XML element under <player>. "kind" drives which fields are meaningful:
//   - clothing: shirt/pants — color (RGBA) and dyeable
//   - boots:    defense/immunity bonuses
//   - ring:     itemId only
//   - hat:      itemId/name only
//
// Clothing slots are always written by the game; hat/boots/rings are omitted
// from the save entirely when nothing is equipped.
type equipmentSlotDef struct {
	slot    string // stable key used by callers/UI
	element string // XML element name under <player>
	kind    string
}

var equipmentSlotDefs = []equipmentSlotDef{
	{"hat", "hat", "hat"},
	{"shirt", "shirtItem", "clothing"},
	{"pants", "pantsItem", "clothing"},
	{"boots", "boots", "boots"},
	{"leftRing", "leftRing", "ring"},
	{"rightRing", "rightRing", "ring"},
}

func equipmentSlotDefFor(slot string) (equipmentSlotDef, bool) {
	for _, d := range equipmentSlotDefs {
		if d.slot == slot {
			return d, true
		}
	}
	return equipmentSlotDef{}, false
}

// EquipmentSlot is the editor-facing view of one worn-gear slot.
type EquipmentSlot struct {
	Slot    string `json:"slot"`
	Kind    string `json:"kind"`
	Present bool   `json:"present"`
	ItemID  string `json:"itemId"`
	Name    string `json:"name"`
	// clothing
	ClothesType string `json:"clothesType"`
	Dyeable     bool   `json:"dyeable"`
	ColorR      int    `json:"colorR"`
	ColorG      int    `json:"colorG"`
	ColorB      int    `json:"colorB"`
	ColorA      int    `json:"colorA"`
	// boots
	DefenseBonus  int `json:"defenseBonus"`
	ImmunityBonus int `json:"immunityBonus"`
}

// GetEquipment returns all six worn-gear slots in a stable order. A slot is
// reported Present=false when its element is missing or carries xsi:nil.
func GetEquipment(root *Node) []EquipmentSlot {
	player := root.Child("player")
	out := make([]EquipmentSlot, 0, len(equipmentSlotDefs))
	for _, def := range equipmentSlotDefs {
		s := EquipmentSlot{Slot: def.slot, Kind: def.kind}
		var node *Node
		if player != nil {
			node = player.Child(def.element)
		}
		if node == nil || node.Attr("nil") == "true" {
			out = append(out, s)
			continue
		}
		s.Present = true
		s.ItemID = textOf(node, "itemId")
		s.Name = textOf(node, "name")
		switch def.kind {
		case "clothing":
			s.ClothesType = textOf(node, "clothesType")
			s.Dyeable = boolOf(node, "dyeable")
			s.ColorR = intOf(node, "clothesColor/R")
			s.ColorG = intOf(node, "clothesColor/G")
			s.ColorB = intOf(node, "clothesColor/B")
			s.ColorA = intOf(node, "clothesColor/A")
		case "boots":
			s.DefenseBonus = intOf(node, "defenseBonus")
			s.ImmunityBonus = intOf(node, "immunityBonus")
		}
		out = append(out, s)
	}
	return out
}

// equipmentNode returns the populated element for a slot, or nil/error.
func equipmentNode(root *Node, slot string) (*Node, equipmentSlotDef, error) {
	def, ok := equipmentSlotDefFor(slot)
	if !ok {
		return nil, def, fmt.Errorf("unknown equipment slot %q", slot)
	}
	player := root.Child("player")
	if player == nil {
		return nil, def, fmt.Errorf("player not found")
	}
	node := player.Child(def.element)
	if node == nil || node.Attr("nil") == "true" {
		return nil, def, fmt.Errorf("%s slot is empty", slot)
	}
	return node, def, nil
}

// packColor encodes an RGBA color the way Stardew stores PackedValue:
// A<<24 | B<<16 | G<<8 | R.
func packColor(r, g, b, a int) uint32 {
	return uint32(a&0xff)<<24 | uint32(b&0xff)<<16 | uint32(g&0xff)<<8 | uint32(r&0xff)
}

// SetEquipmentColor sets the RGBA components of a clothing slot's color and
// keeps PackedValue consistent. Errors if the slot is empty or not clothing.
func SetEquipmentColor(root *Node, slot string, r, g, b, a int) error {
	node, def, err := equipmentNode(root, slot)
	if err != nil {
		return err
	}
	if def.kind != "clothing" {
		return fmt.Errorf("%s slot has no editable color", slot)
	}
	color := node.Child("clothesColor")
	if color == nil {
		return fmt.Errorf("%s slot missing clothesColor", slot)
	}
	if err := setLeaf(color, "R", strconv.Itoa(r)); err != nil {
		return err
	}
	setLeaf(color, "G", strconv.Itoa(g))
	setLeaf(color, "B", strconv.Itoa(b))
	setLeaf(color, "A", strconv.Itoa(a))
	setLeaf(color, "PackedValue", strconv.FormatUint(uint64(packColor(r, g, b, a)), 10))
	return nil
}

// SetEquipmentField patches a single scalar field on a populated slot. Allowed
// fields depend on the slot kind: itemId/name for any; dyeable/clothesType for
// clothing; defenseBonus/immunityBonus for boots.
func SetEquipmentField(root *Node, slot, field, value string) error {
	node, def, err := equipmentNode(root, slot)
	if err != nil {
		return err
	}
	switch field {
	case "itemId", "name":
		return setLeaf(node, field, value)
	case "dyeable", "clothesType":
		if def.kind != "clothing" {
			return fmt.Errorf("field %q not valid for %s slot", field, slot)
		}
		return setLeaf(node, field, value)
	case "defenseBonus", "immunityBonus":
		if def.kind != "boots" {
			return fmt.Errorf("field %q not valid for %s slot", field, slot)
		}
		return setLeaf(node, field, value)
	default:
		return fmt.Errorf("unknown equipment field %q", field)
	}
}

// ClearEquipmentSlot empties a worn-gear slot by removing its element from
// <player>, matching how the game represents an unequipped slot (the element
// is simply absent). Errors if the slot is already empty.
func ClearEquipmentSlot(root *Node, slot string) error {
	def, ok := equipmentSlotDefFor(slot)
	if !ok {
		return fmt.Errorf("unknown equipment slot %q", slot)
	}
	player := root.Child("player")
	if player == nil {
		return fmt.Errorf("player not found")
	}
	for i, c := range player.Children {
		if c.Name == def.element {
			player.Children = append(player.Children[:i], player.Children[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%s slot is already empty", slot)
}

// AddClothing fills an empty clothing slot (shirt or pants) with a fresh
// default-white garment. The node shape is copied from a real 1.6 save, so it
// round-trips cleanly. Hat/boots/ring creation is intentionally not supported
// yet (no schema-verified template); use the in-game wardrobe for those.
func AddClothing(root *Node, slot, itemID, name string) error {
	def, ok := equipmentSlotDefFor(slot)
	if !ok {
		return fmt.Errorf("unknown equipment slot %q", slot)
	}
	if def.kind != "clothing" {
		return fmt.Errorf("%s is not a clothing slot", slot)
	}
	player := root.Child("player")
	if player == nil {
		return fmt.Errorf("player not found")
	}
	if existing := player.Child(def.element); existing != nil && existing.Attr("nil") != "true" {
		return fmt.Errorf("%s slot is already occupied", slot)
	}
	clothesType := "SHIRT"
	if slot == "pants" {
		clothesType = "PANTS"
	}
	if name == "" {
		name = "Shirt"
		if slot == "pants" {
			name = "Pants"
		}
	}
	player.Children = append(player.Children, buildClothingNode(def.element, itemID, name, clothesType))
	return nil
}

func buildClothingNode(element, itemID, name, clothesType string) *Node {
	nilAttr := []xml.Attr{{Name: xml.Name{Space: "http://www.w3.org/2001/XMLSchema-instance", Local: "nil"}, Value: "true"}}
	return &Node{
		Name: element,
		Children: []*Node{
			leaf("isLostItem", "false"),
			leaf("category", "-100"),
			leaf("hasBeenInInventory", "true"),
			leaf("name", name),
			leaf("itemId", itemID),
			leaf("specialItem", "false"),
			leaf("isRecipe", "false"),
			leaf("quality", "0"),
			leaf("stack", "1"),
			leaf("SpecialVariable", "0"),
			leaf("price", "50"),
			leaf("indexInTileSheet", itemID),
			{Name: "indexInTileSheetFemale", Attrs: nilAttr},
			leaf("clothesType", clothesType),
			leaf("dyeable", "true"),
			{Name: "clothesColor", Children: []*Node{
				leaf("B", "255"), leaf("G", "255"), leaf("R", "255"),
				leaf("A", "255"), leaf("PackedValue", "4294967295"),
			}},
			leaf("isPrismatic", "false"),
			leaf("Price", "50"),
		},
	}
}
