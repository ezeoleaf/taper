package storage

// DocKind identifies a per-race markdown file.
type DocKind string

const (
	DocLog       DocKind = "log"
	DocStrategy  DocKind = "strategy"
	DocPacking   DocKind = "packing"
	DocNutrition DocKind = "nutrition"
	DocGear      DocKind = "gear"
	DocWeather   DocKind = "weather"
	DocCrew      DocKind = "crew"
	DocSplits    DocKind = "splits"
)

// PlanDocs are the planning markdown files shown on the Plan tab.
var PlanDocs = []struct {
	Kind  DocKind
	Title string
	File  string
}{
	{DocStrategy, "Strategy & pacing", "strategy.md"},
	{DocNutrition, "Nutrition plan", "nutrition.md"},
	{DocGear, "Gear choices", "gear.md"},
	{DocWeather, "Weather contingency", "weather.md"},
	{DocCrew, "Crew & support", "crew.md"},
	{DocSplits, "Splits / checkpoints", "splits.md"},
}

// AllDocKinds is every per-race markdown file loaded into the UI bundle.
var AllDocKinds = []DocKind{
	DocLog, DocStrategy, DocPacking,
	DocNutrition, DocGear, DocWeather, DocCrew, DocSplits,
}

func AllDocSeeds() map[string]string {
	return map[string]string{
		"log.md":       defaultLog,
		"strategy.md":  defaultStrategy,
		"packing.md":   defaultPacking,
		"nutrition.md": defaultNutrition,
		"gear.md":      defaultGear,
		"weather.md":   defaultWeather,
		"crew.md":      defaultCrew,
		"splits.md":    defaultSplits,
	}
}

const defaultNutrition = `# Nutrition Plan

## Pre-race

-

## During race

| Time / segment | Fuel | Fluid | Caffeine |
| -------------- | ---- | ----- | -------- |
|                |      |       |          |

## Post-race

-
`

const defaultGear = `# Gear Choices

## Shoes

-

## Kit

-

## Watch / tech

-

## Backup gear

-
`

const defaultWeather = `# Weather Contingency

## Forecast notes

-

## Heat plan

-

## Rain / cold plan

-

## Gear swaps

-
`

const defaultCrew = `# Crew & Support

## Contacts

| Name | Role | Phone |
| ---- | ---- | ----- |
|      |      |       |

## Meetup points

-

## Drop bags / aid

-
`

const defaultSplits = `# Splits / Checkpoints

| Segment | Distance | Target pace | Notes |
| ------- | -------- | ----------- | ----- |
|         |          |             |       |
`
