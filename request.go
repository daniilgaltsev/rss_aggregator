package main

import (
	"fmt"
	"errors"
	"net/http"
	"strings"
)

func getAuthorizationHeader(r *http.Request, name string) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("Authorization header required")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return "", errors.New("Malformed Authorization header")
	}

	if parts[0] != name {
		return "", errors.New(fmt.Sprintf("Authorization header must start with %s", name))
	}

	return parts[1], nil
}
