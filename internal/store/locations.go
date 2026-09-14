package store

import (
	"encoding/json"
	"fmt"
	"os"
)

// Location is a saved, geocoded search result, persisted to locations.json.
type Location struct {
	Query      string  `json:"query"`      // original search string used to find it
	Name       string  `json:"name"`       // commune name, as returned by EFS (nom)
	PostalCode string  `json:"postalCode"` // codePostal
	Normalized string  `json:"normalized"` // used as the .ics filename stem
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
}

const DefaultPath = "locations.json"

func Load(path string) ([]Location, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var locs []Location
	if err := json.Unmarshal(data, &locs); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return locs, nil
}

func Save(path string, locs []Location) error {
	data, err := json.MarshalIndent(locs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal locations: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Upsert adds loc, replacing any existing entry with the same Normalized name.
func Upsert(locs []Location, loc Location) []Location {
	for i, l := range locs {
		if l.Normalized == loc.Normalized {
			locs[i] = loc
			return locs
		}
	}
	return append(locs, loc)
}
