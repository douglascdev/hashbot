package seventvapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetUserByConnection(host string, userID string, token string) (*GetUserByConnectionResp, error) {
	query := strings.ReplaceAll(`
query Users {
    users {
        userByConnection(platform: "TWITCH", platformId: "%s") {
            id
            emoteSets {
                id
                name
                capacity
                ownerId
                emotes {
                    totalCount
                    pageCount
                    items {
                        id
                        alias
                    }
                }
            }
			editorFor {
                userId
                editorId
                state
                notes
                addedById
                addedAt
                updatedAt
                searchUpdatedAt
            }
            style {
                activeEmoteSetId
            }
        }
    }
}
`, "\n", "")
	svUrl, err := url.JoinPath(host, "v4", "gql")
	reqBodyMap := map[string]string{
		"query":         fmt.Sprintf(query, userID),
		"authorization": "Bearer " + token,
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
	decoder := json.NewDecoder(bytes.NewReader(body))

	// Disallow unknown fields
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
