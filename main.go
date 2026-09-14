package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"efs-ical/internal/efsapi"
	"efs-ical/internal/ical"
	"efs-ical/internal/normalize"
	"efs-ical/internal/sitegen"
	"efs-ical/internal/store"
)

const outputDir = "dist"

// jitterDelay returns a random delay in [800ms, 2500ms), so multi-location
// builds don't hit the endpoint at a suspiciously constant cadence.
func jitterDelay() time.Duration {
	const minMs, maxMs = 800, 2500
	return time.Duration(minMs+rand.IntN(maxMs-minMs)) * time.Millisecond
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "geocode":
		err = runGeocode(os.Args[2:])
	case "build":
		err = runBuild(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  efs-ical geocode <city or zip code>
  efs-ical build [normalized-name ...]   (omit to rebuild every saved location)`)
}

func runGeocode(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("geocode needs a search string, e.g. `efs-ical geocode 44130`")
	}
	search := strings.Join(args, " ")

	client := &http.Client{Timeout: 15 * time.Second}
	results, err := efsapi.Geocode(client, search)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no matches for %q", search)
	}

	for i, r := range results {
		fmt.Printf("%2d) %-30s %-8s (%.5f, %.5f)\n", i+1, r.Nom, r.CodePostal, r.Lat, r.Lon)
	}

	fmt.Print("Select a location (number): ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read selection: %w", err)
	}
	idx, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || idx < 1 || idx > len(results) {
		return fmt.Errorf("invalid selection %q", strings.TrimSpace(line))
	}
	picked := results[idx-1]

	loc := store.Location{
		Query:      search,
		Name:       picked.Nom,
		PostalCode: picked.CodePostal,
		Normalized: normalize.Name(picked.Nom),
		Lat:        picked.Lat,
		Lon:        picked.Lon,
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", outputDir, err)
	}

	locs, err := store.Load(filepath.Join(outputDir, store.DefaultPath))
	if err != nil {
		return err
	}
	locs = store.Upsert(locs, loc)
	if err := store.Save(filepath.Join(outputDir, store.DefaultPath), locs); err != nil {
		return err
	}

	fmt.Printf("Saved %q to %s/locations.json as %q (run `efs-ical build %s` to generate %s/%s.ics)\n", loc.Name, outputDir, loc.Normalized, loc.Normalized, outputDir, loc.Normalized)
	return nil
}

func runBuild(args []string) error {
	locs, err := store.Load(filepath.Join(outputDir, store.DefaultPath))
	if err != nil {
		return err
	}
	if len(locs) == 0 {
		return fmt.Errorf("no locations saved yet, run `efs-ical geocode <search>` first")
	}

	targets := locs
	if len(args) > 0 {
		targets = nil
		for _, name := range args {
			found := false
			for _, l := range locs {
				if l.Normalized == name {
					targets = append(targets, l)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("no saved location matches %q", name)
			}
		}
	}

	parisLoc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return fmt.Errorf("load Europe/Paris timezone: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	now := time.Now().In(parisLoc)

	for i, l := range targets {
		if i > 0 {
			time.Sleep(jitterDelay())
		}
		raws, err := efsapi.Events(client, l.Name, l.PostalCode, l.Lat, l.Lon)
		if err != nil {
			return fmt.Errorf("%s: fetch events: %w", l.Normalized, err)
		}

		events := make([]efsapi.Event, 0, len(raws))
		for _, raw := range raws {
			evs, err := efsapi.ParseEvents(raw, parisLoc, now)
			if errors.Is(err, efsapi.ErrNoSchedule) {
				continue // permanent site, no single date to add to the calendar
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: %s: skipping unparsable event: %v\n", l.Normalized, err)
				continue
			}
			events = append(events, evs...)
		}

		cal := ical.Build(l.Name, events)
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", outputDir, err)
		}
		filename := filepath.Join(outputDir, l.Normalized+".ics")
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("create %s: %w", filename, err)
		}
		if err := cal.SerializeTo(f); err != nil {
			f.Close()
			return fmt.Errorf("write %s: %w", filename, err)
		}
		f.Close()

		fmt.Printf("%s: %d event(s) -> %s\n", l.Name, len(events), filename)
	}

	baseURL := os.Getenv("EFS_ICAL_BASE_URL")
	if err := sitegen.Generate(outputDir, locs, baseURL); err != nil {
		return fmt.Errorf("generate site: %w", err)
	}

	return nil
}
