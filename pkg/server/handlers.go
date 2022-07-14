package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"store/pkg/kvstore"
)

func ping(writer http.ResponseWriter, request *http.Request, username string,
	kvstore *kvstore.KVStore, logger *log.Logger) {
	fmt.Fprintf(writer, "pong")
}

func storeKey(writer http.ResponseWriter, request *http.Request, username string,
	kvstore *kvstore.KVStore, logger *log.Logger) {
	switch request.Method {
	case http.MethodPut:
		put(writer, request, username, kvstore, getKey(request.URL.Path), logger)
	case http.MethodGet:
		get(writer, request, username, kvstore, getKey(request.URL.Path), logger)
	case http.MethodDelete:
		deleteKey(writer, request, username, kvstore, getKey(request.URL.Path), logger)
	default:
		http.NotFound(writer, request)
	}
}

func put(writer http.ResponseWriter, request *http.Request, username string,
	store *kvstore.KVStore, key string, logger *log.Logger) {
	if key == "" {
		logger.Println("No key specified")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if username == "" {
		logger.Println("No owner specified")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	defer request.Body.Close()

	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		logger.Println("unable to read HTTP body: ", err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	value := string(bytes)

	logger.Printf("put key %s value %s owner %s", key, value, username)

	err = kvstore.Write(store, key, value, username)
	if err == nil {
		fmt.Fprint(writer, "OK")
	} else {
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden")
	}
}

func get(writer http.ResponseWriter, request *http.Request, username string,
	store *kvstore.KVStore, key string, logger *log.Logger) {
	if key == "" {
		logger.Println("No key specified")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	logger.Printf("get key %s", key)

	value, ok := kvstore.Read(store, key)
	if ok {
		fmt.Fprint(writer, value)
	} else {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		fmt.Fprint(writer, "404 key not found")
	}
}

func deleteKey(writer http.ResponseWriter, request *http.Request, username string,
	store *kvstore.KVStore, key string, logger *log.Logger) {
	if key == "" {
		logger.Println("No key specified")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if username == "" {
		logger.Println("No owner specified")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	logger.Printf("delete key %s owner %s", key, username)
	ok, err := kvstore.Delete(store, key, username)

	switch {
	case ok:
		fmt.Fprint(writer, "OK")
	case err != nil:
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden")
	default:
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		fmt.Fprint(writer, "404 key not found")
	}
}

func listKey(writer http.ResponseWriter, request *http.Request, username string,
	store *kvstore.KVStore, logger *log.Logger) {
	key := getKey(request.URL.Path)
	if key == "" {
		logger.Println("No key specified")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	logger.Printf("list key %s", key)

	entry := kvstore.List(store, key)
	if entry == nil {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		fmt.Fprint(writer, "404 key not found")

		return
	}

	bytes, err := json.Marshal(entry)
	if err != nil {
		logger.Print("Error marshalling key entry to JSON: ", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(bytes)
}

func listAll(writer http.ResponseWriter, request *http.Request, username string,
	store *kvstore.KVStore, logger *log.Logger) {
	logger.Print("list all keys")

	entries := kvstore.ListAll(store)

	bytes, err := json.Marshal(entries)
	if err != nil {
		logger.Print("Error marshalling entries to JSON: ", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(bytes)
}

var storeKeyRegex = regexp.MustCompile(`^\/[^\/].*\/(.*)$`)

// getKey extracts the key from the end of the REST path,
// or returns an empty string if not defined.
func getKey(path string) string {
	matches := storeKeyRegex.FindStringSubmatch(path)
	key := ""

	if len(matches) == 2 {
		key = matches[1]
	}

	return key
}
