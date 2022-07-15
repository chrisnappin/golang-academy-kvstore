package server

import (
	"encoding/base64"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// to diable logging change os.Stdout to ioutil.Discard.
var testLogger = log.New(os.Stdout, "Code under test: ", log.Ldate|log.Ltime|log.Lshortfile)

func TestLoginNoCreds(t *testing.T) {
	recorder, request := setupLoginRequest("", "")

	login(recorder, request, "", nil, testLogger)

	checkResponse(t, recorder, 401, "Unauthorized\n")
}

func TestLoginUnknownUser(t *testing.T) {
	recorder, request := setupLoginRequest("noone", "password")

	login(recorder, request, "", nil, testLogger)

	checkResponse(t, recorder, 401, "Unauthorized\n")
}

func TestLoginWrongPassword(t *testing.T) {
	recorder, request := setupLoginRequest("user_a", "wrongpassword")

	login(recorder, request, "", nil, testLogger)

	checkResponse(t, recorder, 401, "Unauthorized\n")
}

func TestLoginValid(t *testing.T) {
	recorder, request := setupLoginRequest("user_a", "passwordA")

	login(recorder, request, "", nil, testLogger)

	checkResponse(t, recorder, 200, "Bearer .*")
}

func setupLoginRequest(username string, password string) (*httptest.ResponseRecorder, *http.Request) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/login", nil)

	if username != "" {
		encodedCreds := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		request.Header.Add("Authorization", "Basic "+encodedCreds)
	}

	return recorder, request
}
