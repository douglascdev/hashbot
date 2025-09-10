package seventvapi

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
