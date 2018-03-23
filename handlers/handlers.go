package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

type errorResponse struct {
	Errors map[string][]string `json:"errors"`
}

func JsonErrorNotFoundResponse() []byte {
	resp := map[string]string{
		"status": "404",
		"error":  "Not Found",
	}

	respMarshalled, err := json.Marshal(&resp)
	if err != nil {
		panic(fmt.Errorf("Can not Marshall: %s", err))
	}
	return respMarshalled
}

func JsonErrorResponse(field, error string) []byte {
	resp := &errorResponse{
		Errors: map[string][]string{
			field: []string{
				error,
			},
		},
	}

	respMarshalled, err := json.Marshal(&resp)
	if err != nil {
		panic(fmt.Errorf("Can not Marshall: %s", err))
	}
	return respMarshalled
}

func JsonErrorResponseFromValidator(err error) []byte {
	var errorMsgs = make(map[string][]string)
	for _, err := range err.(validator.ValidationErrors) {
		errorMsgs[strings.ToLower(err.StructField())] = []string{err.Tag()}
	}

	resp := &errorResponse{
		Errors: errorMsgs,
	}

	respMarshalled, err := json.Marshal(&resp)
	if err != nil {
		panic(fmt.Errorf("Can not Marshall: %s", err))
	}
	return respMarshalled
}
