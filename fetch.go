package main


import (
	"context"
	"encoding/xml"
	"net/http"
	"io"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

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
			fmt.Println("-----(not thread safe logging of a fetched feed)----")
			fmt.Println(rss.Channel.Title)
			fmt.Println(" Number of items:", len(rss.Channel.Item))

			err = DB.UpdateLastFetchedAt(context, feed.ID)
			if err != nil {
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
