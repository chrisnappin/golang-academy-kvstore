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
	// TODO: make this a file, e.g. htaccess.log
	htaccessLogger := log.New(os.Stdout, "HTACCESS", log.Ldate|log.Ltime)

	// TODO: make this a file, e.g. store.log
	appLogger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	appLogger.Println("Starting up...")

	port := flag.Int("port", restServerPort, "HTTP server port to listen on")
	flag.Parse()

	store := kvstore.NewKVStore()
	server.Start(*port, store, htaccessLogger, appLogger)

	appLogger.Println("Shutting down...")
	kvstore.Close(store)
}
