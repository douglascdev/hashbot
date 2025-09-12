package seventvapi

import (
	"testing"
)

func TestRemoveEmote(t *testing.T) {
	////// Create a mock server
	////mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	////	// Set the response you want to return
	////	w.WriteHeader(http.StatusOK)
	////	w.Write([]byte(`
	////{
	////"data": {
	////    "emoteSets": {
	////        "emoteSet": {
	////            "removeEmote": {
	////                "id": "123"
	////            }
	////        }
	////    }
	////},
	////"extensions": {
	////    "analyzer": {
	////        "complexity": 4,
	////        "depth": 4
	////    }
	////}
	////}`))
	////}))
	////defer mockServer.Close()
	////// Call the function with the mock server's URL
	////result, err := RemoveEmote(mockServer.URL, "123", )
	////if err != nil {
	////	t.Fatal(err)
	////}

	// //// Check the result
	// //id := result.Data.Users.UserByConnection.EmoteSets[0].ID
	// //expected := "123"
	// //if id != expected {
	// //	t.Fatalf("invalid result, expected=%q got=%q", expected, id)
	// //}
}
