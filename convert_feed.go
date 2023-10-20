package main

import (
	"time"

	"github.com/google/uuid"

	"github.com/daniilgaltsev/rss_aggregator/internal/database"
)

type Feed struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string
	Url           string
	UserID        uuid.UUID
	LastFetchedAt time.Time
}

func databaseFeedToFeed(f database.Feed) Feed {
	LastFetchedAt := time.Time{}
	if f.LastFetchedAt.Valid {
		LastFetchedAt = f.LastFetchedAt.Time
	}

	return Feed{
		ID: f.ID,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
		Name: f.Name,
		Url: f.Url,
		UserID: f.UserID,
		LastFetchedAt: LastFetchedAt,
	}
}

func databaseFeedsToFeeds(fs []database.Feed) []Feed {
	feeds := make([]Feed, len(fs))
	for i, f := range fs {
		feeds[i] = databaseFeedToFeed(f)
	}
	return feeds
}
