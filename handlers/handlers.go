package handlers

import (
	"encoding/json"
	"fmt"
)

type errorResponse struct {
	Errors map[string][]string `json:"errors"`
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
