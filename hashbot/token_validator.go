package hashbot

import (
	"context"
	"errors"
	"hashbot/twitchapi"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type MyCircuit func(cfg CfgToken, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc)

func DebounceFirstAndTimeout(ctx context.Context, circuit MyCircuit, debounceDuration time.Duration) MyCircuit {
	var threshold time.Time
	var m sync.Mutex

	return func(cfg CfgToken, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc) {
		m.Lock()
		defer func() {
			threshold = time.Now().Add(debounceDuration)
			m.Unlock()
		}()
		if time.Now().Before(threshold) {
			return
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*20)
		defer cancel()

		done := make(chan bool, 1)
		go func() {
			circuit(cfg, refreshToken, validateToken)
			done <- true
		}()

		select {
		case <-done:
			log.Info().Msg("tryRefresh is finished")
		case <-ctx.Done():
			log.Warn().Msg(("tryRefresh context timed out or terminated"))
		}
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

type CfgToken interface {
	twitchapi.CfgIdSecretRefreshToken

	GetTwitchToken() string
	SetTwitchToken(token string)

	GetLogin() string
}

type RefreshTokenFunc func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error)
type ValidateTokenFunc func(token string) (bool, error)

func tryRefresh(cfg CfgToken, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc) {
	refreshTokenWithTimeout := RefreshWithTimeout(time.Second*5, refreshToken)
	validateTokenWithTimeout := ValidateWithTimeout(time.Second*5, validateToken)

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
	}
}

func RunTokenValidator(ctx context.Context, cfg CfgToken, tokenInvalidated chan bool, refreshToken RefreshTokenFunc, validateToken ValidateTokenFunc) {
	debouncedTryRefresh := DebounceFirstAndTimeout(ctx, tryRefresh, time.Second*5)

	debouncedTryRefresh(cfg, refreshToken, validateToken)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Warn().Msg("token validation stopped")
				return
			case <-time.After(time.Hour):
				log.Info().Msg("refreshing token after waiting")
				debouncedTryRefresh(cfg, refreshToken, validateToken)
			case <-tokenInvalidated:
				log.Info().Msg("token invalidated, refreshing")
				debouncedTryRefresh(cfg, refreshToken, validateToken)
			}
		}
	}()
}
