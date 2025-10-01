package seventvapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func AcceptEditorRequest(host, userId, editorId, seventvBearerToken string) error {
	gqlQuery := strings.ReplaceAll(`
mutation UserEditors {
    userEditors {
        editor(
            userId: "%s"
            editorId: "%s"
        ) {
            updateState(state: "ACCEPT") {
                userId
                editorId
                state
                notes
                addedById
                addedAt
                updatedAt
                searchUpdatedAt
            }
        }
    }
}

`, "\n", "")
	svUrl, err := url.JoinPath(host, "v4", "gql")
	reqBodyMap := map[string]string{
		"query": fmt.Sprintf(gqlQuery, userId, editorId),
	}
	m, err := json.Marshal(reqBodyMap)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	reqBody := strings.NewReader(string(m))

	req, err := http.NewRequest("POST", svUrl, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create AcceptEditor request: %w", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", seventvBearerToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("AcceptEditor request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			UserEditors struct {
				Editor struct {
					UpdateState struct {
						UserID          string    `json:"userId"`
						EditorID        string    `json:"editorId"`
						State           string    `json:"state"`
						Notes           any       `json:"notes"`
						AddedByID       string    `json:"addedById"`
						AddedAt         time.Time `json:"addedAt"`
						UpdatedAt       time.Time `json:"updatedAt"`
						SearchUpdatedAt any       `json:"searchUpdatedAt"`
					} `json:"updateState"`
				} `json:"editor"`
			} `json:"userEditors"`
		} `json:"data"`
		Extensions struct {
			Analyzer struct {
				Complexity int `json:"complexity"`
				Depth      int `json:"depth"`
			} `json:"analyzer"`
		} `json:"extensions"`
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	return nil
}
