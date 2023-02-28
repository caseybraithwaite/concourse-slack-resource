package main

type ResponseVersion struct {
	Timestamp string `json:"timestamp"`
}

type ResponseMetadata struct {
}

type Response struct {
	Version ResponseVersion `json:"version"`
}
