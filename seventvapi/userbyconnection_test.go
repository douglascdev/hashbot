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
        "users": {
            "userByConnection": {
                "id": "123",
                "emoteSets": [
                    {
                        "id": "123",
                        "name": "Set",
                        "capacity": 300,
                        "ownerId": "01HEXK7BA0000FJQ1V068WZSDN",
                        "emotes": {
                            "totalCount": 204,
                            "pageCount": 1,
                            "items": [
                                {
                                    "id": "01GEEHRQYG0006MCY6R6BPV3HB",
                                    "alias": "NOTED"
                                },
                                {
                                    "id": "01FYQZVG280006SX8JX4TD7SJA",
                                    "alias": "VIBE"
                                }
                            ]
                        }
                    }
                ],
                "style": {
                    "activeEmoteSetId": ""
                }
            }
        }
    },
    "extensions": {
        "analyzer": {
            "complexity": 16,
            "depth": 6
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
	id := result.Data.Users.UserByConnection.EmoteSets[0].ID
	expected := "123"
	if id != expected {
		t.Fatalf("invalid result, expected=%q got=%q", expected, id)
	}
}
