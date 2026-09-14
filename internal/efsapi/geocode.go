package efsapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const geocodeURL = "https://oudonner.api.efs.sante.fr/carto-api/v3/city/searchbyinput"

// GeocodeResult is one match from the reverse-geocode endpoint.
type GeocodeResult struct {
	Nom        string  `json:"nom"`
	CodePostal string  `json:"codePostal"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
}

// Geocode looks up a city name / zip code and returns all matches.
func Geocode(client *http.Client, search string) ([]GeocodeResult, error) {
	u := geocodeURL + "?" + url.Values{"searchString": {search}}.Encode()

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	setBrowserHeaders(req, "https://dondesang.efs.sante.fr/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode request: unexpected status %s", resp.Status)
	}

	var results []GeocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode geocode response: %w", err)
	}
	return results, nil
}
