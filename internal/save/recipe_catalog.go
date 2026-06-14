package save

// Vanilla recipe catalogs. The names here are the dictionary keys Stardew
// Valley stores under player/cookingRecipes and player/craftingRecipes — a
// recipe is "known" when its key is present, regardless of times-made count.
//
// Like itemCatalog, these lists cover the base-game (vanilla) recipes. Any
// recipe a save already knows is preserved by AddRecipe even if it's not
// listed here, so an unlisted/modded recipe is never dropped.

var cookingRecipeCatalog = []string{
	"Fried Egg", "Omelet", "Salad", "Cheese Cauliflower", "Baked Fish",
	"Parsnip Soup", "Vegetable Medley", "Complete Breakfast", "Fried Calamari",
	"Strange Bun", "Lucky Lunch", "Fried Mushroom", "Pizza", "Bean Hotpot",
	"Glazed Yams", "Carp Surprise", "Hashbrowns", "Pancakes", "Salmon Dinner",
	"Fish Taco", "Crispy Bass", "Pepper Poppers", "Bread", "Tom Kha Soup",
	"Trout Soup", "Chocolate Cake", "Pink Cake", "Rhubarb Pie", "Cookie",
	"Spaghetti", "Fried Eel", "Spicy Eel", "Sashimi", "Maki Roll", "Tortilla",
	"Red Plate", "Eggplant Parmesan", "Rice Pudding", "Ice Cream",
	"Blueberry Tart", "Autumn's Bounty", "Pumpkin Soup", "Super Meal",
	"Cranberry Sauce", "Stuffing", "Farmer's Lunch", "Survival Burger",
	"Dish O' The Sea", "Miner's Treat", "Roots Platter", "Triple Shot Espresso",
	"Seafoam Pudding", "Algae Soup", "Pale Broth", "Plum Pudding",
	"Artichoke Dip", "Stir Fry", "Roasted Hazelnuts", "Pumpkin Pie",
	"Radish Salad", "Fruit Salad", "Blackberry Cobbler", "Cranberry Candy",
	"Bruschetta", "Coleslaw", "Fiddlehead Risotto", "Poppyseed Muffin",
	"Chowder", "Fish Stew", "Escargot", "Lobster Bisque", "Maple Bar",
	"Crab Cakes", "Shrimp Cocktail", "Ginger Ale", "Banana Pudding",
	"Mango Sticky Rice", "Poi", "Tropical Curry", "Squid Ink Ravioli",
}

var craftingRecipeCatalog = []string{
	// explosives
	"Cherry Bomb", "Bomb", "Mega Bomb",
	// fences & gate
	"Gate", "Wood Fence", "Stone Fence", "Iron Fence", "Hardwood Fence",
	// sprinklers
	"Sprinkler", "Quality Sprinkler", "Iridium Sprinkler",
	// fishing tackle
	"Spinner", "Trap Bobber", "Cork Bobber", "Treasure Hunter",
	"Dressed Spinner", "Barbed Hook", "Magnet", "Bait", "Wild Bait",
	"Magic Bait", "Quality Bobber", "Crab Pot",
	// music blocks
	"Drum Block", "Flute Block",
	// storage & machines
	"Chest", "Furnace", "Bee House", "Cask", "Cheese Press", "Keg", "Loom",
	"Mayonnaise Machine", "Oil Maker", "Preserves Jar", "Recycling Machine",
	"Worm Bin", "Seed Maker", "Slime Incubator", "Slime Egg-Press",
	"Crystalarium", "Charcoal Kiln", "Lightning Rod", "Solar Panel", "Tapper",
	"Heavy Tapper", "Geode Crusher", "Ostrich Incubator", "Deconstructor",
	"Mini-Obelisk", "Garden Pot", "Mini-Jukebox",
	// signs & scarecrows
	"Wood Sign", "Stone Sign", "Dark Sign", "Scarecrow", "Deluxe Scarecrow",
	"Rarecrow",
	// lighting
	"Torch", "Campfire", "Wooden Brazier", "Stone Brazier", "Gold Brazier",
	"Carved Brazier", "Stump Brazier", "Barrel Brazier", "Skull Brazier",
	"Marble Brazier", "Wood Lamp-post", "Iron Lamp-post", "Jack-O-Lantern",
	// decoration
	"Tub o' Flowers", "Wicked Statue", "Tea Sapling",
	// totems
	"Warp Totem: Farm", "Warp Totem: Mountains", "Warp Totem: Beach",
	"Warp Totem: Desert", "Warp Totem: Island", "Rain Totem",
	// consumables & fertilizer
	"Field Snack", "Life Elixir", "Oil of Garlic", "Monster Musk",
	"Fairy Dust", "Basic Fertilizer", "Quality Fertilizer", "Deluxe Fertilizer",
	"Speed-Gro", "Deluxe Speed-Gro", "Hyper Speed-Gro", "Basic Retaining Soil",
	"Quality Retaining Soil", "Deluxe Retaining Soil", "Tree Fertilizer",
	"Wild Seeds (Sp)", "Wild Seeds (Su)", "Wild Seeds (Fa)", "Wild Seeds (Wi)",
	"Fiber Seeds", "Grass Starter",
	// flooring & paths
	"Wood Floor", "Rustic Plank Floor", "Straw Floor", "Weathered Floor",
	"Crystal Floor", "Stone Floor", "Stone Walkway Floor", "Brick Floor",
	"Wood Path", "Gravel Path", "Cobblestone Path", "Stepping Stone Path",
	"Crystal Path",
}

// KnownCookingRecipes returns the vanilla cooking recipe names.
func KnownCookingRecipes() []string {
	out := make([]string, len(cookingRecipeCatalog))
	copy(out, cookingRecipeCatalog)
	return out
}

// KnownCraftingRecipes returns the vanilla crafting recipe names.
func KnownCraftingRecipes() []string {
	out := make([]string, len(craftingRecipeCatalog))
	copy(out, craftingRecipeCatalog)
	return out
}
