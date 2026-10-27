package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/thehardcanadian/blogaggregator/internal/database"
	"github.com/thehardcanadian/blogaggregator/internal/rss"
)

/*

Write an aggregation function, I called mine scrapeFeeds. It should:
Get the next feed to fetch from the DB.
Mark it as fetched.
Fetch the feed using the URL (we already wrote this function)
Iterate over the items in the feed and print their titles to the console.

*/

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("fetchng next feed: %w", err)
	}
	now := time.Now().UTC()
	_, err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		ID:            feed.ID,
		LastFetchedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt:     now,
	})

	if err != nil {
		return fmt.Errorf("issue with fetchfeed: %w", err)
	}

	rssFeed, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("fetch rss feed: %w", err)
	}

	fmt.Printf("== %s (%d items) ===\n", feed.Name, len(rssFeed.Channel.Item))
	for _, item := range rssFeed.Channel.Item {
		publishedAt := parsePublishedAt(item.PubDate)
		_, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   now,
			UpdatedAt:   now,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				continue
			}
			fmt.Printf("creating post %q: %v\n", item.Title, err)
			continue

		}
	}

	return nil
}

func parsePublishedAt(raw string) sql.NullTime {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		"Mon, 2 Jan 2006 15:04:05 MST",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, strings.TrimSpace(raw)); err == nil {
			return sql.NullTime{Time: t, Valid: true}
		}
	}
	return sql.NullTime{Valid: false}
}
