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

func TestRunTokenValidator(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg                  hashbot.CfgToken
		invalidateToken      bool
		cancelCtx            bool
		expectedRefreshCount int
		bot                  hashbot.BotReconnect
	}{
		{name: "Cancelling ctx stops token validator", cfg: &testToken{}, invalidateToken: false, cancelCtx: true, expectedRefreshCount: 1, bot: &testBot{}},
		{name: "Invalidate token", cfg: &testToken{}, invalidateToken: true, cancelCtx: false, expectedRefreshCount: 2, bot: &testBot{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			invalidateCh := make(chan bool)

			refreshCount := 0
			refreshFunc := func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error) {
				refreshCount += 1
				return &TokenApi{}, nil
			}

			validateToken := func(token string) (bool, error) {
				return false, nil
			}

			hashbot.RunTokenValidator(ctx, cancel, tt.cfg, invalidateCh, tt.bot, refreshFunc, validateToken)
			if tt.invalidateToken {
				timeout, cancelTimeout := context.WithTimeout(context.Background(), time.Second/10)
				defer cancelTimeout()
				go func() {
					invalidateCh <- true
				}()
				select {
				case <-ctx.Done():
				case <-timeout.Done():
					t.Error("token invalidation timed out")
				}
			}

			if tt.cancelCtx {
				timeout, cancelTimeout := context.WithTimeout(ctx, time.Second/10)
				defer cancelTimeout()
				go func() {
					cancel()
				}()
				select {
				case <-ctx.Done():
				case <-timeout.Done():
					t.Error("cancel ctx timed out")
				}
			}

			expected, got := tt.expectedRefreshCount, refreshCount
			if expected != got {
				t.Errorf("test %q refreshed %d times instead of %d", tt.name, got, expected)
			}
		})
	}
}
