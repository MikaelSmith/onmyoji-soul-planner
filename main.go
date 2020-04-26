package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/MikaelSmith/onmyoji-soul-planner/onmyoji"
	"gopkg.in/yaml.v3"
)

type constraint struct {
	Low, High int
}

func parseConstraint(s string) constraint {
	cons := strings.Split(s, "-")
	if len(cons) > 2 {
		log.Fatalf("Illegal constraint %v, must be a number N or range of the form M-N", s)
	}
	if len(cons) == 1 {
		cons = []string{cons[0], cons[0]}
	}
	var err error
	consf := make([]int, 2)
	for i, v := range cons {
		if v == "" {
			// Included a dash but left one end open. Leave that end uninitialized.
			continue
		}

		if consf[i], err = strconv.Atoi(v); err != nil {
			log.Fatalf("%v could not be parsed as a number: %v", cons[0], err)
		}
	}
	return constraint{Low: consf[0], High: consf[1]}
}

type member struct {
	onmyoji.Shikigami
	Name        string
	Primary     string
	Primaries   []string
	Secondary   string
	Secondaries []string
	Optimize    onmyoji.Optimizer
	Constraints map[string]constraint
	Modifiers   onmyoji.Modifiers
}

var soulsSource = flag.String("soulsdb", "souls.yaml", "A YAML file describing your souls")
var ignoreSetBonus = flag.Bool("ignore-set", false, "Ignore the primary set effect when calculating damage")
var atkMod = flag.Int("modify-atk", 0, "Modify attack to account for buffs and/or debuffs")
var atkBonusMod = flag.Int("modify-atkbonus", 0, "Modify attack bonus to account for buffs and/or debuffs")
var critMod = flag.Int("modify-crit", 0, "Modify crit to account for buffs and/or debuffs")
var critDmgMod = flag.Int("modify-critdmg", 0, "Modify crit damage to account for buffs and/or debuffs")
var orbs = flag.Int("orbs", 5, "Specify how many orbs to assume when attacking")

func splitSouls(arg string) []string {
	if len(arg) == 0 {
		return []string{}
	}
	return strings.Split(arg, "|")
}

func main() {
	flag.Usage = func() {
		fmt.Println(`Usage: onmyoji-soul-planner [options] <team.yaml> OR
       onmyoji-soul-planner [options] <shikigami> <main soul> [<secondary soul>] [<attr>=<constraint>]`)
		flag.PrintDefaults()
	}

	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	var team []member
	if len(args) > 1 {
		name, mainSoul, secondarySoul := args[0], args[1], ""

		rem := args[2:]
		if len(rem) > 0 && !strings.Contains(rem[0], "=") {
			secondarySoul = rem[0]
			rem = rem[1:]
		}

		constraints := make(map[string]constraint)
		allowed := map[string]struct{}{"crit": {}, "spd": {}}
		for _, arg := range rem {
			pair := strings.Split(arg, "=")
			if len(pair) != 2 {
				log.Fatalf("Unknown argument %v, must be of the form <attribute>=<range>, such as spd=117-127 or crit=1.0", arg)
			}
			key := strings.ToLower(pair[0])
			if _, ok := allowed[key]; !ok {
				log.Fatalf("Unsupported attribute constraint %v", key)
			}

			constraints[key] = parseConstraint(pair[1])
		}

		team = append(team, member{
			Name:        name,
			Primaries:   splitSouls(mainSoul),
			Secondaries: splitSouls(secondarySoul),
			Constraints: constraints,
		})
	} else {
		source, err := ioutil.ReadFile(args[0])
		if err != nil {
			log.Fatalf("Error reading %v: %v", args[0], err)
		}

		if err := yaml.Unmarshal(source, &team); err != nil {
			log.Fatalf("Error parsing %v: %v", args[0], err)
		}
	}

	for i, place := range team {
		shiki, err := onmyoji.GetShikigami(place.Name)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		place.Shikigami = shiki

		if place.Primary != "" {
			if len(place.Primaries) > 0 {
				log.Fatalf("Shiki %v: only set one of primary or primaries", place.Name)
			}
			place.Primaries = []string{place.Primary}
		}

		for _, primary := range place.Primaries {
			if _, err = onmyoji.SoulSetBonus(primary); err != nil {
				log.Fatalf("Error with primary soul: %v", err)
			}
		}

		if place.Secondary != "" {
			if len(place.Secondaries) > 0 {
				log.Fatalf("Shiki %v: only set one of secondary or secondaries", place.Name)
			}
			place.Secondaries = []string{place.Secondary}
		}

		for _, secondary := range place.Secondaries {
			if _, err = onmyoji.SoulSetBonus(secondary); err != nil {
				log.Fatalf("Error with secondary soul: %v", err)
			}
		}

		if place.Optimize == "" {
			place.Optimize = onmyoji.Damage
		}

		// Update the team member.
		team[i] = place
	}

	source, err := ioutil.ReadFile(*soulsSource)
	if err != nil {
		log.Fatalf("Error reading %v: %v", *soulsSource, err)
	}

	var soulsDb onmyoji.SoulDb
	if err := yaml.Unmarshal(source, &soulsDb); err != nil {
		log.Fatalf("Error parsing %v: %v", *soulsSource, err)
	}

	// After optimizing each member, remove those souls from the db.
	for _, place := range team {
		fmt.Printf("Finding best souls for %v with %v\n", place.Name, strings.Join(place.Primaries, ", "))
		best := bestSouls(place, soulsDb)

		if best.Souls.Empty() {
			log.Fatal("Unable to find souls that include 4 of the primary soul and satisfy constraints")
			break
		}

		fmt.Println(best)
		soulsDb.Remove(best.Souls)
	}
}

func bestSouls(m member, soulsDb onmyoji.SoulDb) onmyoji.Result {
	return soulsDb.BestSet(m.Primaries, m.Secondaries, m.Optimize, func(souls onmyoji.SoulSet) onmyoji.Result {
		spd := m.Spd
		for _, sl := range souls.Souls() {
			spd += sl.Spd
		}
		if cons, ok := m.Constraints["spd"]; ok {
			if (cons.Low > 0 && spd < cons.Low) || (cons.High > 0 && spd > cons.High) {
				return onmyoji.Result{}
			}
		}

		opts := onmyoji.DamageOptions{IgnoreSetBonus: *ignoreSetBonus, Orbs: *orbs}
		mods := applyCliMods(m.Modifiers)
		crit := souls.ComputeCrit(m.Shikigami, mods.Crit)
		if cons, ok := m.Constraints["crit"]; ok {
			if (cons.Low > 0 && crit < cons.Low) || (cons.High > 0 && crit > cons.High) {
				return onmyoji.Result{}
			}
		}

		hp := souls.HP(m.Shikigami, mods)
		dmg := souls.Damage(m.Shikigami, mods, opts)
		heal := souls.Heal(m.Shikigami, mods)
		return onmyoji.Result{HP: hp, Heal: heal, Damage: dmg, Crit: crit, Spd: spd, Souls: souls}
	})
}

func applyCliMods(mods onmyoji.Modifiers) onmyoji.Modifiers {
	mods.Atk += *atkMod
	mods.AtkBonus += *atkBonusMod
	mods.Crit += *critMod
	mods.CritDmg += *critDmgMod
	return mods
}
