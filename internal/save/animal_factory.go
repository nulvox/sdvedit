package save

import (
	"encoding/xml"
	"fmt"
	"math/rand/v2"
	"strconv"
)

// AnimalDef holds the static properties derived from the animal type name.
type AnimalDef struct {
	IsCoopDweller bool
	Sound         string
	HarvestType   int // 0=drop, 1=tool, 2=pail
	DaysToLay     int
	Price         int
}

var animalDefs = map[string]AnimalDef{
	"Chicken":       {true, "cluck", 0, 1, 800},
	"Duck":          {true, "Duck", 0, 2, 4000},
	"Rabbit":        {true, "rabbit", 1, 4, 8000},
	"Dinosaur":      {true, "Dinosaur", 0, 7, 0},
	"Void Chicken":  {true, "cluck", 0, 1, 5000},
	"Blue Chicken":  {true, "cluck", 0, 1, 800},
	"Golden Chicken":{true, "cluck", 0, 1, 100000},
	"Cow":           {false, "cow", 2, 2, 1500},
	"Goat":          {false, "goat", 2, 2, 4000},
	"Pig":           {false, "pig", 1, 1, 16000},
	"Sheep":         {false, "sheep", 1, 3, 8000},
	"Ostrich":       {false, "ostrich", 0, 7, 0},
}

// CoopTypes and BarnTypes are the building types that house coop/barn animals.
var coopTypes = map[string]bool{
	"Coop": true, "Big Coop": true, "Deluxe Coop": true,
}
var barnTypes = map[string]bool{
	"Barn": true, "Big Barn": true, "Deluxe Barn": true,
}

// KnownAnimalTypes returns all supported animal type names.
func KnownAnimalTypes() []string {
	types := make([]string, 0, len(animalDefs))
	for k := range animalDefs {
		types = append(types, k)
	}
	return types
}

// AddAnimal creates a new FarmAnimal node and appends it to the given building.
// It returns an error if the building isn't found or the animal type is unknown.
func AddAnimal(root *Node, buildingID, animalType, animalName string) error {
	def, ok := animalDefs[animalType]
	if !ok {
		return fmt.Errorf("unknown animal type %q", animalType)
	}

	// locate the building
	farm := farmNode(root)
	if farm == nil {
		return fmt.Errorf("farm not found")
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return fmt.Errorf("buildings not found")
	}
	var target *Node
	for _, b := range bldgs.ChildrenNamed("Building") {
		if textOf(b, "id") == buildingID {
			target = b
			break
		}
	}
	if target == nil {
		return fmt.Errorf("building %q not found", buildingID)
	}

	// validate coop/barn match
	btype := textOf(target, "buildingType")
	if def.IsCoopDweller && !coopTypes[btype] {
		return fmt.Errorf("%s is a coop animal but building is %q", animalType, btype)
	}
	if !def.IsCoopDweller && !barnTypes[btype] {
		return fmt.Errorf("%s is a barn animal but building is %q", animalType, btype)
	}

	// find or create the animal dictionary node
	indoors := target.Child("indoors")
	if indoors == nil {
		return fmt.Errorf("building %q has no indoors node", buildingID)
	}
	animalsNode := indoors.Child("Animals")
	if animalsNode == nil {
		return fmt.Errorf("building %q has no Animals node", buildingID)
	}
	dictNode := animalsNode.Child("SerializableDictionaryOfInt64FarmAnimal")
	if dictNode == nil {
		// create it from scratch
		dictNode = &Node{Name: "SerializableDictionaryOfInt64FarmAnimal"}
		animalsNode.Children = append(animalsNode.Children, dictNode)
	}

	// generate a unique int64 ID (random, very unlikely to collide)
	id := rand.Int64()
	if id < 0 {
		id = -id
	}
	idStr := strconv.FormatInt(id, 10)

	ownerID := textOf(root, "player/UniqueMultiplayerID")
	if ownerID == "" {
		ownerID = "0"
	}

	item := buildAnimalItem(idStr, animalName, animalType, btype, ownerID, def)
	dictNode.Children = append(dictNode.Children, item)
	return nil
}

