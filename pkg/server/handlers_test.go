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
