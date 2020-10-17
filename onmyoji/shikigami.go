package onmyoji

import (
	"fmt"
	"strings"
)

// Shikigami encapsulates a shikigami's damage-related attributes.
type Shikigami struct {
	HP, Atk, Spd, Crit, CritDmg int
	Multihit                    bool
}

// Shikigamis lists the stats for a variety of shikigami.
var shikigamis = map[string]Shikigami{
	"onikiri": {
		HP:       10823,
		Atk:      3350,
		Crit:     11,
		CritDmg:  160,
		Spd:      117,
		Multihit: true,
	},
	"ibaraki doji": {
		HP:      10254,
		Atk:     3216,
		Crit:    10,
		CritDmg: 150,
		Spd:     112,
	},
	"ubume": {
		HP:       10823,
		Atk:      3082,
		Crit:     10,
		CritDmg:  150,
		Spd:      113,
		Multihit: true,
	},
	"kamikui g5": {
		Atk:     1741,
		Crit:    8,
		CritDmg: 150,
		Spd:     118,
	},
	"kamikui": {
		HP:      10709,
		Atk:     2894,
		Crit:    8,
		CritDmg: 150,
		Spd:     118,
	},
	"shuten doji": {
		HP:       11165,
		Atk:      3136,
		Crit:     10,
		CritDmg:  150,
		Spd:      113,
		Multihit: true,
	},
	"tamamonomae": {
		HP:      12532,
		Atk:     3350,
		Crit:    12,
		CritDmg: 160,
		Spd:     110,
	},
	"nekomata": {
		Atk:      3002,
		Crit:     10,
		CritDmg:  150,
		Spd:      118,
		Multihit: true,
	},
	"kisei": {
		HP:       9912,
		Atk:      3002,
		Crit:     8,
		CritDmg:  150,
		Spd:      106,
		Multihit: true,
	},
	"shiranui": {
		HP:      9229,
		Atk:     3457,
		Crit:    10,
		CritDmg: 150,
		Spd:     117,
	},
	"sp ibaraki doji": {
		HP:      10254,
		Atk:     3323,
		Crit:    15,
		CritDmg: 150,
		Spd:     112,
	},
	"ryomen": {
		HP:       10482,
		Atk:      3136,
		Crit:     10,
		CritDmg:  150,
		Spd:      109,
		Multihit: true,
	},
	"bukkuman": {
		HP:      11393,
		Atk:     2680,
		Crit:    8,
		CritDmg: 150,
		Spd:     109,
	},
	"ootengu": {
		HP:       10026,
		Atk:      3136,
		Crit:     10,
		CritDmg:  150,
		Spd:      110,
		Multihit: true,
	},
	"kuro": {
		HP:       9912,
		Atk:      3377,
		Crit:     9,
		CritDmg:  150,
		Spd:      109,
		Multihit: true,
	},
	"orochi": {
		HP:      12418,
		Atk:     4074,
		Crit:    10,
		CritDmg: 150,
		Spd:     118,
	},
	"inuyasha": {
		HP:       11393,
		Atk:      2975,
		Crit:     10,
		CritDmg:  150,
		Spd:      114,
		Multihit: true,
	},
	"sp crimson yoto": {
		HP:      9912,
		Atk:     3377,
		Crit:    12,
		CritDmg: 150,
		Spd:     111,
	},
	"sp blazing tamamanomae": {
		HP:       12532,
		Atk:      3511,
		Crit:     12,
		CritDmg:  160,
		Spd:      115,
		Multihit: true,
	},
	"sp shuten doji": {
		HP:      11963,
		Atk:     3189,
		Crit:    10,
		CritDmg: 150,
		Spd:     109,
	},
	"ushi no toki g5": {
		HP:      7963,
		Atk:     1741,
		Crit:    10,
		CritDmg: 150,
		Spd:     117,
	},
	"ushi no toki": {
		HP:       11165,
		Atk:      2894,
		Crit:     10,
		CritDmg:  150,
		Spd:      117,
		Multihit: true,
	},
	"suzuka gozen": {
		HP:       13216,
		Atk:      3270,
		Crit:     10,
		CritDmg:  150,
		Spd:      110,
		Multihit: true,
	},
	"takiyashahime": {
		HP:       10026,
		Atk:      3511,
		Crit:     10,
		CritDmg:  150,
		Spd:      120,
		Multihit: true,
	},
}

var nicknames = map[string]string{
	"iba":            "ibaraki doji",
	"ibaraki":        "ibaraki doji",
	"shuten":         "shuten doji",
	"oni":            "onikiri",
	"tama":           "tamamonomae",
	"tamamo":         "tamamonomae",
	"sp iba":         "sp ibaraki doji",
	"sp ibaraki":     "sp ibaraki doji",
	"sp yoto":        "sp crimson yoto",
	"sp shuten":      "sp shuten doji",
	"sp tama":        "sp blazing tamamanomae",
	"sp tamamo":      "sp blazing tamamanomae",
	"sp tamamonomae": "sp blazing tamamanomae",
	"ushi":           "ushi no toki",
	"suzuka":         "suzuka gozen",
	"taki":           "takiyashahime",
}

// GetShikigami returns attributes for the named shikigami.
func GetShikigami(name string) (Shikigami, error) {
	name = strings.ToLower(name)
	if nick, ok := nicknames[name]; ok {
		name = nick
	}

	if shiki, ok := shikigamis[name]; ok {
		return shiki, nil
	}
	return Shikigami{}, fmt.Errorf("unknown shikigami %v", name)
}
