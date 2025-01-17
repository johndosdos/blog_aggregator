package commands

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/johndosdos/blog_aggregator/internal/config"
	"github.com/johndosdos/blog_aggregator/internal/database"
	"github.com/johndosdos/blog_aggregator/internal/rss"
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

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(s *State, cmd Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.DB.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		return handler(s, cmd, user)
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

func HandlerReset(s *State, cmd Command) error {
	// reset users table
	err := s.DB.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}

	// reset feeds table
	err = s.DB.DeleteFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete feeds: %w", err)
	}

	// reset users feed follows table
	err = s.DB.DeleteUsersFeedFollows(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete users feed subscriptions: %w", err)
	}

	fmt.Println("users deletion success!")
	fmt.Println("feeds deletion success!")
	fmt.Println("users feed subscriptions deletion success!")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("users database is empty!")
	}

	for _, v := range users {
		if v.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", v.Name)
		} else {
			fmt.Printf("* %s\n", v.Name)
		}
	}

	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("missing feed URL.")
	}
	feedURL := cmd.Args[0]

	feed, err := rss.FetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	fmt.Println(feed)

	return nil
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	switch len(cmd.Args) {
	case 0:
		return fmt.Errorf("missing feed name and URL.")
	case 1:
		return fmt.Errorf("missing feed URL.")
	}

	// the addfeed command accepts two arguments, feed name and feed URL
	feedName := cmd.Args[0]
	feedURL := cmd.Args[1]

	feed, err := s.DB.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name:      feedName,
			Url:       feedURL,
			UserID:    user.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w.", err)
	}

	// return a record of the feed the user recently followed
	_, err = s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow feed: %w", err)
	}

	fmt.Println("feed has been added.")

	fmt.Printf(`
	{
	ID: %s
	Created At: %s
	Updated At: %s
	Name: %s
	URL: %s
	UserID: %s	
	}
	`, feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.UserID)

	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	// this function does not accept any arguments
	// print all feeds in the feeds table

	dbFeeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}

	for _, record := range dbFeeds {
		fmt.Printf("{Name: %v, URL: %v, User: %v}\n", record.Name, record.Url, record.Username)
	}

	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("missing feed URL.")
	}

	feedUrl := cmd.Args[0]

	feed, err := s.DB.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("failed to get feed: %w", err)
	}

	// return a record of the feed the user recently followed
	feedFollowRecord, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow feed: %w", err)
	}

	fmt.Printf("{Feed name: %v, User: %v}\n", feedFollowRecord.Name_2, feedFollowRecord.Name)

	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	// this function does not accept any arguments
	// print all feeds the current user is following

	userFeeds, err := s.DB.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("failed to get user feeds: %w", err)
	}

	if len(userFeeds) == 0 {
		fmt.Println("Feeds: {}")
	} else {
		fmt.Printf("User: %s\n", user.Name)
		fmt.Println("Feeds: {")
		for _, record := range userFeeds {
			fmt.Printf("\t%v,\n", record.Name_2)
		}
		fmt.Println("}")
	}

	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("missing feed URL.")
	}

	feedUrl := cmd.Args[0]

	err := s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    feedUrl,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow feed: %w", err)
	}

	fmt.Println("feed unfollow success!")

	return nil
}
