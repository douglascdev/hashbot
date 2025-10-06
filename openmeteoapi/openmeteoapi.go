package openmeteoapi

import (
	"encoding/json"
	"fmt"
	"hashbot/types"
	"net/http"

	"github.com/rs/zerolog/log"
)

func FindLocation(searchQuery string) (*types.FindLocationResult, error) {
	requestURL := "https://geocoding-api.open-meteo.com/v1/search?name=" + searchQuery
	log.Debug().Str("request", requestURL).Msg("generated FindLocation query")

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get location. Status: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response types.FindLocationResult
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode location: %w", err)
	}

	return &response, nil
}
