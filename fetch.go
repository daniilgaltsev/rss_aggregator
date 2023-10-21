package main


import (
	"database/sql"
	"context"
	"encoding/xml"
	"net/http"
	"io"
	"errors"
	"fmt"
	"time"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"

	"github.com/daniilgaltsev/rss_aggregator/internal/database"
)

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"atom,attr"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description   string `xml:"description"`
		Generator     string `xml:"generator"`
		LastBuildDate string `xml:"lastBuildDate"`
		Item          []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Guid        string `xml:"guid"`
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}


func decodeXmlToRss(s string) (Rss, error) {
	result := Rss{}
	err := xml.Unmarshal([]byte(s), &result)
	return result, err
}

func fetchAndDecodeFeed(f database.Feed) (Rss, error) {
	xmlStr, err := fetchFeed(f)
	if err != nil {
		return Rss{}, err
	}
	return decodeXmlToRss(xmlStr)
}


func fetchFeed(f database.Feed) (string, error) {
	return fetchFeedFromUrl(f.Url)
}

func fetchFeedFromUrl(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode > 299 {
		return "", errors.New(fmt.Sprintf("HTTP error: %d", response.StatusCode))
	}
	if err != nil {
		return "", err
	}

	bodyStr := string(body)



	return bodyStr, nil
}


func updateFeeds(n int32, DB *database.Queries) error {
	var layouts = []string{
		time.Layout,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	}

	context := context.Background()
	feeds, err := DB.GetNextFeedsToFetch(context, n)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var failed int32 = 0
	for _, feed := range feeds {
		wg.Add(1)
		go func(feed database.Feed) {
			defer wg.Done()
			rss, err := fetchAndDecodeFeed(feed)
			if err != nil {
				atomic.AddInt32(&failed, 1)
				DB.UpdateLastFetchedAt(context, feed.ID)
				return
			}

			postsFailed := 0
			for _, item := range rss.Channel.Item {
				now := time.Now()
				var publishedAt sql.NullTime = sql.NullTime{Valid: false}
				for _, Layout := range layouts {
					parsed, err := time.Parse(Layout, item.PubDate)
					if err == nil {
						publishedAt = sql.NullTime{Time: parsed, Valid: true}
						break
					}
				}
				title := sql.NullString{Valid: false}
				if item.Title != "" {
					title = sql.NullString{String: item.Title, Valid: true}
				}
				link := sql.NullString{Valid: false}
				if item.Link != "" {
					link = sql.NullString{String: item.Link, Valid: true}
				}
				description := sql.NullString{Valid: false}
				if item.Description != "" {
					description = sql.NullString{String: item.Description, Valid: true}
				}
				
				params := database.CreatePostParams{
					ID: uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
					Title: title,
					Url: link,
					Description: description,
					PublishedAt: publishedAt,
					FeedID: feed.ID,
				}
				_, err := DB.CreatePost(context, params)
				if err != nil {
					postsFailed += 1
				}
			}

			err = DB.UpdateLastFetchedAt(context, feed.ID)
			if postsFailed > 0 || err != nil {
				atomic.AddInt32(&failed, 1)
			}
		}(feed)
	}
	wg.Wait()

	if failed > 0 {
		return errors.New(fmt.Sprintf("Failed to fetch %d feeds", failed))
	}
	return nil
}
