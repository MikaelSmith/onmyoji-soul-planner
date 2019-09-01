package onmyoji

import (
	"fmt"
	"strings"
)

// Shikigami encapsulates a shikigami's damage-related attributes.
type Shikigami struct {
	Atk, Spd, Crit, CritDmg int
}

// Shikigamis lists the stats for a variety of shikigami.
var shikigamis = map[string]Shikigami{
	"onikiri": Shikigami{
		Atk:     3350,
		Crit:    11,
		CritDmg: 160,
		Spd:     117,
	},
	"ibaraki doji": Shikigami{
		Atk:     3216,
		Crit:    10,
		CritDmg: 150,
		Spd:     112,
	},
	"ubume": Shikigami{
		Atk:     3082,
		Crit:    10,
		CritDmg: 150,
		Spd:     113,
	},
	"kamikui": Shikigami{
		Atk:     1741,
		Crit:    8,
		CritDmg: 150,
		Spd:     118,
	},
	"shuten doji": Shikigami{
		Atk:     3136,
		Crit:    10,
		CritDmg: 150,
		Spd:     113,
	},
}

var nicknames = map[string]string{
	"ibaraki": "ibaraki doji",
	"iba":     "ibaraki doji",
	"shuten":  "shuten doji",
	"oni":     "onikiri",
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
