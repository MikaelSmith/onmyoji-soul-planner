package onmyoji

import (
	"fmt"
	"strconv"
	"strings"
)

// Optimizer represents what to optimize for.
type Optimizer string

// Constants for selecting what to optimize.
const (
	Damage Optimizer = "Damage"
	HP               = "HP"
	Heal             = "Heal"
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
	"tree spirit":        "hpbonus",
	"soul edge":          "hpbonus",
	"priestess":          "hpbonus",
	"mirror lady":        "hpbonus",
	"boroboroton":        "hpbonus",
	"jizo statue":        "hpbonus",
	"holy flame":         "hpbonus",
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
	Type                                           string
	Atk, AtkBonus, Crit, CritDmg, Spd, HP, HPBonus int
}

func (s Soul) String() string {
	attrs := make([]string, 0, 5)
	if s.HP > 0 {
		attrs = append(attrs, "HP="+strconv.Itoa(s.HP))
	}
	if s.HPBonus > 0 {
		attrs = append(attrs, "HPBonus="+strconv.Itoa(s.HPBonus))
	}
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
	Damage, Heal, HP, Crit, Spd int
	Souls                       SoulSet
}

func (r Result) String() string {
	return fmt.Sprintf("dmg = %v, heal = %v, hp = %v, speed = %v, crit = %v\n%v", r.Damage, r.Heal, r.HP, r.Spd, r.Crit, r.Souls)
}

func contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// BestSet constructs a SoulSet for each combination of souls in the database. It calls the fitness
// function on each set that includes at least 4 of the primary soul (if primary is not an empty
// string). It returns the best set.
func (db *SoulDb) BestSet(primaries, secondaries []string, opt Optimizer, fn func(SoulSet) Result) Result {
	candidates := make(chan Result)

	match := func(typ string, prim, sec int) (int, int) {
		if contains(primaries, typ) {
			prim++
		}
		if contains(secondaries, typ) {
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
					if len(primaries) > 0 && primary3 == 0 {
						continue
					}

					for _, sl4 := range db.Slot4 {
						primary4, secondary4 := match(sl4.Type, primary3, secondary3)
						if len(primaries) > 0 && primary4 <= 1 {
							continue
						}

						for _, sl5 := range db.Slot5 {
							primary5, secondary5 := match(sl5.Type, primary4, secondary4)
							if (len(primaries) > 0 && primary5 <= 2) || (len(secondaries) > 0 && secondary5 == 0) {
								continue
							}

							for _, sl6 := range db.Slot6 {
								primary6, secondary6 := match(sl6.Type, primary5, secondary5)
								if (len(primaries) > 0 && primary6 <= 3) || (len(secondaries) > 0 && secondary6 <= 1) {
									continue
								}

								r := fn(NewSoulSet([6]Soul{sl1, sl2, sl3, sl4, sl5, sl6}))
								if (opt == Damage && r.Damage > best.Damage) ||
									(opt == Heal && r.Heal > best.Heal) ||
									(opt == HP && r.HP > best.HP) {
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
	CritMod        int
	IgnoreSetBonus bool
	Orbs           int
}

// ComputeCrit returns the critical hit chance of the shikigami with this soul set.
func (set SoulSet) ComputeCrit(shiki Shikigami, critMod int, opts DamageOptions) int {
	crit := shiki.Crit + critMod + opts.CritMod
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

	if crit > 100 {
		crit = 100
	}
	return crit
}

// Damage computes the shikigami's damage output with this soul set.
func (set SoulSet) Damage(shiki Shikigami, mod Modifiers, opts DamageOptions) int {
	// soul and shikigami numbers are stored as ints to simplify input. Convert to percentages here.
	atkbonus := 1.0 + float64(mod.AtkBonus)/100.0
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

	crit := float64(set.ComputeCrit(shiki, mod.Crit, opts)) / 100.0

	critDmg := float64(shiki.CritDmg) / 100.0
	for _, sl := range set.Souls() {
		critDmg += float64(sl.CritDmg) / 100.0
	}

	dmg := atk * (crit*critDmg + (1.0 - crit))
	if set.Count("Odokuro") >= 2 {
		dmg *= 1.1
	}
	if shiki.Multihit && set.Count("Ghostly Songstress") >= 2 {
		// Every 6th hit deals extra 255% of Atk (up to 20% of target's max HP).
		dmg += (2.55 * atk) / 6
	}
	if !opts.IgnoreSetBonus {
		if set.Count("Seductress") >= 4 {
			dmg += 1.2 * crit * atk
		} else if set.Count("Shadow") >= 4 || set.Count("Watcher") >= 4 {
			dmg *= 1.4
		} else if set.Count("Kyoukotsu") >= 4 {
			dmg *= (1.0 + 0.08*float64(opts.Orbs))
		}
	}
	return int(dmg)
}

// Heal returns the healing prowess of the shikigami, evaluated as HP * Crit * CritDmg
func (set SoulSet) Heal(shiki Shikigami, mod Modifiers, opts DamageOptions) int {
	hp := set.HP(shiki, mod)

	crit := float64(set.ComputeCrit(shiki, mod.Crit, opts)) / 100.0

	critDmg := float64(shiki.CritDmg) / 100.0
	for _, sl := range set.Souls() {
		critDmg += float64(sl.CritDmg) / 100.0
	}

	heal := float64(hp) * (crit*critDmg + (1.0 - crit))
	return int(heal)
}

// HP returns the shikigami's HP with this soul set.
func (set SoulSet) HP(shiki Shikigami, mod Modifiers) int {
	// soul and shikigami numbers are stored as ints to simplify input. Convert to percentages here.
	hpbonus := 1.0 + float64(mod.HPBonus)/100.0
	for _, sl := range set.Souls() {
		hpbonus += float64(sl.HPBonus) / 100.0
	}

	hpSouls := 0
	for name, attr := range soulTypes {
		if attr == "hpbonus" && set.Count(name) >= 2 {
			hpSouls++
		}
	}
	hpbonus += 0.15 * float64(hpSouls)

	hp := float64(shiki.HP) * hpbonus
	for _, sl := range set.Souls() {
		hp += float64(sl.HP)
	}
	return int(hp)
}

func (set SoulSet) String() string {
	var out string
	for i, soul := range set.souls {
		out += "Slot " + strconv.Itoa(i+1) + ": " + soul.String() + "\n"
	}
	return out
}

// Modifiers contains modifications to specific stats.
type Modifiers struct {
	Crit, AtkBonus, HPBonus int
}