// RemoveAnimal deletes the FarmAnimal with the given int64 key from whichever
// building dictionary holds it. Returns an error if no animal matches.
func RemoveAnimal(root *Node, animalID string) error {
	farm := farmNode(root)
	if farm == nil {
		return fmt.Errorf("farm not found")
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return fmt.Errorf("buildings not found")
	}
	for _, b := range bldgs.ChildrenNamed("Building") {
		dict := animalDict(b)
		if dict == nil {
			continue
		}
		for i, item := range dict.Children {
			if item.Name != "item" {
				continue
			}
			key := item.Get("key/long")
			if key != nil && key.Text == animalID {
				dict.Children = append(dict.Children[:i], dict.Children[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("animal %q not found", animalID)
}

// MoveAnimal relocates the animal with the given int64 key to targetBuildingID,
// splicing its <item> out of the source dictionary and appending it to the
// target's. The target must be a barn/coop matching the animal's habitat;
// otherwise the animal is left untouched and an error is returned.
func MoveAnimal(root *Node, animalID, targetBuildingID string) error {
	farm := farmNode(root)
	if farm == nil {
		return fmt.Errorf("farm not found")
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return fmt.Errorf("buildings not found")
	}

	// resolve target building and its dictionary first, so we don't mutate the
	// source on a doomed move.
	var target *Node
	for _, b := range bldgs.ChildrenNamed("Building") {
		if textOf(b, "id") == targetBuildingID {
			target = b
			break
		}
	}
	if target == nil {
		return fmt.Errorf("target building %q not found", targetBuildingID)
	}
	targetType := textOf(target, "buildingType")

	// locate the source item and the animal's type.
	var srcDict, item *Node
	srcIdx := -1
	for _, b := range bldgs.ChildrenNamed("Building") {
		dict := animalDict(b)
		if dict == nil {
			continue
		}
		for i, it := range dict.Children {
			if it.Name != "item" {
				continue
			}
			if key := it.Get("key/long"); key != nil && key.Text == animalID {
				srcDict, item, srcIdx = dict, it, i
				break
			}
		}
		if item != nil {
			break
		}
	}
	if item == nil {
		return fmt.Errorf("animal %q not found", animalID)
	}

	animalType := textOf(item, "value/FarmAnimal/type")
	def, ok := animalDefs[animalType]
	if !ok {
		return fmt.Errorf("unknown animal type %q", animalType)
	}
	if def.IsCoopDweller && !coopTypes[targetType] {
		return fmt.Errorf("%s is a coop animal but target is %q", animalType, targetType)
	}
	if !def.IsCoopDweller && !barnTypes[targetType] {
		return fmt.Errorf("%s is a barn animal but target is %q", animalType, targetType)
	}

	targetDict := animalDict(target)
	if targetDict == nil {
		return fmt.Errorf("target building %q has no animal storage", targetBuildingID)
	}

	srcDict.Children = append(srcDict.Children[:srcIdx], srcDict.Children[srcIdx+1:]...)
	targetDict.Children = append(targetDict.Children, item)
	// keep the animal's record of which building type it lives in consistent.
	setLeaf(item, "value/FarmAnimal/buildingTypeILiveIn", targetType)
	return nil
}

// animalDict returns the SerializableDictionaryOfInt64FarmAnimal node for a
// building, or nil if the building has no indoors animal storage.
func animalDict(b *Node) *Node {
	indoors := b.Child("indoors")
	if indoors == nil {
		return nil
	}
	animals := indoors.Child("Animals")
	if animals == nil {
		return nil
	}
	return animals.Child("SerializableDictionaryOfInt64FarmAnimal")
}

// buildAnimalItem constructs the <item> node for one FarmAnimal entry.
func buildAnimalItem(id, name, animalType, buildingType, ownerID string, def AnimalDef) *Node {
	isCoopStr := strconv.FormatBool(def.IsCoopDweller)
	nilAttr := []xml.Attr{{Name: xml.Name{Space: "http://www.w3.org/2001/XMLSchema-instance", Local: "nil"}, Value: "true"}}

	animal := &Node{
		Name: "FarmAnimal",
		Children: []*Node{
			leaf("name", name),
			leaf("type", animalType),
			leaf("age", "0"),
			leaf("daysOwned", "0"),
			leaf("friendshipTowardFarmer", "0"),
			leaf("happiness", "255"),
			leaf("fullness", "255"),
			leaf("wasPet", "false"),
			leaf("produceQuality", "0"),
			leaf("isCoopDweller", isCoopStr),
			leaf("buildingTypeILiveIn", buildingType),
			leaf("myProduce", "-1"),
			leaf("currentProduce", "-1"),
			leaf("harvestType", strconv.Itoa(def.HarvestType)),
			leaf("daysToLay", strconv.Itoa(def.DaysToLay)),
			leaf("daysSinceLay", "0"),
			leaf("allowReproduction", "true"),
			leaf("meatIndex", "-1"),
			leaf("price", strconv.Itoa(def.Price)),
			leaf("sound", def.Sound),
			leaf("ownerID", ownerID),
			leaf("parentId", "-1"),
			leaf("id", id),
			leaf("speed", "2"),
			leaf("isSheared", "false"),
			leaf("showDifferentTextureWhenReadyForHarvest", "false"),
			{Name: "home", Attrs: nilAttr},
			{Name: "Sprite", Attrs: nilAttr},
			{Name: "Breeder", Attrs: nilAttr},
			{Name: "isSwimming", Attrs: nilAttr},
			{
				Name: "Position",
				Children: []*Node{leaf("X", "0"), leaf("Y", "0")},
			},
		},
	}

	return &Node{
		Name: "item",
		Children: []*Node{
			{Name: "key", Children: []*Node{leaf("long", id)}},
			{Name: "value", Children: []*Node{animal}},
		},
	}
}

func leaf(name, text string) *Node {
	return &Node{Name: name, Text: text}
}
