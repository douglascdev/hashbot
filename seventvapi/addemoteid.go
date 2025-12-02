package seventvapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

var NotAnEditorErr error = errors.New("not and editor")
var EmoteConflictingName error = errors.New("conflicting emote name")

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
	log.Debug().Any("query", string(m)).Msg("sending query for AddEmoteWithID")
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
		Data   any `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	err = decoder.Decode(&result)
	if err == nil && result.Data == nil {
		var errorMsgs []string
		for _, msg := range result.Errors {
			if strings.HasPrefix(msg.Message, "LACKING_PRIVILEGES") {
				return NotAnEditorErr
			} else if msg.Message == "BAD_REQUEST this emote has a conflicting name" {
				return EmoteConflictingName
			}

			errorMsgs = append(errorMsgs, msg.Message)
		}
		return fmt.Errorf("failed with errors: %q", strings.Join(errorMsgs, ", "))
	}

	return nil
}
