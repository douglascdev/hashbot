package seventvapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserByConnection(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the response you want to return
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
	{
		"data": {
			"userByConnection": {
				"emote_sets": [
					{
						"id": "123",
						"name": "hash_table's Emotes",
						"owner_id": "01GBE7YESG000CRDCG0DV7KBEV",
						"capacity": 300,
						"emote_count": 207,
						"emotes": [
							{
								"id": "01GBE7YESG000CRDCG0DV7KBEV",
								"timestamp": "2023-11-26T18:58:15.301+00:00",
								"name": "NOTED",
								"flags": 0,
								"origin_id": null
							}
						]
					}
				]
			}
		}
	}`))
	}))
	defer mockServer.Close()
	// Call the function with the mock server's URL
	result, err := GetUserByConnection(mockServer.URL, "123")
	if err != nil {
		t.Fatal(err)
	}

	// Check the result
	id := result.Data.UserByConnection.EmoteSets[0].ID
	expected := "123"
	if id != expected {
		t.Fatalf("invalid result, expected=%q got=%q", expected, id)
	}
}
