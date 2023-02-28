package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/caseybraithwaite/concourse-slack-resource/structs"
)

func writeTimestamp(outputDirectory string, timestamp string) error {
	return os.WriteFile(
		fmt.Sprintf("%s/%s", outputDirectory, "ts"),
		[]byte(timestamp),
		0644,
	)
}

// essentially a no op
func main() {
	var getRequest structs.GetInput
	err := json.NewDecoder(os.Stdin).Decode(&getRequest)
	if err != nil {
		log.Fatal("couldn't decode stdin: ", err)
	}

	if !(len(os.Args) > 1) {
		log.Fatal("output folder not provided as arg")
	}
	outputDirectory := os.Args[1]

	if getRequest.Version.Timestamp == "" {
		err = json.NewEncoder(os.Stdout).Encode(`{"version":{"timestamp":"none"}}`)
		if err != nil {
			log.Fatal("couldn't encode response", err)
		}
	}

	err = writeTimestamp(outputDirectory, getRequest.Version.Timestamp)
	if err != nil {
		log.Fatal("error writing timestamp to disk")
	}

	rsp := Response{
		Version: ResponseVersion{
			Timestamp: getRequest.Version.Timestamp,
		},
	}
	err = json.NewEncoder(os.Stdout).Encode(rsp)
	if err != nil {
		log.Fatal("couldn't encode response", err)
	}
}
