package seventvapi

import "time"

type GetUserByConnectionResp struct {
	Data struct {
		UserByConnection struct {
			EmoteSets []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				OwnerID    string `json:"owner_id"`
				Capacity   int    `json:"capacity"`
				EmoteCount int    `json:"emote_count"`
				Emotes     []struct {
					ID        string    `json:"id"`
					Timestamp time.Time `json:"timestamp"`
					Name      string    `json:"name"`
					Flags     int       `json:"flags"`
					OriginID  any       `json:"origin_id"`
				} `json:"emotes"`
			} `json:"emote_sets"`
		} `json:"userByConnection"`
	} `json:"data"`
	Extensions struct {
		Analyzer struct {
			Complexity int `json:"complexity"`
			Depth      int `json:"depth"`
		} `json:"analyzer"`
		Tracing struct {
			Version   int       `json:"version"`
			StartTime time.Time `json:"startTime"`
			EndTime   time.Time `json:"endTime"`
			Duration  int       `json:"duration"`
			Execution struct {
				Resolvers []struct {
					Path        []string `json:"path"`
					FieldName   string   `json:"fieldName"`
					ParentType  string   `json:"parentType"`
					ReturnType  string   `json:"returnType"`
					StartOffset int      `json:"startOffset"`
					Duration    int      `json:"duration"`
				} `json:"resolvers"`
			} `json:"execution"`
		} `json:"tracing"`
	} `json:"extensions"`
}
