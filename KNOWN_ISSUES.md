# Known Issues — Read-only or partially-editable menus

This is a survey of places where the UI displays save state but doesn't let
the user mutate it (or only lets them mutate part of it). They were found
while fixing the "can't add new items to empty inventory slots" bug — that
one is now resolved, the rest are open.

Each entry lists the **symptom**, **what's actually editable today**, and a
sketch of the **fix** so anyone picking it up has a starting point.

---

## Fixed

### Inventory: empty slots could not be filled

- **Symptom:** an empty slot rendered as `<td colspan="5">Empty</td>` with no
  controls.
- **Fix shipped:** `save.AddInventoryItem` builds an `<Item xsi:type="Object">`
  node from a small catalog (`KnownItemTypes`), exposed via
  `sdvedit_addInventoryItem` and `sdvedit_itemCatalog`. The inventory row for
  an empty slot now shows a catalog picker, name override, stack, quality, and
  an "Add" button. See `internal/save/accessor.go` and `site/app.js`
  `loadInventory`.

### Inventory: empty slots can now be cleared

- **Symptom:** `SetInventoryItem` only edited `stack`/`quality`; there was no
  inverse to `AddInventoryItem` to turn a populated `<Item>` back into
  `<Item xsi:nil="true"/>`.
- **Fix shipped:** `save.ClearInventorySlot(root, slot)` resets the slot's
  `Attrs` to `xsi:nil="true"` and empties `Children`/`Text`, exposed via
  `sdvedit_clearInventorySlot`. Each populated inventory row now shows a
  "Clear" button. See `internal/save/accessor.go` and `site/app.js`
  `loadInventory`.

### Player: current HP / stamina are now editable

- **Symptom:** the Player tab edited `maxHealth`/`maxStamina` only, not the
  current `<health>`/`<stamina>` values people usually want to refill.
- **Fix shipped:** `PlayerStats` gained `Health`/`Stamina`, read/written from
  `player/health` and `player/stamina`; the Player tab surfaces "Current
  Health" and "Current Stamina" fields. See `GetPlayerStats`/`SetPlayerStats`
  and `loadPlayer`.

### Friendships: pet type and gender are now editable

- **Symptom:** `SetPet` updated name/friendship/timesPet/breed, but the JS
  stripped `petType` and `gender`, leaving them read-only.
- **Fix shipped:** `SetPet` now writes `petType` and `Gender`; the pet form
  surfaces Type (Dog/Cat) and Gender (Male/Female) dropdowns that pass
  through. See `SetPet` and `loadFriendships`.

### Animals: can now be renamed

- **Symptom:** the Animals row showed `name` as plain text even though
  `SetAnimalField` already supported the `"name"` field.
- **Fix shipped:** the name cell is now a text input and `name` is included in
  the per-row apply patch. See `loadAnimals`.

### Inventory: item type can now be swapped in place

- **Symptom:** editing a populated slot only adjusted stack/quality; changing
  Stone → Iridium Ore meant clear-and-re-add.
- **Fix shipped:** `save.ReplaceInventoryItem(root, slot, itemId, name, stack,
  quality)` re-emits the item's child nodes for a new `itemId` (reusing
  `buildObjectItemChildren`/`lookupItemDef`), exposed via
  `sdvedit_replaceInventoryItem`. Populated rows now carry a catalog picker and
  a "Replace" button. See `internal/save/accessor.go` and `loadInventory`.

### Recipes: "Learn All" now unlocks the whole vanilla catalog

- **Symptom:** the Learn-All button only set `timesMade=1` on recipes the
  player had already learned; it could not unlock new ones.
- **Fix shipped:** `recipe_catalog.go` holds the vanilla cooking and crafting
  recipe key lists; `save.AddRecipe(root, section, name)` appends a learned
  entry (`value/int=0`, idempotent) and `save.LearnAllRecipes(root, section)`
  iterates the catalog, returning the count added. Exposed via
  `sdvedit_addRecipe`/`sdvedit_learnAllRecipes`; the "Learn All" buttons now
  unlock every catalog recipe and reload. Unlisted/modded recipes a save
  already knows are preserved. See `loadRecipes`.

### Friendships: pet breed options are filtered by type

- **Symptom:** the breed field was a free number input not tied to Dog vs Cat.
- **Fix shipped:** the breed control is now a dropdown driven by `PET_BREEDS`
  per pet type, rebuilt when the Type dropdown changes. See `petBreedField` /
  `loadFriendships`.

---

## Open issues

### Inventory: equipped gear is not surfaced at all

The save has top-level `<hat>`, `<shirtItem>`, `<pantsItem>`, `<boots>`, and
`<leftRing>`/`<rightRing>` slots distinct from `<items>`. The Inventory tab
ignores them, so the user can't edit ring effects or hat appearance even
though all the data is present in the parsed tree.

