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
	"onikiri": Shikigami{
		HP:       10823,
		Atk:      3350,
		Crit:     11,
		CritDmg:  160,
		Spd:      117,
		Multihit: true,
	},
	"ibaraki doji": Shikigami{
		HP:      10254,
		Atk:     3216,
		Crit:    10,
		CritDmg: 150,
		Spd:     112,
	},
	"ubume": Shikigami{
		HP:       10823,
		Atk:      3082,
		Crit:     10,
		CritDmg:  150,
		Spd:      113,
		Multihit: true,
	},
	"kamikui g5": Shikigami{
		Atk:     1741,
		Crit:    8,
		CritDmg: 150,
		Spd:     118,
	},
	"kamikui": Shikigami{
		HP:      10709,
		Atk:     2894,
		Crit:    8,
		CritDmg: 150,
		Spd:     118,
	},
	"shuten doji": Shikigami{
		HP:       11165,
		Atk:      3136,
		Crit:     10,
		CritDmg:  150,
		Spd:      113,
		Multihit: true,
	},
	"tamamonomae": Shikigami{
		HP:      12532,
		Atk:     3350,
		Crit:    12,
		CritDmg: 160,
		Spd:     110,
	},
	"nekomata": Shikigami{
		Atk:      3002,
		Crit:     10,
		CritDmg:  150,
		Spd:      118,
		Multihit: true,
	},
	"kisei": Shikigami{
		HP:       9912,
		Atk:      3002,
		Crit:     8,
		CritDmg:  150,
		Spd:      106,
		Multihit: true,
	},
	"shiranui": Shikigami{
		HP:      9229,
		Atk:     3457,
		Crit:    10,
		CritDmg: 150,
		Spd:     117,
	},
	"sp ibaraki doji": Shikigami{
		HP:      10254,
		Atk:     3323,
		Crit:    15,
		CritDmg: 150,
		Spd:     112,
	},
	"ryomen": Shikigami{
		HP:       10482,
		Atk:      3136,
		Crit:     10,
		CritDmg:  150,
		Spd:      109,
		Multihit: true,
	},
	"bukkuman": Shikigami{
		HP:      11393,
		Atk:     2680,
		Crit:    8,
		CritDmg: 150,
		Spd:     109,
	},
	"ootengu": Shikigami{
		HP:       10026,
		Atk:      3136,
		Crit:     10,
		CritDmg:  150,
		Spd:      110,
		Multihit: true,
	},
	"kuro": Shikigami{
		HP:       9912,
		Atk:      3377,
		Crit:     9,
		CritDmg:  150,
		Spd:      109,
		Multihit: true,
	},
	"orochi": Shikigami{
		HP:      12418,
		Atk:     4074,
		Crit:    10,
		CritDmg: 150,
		Spd:     118,
	},
	"inuyasha": Shikigami{
		HP:       11393,
		Atk:      2975,
		Crit:     10,
		CritDmg:  150,
		Spd:      114,
		Multihit: true,
	},
	"sp crimson yoto": Shikigami{
		HP:      9912,
		Atk:     3377,
		Crit:    12,
		CritDmg: 150,
		Spd:     111,
	},
	"sp blazing tamamanomae": Shikigami{
		HP:      12532,
		Atk:     3511,
		Crit:    12,
		CritDmg: 160,
		Spd:     115,
	},
	"sp shuten doji": Shikigami{
		HP:      11963,
		Atk:     3189,
		Crit:    10,
		CritDmg: 150,
		Spd:     109,
	},
}

var nicknames = map[string]string{
	"iba":        "ibaraki doji",
	"ibaraki":    "ibaraki doji",
	"shuten":     "shuten doji",
	"oni":        "onikiri",
	"tama":       "tamamonomae",
	"tamamo":     "tamamonomae",
	"sp iba":     "sp ibaraki doji",
	"sp ibaraki": "sp ibaraki doji",
	"sp yoto":    "sp crimson yoto",
	"sp shuten":  "sp shuten doji",
	"sp tama":    "sp blazing tamamanomae",
	"sp tamamo":  "sp blazing tamamanomae",
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
