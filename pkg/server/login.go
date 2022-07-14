package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"store/pkg/hash"
	"store/pkg/kvstore"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	adminUsername = "admin"

	users = map[string]string{
		"user_a":      "$argon2id$v=19$m=65536,t=3,p=2$1j5au/wHwkSi64OrwYSTfQ$zA23lAMgLkoVyNB3QXhF14licOD6M1Nf4Xr6/g4ErDg",
		"user_b":      "$argon2id$v=19$m=65536,t=3,p=2$U3e/x14UmLqn1FsmEsrprw$F0forUk8e9kKgEt4bNcmXxmoWr4t/gFP8MF4pt4uOM4",
		"user_c":      "$argon2id$v=19$m=65536,t=3,p=2$L2yrvKNY1KsPg1C5CvMQ5w$WjrSRwZ33GuhYc5vC8qlZmtvMub8Q4wvUD+rLses2BM",
		adminUsername: "$argon2id$v=19$m=65536,t=3,p=2$j9WX90NZZV/FPoPR5cNOtQ$qLYBlPLnx2n57LcRFjzN41tPMRZkiPf2/9t7quttlPg",
	}
)

const tokenExpiryMins = 5

type claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var jwtKey = []byte("ChangeMeThisIsNotSecure")

func login(writer http.ResponseWriter, request *http.Request, unused string,
	kvstore *kvstore.KVStore, logger *log.Logger) {
	username, password, present := request.BasicAuth()
	if !present {
		logger.Println("Error parsing basic auth credentials")
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	userStorePassword, ok := users[username]
	if !ok {
		logger.Println("Unknown user: ", username)
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	verified, err := hash.VerifyAgainstHash(password, userStorePassword)
	if err != nil {
		logger.Println("Error verifying password: ", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	if !verified {
		logger.Println("Password incorrect")
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	expirationTime := time.Now().Add(tokenExpiryMins * time.Minute)
	claims := &claims{username, jwt.StandardClaims{ExpiresAt: expirationTime.Unix(), Issuer: "MyRESTService"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		logger.Println("Error generating JWT: ", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	logger.Print("Returning token for user: ", username)
	fmt.Fprint(writer, "Bearer ", tokenString)
}

func withAccessLogAndSecurityCheck(store *kvstore.KVStore, accessLog *log.Logger,
	appLog *log.Logger, handlerFunc handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		accessLog.Printf("%s %s %s", request.RemoteAddr, request.Method, request.URL)

		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			appLog.Println("no Authorization header present")
			http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)

			return
		}

		prefix := "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			appLog.Println("invalid format bearer token")
			http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		tokenString := authHeader[len(prefix):]
		claims := &claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				appLog.Println("bearer token signature invalid: ", err)
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

				return
			}

			appLog.Println("bearer token parse error: ", err)
			http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		if !token.Valid {
			appLog.Println("bearer token is invalid")
			http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		handlerFunc(writer, request, claims.Username, store, appLog)
	}
}
