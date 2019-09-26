package onmyoji

import (
	"fmt"
	"strconv"
	"strings"
)

// soulTypes map the name of souls to their 2-soul attribute bonus.
var soulTypes = map[string]string{
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

// SoulSetBonus returns the 2-soul attribute bonus for a set.
func SoulSetBonus(name string) (string, error) {
	if bonus, ok := soulTypes[strings.ToLower(name)]; ok {
		return bonus, nil
	}
	return "", fmt.Errorf("unknown soul type %v", name)
}

// Soul contains the name of the soul and stats relevant to damage output.
type Soul struct {
	Type                              string
	Atk, AtkBonus, Crit, CritDmg, Spd int
}

func (s Soul) String() string {
	attrs := make([]string, 0, 5)
	if s.Atk > 0 {
		attrs = append(attrs, "Atk="+strconv.Itoa(s.Atk))
	}
	if s.AtkBonus > 0 {
		attrs = append(attrs, "AtkBonus="+strconv.Itoa(s.AtkBonus)+"%")
	}
	if s.Crit > 0 {
		attrs = append(attrs, "Crit="+strconv.Itoa(s.Crit)+"%")
	}
	if s.CritDmg > 0 {
		attrs = append(attrs, "CritDmg="+strconv.Itoa(s.CritDmg)+"%")
	}
	if s.Spd > 0 {
		attrs = append(attrs, "Spd="+strconv.Itoa(s.Spd))
	}
	return s.Type + " | " + strings.Join(attrs, ", ")
}

// SoulDb represents all your souls.
type SoulDb struct {
	Slot1, Slot2, Slot3, Slot4, Slot5, Slot6 []Soul
}

// Result contains the outcome of applying a soulset to a shikigami.
type Result struct {
	Damage    float64
	Crit, Spd int
	Souls     SoulSet
}

// BestSet constructs a SoulSet for each combination of souls in the database. It calls the fitness
// function on each set that includes at least 4 of the primary soul (if primary is not an empty
// string). It returns the best set.
func (db *SoulDb) BestSet(primary, secondary string, fn func(SoulSet) Result) Result {
	candidates := make(chan Result)

	match := func(typ string, prim, sec int) (int, int) {
		if typ == primary {
			prim++
		}
		if typ == secondary {
			sec++
		}
		return prim, sec
	}

	for _, sl1 := range db.Slot1 {
		primary1, secondary1 := match(sl1.Type, 0, 0)

		go func(sl1 Soul) {
			var best Result
			for _, sl2 := range db.Slot2 {
				primary2, secondary2 := match(sl2.Type, primary1, secondary1)

				for _, sl3 := range db.Slot3 {
					primary3, secondary3 := match(sl3.Type, primary2, secondary2)
					// Starting once we have 3 souls, test that we have sufficient copies of the primary soul
					// type to complete a set of 4. If not, skip this set of combinations.
					if primary != "" && primary3 == 0 {
						continue
					}

					for _, sl4 := range db.Slot4 {
						primary4, secondary4 := match(sl4.Type, primary3, secondary3)
						if primary != "" && primary4 <= 1 {
							continue
						}

						for _, sl5 := range db.Slot5 {
							primary5, secondary5 := match(sl5.Type, primary4, secondary4)
							if (primary != "" && primary5 <= 2) || (secondary != "" && secondary5 == 0) {
								continue
							}

							for _, sl6 := range db.Slot6 {
								primary6, secondary6 := match(sl6.Type, primary5, secondary5)
								if (primary != "" && primary6 <= 3) || (secondary != "" && secondary6 <= 1) {
									continue
								}

								r := fn(NewSoulSet([6]Soul{sl1, sl2, sl3, sl4, sl5, sl6}))
								if r.Damage > best.Damage {
									best = r
								}
							}
						}
					}
				}
			}
			candidates <- best
		}(sl1)
	}

	var best Result
	for i := 0; i < len(db.Slot1); i++ {
		r := <-candidates
		if r.Damage > best.Damage {
			best = r
		}
	}
	close(candidates)
	return best
}

// Remove all souls in the SoulSet from the database.
func (db *SoulDb) Remove(set SoulSet) {
	db.Slot1 = removeFirst(db.Slot1, set.souls[0])
	db.Slot2 = removeFirst(db.Slot2, set.souls[1])
	db.Slot3 = removeFirst(db.Slot3, set.souls[2])
	db.Slot4 = removeFirst(db.Slot4, set.souls[3])
	db.Slot5 = removeFirst(db.Slot5, set.souls[4])
	db.Slot6 = removeFirst(db.Slot6, set.souls[5])
}

func removeFirst(s []Soul, x Soul) []Soul {
	i := find(s, x)
	return remove(s, i)
}

func remove(s []Soul, i int) []Soul {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func find(a []Soul, x Soul) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

// SoulSet represents a set of 6 souls, slots 1-6.
type SoulSet struct {
	souls  [6]Soul
	counts map[string]int
}

// NewSoulSet constructs a new soul set and computes counts of different soul types.
func NewSoulSet(souls [6]Soul) SoulSet {
	return SoulSet{souls: souls, counts: soulCounts(souls[:])}
}

func soulCounts(soulSet []Soul) map[string]int {
	counts := make(map[string]int)
	for _, sl := range soulSet {
		counts[strings.ToLower(sl.Type)]++
	}
	return counts
}

// Empty returns true if the set has no souls.
func (set SoulSet) Empty() bool {
	return set.souls == [6]Soul{}
}

// Souls returns the list of souls
func (set SoulSet) Souls() [6]Soul {
	return set.souls
}

// Count returns the count of a particular soul type in the set.
func (set SoulSet) Count(name string) int {
	return set.counts[strings.ToLower(name)]
}

// DamageOptions is used to pass options that change how damage is calculated.
type DamageOptions struct {
	IgnoreCrit, IgnoreSeductress, YellowImp bool
}

// ComputeCrit returns the critical hit chance of the shikigami with this soul set.
func (set SoulSet) ComputeCrit(shiki Shikigami, critMod int, opts DamageOptions) int {
	crit := shiki.Crit + critMod
	for _, sl := range set.Souls() {
		crit += sl.Crit
	}

	critSouls := 0
	for name, attr := range soulTypes {
		if attr == "crit" && set.Count(name) >= 2 {
			critSouls++
		}
	}
	crit += 15 * critSouls

	if opts.YellowImp {
		crit += 15
	}

	if crit > 100 {
		crit = 100
	}
	return crit
}

// Damage computes the shikigami's damage output with this soul set.
func (set SoulSet) Damage(shiki Shikigami, mod Modifiers, opts DamageOptions) float64 {
	// soul and shikigami numbers are stored as ints to simplify input. Convert to percentages here.
	atkbonus := 1.0
	for _, sl := range set.Souls() {
		atkbonus += float64(sl.AtkBonus) / 100.0
	}

	atkSouls := 0
	for name, attr := range soulTypes {
		if attr == "atkbonus" && set.Count(name) >= 2 {
			atkSouls++
		}
	}
	atkbonus += 0.15 * float64(atkSouls)

	atk := float64(shiki.Atk) * atkbonus
	for _, sl := range set.Souls() {
		atk += float64(sl.Atk)
	}

	crit := 0.0
	if !opts.IgnoreCrit {
		crit = float64(set.ComputeCrit(shiki, mod.Crit, opts)) / 100.0
	}

	critDmg := float64(shiki.CritDmg) / 100.0
	for _, sl := range set.Souls() {
		critDmg += float64(sl.CritDmg) / 100.0
	}

	dmg := atk * (crit*critDmg + (1.0 - crit))
	if set.Count("Odokuro") >= 2 {
		dmg *= 1.1
	}
	if set.Count("Seductress") >= 4 && !opts.IgnoreSeductress {
		dmg += 1.2 * crit * atk
	}
	return dmg
}

func (set SoulSet) String() string {
	var out string
	for i, soul := range set.souls {
		out += "Slot " + strconv.Itoa(i+1) + ": " + soul.String() + "\n"
	}
	return out
}

type Modifiers struct {
	Crit int
}
