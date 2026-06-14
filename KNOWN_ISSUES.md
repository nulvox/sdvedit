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

### Animals: can now be deleted

- **Symptom:** `AddAnimal`/`SetAnimalField` existed but there was no inverse,
  so a typo'd or replaced animal couldn't be removed.
- **Fix shipped:** `save.RemoveAnimal(root, animalID)` drops the matching
  `<item>` from its building's `SerializableDictionaryOfInt64FarmAnimal`,
  exposed via `sdvedit_removeAnimal`. Each animal row now has a "Delete"
  button (with a confirm). See `animal_factory.go` and `loadAnimals`.

### Animals: can now be moved between buildings

- **Symptom:** no way to reassign an animal to a different barn/coop after a
  new building was added.
- **Fix shipped:** `save.MoveAnimal(root, animalID, targetBuildingID)` splices
  the `<item>` out of the source dictionary into the target's, validating that
  the target's habitat (coop vs barn) matches the animal and updating
  `buildingTypeILiveIn`. Exposed via `sdvedit_moveAnimal`; each row shows a
  habitat-filtered target picker and a "Move" button. See `MoveAnimal` and
  `loadAnimals`.

### Buildings: can now be deleted

- **Symptom:** `AddBuilding` existed with no inverse, so a misplaced building
  was stuck.
- **Fix shipped:** `save.RemoveBuilding(root, buildingID)` drops the matching
  `<Building>` from `farm/buildings`, refusing (with no mutation) if the
  building still houses animals. Exposed via `sdvedit_removeBuilding`; each
  building card has a "Delete Building" button. See `building_factory.go` and
  `loadBuildings`.

### Friendships: can now add a missing NPC

- **Symptom:** `SetFriendships` silently skipped NPCs the player hadn't met,
  and there was no way to create the `<item>` under `friendshipData`.
- **Fix shipped:** `save.AddFriendship(root, npcName)` appends a fresh
  default-zero `<Friendship>` entry (idempotent), backed by a vanilla villager
  list (`npc_catalog.go` / `KnownNPCs`). Exposed via `sdvedit_addFriendship` /
  `sdvedit_knownNPCs`; the Friendships tab has an "Add NPC" picker of villagers
  not yet present. See `accessor.go` and `loadFriendships`.

### Friendships: can now add a pet when none exists

- **Symptom:** `loadFriendships` only rendered the pet section when a pet was
  present; new-game/pet-less saves never saw the form.
- **Fix shipped:** `save.AddPet(root, petType, name, breed)` builds an
  `<NPC xsi:type="Pet">` under `farm/characters` (refusing if one already
  exists), exposed via `sdvedit_addPet`. When no pet is present the tab now
  shows an "Add Pet" form. See `AddPet`/`buildPetNode` and `loadFriendships`.

### Inventory: equipped gear is now surfaced

- **Symptom:** the top-level `<hat>`, `<shirtItem>`, `<pantsItem>`, `<boots>`,
  and `<leftRing>`/`<rightRing>` slots were ignored by the Inventory tab, so
  ring effects, boot bonuses, and clothing colour couldn't be edited.
- **Fix shipped:** `equipment.go` adds `GetEquipment` (six slots, present/absent
  aware), `SetEquipmentField`, `SetEquipmentColor` (keeps `PackedValue`
  consistent via `A<<24|B<<16|G<<8|R`), `ClearEquipmentSlot` (removes the
  element, matching how the game represents an unequipped slot), and
  `AddClothing`. Exposed via `sdvedit_getEquipment` /
  `sdvedit_setEquipmentField` / `sdvedit_setEquipmentColor` /
  `sdvedit_clearEquipmentSlot` / `sdvedit_addClothing`; the Inventory tab grows
  an "Equipped Gear" section. See `internal/save/equipment.go` and `site/app.js`
  `equipmentSectionHtml`/`wireEquipment`.
- **Note / scope:** real 1.6 saves *omit* `hat`/`boots`/`leftRing`/`rightRing`
  entirely when nothing is equipped (they are not `xsi:nil` placeholders), and
  none appear in any reference save, so there is no schema-verified template to
  build one from scratch. Read/edit/clear work for all six slots and creation
  works for clothing (ground-truthed from a real save); **creating a hat, pair
  of boots, or ring from an empty slot is deliberately not implemented** — equip
  one in-game first, then edit it here. See the open issue below.

### Buildings: building type can now be changed

- **Symptom:** the edit form covered tile position and paint but not type;
  changing type in-game requires demolition.
- **Fix shipped:** `save.ChangeBuildingType(root, id, newType, recomputeStructural)`
  sets `<buildingType>` and, when `recomputeStructural` is true, re-derives
  `tilesWide/tilesHigh/maxOccupants/hayCapacity` from `buildingDefs`. While the
  building houses animals the type may only change within the same habitat group
  (barn↔barn, coop↔coop), so a Barn→Big Barn upgrade is allowed but switching to
  a coop or a non-housing type is refused; occupants' `buildingTypeILiveIn` is
  updated to match. Exposed via `sdvedit_changeBuildingType`; each building card
  gains a Type dropdown, a "recompute size/occupants" checkbox, and a "Change
  Type" button. See `internal/save/building_factory.go` and `loadBuildings`.

---

## Open issues

### Equipment: creating a hat, boots, or ring from an empty slot

Read/edit/clear are supported for all six worn-gear slots, and clothing can be
created, but conjuring a brand-new hat/boots/ring node is not. The blocker is
schema fidelity: those slots are absent (not `xsi:nil`) when unequipped, and no
reference save exposes a populated example, so the exact element order the
game's `XmlSerializer` expects can't be verified — a fabricated node risks being
silently dropped or rejected on load.

**Fix sketch:** capture a real populated `<hat>`/`<boots>`/`<leftRing>` from a
save that has them equipped, model `buildHatNode`/`buildBootsNode`/
`buildRingNode` on that exact field order, and add per-type "Add" forms (mirror
`AddClothing`). Verify by loading an edited save in-game, not just round-tripping
through the parser.

### Quests and Bundles tabs are intentionally read-only

Both panels render a hint explaining why
(`Quest editing is read-only — completing quests mid-game can cause script state
issues.` / `Editing bundle completion requires modifying netWorldState/completedBundles
which varies by game version`). These are documented design decisions, not
bugs — listed here so anyone scanning the menus knows they were considered.
