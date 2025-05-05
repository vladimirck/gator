package main

import (
	"context"
	"database/sql"
	"html"

	//"strconv"

	//"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
	"github.com/vladimirck/gator/internal/config"
	"github.com/vladimirck/gator/internal/database"

	"errors"
	"os"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmdNames    []string
	handlersMap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	fn, cmd_exist := c.handlersMap[cmd.name]
	if cmd_exist {
		return fn(s, cmd)
	}

	return errors.New("The command " + cmd.name + " does not exist")
}

func (c *commands) register(name string, f func(*state, command) error) error {
	for i := 0; i < len(c.cmdNames); i++ {
		if name == c.cmdNames[i] {
			return errors.New("this command is already registered")
		}
	}
	c.cmdNames = append(c.cmdNames, name)
	c.handlersMap[name] = f
	return nil
}

func handlerLogin(s *state, cmd command) error {

	if len(cmd.args) != 2 {
		return errors.New("the command login expect a single argument")
	}

	if _, err := s.db.GetUserByName(context.Background(), cmd.args[1]); err != nil {
		return errors.New("the user is not registered")
	}

	if err := s.cfg.SetUser(cmd.args[1]); err != nil {
		return errors.New("the user could not be set")
	}

	fmt.Printf("the user %s has been set", cmd.args[1])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return errors.New("the command login expect a single argument")
	}

	userData, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      cmd.args[1],
		},
	)

	if err != nil {
		fmt.Printf("no se pudo registrar el usuario %s: %v\n", cmd.args[1], err)
		os.Exit(1)
		return err
	}

	if err := s.cfg.SetUser(cmd.args[1]); err != nil {
		return err
	}

	fmt.Printf("User registered: %v", userData)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("the command login expect no argument")
	}

	if err := s.db.Reset(context.Background()); err != nil {
		return err
	}

	fmt.Print("all user has been erased from the database\n")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("the command login expect no argument")
	}

	users, err := s.db.GetUsers(context.Background())

	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 3 {
		return errors.New("the command addFeed expect two arguments")
	}

	feed, err := s.db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:     uuid.New(),
			Name:   cmd.args[1],
			Url:    cmd.args[2],
			UserID: user.ID,
		},
	)

	if err != nil {
		return fmt.Errorf("The RSS feed could not be save in the database: %v", err)
	}

	_, err = s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:     uuid.New(),
			FeedID: feed.ID,
			UserID: feed.UserID,
		},
	)

	if err != nil {
		return fmt.Errorf("The feedfollow was not created: %v", err)
	}

	fmt.Println("The RSS feedwas saved successfully!")
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("the command feeds expect no argument")
	}

	feeds, err := s.db.GetFeeds(context.Background())

	if err != nil {
		return err
	}

	fmt.Printf("******List of all feeds in the database**********\n\n")

	for _, feed := range feeds {
		fmt.Printf("Name of the RSS feed: %s\n", feed.RssName)
		fmt.Printf("                 URL: %s\n", feed.RssUrl)
		fmt.Printf("  User who create it: %s\n", feed.UserName)
		fmt.Printf("-----------------\n\n")
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return errors.New("the command login expect one argument")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[1])
	if err != nil {
		fmt.Printf("Error during parsin the time duration: %v", err)
		return err
	}

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		fmt.Printf("Scraping the web: %s\n", time.Now().GoString())
		err := s.scrapeFeeds()

		if err != nil {
			return err
		}
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return errors.New("the command follow expect one argument")
	}

	rssFeed, err := s.db.GetFeedByURL(context.Background(), cmd.args[1])

	if err != nil {
		return fmt.Errorf("The URL was not found in the database: %v", err)
	}

	feedFollows, err := s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:     uuid.New(),
			FeedID: rssFeed.ID,
			UserID: user.ID,
		},
	)

	if err != nil {
		return fmt.Errorf("The feed follow could not be created: %v", err)
	}

	fmt.Printf("--list of all feed follows---\n\n")

	for _, feedFollow := range feedFollows {
		fmt.Printf("name of the feed: %s\n", feedFollow.FeedName)
		fmt.Printf("URL of the feed: %s\n", feedFollow.FeedUrl)
		fmt.Printf("URL of the feed: %s\n", feedFollow.UserName)
		fmt.Printf("-------------\n\n")
	}
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("the command following expect no argument")
	}

	feedFollows, err := s.db.GetFeedFollowForUser(context.Background(), user.ID)

	if err != nil {
		return fmt.Errorf("The user wasnt found in the database: %v", err)
	}

	fmt.Printf("--list of all feed follows---\n\n")

	for _, feedFollow := range feedFollows {
		fmt.Printf("  name of the feed: %s\n", feedFollow.FeedName)
		fmt.Printf("   URL of the feed: %s\n", feedFollow.FeedUrl)
		fmt.Printf("  user of the feed: %s\n", feedFollow.UserName)
		fmt.Printf(" Last time fetched: %v\n", feedFollow.LastFetchedAt)
		fmt.Printf("-------------\n\n")
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return errors.New("the command following expect one argument")
	}

	err := s.db.DeleteFeedFollow(context.Background(),
		database.DeleteFeedFollowParams{
			Url:    cmd.args[1],
			UserID: user.ID,
		},
	)

	if err != nil {
		return fmt.Errorf("The URL is not in the database: %v", err)
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 2 || len(cmd.args) < 1 {
		return errors.New("the command following expect one or two argument")
	}

	posts, err := s.db.GetPostsForUser(context.Background())

	if err != nil {
		fmt.Printf("Posts could no be loaded from the database\n", err)
		return err
	}

	for _, post := range posts {
		fmt.Printf("      Title: %s\n", post.Title)
		fmt.Printf("Description: %s\n", post.Description)
		fmt.Printf("        URL: %s\n", post.Url)
		fmt.Printf("------------------: %s\n\n")
	}

	if err != nil {
		return fmt.Errorf("The URL is not in the database: %v", err)
	}

	return nil
}

func middleWareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			fmt.Printf("The user could not be authenticated.\n")
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}

func (s *state) scrapeFeeds() error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())

	if err != nil {
		fmt.Printf("No more RSS feed to fetch")
		return err
	}

	if err := s.db.MarkFeedFetched(context.Background(), feed.ID); err != nil {
		fmt.Printf("Could not marked the feed as fetch!: %s", err)
		return err
	}

	var rssFeed *RSSFeed

	rssFeed, err = fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("The URL could not be fetch: %", err)
		return err
	}

	fmt.Printf("RSS feed title: %s\n\n", html.UnescapeString(rssFeed.Channel.Title))
	for _, item := range rssFeed.Channel.Item {
		pubTime, _ := time.Parse(time.RFC3339, item.PubDate)
		_ = s.db.CreatePost(context.Background(),
			database.CreatePostParams{
				ID:          uuid.New(),
				Title:       html.UnescapeString(item.Title),
				Url:         item.Link,
				Description: item.Description,
				FeedID:      feed.ID,
				PublishedAt: sql.NullTime{Time: pubTime, Valid: true},
			},
		)
	}

	return nil

}

func main() {
	gatorState := state{}
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("The configuration file could no be read: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Printf("The database could not be opened: %v\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	gatorState.cfg = &cfg
	gatorState.db = dbQueries

	gatorCommands := commands{
		cmdNames:    []string{},
		handlersMap: map[string]func(*state, command) error{},
	}

	if err := gatorCommands.register("login", handlerLogin); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("register", handlerRegister); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("reset", handlerReset); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("users", handlerUsers); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("agg", handlerAgg); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("addfeed", middleWareLoggedIn(handlerAddFeed)); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("feeds", handlerFeeds); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("follow", middleWareLoggedIn(handlerFollow)); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("following", middleWareLoggedIn(handlerFollowing)); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("unfollow", middleWareLoggedIn(handlerUnfollow)); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if err := gatorCommands.register("browse", middleWareLoggedIn(handlerBrowse)); err != nil {
		fmt.Printf("The command could not be registeres\n")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Printf("No commando to run\n")
		os.Exit(1)
	}

	if err := gatorCommands.run(&gatorState, command{name: os.Args[1], args: os.Args[1:]}); err != nil {
		fmt.Printf("Error while executing the command: %v\n", err)
		os.Exit(1)
	}

}
