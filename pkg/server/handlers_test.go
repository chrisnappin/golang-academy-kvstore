package server

import (
	"testing"
)

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
}
