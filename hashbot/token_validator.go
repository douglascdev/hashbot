package hashbot

import (
	"context"
	"errors"
	"hashbot/twitchapi"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type MyCircuit func(cfg CfgToken, connectClient ConnectClientFunc, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc)

func DebounceFirst(circuit MyCircuit, d time.Duration) MyCircuit {
	var threshold time.Time
	var m sync.Mutex

	return func(cfg CfgToken, connectClient ConnectClientFunc, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc) {
		m.Lock()
		defer func() {
			threshold = time.Now().Add(d)
			m.Unlock()
		}()
		if time.Now().Before(threshold) {
			return
		}
		circuit(cfg, connectClient, refreshToken, validateToken)
	}
}

var RefreshTimedOut = errors.New("refresh function timed out")

func RefreshWithTimeout(timeout time.Duration, refresh RefreshTokenFunc) RefreshTokenFunc {
	return func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error) {
		type result struct {
			value twitchapi.TokenGetter
			err   error
		}
		done := make(chan result, 1)
		go func() {
			var res result
			res.value, res.err = refresh(cfg)
			done <- res
		}()
		select {
		case res := <-done:
			return res.value, res.err
		case <-time.After(timeout):
			return nil, RefreshTimedOut
		}
	}
}

var ValidateTimedOut = errors.New("validate token function timed out")

func ValidateWithTimeout(timeout time.Duration, validate ValidateTokenFunc) ValidateTokenFunc {
	return func(token string) (bool, error) {
		type result struct {
			value bool
			err   error
		}
		done := make(chan result, 1)
		go func() {
			var res result
			res.value, res.err = validate(token)
			done <- res
		}()
		select {
		case res := <-done:
			return res.value, res.err
		case <-time.After(timeout):
			return false, ValidateTimedOut
		}
	}
}

var ConnectClientTimedOut = errors.New("connect client function timed out")

func ConnectClientWithTimeout(timeout time.Duration, connectClient ConnectClientFunc) ConnectClientFunc {
	return func(login, token string) error {
		done := make(chan error, 1)
		go func() {
			done <- connectClient(login, token)
		}()
		select {
		case res := <-done:
			return res
		case <-time.After(timeout):
			return ConnectClientTimedOut
		}
	}
}

type CfgToken interface {
	twitchapi.CfgIdSecretRefreshToken

	GetTwitchToken() string
	SetTwitchToken(token string)

	GetLogin() string
}

type RefreshTokenFunc func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error)
type ValidateTokenFunc func(token string) (bool, error)
type ConnectClientFunc func(login string, token string) error

func tryRefresh(cfg CfgToken, connectClient ConnectClientFunc, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc) {
	refreshTokenWithTimeout := RefreshWithTimeout(time.Second*5, refreshToken)
	validateTokenWithTimeout := ValidateWithTimeout(time.Second*5, validateToken)
	connectClientWithTimeout := ConnectClientWithTimeout(time.Second*5, connectClient)

	valid, err := validateTokenWithTimeout(cfg.GetTwitchToken())
	if err != nil {
		log.Error().Err(err).Msg("failed to validate token")
	}

	log.Info().Bool("validToken", valid).Msg("token validation")
	if !valid {
		token, err := refreshTokenWithTimeout(cfg)
		if err != nil {
			log.Error().Err(err).Msg("failed to refresh invalidated token")
			return
		}
		log.Info().Msg("succesfully obtained refreshed token, reconnecting with new twitch client")
		cfg.SetTwitchToken(token.GetToken())
		err = connectClientWithTimeout(cfg.GetLogin(), token.GetToken())
		if err != nil {
			log.Error().Err(err).Msg("client failed to reconnect")
		}
	}
}

func RunTokenValidator(ctx context.Context, cfg CfgToken, tokenInvalidated chan bool, connectClient ConnectClientFunc, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc) {
	debouncedTryRefresh := DebounceFirst(tryRefresh, time.Second*5)

	debouncedTryRefresh(cfg, connectClient, refreshToken, validateToken)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Warn().Msg("token validation stopped")
				return
			case <-time.After(time.Hour):
				log.Info().Msg("refreshing token after waiting")
				debouncedTryRefresh(cfg, connectClient, refreshToken, validateToken)
			case <-tokenInvalidated:
				log.Info().Msg("token invalidated, refreshing")
				debouncedTryRefresh(cfg, connectClient, refreshToken, validateToken)
			}
		}
	}()
}
