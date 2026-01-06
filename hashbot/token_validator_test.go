package hashbot_test

import (
	"context"
	"hashbot/hashbot"
	"hashbot/twitchapi"
	"sync"
	"sync/atomic"
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
	}{
		{name: "Cancelling ctx stops token validator", cfg: &testToken{}, cancelCtx: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			invalidateCh := make(chan bool)

			refreshFunc := func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error) {
				return &TokenApi{}, nil
			}
			validateToken := func(token string) (bool, error) {
				return false, nil
			}

			hashbot.RunTokenValidator(ctx, tt.cfg, invalidateCh, refreshFunc, validateToken)
			if tt.cancelCtx {
				timeout, cancelTimeout := context.WithTimeout(context.Background(), time.Second/10)
				defer cancelTimeout()
				go func() {
					cancel()
				}()
				select {
				case <-ctx.Done():
				case <-timeout.Done():
					t.Errorf("test %q cancel ctx timed out", tt.name)
				}
			}
		})
	}
}

func TestTokenValidator_Concurrency(t *testing.T) {
	var refreshCount atomic.Int32

	refreshFunc := func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error) {
		refreshCount.Add(1)
		time.Sleep(time.Millisecond * 10) // simulate network latency
		return &TokenApi{}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	invalidateCh := make(chan bool)

	validateToken := func(token string) (bool, error) {
		time.Sleep(time.Millisecond * 10)
		return false, nil
	}

	hashbot.RunTokenValidator(ctx, &testToken{}, invalidateCh, refreshFunc, validateToken)

	var wg sync.WaitGroup
	numGoroutines := 10

	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case invalidateCh <- true:
			case <-ctx.Done():
				t.Errorf("invalidate timed out")
			}
		}()
	}

	wg.Wait() // wait for all goroutines to finish

	// check if singleflight is correctly preventing the load
	if refreshCount.Load() > 1 {
		t.Errorf("refresh ran %d times", refreshCount.Load())
	}
}

func TestRefreshWithTimeout(t *testing.T) {
	refresh := func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error) {
		time.Sleep(time.Millisecond * 10)
		return nil, nil
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		timeout       time.Duration
		refresh       hashbot.RefreshTokenFunc
		expectTimeout bool
	}{
		{
			name:          "Times out",
			timeout:       time.Millisecond * 5,
			refresh:       refresh,
			expectTimeout: true,
		},
		{
			name:          "Does not time out",
			timeout:       time.Millisecond * 25,
			refresh:       refresh,
			expectTimeout: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshWithTimeout := hashbot.RefreshWithTimeout(tt.timeout, tt.refresh)
			_, err := refreshWithTimeout(&testToken{})

			if tt.expectTimeout {
				if err != hashbot.RefreshTimedOut {
					t.Errorf("RefreshWithTimeout() error = %v, want %v", err, hashbot.RefreshTimedOut)
				}
			} else {
				if err != nil {
					t.Errorf("RefreshWithTimeout() error = %v, want nil", err)
				}
			}
		})
	}
}

func TestValidateWithTimeout(t *testing.T) {
	refresh := func(token string) (bool, error) {
		time.Sleep(time.Millisecond * 10)
		return false, nil
	}
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		timeout       time.Duration
		refresh       hashbot.ValidateTokenFunc
		expectTimeout bool
	}{
		{
			name:          "Times out",
			timeout:       time.Millisecond * 5,
			refresh:       refresh,
			expectTimeout: true,
		},
		{
			name:          "Does not time out",
			timeout:       time.Millisecond * 25,
			refresh:       refresh,
			expectTimeout: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshWithTimeout := hashbot.ValidateWithTimeout(tt.timeout, tt.refresh)
			_, err := refreshWithTimeout("token")

			if tt.expectTimeout {
				if err != hashbot.ValidateTimedOut {
					t.Errorf("ValidateWithTimeout() error = %v, want %v", err, hashbot.ValidateTimedOut)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateWithTimeout() error = %v, want nil", err)
				}
			}
		})
	}

}
