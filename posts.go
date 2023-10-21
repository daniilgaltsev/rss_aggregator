package main

import (
	"context"
	"net/http"
	"strconv"
	"fmt"

	"github.com/daniilgaltsev/rss_aggregator/internal/database"
)


func handleGetPosts(w http.ResponseWriter, r *http.Request, DB *database.Queries) {
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

	nStr := r.URL.Query().Get("n")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		n = 10
	}


	posts, err := DB.GetPostsByUser(context.Background(), database.GetPostsByUserParams{
		UserID: user.ID,
		Limit: int32(n),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		fmt.Println(err)
		return
	}
	respondWithJson(w, http.StatusOK, posts)
}
