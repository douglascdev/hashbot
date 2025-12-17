package hashbot

import (
	"context"
	"hashbot/twitchapi"
	"time"

	"github.com/rs/zerolog/log"
)

type BotReconnect interface {
	ConnectClient(login string, token string) error
}

type CfgToken interface {
	twitchapi.CfgIdSecretRefreshToken

	GetTwitchToken() string
	SetTwitchToken(token string)

	GetLogin() string
}

func RunTokenValidator(ctx context.Context, cancelFn context.CancelFunc, cfg CfgToken, tokenInvalidated chan bool, bot BotReconnect) {
	tryRefresh := func() {
		valid, err := twitchapi.ValidateToken(cfg.GetTwitchToken())
		if err != nil {
			log.Error().Err(err)
		}
		log.Info().Bool("validToken", valid).Msg("token validation")
		if !valid {
			token, err := twitchapi.RefreshTwitchToken(cfg)
			if err != nil {
				log.Error().Err(err).Msg("failed to refresh invalidated token")
				cancelFn()
			}
			log.Info().Msg("succesfully obtained refreshed token, reconnecting with new twitch client")
			cfg.SetTwitchToken(token.AccessToken)
			err = bot.ConnectClient(cfg.GetLogin(), token.AccessToken)
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
				tryRefresh()
			case <-tokenInvalidated:
				tryRefresh()
			}
		}
	}()
}
