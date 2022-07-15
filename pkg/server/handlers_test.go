package server

import (
	"io"
	"net/http/httptest"
	"regexp"
	"store/pkg/kvstore"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/ping", nil)

	ping(recorder, request, "", nil, testLogger)

	checkResponse(t, recorder, 200, "pong")
}

func TestGetNoKeySpecified(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/store", nil) // no key specified
	store := kvstore.NewKVStore()

	storeKey(recorder, request, "", store, testLogger)

	checkResponse(t, recorder, 400, "Bad Request")

	kvstore.Close(store)
}

func TestGetNoSuchKey(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/store/abc", nil)
	store := kvstore.NewKVStore()

	storeKey(recorder, request, "", store, testLogger)

	checkResponse(t, recorder, 404, "Not Found")

	kvstore.Close(store)
}

func TestGetPopulatedKey(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/store/abc", nil)
	store := kvstore.NewKVStore()
	if err := kvstore.Write(store, "abc", "123", "my_user"); err != nil {
		t.Fatal("Error setting key: ", err)
	}

	storeKey(recorder, request, "", store, testLogger)

	checkResponse(t, recorder, 200, "123")

	kvstore.Close(store)
}

func TestPutNoKeySpecified(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("PUT", "/store", nil) // no key specified
	store := kvstore.NewKVStore()

	storeKey(recorder, request, "", store, testLogger)

	checkResponse(t, recorder, 400, "Bad Request")

	kvstore.Close(store)
}

func TestPutNoOwnerSpecified(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("PUT", "/store/abc", nil)
	store := kvstore.NewKVStore()

	storeKey(recorder, request, "", store, testLogger) // no owner

	checkResponse(t, recorder, 400, "Bad Request")

	kvstore.Close(store)
}

func TestPutValid(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("PUT", "/store/abc", strings.NewReader("123"))
	store := kvstore.NewKVStore()

	storeKey(recorder, request, "user_a", store, testLogger)

	checkResponse(t, recorder, 200, "OK")

	value, present := kvstore.Read(store, "abc")
	if !present {
		t.Fatal("Key PUT didn't write to store")
	}
	if value != "123" {
		t.Fatal("Invalid key value: ", value)
	}

	kvstore.Close(store)
}

func TestPutWrongOwner(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("PUT", "/store/abc", strings.NewReader("123"))
	store := kvstore.NewKVStore()
	kvstore.Write(store, "abc", "123", "user_b") // key already exists, owned by user_b

	storeKey(recorder, request, "user_a", store, testLogger) // attempt to write by user_a

	checkResponse(t, recorder, 403, "Forbidden\n")

	kvstore.Close(store)
}

func TestDeleteNoKeySpecified(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/store", nil) // no key specified
	store := kvstore.NewKVStore()

	storeKey(recorder, request, "", store, testLogger)

	checkResponse(t, recorder, 400, "Bad Request")

	kvstore.Close(store)
}

func TestDeleteNoOwnerSpecified(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/store/abc", nil)
	store := kvstore.NewKVStore()

	storeKey(recorder, request, "", store, testLogger) // no owner

	checkResponse(t, recorder, 400, "Bad Request")

	kvstore.Close(store)
}

func TestDeleteNoSuchKey(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/store/abc", nil)
	store := kvstore.NewKVStore()
	// key not in store

	storeKey(recorder, request, "user_a", store, testLogger)

	checkResponse(t, recorder, 404, "Not Found\n")

	kvstore.Close(store)
}

func TestDeleteValid(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/store/abc", nil)
	store := kvstore.NewKVStore()
	kvstore.Write(store, "abc", "123", "user_a")

	storeKey(recorder, request, "user_a", store, testLogger)

	checkResponse(t, recorder, 200, "OK")

	_, present := kvstore.Read(store, "abc")
	if present {
		t.Fatal("Key DELETE didn't delete in store")
	}

	kvstore.Close(store)
}

func TestDeleteWrongUser(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/store/abc", nil)
	store := kvstore.NewKVStore()
	kvstore.Write(store, "abc", "123", "user_a")

	storeKey(recorder, request, "user_b", store, testLogger) // different user

	checkResponse(t, recorder, 403, "Forbidden\n")

	kvstore.Close(store)
}

func TestListNoKeySpecified(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/list", nil) // no key specified
	store := kvstore.NewKVStore()

	listKey(recorder, request, "", store, testLogger)

	checkResponse(t, recorder, 400, "Bad Request")

	kvstore.Close(store)
}

func TestListNoSuchKey(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/list/abc", nil)
	store := kvstore.NewKVStore()
	// no key populated

	listKey(recorder, request, "user_a", store, testLogger)

	checkResponse(t, recorder, 404, "Not Found\n")

	kvstore.Close(store)
}

func TestListValid(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/list/abc", nil)
	store := kvstore.NewKVStore()
	kvstore.Write(store, "abc", "123", "user_a")

	listKey(recorder, request, "user_a", store, testLogger)

	checkResponse(t, recorder, 200, `{"key":"abc","owner":"user_a","writes":1,"reads":0,"age":0}`)

	kvstore.Close(store)
}

func TestListAllValid(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/list", nil)
	store := kvstore.NewKVStore()
	kvstore.Write(store, "abc", "123", "user_a")
	kvstore.Write(store, "def", "456", "user_b")
	kvstore.Write(store, "ghi", "789", "user_a")

	listAll(recorder, request, "user_a", store, testLogger)

	checkResponse(t, recorder, 200, `[
		{"key":"abc","owner":"user_a","writes":1,"reads":0,"age":0},
		{"key":"def","owner":"user_b","writes":1,"reads":0,"age":0},
		{"key":"ghi","owner":"user_a","writes":1,"reads":0,"age":0}
	]`)

	kvstore.Close(store)
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
