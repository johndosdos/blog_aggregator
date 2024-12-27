package commands

import (
	"fmt"

	"github.com/johndosdos/blog_aggregator/internal/config"
)

type State struct {
	// store the state for each user
	Config *config.Config
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

	// set current user to config
	if err := s.Config.SetUser(s.Config.GetFilename(), username); err != nil {
		return err
	}

	// print success message
	fmt.Printf("user has been set: %s.\n", s.Config.CurrentUserName)

	return nil
}