**Fix sketch:** new accessors `GetEquipment`/`SetEquipment` reading those
top-level slots; new section under the Inventory tab.

### Friendships: cannot add a missing NPC

`SetFriendships` loops `idx[e.NPC]` and silently skips unknown NPCs
(`accessor.go` `SetFriendships`). If the player hasn't met an NPC yet, the
NPC simply doesn't appear in the list — and the editor has no way to create
the `<item>` under `friendshipData`.

**Fix sketch:** new `save.AddFriendship(root, npcName)` that appends a fresh
`<item><key><string>NPC</string></key><value><Friendship>…</Friendship></value></item>`
with default-zero fields, plus a hard-coded list of vanilla NPC names for the
add UI.

### Friendships: cannot add a pet when none exists

`loadFriendships` only renders the pet section when `getJSON(sdvedit_getPet)`
is truthy. New-game saves and farms without a pet never see the form. There
is no `AddPet` accessor.

**Fix sketch:** add `save.AddPet(root, petType, name, breed)` that builds the
required `<NPC xsi:type="Pet">` node under `farm/characters`. Guard against
duplicates by checking for an existing Pet NPC first.

### Friendships: pet type and gender are not editable

`SetPet` updates name, friendship, timesPet, breed — but the JS strips
`petType` and `gender` from the payload (`app.js` ~line 254 keeps them as
read-only).

**Fix sketch:** add fields to the UI and pass them through. `SetPet` already
ignores them, so the Go side also needs `setLeaf(npc, "petType", …)` and
`setLeaf(npc, "Gender", …)`.

### Animals: cannot delete an animal

`AddAnimal` and `SetAnimalField` exist; there is no `RemoveAnimal`. The
Animals tab has no delete control. Useful for cleaning up a typo'd add or
removing an animal that's been replaced in-game already.

**Fix sketch:** `save.RemoveAnimal(root, animalID)` that walks each
building's `Animals/SerializableDictionaryOfInt64FarmAnimal` and drops the
matching `<item>` from `Children`.

### Animals: cannot rename an animal

The Animals row shows `name` as plain text. `SetAnimalField` already supports
arbitrary field names (including `"name"`), so this is purely a missing UI
control.

**Fix sketch:** swap the `<td>${esc(a.name)}</td>` for a text input and add
`name` to the patch list in the per-row apply handler.

### Animals: cannot move an animal between buildings

There's no way to reassign an animal to a different barn/coop, which matters
when the user adds a new building and wants to migrate occupants.

**Fix sketch:** `save.MoveAnimal(root, animalID, targetBuildingID)`: locate
the source `<item>` under its current building, splice it out of that
dictionary, append into the target's dictionary. Validate that the target
building type matches `def.IsCoopDweller`.

### Buildings: cannot delete a building

`AddBuilding` exists, no `RemoveBuilding`. Misplaced or unwanted buildings
can be edited (position, paint) but not removed. Deleting through this tool
would let users undo a bad placement without re-loading their save.

**Fix sketch:** `save.RemoveBuilding(root, buildingID)` that drops the
matching `<Building>` from `farm/buildings/Children`. Refuse if the building
still has occupants, or surface a warning.

### Buildings: cannot change building type

Edit form covers tile position and paint but not type. Changing type
in-game requires demolition; in the save it's a single `<buildingType>`
field, but tilesWide/Tall/maxOccupants/hayCapacity all need to stay
consistent.

**Fix sketch:** add a type dropdown; on apply, re-derive the def-driven
fields from `buildingDefs`. Probably gate behind a "recompute structural
fields" checkbox so users who hand-tuned values aren't surprised.

### Recipes: "Learn All" only updates already-known recipes

`getRecipes` iterates `<item>` children of `player/cookingRecipes` /
`player/craftingRecipes` — only recipes the player has already learned. The
"Learn All (set 1)" button does not actually unlock new recipes; it just
sets `timesMade=1` on the ones already present.

**Fix sketch:** maintain a vanilla recipe whitelist (cooking + crafting), and
add `save.AddRecipe(root, section, recipeName)` that appends a new `<item>`
with `value/int=0`. The Learn-All button can then iterate the full list.
Mirror logic for an "unlearn" / remove control if desired.

### Player: cannot edit current HP / current stamina

The Player tab edits `maxHealth` and `maxStamina` only. The save also has
the current `<health>` and `<stamina>` values, which is what people usually
want to refill mid-day.

**Fix sketch:** extend `PlayerStats` with `Health`/`Stamina`, read/write
`player/health` and `player/stamina`, surface two extra number fields.

### Quests and Bundles tabs are intentionally read-only

Both panels render a hint explaining why
(`Quest editing is read-only — completing quests mid-game can cause script state
issues.` / `Editing bundle completion requires modifying netWorldState/completedBundles
which varies by game version`). These are documented design decisions, not
bugs — listed here so anyone scanning the menus knows they were considered.
