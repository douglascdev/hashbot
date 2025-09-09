package seventvapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetUserByConnection(host string, userID string) (*GetUserByConnectionResp, error) {
	query := fmt.Sprintf(`
	query UserByConnection {
		userByConnection(platform: "TWITCH", id: "%s") {
			emote_sets {
				id
				name
				owner_id
				capacity
				emote_count
				emotes {
					id
					timestamp
					name
					flags
					origin_id
				}
			}
		}
	}
`, userID)
	svUrl, err := url.JoinPath(host, "v3", "gql")
	reqBodyMap := map[string]string{
		"operationName": "GetUserByConnection",
		"query":         query,
	}
	m, err := json.Marshal(reqBodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	reqBody := strings.NewReader(string(m))

	req, err := http.NewRequest("POST", svUrl, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetUserByConnection request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetUserByConnection request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result GetUserByConnectionResp
	err = json.Unmarshal(body, &result)
	if err != nil {
		return &result, err
	}

	return &result, nil
}
