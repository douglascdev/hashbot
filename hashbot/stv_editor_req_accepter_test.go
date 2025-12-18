package hashbot_test

import (
	"context"
	"hashbot/hashbot"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"testing"
	"time"
)

type TestCfg struct {
}

func (t *TestCfg) GetClientID() string {
	return "ID"
}

func (t *TestCfg) GetLogin() string {
	return "Login"
}

func (t *TestCfg) GetSevenTVToken() string {
	return "STVToken"
}

func (t *TestCfg) GetTwitchToken() string {
	return "TTVToken"
}

func GetUserByName(config twitchapi.TwitchTokenClientIDGetter, names ...string) ([]twitchapi.HelixUser, error) {
	return []twitchapi.HelixUser{{}}, nil
}

func GetUserByConnection(host string, userID string, token string) (*seventvapi.GetUserByConnectionResp, error) {
	return &seventvapi.GetUserByConnectionResp{Data: struct {
		Users struct {
			UserByConnection struct {
				ID        string "json:\"id\""
				EmoteSets []struct {
					ID       string "json:\"id\""
					Name     string "json:\"name\""
					Capacity int    "json:\"capacity\""
					OwnerID  string "json:\"ownerId\""
					Emotes   struct {
						TotalCount int "json:\"totalCount\""
						PageCount  int "json:\"pageCount\""
						Items      []struct {
							ID    string "json:\"id\""
							Alias string "json:\"alias\""
						} "json:\"items\""
					} "json:\"emotes\""
				} "json:\"emoteSets\""
				EditorFor []struct {
					UserID          string    "json:\"userId\""
					EditorID        string    "json:\"editorId\""
					State           string    "json:\"state\""
					Notes           any       "json:\"notes\""
					AddedByID       string    "json:\"addedById\""
					AddedAt         time.Time "json:\"addedAt\""
					UpdatedAt       time.Time "json:\"updatedAt\""
					SearchUpdatedAt time.Time "json:\"searchUpdatedAt\""
				} "json:\"editorFor\""
				Style struct {
					ActiveEmoteSetID string "json:\"activeEmoteSetId\""
				} "json:\"style\""
			} "json:\"userByConnection\""
		} "json:\"users\""
	}{}}, nil
}

func AcceptEditorRequest(host, userId, editorId, seventvBearerToken string) error {
	return nil
}

func TestRunSevenTVEditorReqAccepter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		ctx                 context.Context
		cfg                 hashbot.CfgSTVTokenIdLoginGetter
		tryCancel           bool
		tokenInvalidated    chan bool
		getUserByName       hashbot.GetUserByName
		getUserByConnection hashbot.GetUserByConnection
		acceptEditorRequest hashbot.AcceptEditorRequest
	}{
		{name: "Stops on context cancel", ctx: ctx, cfg: &TestCfg{}, tryCancel: true, tokenInvalidated: make(chan bool), getUserByName: GetUserByName, getUserByConnection: GetUserByConnection, acceptEditorRequest: AcceptEditorRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashbot.RunSevenTVEditorReqAccepter(tt.ctx, tt.cfg, tt.tokenInvalidated, tt.getUserByName, tt.getUserByConnection, tt.acceptEditorRequest)

			if tt.tryCancel {
				timeout, timeoutCancel := context.WithTimeout(context.Background(), time.Second/10)
				defer timeoutCancel()
				go func() {
					cancel()
				}()
				select {
				case <-timeout.Done():
					t.Errorf("test %q timed out trying to cancel context", tt.name)
				case <-ctx.Done():
				}
			}
		})
	}
}
