package save

import (
	"fmt"
	"sort"
	"strings"
)

// PlacementWarning is an informational message about a proposed building placement.
// Severity is "warning" (terrain/bounds/clearable) or "conflict" (overlaps existing building).
// All warnings are non-blocking — the caller decides whether to proceed.
type PlacementWarning struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type tileKey = [2]int

// farmSaveObstacles reads all current-state obstacles from the farm location.
// Returns a map of occupied tile positions → kind label, plus multi-tile resource clumps.
func farmSaveObstacles(root *Node) (map[tileKey]string, []Rect) {
	tiles := make(map[tileKey]string)
	var clumps []Rect

	farm := farmNode(root)
	if farm == nil {
		return tiles, clumps
	}

	// Terrain features: trees, fruit trees, crops (HoeDirt), grass, paths (Flooring), bushes.
	readTileDict := func(parent *Node, dictName, valuePath, fallbackKind string) {
		if parent == nil {
			return
		}
		dict := parent.Child(dictName)
		if dict == nil {
			dict = parent // some saves omit the wrapper element
		}
		for _, item := range dict.ChildrenNamed("item") {
			x := intOf(item, "key/Vector2/X")
			y := intOf(item, "key/Vector2/Y")
			kind := fallbackKind
			if vn := item.Get(valuePath); vn != nil {
				if t := vn.Attr("type"); t != "" {
					kind = t
				}
			}
			tiles[tileKey{x, y}] = kind
		}
	}

	readTileDict(
		farm.Child("terrainFeatures"),
		"SerializableDictionaryOfVector2TerrainFeature",
		"value/TerrainFeature",
		"terrain feature",
	)
	readTileDict(
		farm.Child("largeTerrainFeatures"),
		"SerializableDictionaryOfVector2LargeTerrainFeature",
		"value/LargeTerrainFeature",
		"large terrain feature",
	)

	// Placed objects: stones, weeds, machines, scarecrows, etc.
	if objs := farm.Child("objects"); objs != nil {
		dict := objs.Child("SerializableDictionaryOfVector2Object")
		if dict == nil {
			dict = objs
		}
		for _, item := range dict.ChildrenNamed("item") {
			x := intOf(item, "key/Vector2/X")
			y := intOf(item, "key/Vector2/Y")
			name := textOf(item, "value/Object/name")
			if name == "" {
				name = "Object"
			}
			tiles[tileKey{x, y}] = name
		}
	}

	// Resource clumps: large boulders, large logs, large stumps (multi-tile).
	if rc := farm.Child("resourceClumps"); rc != nil {
		for _, clump := range rc.ChildrenNamed("ResourceClump") {
			cx := intOf(clump, "tile/X")
			cy := intOf(clump, "tile/Y")
			cw := intOf(clump, "width")
			ch := intOf(clump, "height")
			if cw <= 0 {
				cw = 2
			}
			if ch <= 0 {
				ch = 2
			}
			clumps = append(clumps, Rect{cx, cy, cx + cw, cy + ch})
		}
	}

	return tiles, clumps
}

// ValidatePlacement checks a proposed building placement and returns any warnings.
// It never prevents placement; all issues are informational.
func ValidatePlacement(root *Node, buildingType string, tileX, tileY int) []PlacementWarning {
	out := []PlacementWarning{}

	def, ok := buildingDefs[buildingType]
	if !ok {
		out = append(out, PlacementWarning{"warning", fmt.Sprintf("unknown building type %q", buildingType)})
		return out
	}

	w, h := def.TilesWide, def.TilesHigh
	farmType := GetFarmType(root)
	layout := farmLayoutFor(farmType)

	// Bounds check.
	if !layout.Buildable.containsBuilding(tileX, tileY, w, h) {
		out = append(out, PlacementWarning{
			"warning",
			fmt.Sprintf("placement (%d,%d) extends outside the %s farm boundary (%d,%d)–(%d,%d)",
				tileX, tileY,
				layout.Name,
				layout.Buildable.X1, layout.Buildable.Y1,
				layout.Buildable.X2, layout.Buildable.Y2),
		})
	}

	// Permanent terrain obstacle check (water, cliffs, forest — per farm type).
	for _, obs := range layout.Obstacles {
		if obs.overlapsBuilding(tileX, tileY, w, h) {
			out = append(out, PlacementWarning{
				"warning",
				fmt.Sprintf("placement (%d,%d) overlaps a known impassable terrain zone on the %s farm",
					tileX, tileY, layout.Name),
			})
			break
		}
	}

	// Save-state obstacle check: terrain features, objects, resource clumps.
	// These are all clearable but currently block placement.
	tilemap, clumps := farmSaveObstacles(root)

	kindCounts := map[string]int{}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			if kind, blocked := tilemap[tileKey{tileX + dx, tileY + dy}]; blocked {
				kindCounts[kind]++
			}
		}
	}
	if len(kindCounts) > 0 {
		parts := make([]string, 0, len(kindCounts))
		for kind, count := range kindCounts {
			parts = append(parts, fmt.Sprintf("%d×%s", count, kind))
		}
		sort.Strings(parts)
		out = append(out, PlacementWarning{
			"warning",
			"clearable obstacles in footprint: " + strings.Join(parts, ", "),
		})
	}

	for _, clump := range clumps {
		if clump.overlapsBuilding(tileX, tileY, w, h) {
			out = append(out, PlacementWarning{
				"warning",
				fmt.Sprintf("large boulder/log/stump at (%d,%d) blocks footprint (clearable)", clump.X1, clump.Y1),
			})
		}
	}

	// Building-to-building conflict check (exact, non-clearable).
	for _, b := range GetBuildings(root) {
		def2, ok2 := buildingDefs[b.BuildingType]
		if !ok2 {
			continue
		}
		existing := Rect{b.TileX, b.TileY, b.TileX + def2.TilesWide, b.TileY + def2.TilesHigh}
		if existing.overlapsBuilding(tileX, tileY, w, h) {
			out = append(out, PlacementWarning{
				"conflict",
				fmt.Sprintf("overlaps existing %s at (%d,%d)", b.BuildingType, b.TileX, b.TileY),
			})
		}
	}

	return out
}

// SuggestPlacement scans the buildable area and returns the first tile position
// that has no warnings for the given building type. found is false if none exists.
func SuggestPlacement(root *Node, buildingType string) (tileX, tileY int, found bool) {
	def, ok := buildingDefs[buildingType]
	if !ok {
		return 0, 0, false
	}
	farmType := GetFarmType(root)
	layout := farmLayoutFor(farmType)

	w, h := def.TilesWide, def.TilesHigh
	b := layout.Buildable

	for y := b.Y1; y+h <= b.Y2; y++ {
		for x := b.X1; x+w <= b.X2; x++ {
			if len(ValidatePlacement(root, buildingType, x, y)) == 0 {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}
