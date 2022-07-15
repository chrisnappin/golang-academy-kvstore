// Package server provides a REST API for a key value store.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"store/pkg/kvstore"
	"time"
)

const forceIdleRequestsClosedAfterSecs = 5

// common arguments needed by most of the handlers.
type handler func(w http.ResponseWriter, r *http.Request, username string, kvstore *kvstore.KVStore, logger *log.Logger)

func withAccessLog(store *kvstore.KVStore, accessLog *log.Logger, appLog *log.Logger, h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessLog.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		h(w, r, "", store, appLog)
	}
}

func shutdown(writer http.ResponseWriter, r *http.Request, username string,
	ksstore *kvstore.KVStore, logger *log.Logger, c chan<- int) {
	if username == adminUsername {
		fmt.Fprintf(writer, "OK")
		logger.Println("Requesting shut down of REST server")
		c <- 1
	} else {
		logger.Println("Ignoring shutdown request not from admin user")
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		fmt.Fprintf(writer, "Forbidden")
	}
}

// Start sets up the REST server and starts it going. This function only returns after the server has been shutdown.
func Start(port int, store *kvstore.KVStore, accessLog *log.Logger, appLog *log.Logger) {
	server := &http.Server{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	gracefulShutdown := make(chan int)

	// endpoints that don't require JWT bearer tokens
	http.HandleFunc("/ping", withAccessLog(store, accessLog, appLog, ping))
	http.HandleFunc("/login", withAccessLog(store, accessLog, appLog, login))

	// endpoints that do require JWT bearer tokens
	http.HandleFunc("/store/", withAccessLogAndSecurityCheck(store, accessLog, appLog, storeKey))
	http.HandleFunc("/list/", withAccessLogAndSecurityCheck(store, accessLog, appLog, listKey))
	http.HandleFunc("/list", withAccessLogAndSecurityCheck(store, accessLog, appLog, listAll))
	http.HandleFunc("/shutdown", withAccessLogAndSecurityCheck(store, accessLog, appLog,
		func(w http.ResponseWriter, r *http.Request, username string, s *kvstore.KVStore, logger *log.Logger) {
			shutdown(w, r, username, s, logger, gracefulShutdown)
		}))

	appLog.Printf("Starting REST server on port %d", port)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLog.Println(err)
			os.Exit(-2)
		}
	}()

	<-gracefulShutdown

	appLog.Println("Gracefully shutting down REST server")

	// trigger graceful HTTP server shutdown, but force shutdown
	// if active handlers are still processing/stuck after a set time limit
	ctx, cancel := context.WithTimeout(context.Background(), forceIdleRequestsClosedAfterSecs*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		appLog.Fatalf("REST server shutdown failed: %+v", err)
	}

	appLog.Println("REST server showdown completed")
}
