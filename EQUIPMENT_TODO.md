# Outstanding work: equipment creation (hat / boots / rings)

## Status

The Inventory tab surfaces all six worn-gear slots — `hat`, `shirtItem`,
`pantsItem`, `boots`, `leftRing`, `rightRing`. For every slot we can **read,
edit-in-place, and clear**. We can also **create clothing** (shirt/pants) from
scratch, because a real save gave us the exact node shape.

What's missing: **creating a hat, pair of boots, or ring from an empty slot.**

| Slot type | Read | Edit | Clear | Create from empty |
|-----------|:----:|:----:|:-----:|:-----------------:|
| Shirt / Pants (clothing) | ✅ | ✅ | ✅ | ✅ |
| Boots | ✅ | ✅ | ✅ | ❌ |
| Rings (left/right) | ✅ | ✅ | ✅ | ❌ |
| Hat | ✅ | ✅ | ✅ | ❌ |

Relevant code: `internal/save/equipment.go`, `cmd/sdvedit/main.go`,
`site/app.js` (`equipmentSectionHtml` / `wireEquipment`).

## Why creation is blocked

The save is written by .NET's `XmlSerializer`, which is **order-sensitive** and
rejects unknown or misordered elements. In real Stardew 1.6 saves, an
unequipped hat/boots/ring is **omitted entirely** — there is no `<hat>` element
at all (it is *not* an `xsi:nil` placeholder). To create one we have to emit a
full element from nothing.

We have a ground-truth template for clothing because `shirtItem`/`pantsItem` are
always present in a save. We have **no reference** for hat/boots/ring: neither
local save has any equipped, and they don't appear anywhere in the file. Without
a real example we'd be guessing the exact field set and order, and a wrong guess
produces an item the game silently drops or errors on at load — a bad outcome
for a save editor.

## The fix (once we have a reference)

For each type, mirror what `AddClothing` / `buildClothingNode` already do:

1. Capture a populated `<hat>` / `<boots>` / `<leftRing>` from a real save (see
   below) and record the exact child element order.
2. Add `buildHatNode` / `buildBootsNode` / `buildRingNode` in
   `internal/save/equipment.go`, modelled byte-for-byte on that order.
3. Add `AddHat` / `AddBoots` / `AddRing` accessors (errors if the slot is
   already occupied; appends the element under `<player>`), with parse→serialize
   round-trip tests like `TestAddClothing_FillsEmptySlotAndRoundTrips`.
4. Expose them in `cmd/sdvedit/main.go` (`sdvedit_addHat`, …) and replace the
   "creation not supported" hint in `site/app.js` with per-type Add forms.
5. **Verify in-game**, not just through the parser — round-tripping proves our
   parser is happy, not that Stardew accepts the item.

A small item catalog (a few common hats/boots/rings by ID + name, like the
existing `itemCatalog`) would make the Add form a picker instead of a raw ID
box.

## What you can do to empower this

The one thing blocking us is **a real save with these items equipped.** If you
can do the following in-game and hand back the result, we can finish the
feature:

1. In Stardew Valley, equip your farmer with **a hat, a pair of boots, a ring in
   the left slot, and a different ring in the right slot**, then save and quit.
2. Run this from the repo root — it pulls just the four relevant XML snippets
   out of your latest save (it does **not** copy or upload the whole save):

   ```bash
   python3 - <<'PY'
   import glob, os, re
   saves = glob.glob(os.path.expanduser('~/.config/StardewValley/Saves/*/*'))
   saves = [f for f in saves if os.path.isfile(f) and 'SaveGameInfo' not in f]
   newest = max(saves, key=os.path.getmtime)
   data = open(newest, encoding='utf-8', errors='ignore').read()
   for tag in ('hat', 'boots', 'leftRing', 'rightRing'):
       m = re.search(r'<%s\b[^>]*?>.*?</%s>' % (tag, tag), data) \
           or re.search(r'<%s\b[^>]*?/>' % tag, data)
       print('==== %s ====' % tag)
       print(m.group(0) if m else 'NOT FOUND (still unequipped?)')
       print()
   PY
   ```

3. Paste the output back here (or drop it into a file in the repo). That gives
   us the exact element order to build against.

If grabbing a save isn't convenient, a decompiled reference for the field order
of `StardewValley.Objects.Hat`, `Boots`, and `Ring` (their serialized
`[XmlElement]` members, in declaration order) would also unblock us — but a real
save is the most reliable source.
