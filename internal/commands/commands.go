package commands

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/johndosdos/blog_aggregator/internal/config"
	"github.com/johndosdos/blog_aggregator/internal/database"
)

type State struct {
	// store the state for each user
	Config *config.Config
	DB     *database.Queries
}

type Command struct {
	// e.g. gator <command name> [arguments]. store them here
	Name string
	Args []string
}

type Commands struct {
	// a handler is a function handling the <command name> argument
	Handlers map[string]func(*State, Command) error
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Handlers[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	if handler, ok := c.Handlers[cmd.Name]; !ok {
		return fmt.Errorf("unable to find command.")
	} else {
		return handler(s, cmd)
	}
}

func HandlerLogin(s *State, cmd Command) error {
	// cmd.Args is the username
	if len(cmd.Args) == 0 {
		return fmt.Errorf("username is required for login.")
	}

	username := cmd.Args[0]
	if username == "" {
		return fmt.Errorf("invalid username provided: %s", s.Config.CurrentUserName)
	}

	_, err := s.DB.GetUser(context.Background(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found, login failed. %w", err)
		}
		return err
	}

	// set current user to config
	if err := s.Config.SetUser(s.Config.GetFilename(), username); err != nil {
		return err
	}

	// print success message
	fmt.Printf("user has been logged in: %s.\n", s.Config.CurrentUserName)

	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("username is required for login.")
	}

	username := cmd.Args[0]

	// check if user exists in the database before creating a new entry
	_, err := s.DB.GetUser(context.Background(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := s.DB.CreateUser(
				context.Background(),
				database.CreateUserParams{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Name:      username,
				})
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
		}
	} else {
		return fmt.Errorf("user already exist.")
	}

	if err := s.Config.SetUser(s.Config.GetFilename(), username); err != nil {
		return err
	}
	fmt.Printf("user has been set: %s.\n", s.Config.CurrentUserName)

	return nil
}
