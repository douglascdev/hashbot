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

type SearchEmoteResponse struct {
	Data struct {
		Emotes struct {
			Search struct {
				Items []struct {
					ID          string `json:"id"`
					DefaultName string `json:"defaultName"`
				} `json:"items"`
			} `json:"search"`
		} `json:"emotes"`
	} `json:"data"`
	Extensions struct {
		Analyzer struct {
			Complexity int `json:"complexity"`
			Depth      int `json:"depth"`
		} `json:"analyzer"`
	} `json:"extensions"`
}

func SearchEmote(host, searchQuery string) (*SearchEmoteResponse, error) {
	var result SearchEmoteResponse

	emoteSearchQuery := strings.ReplaceAll(`
query Emotes {
    emotes {
        search(
            page: 1
            perPage: 1
            query: "%s"
            sort: { sortBy: "TRENDING_MONTHLY", order: "DESCENDING" }
            filters: {
              exactMatch: true
            }
        ) {
            items {
                id
                defaultName
            }
        }
    }
}
`, "\n", "")
	svUrl, err := url.JoinPath(host, "v4", "gql")
	reqBodyMap := map[string]string{
		"query": fmt.Sprintf(emoteSearchQuery, searchQuery),
	}
	m, err := json.Marshal(reqBodyMap)
	if err != nil {
		return &result, fmt.Errorf("failed to marshal request body: %w", err)
	}
	reqBody := strings.NewReader(string(m))

	req, err := http.NewRequest("POST", svUrl, reqBody)
	if err != nil {
		return &result, fmt.Errorf("failed to create SearchEmote request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &result, fmt.Errorf("SearchEmote request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &result, err
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	if err != nil {
		return &result, err
	}

	return &result, err
}
