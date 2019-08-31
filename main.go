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

// Set to true to enable debugging output.
const debug = false

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
	Constraints map[string]constraint
}

var soulsSource = flag.String("soulsdb", "souls.yaml", "A YAML file describing your souls")
var ignoreCrit = flag.Bool("ignore-crit", false, "Ignore crit when calculating damage, useful for fights that negate crit")

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println(`Usage: onmyoji-soul-planner [options] <team.yaml> OR
       onmyoji-soul-planner [options] <shikigami> <main soul> [<attr>=<constraint>]`)
		flag.PrintDefaults()
		os.Exit(0)
	}

	var team []member
	if len(args) > 1 {
		name, mainSoul := args[0], args[1]

		constraints := make(map[string]constraint)
		allowed := map[string]struct{}{"crit": struct{}{}, "spd": struct{}{}}
		for _, arg := range args[2:] {
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

		team = append(team, member{Name: name, Primary: mainSoul, Constraints: constraints})
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
		shiki, ok := onmyoji.Shikigamis[strings.ToLower(place.Name)]
		if !ok {
			log.Fatalf("Unknown shikigami %v", place.Name)
		}
		place.Shikigami = shiki

		// Primary needs to be lower-case when used later. We preserve case until now
		// for clearer error messages.
		rawPrimary := place.Primary
		place.Primary = strings.ToLower(place.Primary)
		if _, ok := onmyoji.SoulTypes[place.Primary]; !ok {
			log.Fatalf("Unknown main soul type %v", rawPrimary)
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
		fmt.Printf("Finding best souls for %v\n", place.Name)
		souls := bestSouls(place, soulsDb)
		out, err := yaml.Marshal(souls)
		if err != nil {
			panic(fmt.Sprintf("Unable to marshal souls %v to yaml: %v", souls, err))
		}
		os.Stderr.Write(out)

		if len(souls) == 0 {
			log.Fatal("Unable to find souls that include 4 of the primary soul and satisfy constraints")
			break
		}

		soulsDb.Slot1 = removeFirst(soulsDb.Slot1, souls[0])
		soulsDb.Slot2 = removeFirst(soulsDb.Slot2, souls[1])
		soulsDb.Slot3 = removeFirst(soulsDb.Slot3, souls[2])
		soulsDb.Slot4 = removeFirst(soulsDb.Slot4, souls[3])
		soulsDb.Slot5 = removeFirst(soulsDb.Slot5, souls[4])
		soulsDb.Slot6 = removeFirst(soulsDb.Slot6, souls[5])
	}
}

func removeFirst(s []onmyoji.Soul, x onmyoji.Soul) []onmyoji.Soul {
	i := find(s, x)
	return remove(s, i)
}

func remove(s []onmyoji.Soul, i int) []onmyoji.Soul {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func find(a []onmyoji.Soul, x onmyoji.Soul) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

func soulCounts(soulSet []onmyoji.Soul) map[string]int {
	counts := make(map[string]int)
	for _, sl := range soulSet {
		counts[strings.ToLower(sl.Type)]++
	}
	return counts
}

func computeCrit(shiki member, soulSet []onmyoji.Soul, types map[string]int) int {
	if *ignoreCrit {
		// Assume crit is worthless
		return 0
	}

	crit := shiki.Crit
	for _, sl := range soulSet {
		crit += sl.Crit
	}

	critSouls := 0
	for name, attr := range onmyoji.SoulTypes {
		if attr == "crit" && types[name] >= 2 {
			critSouls++
		}
	}
	crit += 15 * critSouls
	if crit > 100 {
		crit = 100
	}
	return crit
}

func damage(shiki member, soulSet []onmyoji.Soul, types map[string]int) float64 {
	if types[shiki.Primary] < 4 {
		return 0.0
	}

	// soul and shikigami numbers are stored as ints to simplify input. Convert to percentages here.
	atkbonus := 1.0
	for _, sl := range soulSet {
		atkbonus += float64(sl.AtkBonus) / 100.0
	}

	atkSouls := 0
	for name, attr := range onmyoji.SoulTypes {
		if attr == "atkbonus" && types[name] >= 2 {
			atkSouls++
		}
	}
	atkbonus += 0.15 * float64(atkSouls)

	atk := float64(shiki.Atk) * atkbonus
	for _, sl := range soulSet {
		atk += float64(sl.Atk)
	}
	if debug {
		log.Printf("Attack = %v", atk)
	}

	crit := float64(computeCrit(shiki, soulSet, types)) / 100.0
	if debug {
		log.Printf("Crit = %v", crit)
	}

	critDmg := float64(shiki.CritDmg) / 100.0
	for _, sl := range soulSet {
		critDmg += float64(sl.CritDmg) / 100.0
	}
	if debug {
		log.Printf("CritDmg = %v", critDmg)
	}

	dmg := atk * (crit*critDmg + (1.0 - crit))
	if types["odokuro"] >= 2 {
		dmg *= 1.1
	}
	if types["seductress"] >= 4 {
		dmg += 1.2 * crit * atk
	}
	return dmg
}

func bestSouls(m member, soulsDb onmyoji.SoulDb) []onmyoji.Soul {
	var bestDmg float64
	var finalCrit, finalSpeed int
	var bestSouls []onmyoji.Soul

	for _, sl1 := range soulsDb.Slot1 {
		for _, sl2 := range soulsDb.Slot2 {
			for _, sl3 := range soulsDb.Slot3 {
				for _, sl4 := range soulsDb.Slot4 {
					for _, sl5 := range soulsDb.Slot5 {
						for _, sl6 := range soulsDb.Slot6 {
							souls := []onmyoji.Soul{sl1, sl2, sl3, sl4, sl5, sl6}

							spd := m.Spd
							for _, sl := range souls {
								spd += sl.Spd
							}
							if cons, ok := m.Constraints["spd"]; ok {
								if spd < cons.Low || spd > cons.High {
									continue
								}
							}

							types := soulCounts(souls)
							crit := computeCrit(m, souls, types)
							if cons, ok := m.Constraints["crit"]; ok {
								if crit < cons.Low || crit > cons.High {
									continue
								}
							}

							if dmg := damage(m, souls, types); dmg > bestDmg {
								bestDmg = dmg
								bestSouls = souls
								finalSpeed = spd
								finalCrit = crit
							}
						}
					}
				}
			}
		}
	}

	log.Printf("dmg = %v, speed = %v, crit = %v", bestDmg, finalSpeed, finalCrit)
	return bestSouls
}
