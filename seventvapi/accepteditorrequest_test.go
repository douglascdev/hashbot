package seventvapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAcceptEditorRequestSuccess(t *testing.T) {
	addEmoteResult := `	{
    "data": {
        "userEditors": {
            "editor": {
                "updateState": {
                    "userId": "01GBE7XFHG000E7PMG5MJNHQ8M",
                    "editorId": "01K61PCMXDQD90FQDACGEM07ZV",
                    "state": "ACCEPTED",
                    "notes": null,
                    "addedById": "01GBE7XFHG000E7PMG5MJNHQ8M",
                    "addedAt": "2025-09-30T23:29:30.723+00:00",
                    "updatedAt": "2025-09-30T23:30:28.752+00:00",
                    "searchUpdatedAt": null
                }
            }
        }
    },
    "extensions": {
        "analyzer": {
            "complexity": 11,
            "depth": 4
        }
    }
}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(addEmoteResult))
	}))
	defer mockServer.Close()

	err := AcceptEditorRequest(mockServer.URL, "123", "", "")

	if err != nil {
		t.Fatal(err)
	}
}

func TestAcceptEditorRequestFail(t *testing.T) {
	addEmoteResult := `{
    "data": null,
    "extensions": {
        "analyzer": {
            "complexity": 11,
            "depth": 4
        }
    },
    "errors": [
        {
            "message": "BAD_REQUEST editor is not pending",
            "locations": [
                {
                    "line": 7,
                    "column": 13
                }
            ],
            "path": [
                "userEditors",
                "editor",
                "updateState"
            ],
            "extensions": {
                "code": "BAD_REQUEST",
                "fields": {},
                "message": "BAD_REQUEST editor is not pending",
                "status": 400
            }
        }
    ]
}`
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(addEmoteResult))
	}))
	defer mockServer.Close()

	err := AcceptEditorRequest(mockServer.URL, "123", "", "")

	if err == nil {
		t.Fatal(err)
	}
}
