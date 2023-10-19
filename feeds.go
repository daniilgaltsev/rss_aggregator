package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/daniilgaltsev/rss_aggregator/internal/database"
)

func handlePostFeeds(w http.ResponseWriter, r *http.Request, DB *database.Queries) {
	type request struct {
		Name string `json:"name"`
		Url string `json:"url"`
	}

	apiKey, err := getAuthorizationHeader(r, "ApiKey")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid authorization header")
		return
	}
	
	user, err := DB.GetUser(context.Background(), apiKey)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	var req request
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	now := time.Now()
	feed, err := DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name: req.Name,
		Url: req.Url,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		fmt.Println(err)
		return
	}

	respondWithJson(w, http.StatusCreated, feed)
}

func handleGetFeeds(w http.ResponseWriter, r *http.Request, DB *database.Queries) {

	feeds, err := DB.GetFeeds(context.Background())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		fmt.Println(err)
		return
	}

	respondWithJson(w, http.StatusOK, feeds)
}
