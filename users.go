package main

import (
	"context"
	"net/http"
	"encoding/json"
	"time"
	"fmt"

	"github.com/google/uuid"

	"github.com/daniilgaltsev/rss_aggregator/internal/database"
)

func handlePostUsers(w http.ResponseWriter, r *http.Request, DB *database.Queries) {
	type request struct {
		Name string `json:"name"`
	}

	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	c := context.Background()
	now := time.Now()
	user, err := DB.CreateUser(c, database.CreateUserParams{
		Name: req.Name,
		CreatedAt: now,
		UpdatedAt: now,
		ID: uuid.New(),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		fmt.Println(err)
		return
	}

	respondWithJson(w, http.StatusCreated, user)
}
