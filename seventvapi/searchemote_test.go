package seventvapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchEmote(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the response you want to return
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
{
    "data": {
        "emotes": {
            "search": {
                "items": [
                    {
                        "id": "123",
                        "defaultName": "PAUSE"
                    }
                ]
            }
        }
    },
    "extensions": {
        "analyzer": {
            "complexity": 5,
            "depth": 4
        }
    }
}`))
	}))
	defer mockServer.Close()
	// Call the function with the mock server's URL
	result, err := SearchEmote(mockServer.URL, "123")
	if err != nil {
		t.Fatal(err)
	}

	// Check the result
	id := result.Data.Emotes.Search.Items[0].ID
	expected := "123"
	if id != expected {
		t.Fatalf("invalid result, expected=%q got=%q", expected, id)
	}
}
