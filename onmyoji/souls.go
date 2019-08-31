package onmyoji

// SoulTypes map the name of souls to their 2-soul attribute bonus.
var SoulTypes = map[string]string{
	"harpy":              "atkbonus",
	"watcher":            "atkbonus",
	"house imp":          "atkbonus",
	"scarlet":            "atkbonus",
	"soultaker":          "atkbonus",
	"nightwing":          "atkbonus",
	"kyoukotsu":          "atkbonus",
	"tomb guard":         "crit",
	"shadow":             "crit",
	"fenikkusu":          "crit",
	"claws":              "crit",
	"samisen":            "crit",
	"seductress":         "crit",
	"namazu":             "",
	"odokuro":            "",
	"tsuchigumo":         "",
	"ghostly songstress": "",
}

// Soul contains the name of the soul and stats relevant to damage output.
type Soul struct {
	Type                              string
	Atk, AtkBonus, Crit, CritDmg, Spd int
}

// SoulDb represents all your souls.
type SoulDb struct {
	Slot1, Slot2, Slot3, Slot4, Slot5, Slot6 []Soul
}
