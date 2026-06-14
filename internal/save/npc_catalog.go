package save

// vanillaNPCs is the list of giftable/befriendable villagers in the base game.
// It drives the "add a missing NPC" UI; a save that already knows an NPC keeps
// its entry untouched, and unlisted/modded NPCs already present are preserved.
var vanillaNPCs = []string{
	"Abigail", "Alex", "Caroline", "Clint", "Demetrius", "Dwarf", "Elliott",
	"Emily", "Evelyn", "George", "Gus", "Haley", "Harvey", "Jas", "Jodi",
	"Kent", "Krobus", "Leah", "Leo", "Lewis", "Linus", "Marnie", "Maru",
	"Pam", "Penny", "Pierre", "Robin", "Sam", "Sandy", "Sebastian", "Shane",
	"Vincent", "Willy", "Wizard",
}

// KnownNPCs returns the vanilla villager names known to the editor.
func KnownNPCs() []string {
	out := make([]string, len(vanillaNPCs))
	copy(out, vanillaNPCs)
	return out
}
