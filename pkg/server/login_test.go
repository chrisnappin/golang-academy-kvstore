package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"store/pkg/kvstore"
	"testing"
)

// to enable logging change ioutil.Discard to os.Stdout.
var testLogger = log.New(ioutil.Discard, "Code under test: ", log.Ldate|log.Ltime|log.Lshortfile)

func TestLoginNoCreds(t *testing.T) {
	recorder, request := setupLoginRequest("", "") // no auth header

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

func TestWithAccessLogAndSecurityCheckNoHeader(t *testing.T) {
	recorder, request := setupBearerTokenRequest("") // no auth header
	handlerCalled := false

	withAccessLogAndSecurityCheck(nil, testLogger, testLogger,
		func(w http.ResponseWriter, r *http.Request, username string, kvstore *kvstore.KVStore, logger *log.Logger) {
			handlerCalled = true
		})(recorder, request)

	checkResponse(t, recorder, 403, "Forbidden")

	if handlerCalled {
		t.Fatal("Handler should not have been called")
	}
}

func TestWithAccessLogAndSecurityCheckInvalidHeader(t *testing.T) {
	recorder, request := setupBearerTokenRequest("wibble") // not "Bearer <token>"
	handlerCalled := false

	withAccessLogAndSecurityCheck(nil, testLogger, testLogger,
		func(w http.ResponseWriter, r *http.Request, username string, kvstore *kvstore.KVStore, logger *log.Logger) {
			handlerCalled = true
		})(recorder, request)

	checkResponse(t, recorder, 401, "Unauthorized")

	if handlerCalled {
		t.Fatal("Handler should not have been called")
	}
}

func TestWithAccessLogAndSecurityCheckInvalidToken(t *testing.T) {
	recorder, request := setupBearerTokenRequest("Bearer wibble")
	handlerCalled := false

	withAccessLogAndSecurityCheck(nil, testLogger, testLogger,
		func(w http.ResponseWriter, r *http.Request, username string, kvstore *kvstore.KVStore, logger *log.Logger) {
			handlerCalled = true
		})(recorder, request)

	checkResponse(t, recorder, 401, "Unauthorized")

	if handlerCalled {
		t.Fatal("Handler should not have been called")
	}
}

func TestWithAccessLogAndSecurityCheckValid(t *testing.T) {
	// first call login to generate a valid token
	recorder, request := setupLoginRequest("user_a", "passwordA")
	login(recorder, request, "", nil, testLogger)
	defer recorder.Result().Body.Close()
	bytes, err := io.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal("Error reading response: ", err)
	}

	actualBody := string(bytes)
	token := actualBody[len("Bearer "):]

	// then use the token in a subsequent call
	recorder, request = setupBearerTokenRequest("Bearer " + token)
	handlerCalled := false
	usernamePassed := ""

	withAccessLogAndSecurityCheck(nil, testLogger, testLogger,
		func(w http.ResponseWriter, r *http.Request, username string, kvstore *kvstore.KVStore, logger *log.Logger) {
			handlerCalled = true
			usernamePassed = username
			fmt.Fprint(w, "All is ok")
		})(recorder, request)

	checkResponse(t, recorder, 200, "All is ok")

	if !handlerCalled {
		t.Fatal("Handler should have been called")
	}

	if usernamePassed != "user_a" {
		t.Fatalf("Handler passed wrong username, expecting %s but got %s", "user_a", usernamePassed)
	}
}

func setupBearerTokenRequest(auth string) (*httptest.ResponseRecorder, *http.Request) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/list", nil)

	if auth != "" {
		request.Header.Add("Authorization", auth)
	}

	return recorder, request
}
