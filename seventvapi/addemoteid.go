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

func AddEmoteWithID(host, userTwitchID, emoteID, alias, seventvBearerToken string) error {
	sevenTVUserData, err := GetUserByConnection(host, userTwitchID, seventvBearerToken)
	if err != nil {
		return err
	}
	activeSetID := sevenTVUserData.Data.Users.UserByConnection.Style.ActiveEmoteSetID

	gqlQuery := strings.ReplaceAll(`
mutation EmoteSets {
    emoteSets {
        emoteSet(id: "%s") {
            addEmote(id: { emoteId: "%s", alias: "%s" }) {
                id
            }
        }
    }
}
`, "\n", "")
	svUrl, err := url.JoinPath(host, "v4", "gql")
	reqBodyMap := map[string]string{
		"query": fmt.Sprintf(gqlQuery, activeSetID, emoteID, alias),
	}
	m, err := json.Marshal(reqBodyMap)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	reqBody := strings.NewReader(string(m))

	req, err := http.NewRequest("POST", svUrl, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create AddEmote request: %w", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", seventvBearerToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("AddEmote request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)
	if err == nil {
		var errorMsgs []string
		for _, msg := range result.Errors {
			errorMsgs = append(errorMsgs, msg.Message)
		}
		return fmt.Errorf("failed with errors: %q", strings.Join(errorMsgs, ", "))
	}

	return nil
}
