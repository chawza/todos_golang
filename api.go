package main

import (
	"encoding/json"
	"net/http"
)

type ClientError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (r *ClientError) New(message string) ClientError {
	return ClientError{
		Message: message,
		Status:  400,
	}
}

func WriteClientError(w http.ResponseWriter, error ClientError) {
	responsePayload, err := json.Marshal(error)

	if err != nil {
		panic(err)
	}

	w.Write(responsePayload)
}
