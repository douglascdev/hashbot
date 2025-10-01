package seventvapi

import "time"

type GetUserByConnectionResp struct {
	Data struct {
		Users struct {
			UserByConnection struct {
				ID        string `json:"id"`
				EmoteSets []struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Capacity int    `json:"capacity"`
					OwnerID  string `json:"ownerId"`
					Emotes   struct {
						TotalCount int `json:"totalCount"`
						PageCount  int `json:"pageCount"`
						Items      []struct {
							ID    string `json:"id"`
							Alias string `json:"alias"`
						} `json:"items"`
					} `json:"emotes"`
				} `json:"emoteSets"`
				EditorFor []struct {
					UserID          string    `json:"userId"`
					EditorID        string    `json:"editorId"`
					State           string    `json:"state"`
					Notes           any       `json:"notes"`
					AddedByID       string    `json:"addedById"`
					AddedAt         time.Time `json:"addedAt"`
					UpdatedAt       time.Time `json:"updatedAt"`
					SearchUpdatedAt time.Time `json:"searchUpdatedAt"`
				} `json:"editorFor"`
				Style struct {
					ActiveEmoteSetID string `json:"activeEmoteSetId"`
				} `json:"style"`
			} `json:"userByConnection"`
		} `json:"users"`
	} `json:"data"`
	Extensions struct {
		Analyzer struct {
			Complexity int `json:"complexity"`
			Depth      int `json:"depth"`
		} `json:"analyzer"`
	} `json:"extensions"`
}
