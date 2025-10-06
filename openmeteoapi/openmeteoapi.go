package openmeteoapi

import (
	"encoding/json"
	"fmt"
	"hashbot/types"
	"net/http"
	"net/url"
	"strings"

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

func GetWeather(latitute, longitude string) (*types.GetWeatherResult, error) {
	requestURL := "https://api.open-meteo.com/v1/forecast?"
	log.Debug().Str("request", requestURL).Msg("generated FindLocation query")

	params := url.Values{}
	params.Add("current", strings.Join([]string{"apparent_temperature", "precipitation", "rain", "snowfall", "wind_speed_10m", "cloud_cover", "temperature_2m"}, ","))
	params.Add("latitude", latitute)
	params.Add("longitude", longitude)

	finalURL := requestURL + params.Encode()
	log.Debug().Str("requestURL", finalURL).Msg("GetWeather generated URL")
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get weather. Status: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response types.GetWeatherResult
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode weather: %w", err)
	}

	return &response, nil
}
