package efsapi

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrNoSchedule is returned for cards with no single-date schedule to
// parse, e.g. permanent "Maison du Don" sites (ongoing opening hours,
// not a one-off mobile collection). Expected and skippable, not a bug.
var ErrNoSchedule = errors.New("no single-date schedule on this card (permanent site?)")

// Event is a single upcoming blood donation collection, ready for ICS export.
type Event struct {
	RdvID   string
	Title   string
	Address string
	Start   time.Time
	End     time.Time
	URL     string
	Lat     float64
	Lon     float64
}

var (
	reRdvID = regexp.MustCompile(`marker-rdv-id="(\d+)"`)
	reTitle = regexp.MustCompile(`card-location__title">([^<]*)<span>\(([^)]*)\)</span>`)
	reDate  = regexp.MustCompile(`class="calendar-text"><span[^>]*></span> ([^<]*)</div>`)
	reHref  = regexp.MustCompile(`href="([^"]*trouver-une-collecte[^"]*)"`)
	reWhen  = regexp.MustCompile(`(\d{1,2})h(\d{2})? à (\d{1,2})h(\d{2})?`)
	reDay   = regexp.MustCompile(`(\d{1,2}) (\p{L}+)`)
	reRange = regexp.MustCompile(`Du \p{L}+ (\d{1,2}) au \p{L}+ (\d{1,2}) (\p{L}+)`)
)

var frenchMonths = map[string]time.Month{
	"janvier":   time.January,
	"fevrier":   time.February, // accents stripped before lookup
	"mars":      time.March,
	"avril":     time.April,
	"mai":       time.May,
	"juin":      time.June,
	"juillet":   time.July,
	"aout":      time.August,
	"septembre": time.September,
	"octobre":   time.October,
	"novembre":  time.November,
	"decembre":  time.December,
}

var accentReplacer = strings.NewReplacer(
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"à", "a", "â", "a",
	"û", "u", "ù", "u",
	"ô", "o",
	"î", "i", "ï", "i",
	"ç", "c",
)

// ParseEvents extracts one or more Events from one results_list[].view HTML
// fragment. Most cards produce a single Event; a "Du X au Y" date range
// produces one Event per day in the range. loc is the Europe/Paris location
// used to interpret the (year-less) date text found on the page.
func ParseEvents(raw RawEvent, loc *time.Location, now time.Time) ([]Event, error) {
	view := raw.View

	rdvID := reRdvID.FindStringSubmatch(view)
	if rdvID == nil {
		return nil, fmt.Errorf("no marker-rdv-id found")
	}

	title := reTitle.FindStringSubmatch(view)
	if title == nil {
		return nil, fmt.Errorf("no title found")
	}

	dateText := reDate.FindStringSubmatch(view)
	if dateText == nil {
		return nil, ErrNoSchedule
	}

	ranges, err := parseWhen(dateText[1], loc, now)
	if err != nil {
		return nil, fmt.Errorf("parse date %q: %w", dateText[1], err)
	}

	href := reHref.FindStringSubmatch(view)
	url := ""
	if href != nil {
		url = href[1]
	}

	events := make([]Event, 0, len(ranges))
	for i, r := range ranges {
		id := rdvID[1]
		if len(ranges) > 1 {
			id = fmt.Sprintf("%s-%d", id, i+1)
		}
		events = append(events, Event{
			RdvID:   id,
			Title:   strings.TrimSpace(title[1]),
			Address: strings.TrimSpace(title[2]),
			Start:   r.start,
			End:     r.end,
			URL:     url,
			Lat:     raw.Lat,
			Lon:     raw.Lon,
		})
	}

	return events, nil
}

// dateRange is a single day's start/end time.
type dateRange struct {
	start, end time.Time
}

// parseWhen parses text like "Le mercredi 07 octobre de 16h à 19h30" into a
// single-day range, or a range like "Du mardi 22 au mercredi 23 septembre de
// 12h à 17h" into one range per day. The year isn't present on the page, so
// it's inferred: the current year, rolled forward one year if that would
// land in the past.
func parseWhen(text string, loc *time.Location, now time.Time) ([]dateRange, error) {
	startHour, startMin, endHour, endMin, err := parseHours(text)
	if err != nil {
		return nil, err
	}

	if rangeMatch := reRange.FindStringSubmatch(text); rangeMatch != nil {
		day1, err := strconv.Atoi(rangeMatch[1])
		if err != nil {
			return nil, fmt.Errorf("invalid day: %w", err)
		}
		day2, err := strconv.Atoi(rangeMatch[2])
		if err != nil {
			return nil, fmt.Errorf("invalid day: %w", err)
		}
		month, err := lookupMonth(rangeMatch[3])
		if err != nil {
			return nil, err
		}

		var ranges []dateRange
		for day := day1; day <= day2; day++ {
			ranges = append(ranges, buildRange(now, loc, month, day, startHour, startMin, endHour, endMin))
		}
		return ranges, nil
	}

	dayMatch := reDay.FindStringSubmatch(text)
	if dayMatch == nil {
		return nil, fmt.Errorf("no day/month found")
	}
	day, err := strconv.Atoi(dayMatch[1])
	if err != nil {
		return nil, fmt.Errorf("invalid day: %w", err)
	}
	month, err := lookupMonth(dayMatch[2])
	if err != nil {
		return nil, err
	}

	return []dateRange{buildRange(now, loc, month, day, startHour, startMin, endHour, endMin)}, nil
}

// lookupMonth resolves a French month name (accents optional) to a time.Month.
func lookupMonth(name string) (time.Month, error) {
	monthKey := accentReplacer.Replace(strings.ToLower(name))
	month, ok := frenchMonths[monthKey]
	if !ok {
		return 0, fmt.Errorf("unknown month %q", name)
	}
	return month, nil
}

// parseHours extracts the "12h à 17h30"-style hour range shared by both the
// single-day and date-range formats.
func parseHours(text string) (startHour, startMin, endHour, endMin int, err error) {
	whenMatch := reWhen.FindStringSubmatch(text)
	if whenMatch == nil {
		return 0, 0, 0, 0, fmt.Errorf("no time range found")
	}
	startHour, _ = strconv.Atoi(whenMatch[1])
	if whenMatch[2] != "" {
		startMin, _ = strconv.Atoi(whenMatch[2])
	}
	endHour, _ = strconv.Atoi(whenMatch[3])
	if whenMatch[4] != "" {
		endMin, _ = strconv.Atoi(whenMatch[4])
	}
	return startHour, startMin, endHour, endMin, nil
}

// buildRange turns a day/month/hours into a start/end time, inferring the
// year as now's year, rolled forward one year if that would land in the past.
func buildRange(now time.Time, loc *time.Location, month time.Month, day, startHour, startMin, endHour, endMin int) dateRange {
	year := now.Year()
	start := time.Date(year, month, day, startHour, startMin, 0, 0, loc)
	if start.Before(now.AddDate(0, 0, -1)) {
		year++
		start = time.Date(year, month, day, startHour, startMin, 0, 0, loc)
	}
	end := time.Date(year, month, day, endHour, endMin, 0, 0, loc)
	return dateRange{start: start, end: end}
}
