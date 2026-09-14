package efsapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const eventsURL = "https://dondesang.efs.sante.fr/get-collects-ajax"

// France-wide bounding box, as captured from a real page load. Filtering is
// actually done server-side via the ville param, not by clipping to this
// box, so it's kept constant across all locations.
const (
	neLon = "12.255102809523605"
	neLat = "51.52811834794"
	swLon = "-8.980334886476948"
	swLat = "42.20342456882406"
	wmap  = "795"
	hmap  = "547"
)

type eventsResponse struct {
	Results []struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"results"`
	NumResults  int `json:"num_results"`
	ResultsList []struct {
		View string `json:"view"`
	} `json:"results_list"`
}

// RawEvent pairs one results_list HTML card with its matching marker
// coordinates (same index in the results array), ready for parsing.
type RawEvent struct {
	View string
	Lat  float64
	Lon  float64
}

// Events fetches upcoming collection cards near the given commune.
func Events(client *http.Client, nom, codePostal string, lat, lon float64) ([]RawEvent, error) {
	ville := fmt.Sprintf("%s, %s", nom, codePostal)

	q := url.Values{
		"NorthEastLongitude": {neLon},
		"NorthEastLatitude":  {neLat},
		"SouthWestLongitude": {swLon},
		"SouthWestLatitude":  {swLat},
		"CenterLongitude":    {strconv.FormatFloat(lon, 'f', -1, 64)},
		"CenterLatitude":     {strconv.FormatFloat(lat, 'f', -1, 64)},
		"ville":              {ville},
		"selected":           {"selected"},
		"wmap":               {wmap},
		"hmap":               {hmap},
	}

	req, err := http.NewRequest(http.MethodGet, eventsURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	setBrowserHeaders(req, "https://dondesang.efs.sante.fr/trouver-une-collecte")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("events request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("events request: unexpected status %s", resp.Status)
	}

	var parsed eventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode events response: %w", err)
	}

	raw := make([]RawEvent, 0, len(parsed.ResultsList))
	for i, item := range parsed.ResultsList {
		ev := RawEvent{View: item.View}
		if i < len(parsed.Results) {
			ev.Lat = parsed.Results[i].Lat
			ev.Lon = parsed.Results[i].Lon
		}
		raw = append(raw, ev)
	}
	return raw, nil
}
