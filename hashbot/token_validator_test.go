package hashbot_test

import (
	"context"
	"hashbot/hashbot"
	"hashbot/twitchapi"
	"testing"
	"time"
)

type testToken struct {
	ClientSecret string
	ClientID     string
	Token        string
	Login        string
	RefreshToken string
}

func (t *testToken) GetTwitchToken() string {
	return "123"
}
func (t *testToken) SetTwitchToken(token string) {
	t.Token = token
}
func (t *testToken) GetLogin() string {
	return t.Login
}

func (t *testToken) GetClientID() string {
	return t.ClientID
}
func (t *testToken) GetClientSecret() string {
	return t.ClientSecret
}
func (t *testToken) GetRefreshToken() string {
	return t.RefreshToken
}

type testBot struct {
}

func (t *testBot) ConnectClient(login string, token string) error {
	return nil
}

type TokenApi struct {
}

func (t *TokenApi) GetToken() string {
	return "GetToken"
}

func RefreshTwitchToken(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error) {
	return &TokenApi{}, nil
}

func TestRunTokenValidator(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg             hashbot.CfgToken
		invalidateToken bool
		bot             hashbot.BotReconnect
	}{
		{name: "Invalidate token", cfg: &testToken{}, invalidateToken: true, bot: &testBot{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			invalidateCh := make(chan bool)
			hashbot.RunTokenValidator(ctx, cancel, tt.cfg, invalidateCh, tt.bot, RefreshTwitchToken)
			if tt.invalidateToken {
				invalidateCh <- true
				timeout, cancelTimeout := context.WithTimeout(ctx, time.Second/4)
				defer cancelTimeout()
				select {
				case <-ctx.Done():
				case <-timeout.Done():
					t.Error("token invalidation timed out")
				}
			}
		})
	}
}
