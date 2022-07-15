package server

import (
	"io"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestPing(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/ping", nil)

	ping(recorder, request, "", nil, nil)

	checkResponse(t, recorder, 200, "pong")
}

// checkResponse checks for the expected response code and body. If expectedBodyRegex is empty this is
// used to mean to check the response body is empty.
func checkResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, expectedBodyRegex string) {
	t.Helper()
	if actualCode := recorder.Result().StatusCode; actualCode != expectedCode {
		t.Fatalf("Wrong response code, expected %d, got %d", expectedCode, actualCode)
	}

	defer recorder.Result().Body.Close()
	bytes, err := io.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal("Error reading response: ", err)
	}

	actualBody := string(bytes)
	if len(expectedBodyRegex) > 0 {
		regex := regexp.MustCompile(expectedBodyRegex)
		if regex.Find([]byte(actualBody)) == nil {
			t.Fatalf("Wrong response body, expected to match regex \"%s\" but got \"%s\"", expectedBodyRegex, actualBody)
		}
	} else if len(actualBody) > 0 {
		t.Fatalf("Wrong response body, expected empty but got \"%s\"", actualBody)
	}
}

func TestPathWithKey(t *testing.T) {
	checkGetKey(t, "/store/abc", "abc")
	checkGetKey(t, "/store/", "")
	checkGetKey(t, "/store", "")
	checkGetKey(t, "/list/123", "123")
	checkGetKey(t, "/list", "")
}

func checkGetKey(t *testing.T, path string, expectedKey string) {
	t.Helper()
	if key := getKey(path); key != expectedKey {
		t.Fatalf("Error - expecting %s but got %s", expectedKey, key)
	}

	if key := getKeyAlt(path); key != expectedKey {
		t.Fatalf("Error - expecting %s but got %s", expectedKey, key)
	}
}

func BenchmarkGetKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getKey("/store/123")
	}
}

func BenchmarkGetKeyAlt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getKeyAlt("/store/123")
	}
}
