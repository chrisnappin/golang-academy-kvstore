package main

import (
	"flag"
	"log"
	"os"

	"store/pkg/kvstore"
	"store/pkg/server"
)

const restServerPort = 8000

func main() {
	htaccessFile, err := os.OpenFile("htaccess.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	htaccessLogger := log.New(htaccessFile, "HTACCESS", log.Ldate|log.Ltime)

	storeFile, err := os.OpenFile("store.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	appLogger := log.New(storeFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	appLogger.Println("Starting up...")

	port := flag.Int("port", restServerPort, "HTTP server port to listen on")
	flag.Parse()

	store := kvstore.NewKVStore()
	server.Start(*port, store, htaccessLogger, appLogger)

	appLogger.Println("Shutting down...")
	kvstore.Close(store)

	htaccessFile.Close()
	storeFile.Close()
}
