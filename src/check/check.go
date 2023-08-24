package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	err := json.NewEncoder(os.Stdout).Encode([]string{})
	if err != nil {
		log.Fatal("couldn't encode response", err)
	}
}
