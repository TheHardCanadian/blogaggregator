package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	config "github.com/thehardcanadian/blogaggregator/internal"
	"github.com/thehardcanadian/blogaggregator/internal/database"
)

/*
//This is one of my least favorite things working with SQL in Go currently.
  You have to import the driver, but you don't use it directly anywhere in your code.
  The underscore tells Go that you're importing it for its side effects, not because you need to use it.
*/

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()

	if err != nil {
		fmt.Printf("Reading config: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)

	if err != nil {
		fmt.Printf("Error opening postgres database: %v\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	s := state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		handlers: map[string]func(*state, command) error{
			"login":    handlerLogin,
			"register": handlerRegister,
			"reset":    handlerRemoveUsers,
			"users":    handlerGetUsers,
		},
	}

	if len(os.Args) < 2 {
		fmt.Println("expected a command")
		os.Exit(1)
	}

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err := cmds.run(&s, cmd); err != nil {
		fmt.Println("error running command: %w", err)
		os.Exit(1)
	}
}
