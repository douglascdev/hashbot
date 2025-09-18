package seventvapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddEmote(t *testing.T) {
	userByConnectionResult := `{
    "data": {
        "users": {
            "userByConnection": {
                "id": "123",
                "emoteSets": [
                    {
                        "id": "123",
                        "name": "User's Emotes",
                        "capacity": 300,
                        "ownerId": "123",
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
                                },
                                {
                                    "id": "01F6P1E7QR0002RDNAW6FFQ1E0",
                                    "alias": "TROLL"
                                },
                                {
                                    "id": "01EZZ5X4VR000CYST6006V20TK",
                                    "alias": "Binoculars"
                                },
                                {
                                    "id": "01F6ME7ADR0000WDA7ERT9H30R",
                                    "alias": "COPIUM"
                                },
                                {
                                    "id": "01F6NACCD80006SZ7ZW5FMWKWK",
                                    "alias": "Prayge"
                                },
                                {
                                    "id": "01HS0YC6PR00053R068FSDQXNG",
                                    "alias": "Wowie"
                                }
                            ]
                        }
                    }
                ],
                "style": {
                    "activeEmoteSetId": "123"
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
}`
	searchEmoteResult := `{
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
}`
	addEmoteResult := `	{
    "data": {
        "emotes": {
            "search": {
                "items": [
                    {
                        "id": "01H3G2Q7JR000243E9YKWJZ3TD",
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
}`
	count := 1
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the response you want to return
		w.WriteHeader(http.StatusOK)
		switch count {
		case 1:
			w.Write([]byte(userByConnectionResult))
		case 2:
			w.Write([]byte(searchEmoteResult))
		default:
			w.Write([]byte(addEmoteResult))
		}
		count += 1
	}))
	defer mockServer.Close()
	// Call the function with the mock server's URL
	err := AddEmoteWithQuery(mockServer.URL, "123", "", "")

	if err != nil {
		t.Fatal(err)
	}
}
