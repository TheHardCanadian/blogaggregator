package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
		fmt.Println(item.Title)
	}

	return nil
}
