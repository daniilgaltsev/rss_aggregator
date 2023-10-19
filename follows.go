package main


import (
	"context"
	"time"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"


	"github.com/daniilgaltsev/rss_aggregator/internal/database"
)


func follow(feedid, userid uuid.UUID, DB *database.Queries) (database.Follow, error) {
	now := time.Now()

	follow, err := DB.CreateFollow(context.Background(), database.CreateFollowParams{
		ID: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		FeedID: feedid,
		UserID: userid,
	})

	return follow, err
}

func handlePostFeedFollows(w http.ResponseWriter, r *http.Request, DB *database.Queries) {
	type request struct {
		FeedID uuid.UUID `json:"feed_id"`
	}

	apiKey, err := getAuthorizationHeader(r, "ApiKey")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid authorization header")
		return
	}

	var req request
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	feedId := req.FeedID

	feed, err := DB.GetFeed(context.Background(), feedId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Feed not found")
		return
	}

	user, err := DB.GetUser(context.Background(), apiKey)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	follow, err := follow(feed.ID, user.ID, DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJson(w, http.StatusCreated, follow)
}

func deleteFollow(followId uuid.UUID, DB *database.Queries) error {
	return DB.DeleteFollow(context.Background(), followId)
}

func handleDeleteFeedFollows(w http.ResponseWriter, r *http.Request, DB *database.Queries) {
	apiKey, err := getAuthorizationHeader(r, "ApiKey")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid authorization header")
		return
	}

	followIdStr := chi.URLParam(r, "feedFollowId")
	followId, err := uuid.Parse(followIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid feed follow id")
		return
	}
	follow, err := DB.GetFollow(context.Background(), followId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Feed follow not found")
		return
	}

	user, err := DB.GetUser(context.Background(), apiKey)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	if follow.UserID != user.ID {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	err = deleteFollow(follow.ID, DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}


func handleGetFeedFollows(w http.ResponseWriter, r *http.Request, DB *database.Queries) {
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

	follows, err := DB.GetFollowsForUser(context.Background(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJson(w, http.StatusOK, follows)
}
