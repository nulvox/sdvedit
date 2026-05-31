package save

import (
	"encoding/xml"
	"fmt"
	"math/rand/v2"
)

// BuildingDef holds the static layout properties for each building type.
type BuildingDef struct {
	TilesWide    int
	TilesHigh    int
	HayCapacity  int
	MaxOccupants int
	HumanDoorX   int
	HumanDoorY   int
	AnimalDoorX  int
	AnimalDoorY  int
	NeedsIndoors bool // true for barns, coops, slime hutch, shed
}

var buildingDefs = map[string]BuildingDef{
	"Barn":         {7, 4, 240, 4, 2, 3, 3, 3, true},
	"Big Barn":     {7, 4, 240, 8, 2, 3, 3, 3, true},
	"Deluxe Barn":  {7, 4, 240, 12, 2, 3, 3, 3, true},
	"Coop":         {6, 3, 0, 4, 2, 2, 4, 2, true},
	"Big Coop":     {6, 3, 0, 8, 2, 2, 4, 2, true},
	"Deluxe Coop":  {6, 3, 0, 12, 2, 2, 4, 2, true},
	"Slime Hutch":  {8, 4, 0, 20, 3, 3, -1, -1, true},
	"Shed":         {7, 3, 0, 0, 2, 2, -1, -1, true},
	"Big Shed":     {11, 5, 0, 0, 4, 4, -1, -1, true},
	"Silo":         {3, 3, 240, 0, 1, 2, -1, -1, false},
	"Well":         {4, 4, 0, 0, 2, 3, -1, -1, false},
	"Mill":         {4, 4, 0, 0, 2, 3, -1, -1, false},
	"Fish Pond":    {5, 5, 0, 10, 2, 4, -1, -1, false},
	"Shipping Bin": {2, 2, 0, 0, 0, 1, -1, -1, false},
}

// KnownBuildingTypes returns all supported building type names.
func KnownBuildingTypes() []string {
	types := make([]string, 0, len(buildingDefs))
	for k := range buildingDefs {
		types = append(types, k)
	}
	return types
}

// AddBuilding creates a new Building node and appends it to the farm's buildings list.
func AddBuilding(root *Node, buildingType string, tileX, tileY int) error {
	def, ok := buildingDefs[buildingType]
	if !ok {
		return fmt.Errorf("unknown building type %q", buildingType)
	}

	farm := farmNode(root)
	if farm == nil {
		return fmt.Errorf("farm not found")
	}
	bldgs := farm.Child("buildings")
	if bldgs == nil {
		return fmt.Errorf("buildings node not found")
	}

	id := newUUID()
	bldgs.Children = append(bldgs.Children, buildBuildingNode(id, buildingType, tileX, tileY, def))
	return nil
}

func buildBuildingNode(id, buildingType string, tileX, tileY int, def BuildingDef) *Node {
	nilAttr := []xml.Attr{{Name: xml.Name{Space: "http://www.w3.org/2001/XMLSchema-instance", Local: "nil"}, Value: "true"}}

	b := &Node{
		Name: "Building",
		Children: []*Node{
			leaf("id", id),
			{Name: "skinId", Children: []*Node{{Name: "string", Attrs: nilAttr}}},
			{Name: "nonInstancedIndoorsName", Children: []*Node{{Name: "string", Attrs: nilAttr}}},
			leaf("tileX", itoa(tileX)),
			leaf("tileY", itoa(tileY)),
			leaf("tilesWide", itoa(def.TilesWide)),
			leaf("tilesHigh", itoa(def.TilesHigh)),
			leaf("maxOccupants", itoa(def.MaxOccupants)),
			leaf("currentOccupants", "0"),
			leaf("daysOfConstructionLeft", "0"),
			leaf("daysUntilUpgrade", "0"),
			leaf("buildingType", buildingType),
			buildPaintColorNode(),
			leaf("hayCapacity", itoa(def.HayCapacity)),
			{Name: "buildingChests"},
			{Name: "humanDoor", Children: []*Node{leaf("X", itoa(def.HumanDoorX)), leaf("Y", itoa(def.HumanDoorY))}},
			{Name: "animalDoor", Children: []*Node{leaf("X", itoa(def.AnimalDoorX)), leaf("Y", itoa(def.AnimalDoorY))}},
			leaf("animalDoorOpen", "false"),
			leaf("animalDoorOpenAmount", "0"),
			leaf("magical", "false"),
			leaf("fadeWhenPlayerIsBehind", "true"),
			leaf("owner", "0"),
			leaf("isMoving", "false"),
		},
	}

	if def.NeedsIndoors {
		b.Children = append(b.Children, buildIndoorsNode(buildingType))
	}

	return b
}

