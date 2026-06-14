//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	"github.com/nulvox/sdvedit/internal/save"
)

// state holds the currently loaded save document.
var state struct {
	root *save.Node
}

func main() {
	g := js.Global()
	g.Set("sdvedit_load", js.FuncOf(jsLoad))
	g.Set("sdvedit_export", js.FuncOf(jsExport))
	g.Set("sdvedit_getPlayer", js.FuncOf(jsGetPlayer))
	g.Set("sdvedit_setPlayer", js.FuncOf(jsSetPlayer))
	g.Set("sdvedit_getFriendships", js.FuncOf(jsGetFriendships))
	g.Set("sdvedit_setFriendships", js.FuncOf(jsSetFriendships))
	g.Set("sdvedit_getWorldState", js.FuncOf(jsGetWorldState))
	g.Set("sdvedit_setWorldState", js.FuncOf(jsSetWorldState))
	g.Set("sdvedit_getBuildings", js.FuncOf(jsGetBuildings))
	g.Set("sdvedit_setBuildingField", js.FuncOf(jsSetBuildingField))
	g.Set("sdvedit_addBuilding", js.FuncOf(jsAddBuilding))
	g.Set("sdvedit_buildingTypes", js.FuncOf(jsBuildingTypes))
	g.Set("sdvedit_validatePlacement", js.FuncOf(jsValidatePlacement))
	g.Set("sdvedit_suggestPlacement", js.FuncOf(jsSuggestPlacement))
	g.Set("sdvedit_getAnimals", js.FuncOf(jsGetAnimals))
	g.Set("sdvedit_setAnimalField", js.FuncOf(jsSetAnimalField))
	g.Set("sdvedit_addAnimal", js.FuncOf(jsAddAnimal))
	g.Set("sdvedit_animalTypes", js.FuncOf(jsAnimalTypes))
	g.Set("sdvedit_getPet", js.FuncOf(jsGetPet))
	g.Set("sdvedit_setPet", js.FuncOf(jsSetPet))
	g.Set("sdvedit_getInventory", js.FuncOf(jsGetInventory))
	g.Set("sdvedit_setInventoryItem", js.FuncOf(jsSetInventoryItem))
	g.Set("sdvedit_addInventoryItem", js.FuncOf(jsAddInventoryItem))
	g.Set("sdvedit_itemCatalog", js.FuncOf(jsItemCatalog))
	g.Set("sdvedit_getCookingRecipes", js.FuncOf(jsGetCookingRecipes))
	g.Set("sdvedit_setCookingRecipes", js.FuncOf(jsSetCookingRecipes))
	g.Set("sdvedit_getCraftingRecipes", js.FuncOf(jsGetCraftingRecipes))
	g.Set("sdvedit_setCraftingRecipes", js.FuncOf(jsSetCraftingRecipes))
	g.Set("sdvedit_getMail", js.FuncOf(jsGetMail))
	g.Set("sdvedit_addMail", js.FuncOf(jsAddMail))
	g.Set("sdvedit_removeMail", js.FuncOf(jsRemoveMail))
	g.Set("sdvedit_getQuests", js.FuncOf(jsGetQuests))
	g.Set("sdvedit_getBundles", js.FuncOf(jsGetBundles))
	g.Set("sdvedit_ready", js.ValueOf(true))

	// block forever
	select {}
}

// jsLoad accepts a Uint8Array from JS, parses it as XML.
func jsLoad(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errResult("expected Uint8Array argument")
	}
	buf := make([]byte, args[0].Length())
	js.CopyBytesToGo(buf, args[0])

	root, err := save.Parse(bytes.NewReader(buf))
	if err != nil {
		return errResult(err.Error())
	}
	state.root = root
	return okResult(nil)
}

// jsExport serializes the current state back to XML and returns a Uint8Array.
func jsExport(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	var buf bytes.Buffer
	if err := save.Serialize(state.root, &buf); err != nil {
		return errResult(err.Error())
	}
	dst := js.Global().Get("Uint8Array").New(buf.Len())
	js.CopyBytesToJS(dst, buf.Bytes())
	return okResultRaw(dst)
}

func jsGetPlayer(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	s, err := save.GetPlayerStats(state.root)
	if err != nil {
		return errResult(err.Error())
	}
	return jsonResult(s)
}

func jsSetPlayer(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 1 {
		return errResult("expected JSON string")
	}
	var s save.PlayerStats
	if err := json.Unmarshal([]byte(args[0].String()), &s); err != nil {
		return errResult(err.Error())
	}
	if err := save.SetPlayerStats(state.root, s); err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsGetFriendships(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetFriendships(state.root))
}

func jsSetFriendships(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	var entries []save.FriendshipEntry
	if err := json.Unmarshal([]byte(args[0].String()), &entries); err != nil {
		return errResult(err.Error())
	}
	save.SetFriendships(state.root, entries)
	return okResult(nil)
}

