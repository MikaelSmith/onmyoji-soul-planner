# Onmyoji Soul Planner

Build with `go build`. This will create a binary called `onmyoji-soul-planner`.

To use the planner, you must first create a souls database. An example is provided in [examples/souls.yaml](examples/souls.yaml). By default `onmyoji-soul-planner` will look for `souls.yaml` in the directory where you run it. You can change that by supplying the `-soulsdb` option.

The souls database has 6 keys - `slot1-6` - that map to arrays of souls. Each soul must have a `type`, and can have any of `atk`, `atkbonus`, `crit`, `critdmg`, `spd`. Other attributes are currently ignored.

> Note that this tool just checks all combinations of souls. So it gets slower the more souls you add to the souls database.

## Solo

You can select a set of souls for a single Shikigami with
```
onmyoji-soul-planner [options] <shikigami> <main soul> [spd=<constraint>] [crit=<constraint>]
```
Constraints are an exact integer number or a range, such as `95-100`. For example
```
onmyoji-soul-planner Onikiri Seductress spd=117-127
```
would optimize for average damage while going after Seimei, while
```
onmyoji-soul-planner "Ibaraki Doji" Shadow spd=125-128 crit=99-100
```
would be a good selection for the Ibaraki Doji + Kamikui Souls 10 team.

## Team

You can also supply a file describing a whole team to optimize
```
onmyoji-soul-planner [options] <team.yaml>
```
You can try this out with [examples/team.yaml](examples/team.yaml) using
```
onmyoji-soul-planner -soulsdb examples/souls.yaml examples/team.yaml
```

## Options

* *-ignore-crit*: Ignore crit when calculating damage, useful for fights that negate crit
* *-soulsdb string*: A YAML file describing your souls (default "souls.yaml")
