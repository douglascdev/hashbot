package twitchapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hashbot/config"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type HelixUser struct {
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"`      // Deprecated
	Email           string    `json:"email,omitempty"` // Optional field, omitted if empty
	CreatedAt       time.Time `json:"created_at"`
}

type helixUserResponse struct {
	Data []HelixUser `json:"data"`
}

type AuthorizationCodeResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

type RefreshTokenResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

func (r *RefreshTokenResponse) GetToken() string {
	return r.AccessToken
}

func GetUserByName(config *config.Config, names ...string) ([]HelixUser, error) {
	if len(names) == 0 {
		return nil, nil
	}
	if len(names) > 100 {
		return nil, fmt.Errorf("exceeded maximum number of names (100)")
	}
	var nameParams []string
	for _, name := range names {
		nameParams = append(nameParams, "login="+name)
	}
	requestURL := "https://api.twitch.tv/helix/users?" + strings.Join(nameParams, "&")
	log.Debug().Str("request", requestURL).Strs("names", names).Msg("generated helix user request")

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+config.TwitchToken)
	req.Header.Add("Client-Id", config.ClientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user. Status: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response helixUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func GetUserByID(config *config.Config, ids ...string) (idToUser map[string]*HelixUser, err error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > 100 {
		return nil, fmt.Errorf("exceeded maximum number of names (100)")
	}
	var idsParams []string
	for _, id := range ids {
		idsParams = append(idsParams, "id="+id)
	}
	requestURL := "https://api.twitch.tv/helix/users?" + strings.Join(idsParams, "&")
	log.Debug().Str("request", requestURL).Strs("names", ids).Msg("generated helix user request")
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+config.TwitchToken)
	req.Header.Add("Client-Id", config.ClientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user. Status: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response helixUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	idToUser = make(map[string]*HelixUser)
	for _, user := range response.Data {
		idToUser[user.ID] = &user
	}

	return idToUser, nil
}

type CfgIdSecretRefreshToken interface {
	GetClientID() string
	GetClientSecret() string
	GetRefreshToken() string
}

type TokenGetter interface {
	GetToken() string
}

func RefreshTwitchToken(cfg CfgIdSecretRefreshToken) (TokenGetter, error) {
	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {cfg.GetClientID()},
		"client_secret": {cfg.GetClientSecret()},
		"refresh_token": {cfg.GetRefreshToken()},
		"grant_type":    {"refresh_token"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request refreshed token: %w", err)
	}

	defer resp.Body.Close()
	var result RefreshTokenResponse
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read token response body: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func AuthorizationCode(clientID, clientSecret, code, redirectURI string) (*AuthorizationCodeResponse, error) {
	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch oauth token from twitch: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read token response body: %w", err)
	}

	var result AuthorizationCodeResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func ValidateToken(token string) (bool, error) {
	requestURL := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", "OAuth "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, errors.New("invalid access token")
	}

	return true, nil
}
