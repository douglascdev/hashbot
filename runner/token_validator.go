package runner

import (
	"context"
	"hashbot/twitchapi"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/rs/zerolog/log"
)

type BotConnectBind interface {
	Connect() error
	Disconnect() error
	BindClientFunctions()
	SetClient(client *twitch.Client)
}

type CfgToken interface {
	twitchapi.CfgIdSecretRefreshToken

	GetTwitchToken() string
	SetTwitchToken(token string)

	GetLogin() string
}

func RunTokenValidator(ctx context.Context, cancelFn context.CancelFunc, cfg CfgToken, tokenInvalidated chan bool, bot BotConnectBind) {
	tryRefresh := func() {
		valid, err := twitchapi.ValidateToken(cfg.GetTwitchToken())
		if err != nil {
			log.Error().Err(err)
		}
		log.Info().Bool("validToken", valid).Msg("")
		if !valid {
			token, err := twitchapi.RefreshTwitchToken(cfg)
			if err != nil {
				log.Error().Err(err).Msg("failed to refresh invalidated token")
				cancelFn()
			}
			log.Info().Msg("succesfully obtained refreshed token, disconnecting")
			cfg.SetTwitchToken(token.AccessToken)
			bot.Disconnect()
			bot.SetClient(twitch.NewClient(cfg.GetLogin(), "oauth:"+cfg.GetTwitchToken()))
			bot.BindClientFunctions()
			bot.Connect()
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
