package main

import (
	"log"
	"os"
	"store/pkg/hash"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	if len(os.Args) != 2 {
		logger.Fatal("Usage: hash <password>")
	}

	password := os.Args[1]

	logger.Printf("Hashing password \"%s\"", password)

	hashedPassword, err := hash.GenerateHash(password)
	if err != nil {
		logger.Fatalln("Error generating hash: ", err)
	}

	logger.Printf("Hashed password \"%s\"", hashedPassword)

	equal, err := hash.VerifyAgainstHash(password, hashedPassword)
	if err != nil {
		logger.Fatalln("Error verifying hash: ", err)
	}

	logger.Println("Verified: ", equal)
}
