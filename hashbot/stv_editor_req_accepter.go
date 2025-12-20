package hashbot

import (
	"context"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type CfgSTVTokenIdLoginGetter interface {
	twitchapi.TwitchTokenClientIDGetter

	GetLogin() string
	GetSevenTVToken() string
}

type GetUserByName func(config twitchapi.TwitchTokenClientIDGetter, names ...string) ([]twitchapi.HelixUser, error)
type GetUserByConnection func(host string, userID string, token string) (*seventvapi.GetUserByConnectionResp, error)
type AcceptEditorRequest func(host, userId, editorId, seventvBearerToken string) error

func RunSevenTVEditorReqAccepter(
	ctx context.Context,
	cfg CfgSTVTokenIdLoginGetter,
	tokenInvalidated chan bool,
	getUserByName GetUserByName,
	getUserByConnection GetUserByConnection,
	acceptEditorRequest AcceptEditorRequest,
) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Warn().Msg("sevenTV editor request accepter stopped")
				return
			case <-time.After(time.Minute * 2):
				users, err := getUserByName(cfg, cfg.GetLogin())
				if err != nil {
					if strings.Contains(err.Error(), "401") {
						log.Err(err).Msg("sevenTV editor request accepter failed with 401 trying to get twitch user, invalidating token")
						select {
						case tokenInvalidated <- true:
						default:
							log.Warn().Msg("skipped sending to tokenInvalidated, receiver might be blocked")
						}
						continue
					}
					log.Err(err).Msg("sevenTV editor request accepter failed to get twitch user for the bot")
					continue
				}
				twitchUser := users[0]
				resp, err := getUserByConnection("https://7tv.io", twitchUser.ID, cfg.GetSevenTVToken())
				if err != nil {
					log.Err(err).Msg("sevenTV editor request accepter failed to get 7TV user for the bot")
					continue
				}
				for _, r := range resp.Data.Users.UserByConnection.EditorFor {
					if r.State == "PENDING" {
						err := acceptEditorRequest("https://7tv.io", r.UserID, r.EditorID, cfg.GetSevenTVToken())
						if err != nil {
							log.Err(err).Str("userId", r.UserID).Msg("failed to accept editor request")
						}
						log.Info().Str("userId", r.UserID).Msg("accepted editor request")
					}
				}

			}
		}
	}()
}
