package hashbot

import (
	"context"
	"hashbot/twitchapi"
	"time"

	"github.com/rs/zerolog/log"
)

type CfgToken interface {
	twitchapi.CfgIdSecretRefreshToken

	GetTwitchToken() string
	SetTwitchToken(token string)

	GetLogin() string
}

type RefreshTokenFunc func(cfg twitchapi.CfgIdSecretRefreshToken) (twitchapi.TokenGetter, error)
type ValidateToken func(token string) (bool, error)
type ConnectClient func(login string, token string) error

func RunTokenValidator(ctx context.Context, cfg CfgToken, tokenInvalidated chan bool, connectClient ConnectClient, refreshToken RefreshTokenFunc, validateToken ValidateToken) {
	tryRefresh := func() {
		valid, err := validateToken(cfg.GetTwitchToken())
		if err != nil {
			log.Error().Err(err)
		}
		log.Info().Bool("validToken", valid).Msg("token validation")
		if !valid {
			token, err := refreshToken(cfg)
			if err != nil {
				log.Error().Err(err).Msg("failed to refresh invalidated token")
				return
			}
			log.Info().Msg("succesfully obtained refreshed token, reconnecting with new twitch client")
			cfg.SetTwitchToken(token.GetToken())
			err = connectClient(cfg.GetLogin(), token.GetToken())
			if err != nil {
				log.Error().Err(err).Msg("client failed to reconnect")
			}
		}
	}

	tryRefresh()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Warn().Msg("token validation stopped")
				return
			case <-time.After(time.Hour):
				log.Info().Msg("refreshing token after waiting")
				tryRefresh()
			case <-tokenInvalidated:
				log.Info().Msg("token invalidated, refreshing")
				tryRefresh()
			}
		}
	}()
}
