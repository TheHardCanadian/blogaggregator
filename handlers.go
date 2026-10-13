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
		return fmt.Errorf("Error retrieving users: %w", err)
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
