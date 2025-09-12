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

var EmoteNotFound = errors.New("emote not found")

func RemoveEmote(host string, userTwitchID string, emoteAlias string, seventvBearerToken string) error {
	sevenTVUserData, err := GetUserByConnection(host, userTwitchID)
	if err != nil {
		return err
	}

	var emoteID string
	log.Debug().Int("amount of returned 7tv sets", len(sevenTVUserData.Data.Users.UserByConnection.EmoteSets))
outerLoop:
	for _, set := range sevenTVUserData.Data.Users.UserByConnection.EmoteSets {
		log.Debug().Str("current iteration set", set.Name)
		if set.ID != sevenTVUserData.Data.Users.UserByConnection.Style.ActiveEmoteSetID {
			log.Debug().Str("skipped set", set.Name).Str("active set", sevenTVUserData.Data.Users.UserByConnection.Style.ActiveEmoteSetID)
			continue
		}

		for _, emote := range set.Emotes.Items {
			if emote.Alias == emoteAlias {
				log.Debug().Str("found emote", emote.Alias)
				emoteID = emote.ID
				break outerLoop
			}
			log.Debug().Str("expected emote", emoteAlias).Str("skipped emote", emote.Alias)
		}
	}
	if emoteID == "" {
		return EmoteNotFound
	}

	query := strings.ReplaceAll(`
mutation EmoteSets {
    emoteSets {
        emoteSet(id:"%s") {
            removeEmote(id: { emoteId: "%s", alias: "%s" }) {
                id
            }
        }
    }
}
`, "\n", "")
	svUrl, err := url.JoinPath(host, "v4", "gql")
	reqBodyMap := map[string]string{
		"query": fmt.Sprintf(query, sevenTVUserData.Data.Users.UserByConnection.Style.ActiveEmoteSetID, emoteID, emoteAlias),
	}
	m, err := json.Marshal(reqBodyMap)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	reqBody := strings.NewReader(string(m))

	req, err := http.NewRequest("POST", svUrl, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create RemoveEmote request: %w", err)
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", seventvBearerToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("RemoveEmote request failed: %w", err)
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
