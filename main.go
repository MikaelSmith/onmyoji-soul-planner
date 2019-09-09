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
	Constraints map[string]constraint
}

var soulsSource = flag.String("soulsdb", "souls.yaml", "A YAML file describing your souls")
var ignoreCrit = flag.Bool("ignore-crit", false, "Ignore crit when calculating damage, useful for fights that negate crit")
var ignoreSeduc = flag.Bool("ignore-seductress", false, "Ignore the seductress set effect when calculating damage")
var yellowImp = flag.Bool("yellow-imp", false, "Add 15% crit because you plan to use Yellow Imp's boost")

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
		shiki, err := onmyoji.GetShikigami(place.Name)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		place.Shikigami = shiki

		if place.Primary != "" {
			if _, err = onmyoji.SoulSetBonus(place.Primary); err != nil {
				log.Fatalf("Error with primary soul: %v", err)
			}
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

		if souls.Empty() {
			log.Fatal("Unable to find souls that include 4 of the primary soul and satisfy constraints")
			break
		}

		fmt.Println(souls)
		soulsDb.Remove(souls)
	}
}

func bestSouls(m member, soulsDb onmyoji.SoulDb) onmyoji.SoulSet {
	best := soulsDb.BestSet(m.Primary, func(souls onmyoji.SoulSet) onmyoji.Result {
		spd := m.Spd
		for _, sl := range souls.Souls() {
			spd += sl.Spd
		}
		if cons, ok := m.Constraints["spd"]; ok {
			if (cons.Low > 0 && spd < cons.Low) || (cons.High > 0 && spd > cons.High) {
				return onmyoji.Result{}
			}
		}

		opts := onmyoji.DamageOptions{IgnoreCrit: *ignoreCrit, IgnoreSeductress: *ignoreSeduc, YellowImp: *yellowImp}
		crit := souls.ComputeCrit(m.Shikigami, opts)
		if cons, ok := m.Constraints["crit"]; ok {
			if (cons.Low > 0 && crit < cons.Low) || (cons.High > 0 && crit > cons.High) {
				return onmyoji.Result{}
			}
		}

		dmg := souls.Damage(m.Shikigami, opts)
		return onmyoji.Result{Damage: dmg, Crit: crit, Spd: spd, Souls: souls}
	})

	log.Printf("dmg = %v, speed = %v, crit = %v", best.Damage, best.Spd, best.Crit)
	return best.Souls
}
