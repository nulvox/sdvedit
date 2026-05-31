package save

// Farm type constants matching the game's whichFarm field.
const (
	FarmStandard    = 0
	FarmRiverland   = 1
	FarmForest      = 2
	FarmHilltop     = 3
	FarmWilderness  = 4
	FarmFourCorners = 5
	FarmBeach       = 6
)

// FarmNames maps whichFarm values to display names.
var FarmNames = map[int]string{
	0: "Standard",
	1: "Riverland",
	2: "Forest",
	3: "Hill-top",
	4: "Wilderness",
	5: "Four Corners",
	6: "Beach",
}

// Rect is an axis-aligned inclusive tile rectangle [X1,X2) × [Y1,Y2).
// X2/Y2 are exclusive (one past the last occupied tile).
type Rect struct{ X1, Y1, X2, Y2 int }

// overlapsBuilding returns true when the rect overlaps the given building footprint.
func (r Rect) overlapsBuilding(bx, by, bw, bh int) bool {
	return bx < r.X2 && bx+bw > r.X1 && by < r.Y2 && by+bh > r.Y1
}

// containsBuilding returns true when the building fits entirely within the rect.
func (r Rect) containsBuilding(bx, by, bw, bh int) bool {
	return bx >= r.X1 && by >= r.Y1 && bx+bw <= r.X2 && by+bh <= r.Y2
}

// farmLayout describes the usable bounds and known obstacle zones for one farm type.
// Obstacle data is approximate; warnings are informational only.
type farmLayout struct {
	Name      string
	Buildable Rect   // outer limit of the buildable area
	Obstacles []Rect // known impassable zones (water, cliffs, forest)
}

// farmLayouts holds per-farm-type layout information.
// Coordinates are in tile space; the full farm map is 80×65 tiles.
var farmLayouts = map[int]farmLayout{
	FarmStandard: {
		Name:      "Standard",
		Buildable: Rect{3, 3, 77, 57},
		Obstacles: []Rect{
			{66, 3, 77, 38}, // right-side cliff / mountain
		},
	},
	FarmRiverland: {
		Name:      "Riverland",
		Buildable: Rect{3, 3, 77, 57},
		// Rivers carve the farm into islands; these are the major water bodies.
		Obstacles: []Rect{
			{12, 3, 29, 14},
			{29, 3, 51, 20},
			{3, 27, 21, 57},
			{52, 19, 77, 40},
			{21, 40, 52, 57},
		},
	},
	FarmForest: {
		Name:      "Forest",
		Buildable: Rect{3, 3, 77, 57},
		Obstacles: []Rect{
			{50, 3, 77, 50}, // hardwood forest on the right half
		},
	},
	FarmHilltop: {
		Name:      "Hill-top",
		Buildable: Rect{3, 3, 77, 57},
		Obstacles: []Rect{
			{46, 3, 77, 38}, // quarry / cliff zone
		},
	},
	FarmWilderness: {
		Name:      "Wilderness",
		Buildable: Rect{3, 3, 77, 57},
		Obstacles: []Rect{
			{65, 3, 77, 35}, // right cliff
		},
	},
	FarmFourCorners: {
		Name:      "Four Corners",
		Buildable: Rect{3, 3, 77, 57},
		// Rivers run in a + pattern dividing the farm into four quadrants.
		// The centre clearing (~38-42, 29-34) holds the greenhouse.
		Obstacles: []Rect{
			{38, 3, 42, 29},  // north river arm
			{38, 34, 42, 57}, // south river arm
			{3, 29, 37, 34},  // west river arm
			{43, 29, 77, 34}, // east river arm
		},
	},
	FarmBeach: {
		Name:      "Beach",
		Buildable: Rect{3, 3, 77, 57},
		Obstacles: []Rect{
			{54, 3, 77, 57}, // ocean / water on the right side
		},
	},
}

// farmLayoutFor returns the layout for the given farm type, defaulting to Standard.
func farmLayoutFor(farmType int) farmLayout {
	if l, ok := farmLayouts[farmType]; ok {
		return l
	}
	return farmLayouts[FarmStandard]
}

// GetFarmType returns the whichFarm value stored in the save root.
func GetFarmType(root *Node) int {
	return intOf(root, "whichFarm")
}
