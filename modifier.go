package recipe

type ModifierString string

type ModifierSet struct {
	ExecMods map[ModifierString]OpMod
	OpMods   map[ModifierString]OpMod
}