func jsGetWorldState(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetWorldState(state.root))
}

func jsSetWorldState(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	var ws save.WorldState
	if err := json.Unmarshal([]byte(args[0].String()), &ws); err != nil {
		return errResult(err.Error())
	}
	save.SetWorldState(state.root, ws)
	return okResult(nil)
}

func jsGetBuildings(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetBuildings(state.root))
}

func jsSetBuildingField(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 3 {
		return errResult("expected (id, field, value)")
	}
	err := save.SetBuildingField(state.root, args[0].String(), args[1].String(), args[2].String())
	if err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsBuildingTypes(_ js.Value, _ []js.Value) any {
	return jsonResult(save.KnownBuildingTypes())
}

func jsAddBuilding(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 3 {
		return errResult("expected (buildingType, tileX, tileY)")
	}
	if err := save.AddBuilding(state.root, args[0].String(), args[1].Int(), args[2].Int()); err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsValidatePlacement(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 3 {
		return errResult("expected (buildingType, tileX, tileY)")
	}
	warnings := save.ValidatePlacement(state.root, args[0].String(), args[1].Int(), args[2].Int())
	return jsonResult(warnings)
}

func jsSuggestPlacement(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 1 {
		return errResult("expected (buildingType)")
	}
	x, y, found := save.SuggestPlacement(state.root, args[0].String())
	if !found {
		return errResult("no valid placement found")
	}
	type result struct {
		TileX int `json:"tileX"`
		TileY int `json:"tileY"`
	}
	return jsonResult(result{x, y})
}

func jsAnimalTypes(_ js.Value, _ []js.Value) any {
	return jsonResult(save.KnownAnimalTypes())
}

func jsAddAnimal(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 3 {
		return errResult("expected (buildingId, animalType, animalName)")
	}
	if err := save.AddAnimal(state.root, args[0].String(), args[1].String(), args[2].String()); err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsGetAnimals(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetAnimals(state.root))
}

func jsSetAnimalField(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 3 {
		return errResult("expected (id, field, value)")
	}
	err := save.SetAnimalField(state.root, args[0].String(), args[1].String(), args[2].String())
	if err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsGetPet(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetPet(state.root))
}

func jsSetPet(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 1 {
		return errResult("expected JSON string")
	}
	var e save.PetEntry
	if err := json.Unmarshal([]byte(args[0].String()), &e); err != nil {
		return errResult(err.Error())
	}
	if err := save.SetPet(state.root, e); err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsGetInventory(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetInventory(state.root))
}

func jsSetInventoryItem(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 3 {
		return errResult("expected (slot, stack, quality)")
	}
	err := save.SetInventoryItem(state.root, args[0].Int(), args[1].Int(), args[2].Int())
	if err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsAddInventoryItem(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	if len(args) < 5 {
		return errResult("expected (slot, itemId, name, stack, quality)")
	}
	err := save.AddInventoryItem(state.root,
		args[0].Int(), args[1].String(), args[2].String(),
		args[3].Int(), args[4].Int())
	if err != nil {
		return errResult(err.Error())
	}
	return okResult(nil)
}

func jsItemCatalog(_ js.Value, _ []js.Value) any {
	return jsonResult(save.KnownItemTypes())
}

func jsGetCookingRecipes(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetCookingRecipes(state.root))
}

func jsSetCookingRecipes(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	var entries []save.RecipeEntry
	if err := json.Unmarshal([]byte(args[0].String()), &entries); err != nil {
		return errResult(err.Error())
	}
	save.SetCookingRecipes(state.root, entries)
	return okResult(nil)
}

func jsGetCraftingRecipes(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetCraftingRecipes(state.root))
}

func jsSetCraftingRecipes(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	var entries []save.RecipeEntry
	if err := json.Unmarshal([]byte(args[0].String()), &entries); err != nil {
		return errResult(err.Error())
	}
	save.SetCraftingRecipes(state.root, entries)
	return okResult(nil)
}

func jsGetMail(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetMailReceived(state.root))
}

func jsAddMail(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	save.AddMailFlag(state.root, args[0].String())
	return okResult(nil)
}

func jsRemoveMail(_ js.Value, args []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	save.RemoveMailFlag(state.root, args[0].String())
	return okResult(nil)
}

func jsGetQuests(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetQuests(state.root))
}

func jsGetBundles(_ js.Value, _ []js.Value) any {
	if state.root == nil {
		return errResult("no save loaded")
	}
	return jsonResult(save.GetBundles(state.root))
}

// --- result helpers ---

func okResult(data any) map[string]any {
	return map[string]any{"ok": true, "data": data}
}

func okResultRaw(v js.Value) map[string]any {
	return map[string]any{"ok": true, "data": v}
}

func errResult(msg string) map[string]any {
	return map[string]any{"ok": false, "error": msg}
}

func jsonResult(v any) map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		return errResult(err.Error())
	}
	return okResult(string(b))
}
