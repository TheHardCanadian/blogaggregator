package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/thehardcanadian/blogaggregator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}

		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf(("login requires a username"))
	}
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Printf("user %q does not exist", cmd.args[0])
		os.Exit(1)
	}
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("save username: %w", err)
	}

	fmt.Printf("Username %s has been set", s.cfg.CurrentUserName)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register requires a username")
	}
	name := cmd.args[0]
	now := time.Now().UTC()
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
	})

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			fmt.Printf("user %q already exists", name)
			os.Exit(1)
		}
		return fmt.Errorf("create user: %w", err)
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("Save current user: %w", err)
	}
	fmt.Printf("User created %s\n", user.Name)
	return nil
}

func handlerRemoveUsers(s *state, cmd command) error {
	if err := s.db.RemoveAllUsers(context.Background()); err != nil {
		return fmt.Errorf("remove users: %w", err)
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())

	if err != nil {
		return fmt.Errorf("error retrieving users: %w", err)
	}

	for _, user := range users {
		if user == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("agg needs a time_between_reqs argument, e.g. 30s, 1m, 1h")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("parsing time_between_reqs: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Printf("error scraping feeds %v\n", err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("register requires a name and url")
	}

	name, url := cmd.args[0], cmd.args[1]
	now := time.Now().UTC()

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})

	if err != nil {
		return fmt.Errorf("create feed error: %w", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("create feed follow error: %w", err)
	}

	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())

	if err != nil {
		return fmt.Errorf("Get feeds error %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Name of Feed: %s, URL: %s, User: %s ", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("need url to run command")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("get feed by url error: %w", err)
	}

	time := time.Now().UTC()

	feed_follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time,
		UpdatedAt: time,
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("create feed follow error: %w", err)
	}

	fmt.Printf("* Feed: %s\n * User: %s", feed_follow.FeedName, feed_follow.UserName)

	return nil

}

func handlerFollowing(s *state, cmd command, user database.User) error {

	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("get feed follow error %w", err)
	}

	fmt.Println("Currently subscribed to:")
	for _, feed := range following {
		fmt.Printf("* %s", feed.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("unfollow requires a url")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("get feed error: %w", err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("delete feed follow error %w", err)
	}

	fmt.Printf("Unfollowed %s\n", feed.Name)

	return nil
}