func buildPaintColorNode() *Node {
	nilAttr := []xml.Attr{{Name: xml.Name{Space: "http://www.w3.org/2001/XMLSchema-instance", Local: "nil"}, Value: "true"}}
	boolTrue := func(v string) *Node { return &Node{Name: "boolean", Text: v} }
	intNode := func(v string) *Node { return &Node{Name: "int", Text: v} }
	return &Node{
		Name: "buildingPaintColor",
		Children: []*Node{
			{Name: "ColorName", Children: []*Node{{Name: "string", Attrs: nilAttr}}},
			{Name: "Color1Default", Children: []*Node{boolTrue("true")}},
			{Name: "Color1Hue", Children: []*Node{intNode("0")}},
			{Name: "Color1Saturation", Children: []*Node{intNode("0")}},
			{Name: "Color1Lightness", Children: []*Node{intNode("0")}},
			{Name: "Color2Default", Children: []*Node{boolTrue("true")}},
			{Name: "Color2Hue", Children: []*Node{intNode("0")}},
			{Name: "Color2Saturation", Children: []*Node{intNode("0")}},
			{Name: "Color2Lightness", Children: []*Node{intNode("0")}},
			{Name: "Color3Default", Children: []*Node{boolTrue("true")}},
			{Name: "Color3Hue", Children: []*Node{intNode("0")}},
			{Name: "Color3Saturation", Children: []*Node{intNode("0")}},
			{Name: "Color3Lightness", Children: []*Node{intNode("0")}},
		},
	}
}

func buildIndoorsNode(buildingType string) *Node {
	nilAttr := []xml.Attr{{Name: xml.Name{Space: "http://www.w3.org/2001/XMLSchema-instance", Local: "nil"}, Value: "true"}}
	waterColor := &Node{
		Name: "waterColor",
		Children: []*Node{
			leaf("B", "127"), leaf("G", "100"), leaf("R", "60"), leaf("A", "127"),
			leaf("PackedValue", "2139055164"),
		},
	}
	indoors := &Node{
		Name: "indoors",
		Children: []*Node{
			{Name: "buildings"},
			{Name: "animals"},
			{Name: "Animals", Children: []*Node{
				{Name: "SerializableDictionaryOfInt64FarmAnimal"},
			}},
			leaf("piecesOfHay", "0"),
			{Name: "characters"},
			{Name: "objects"},
			{Name: "resourceClumps"},
			{Name: "largeTerrainFeatures"},
			{Name: "terrainFeatures"},
			leaf("name", buildingType),
			waterColor,
			leaf("isFarm", "false"),
			leaf("isOutdoors", "false"),
			leaf("isStructure", "true"),
			leaf("ignoreDebrisWeather", "false"),
			leaf("ignoreOutdoorLighting", "false"),
			leaf("ignoreLights", "false"),
			leaf("treatAsOutdoors", "false"),
			leaf("numberOfSpawnedObjectsOnMap", "0"),
			leaf("miniJukeboxCount", "0"),
			{Name: "miniJukeboxTrack"},
			{Name: "furniture"},
			{Name: "isThereABed", Attrs: nilAttr},
		},
	}
	return indoors
}

// newUUID generates a random UUID v4 string.
func newUUID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(rand.IntN(256))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
