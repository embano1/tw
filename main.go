package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
)

var version string
var commit string
var date string

type secret struct {
	ConsumerKey       string `json:"Consumer_Key"`
	ConsumerSecret    string `json:"Consumer_Secret"`
	AccessToken       string `json:"Access_Token"`
	AccessTokenSecret string `json:"Access_Token_Secret"`
}

func getLikes(a *anaconda.TwitterApi, pretty *bool) ([]anaconda.Tweet, error) {

	// https://api.twitter.com/1.1
	// Specifies the number of Tweets to try and retrieve, up to a maximum of 200 per distinct request.
	// The value of count is best thought of as a limit to the number of Tweets to return
	v := url.Values{}
	v.Set("count", "200")

	var maxID int64
	var fav []anaconda.Tweet

	w := wow.New(os.Stdout, spin.Get(spin.Dots), " Working...")

	if *pretty {
		w.Start()
		defer func() {
			w.Stop()
			w.PersistWith(spin.Spinner{Frames: []string{"👍"}}, "  We're good...")
		}()
	}

	for {
		f, err := a.GetFavorites(v)
		if err != nil {
			return nil, err
		}

		if len(f) == 0 {
			break
		}

		// Remember the lowest max_id of tweets returned so we can start from there in the next round, if any
		maxID = f[len(f)-1].Id - 1
		v.Set("max_id", strconv.FormatInt(maxID, 10))

		fav = append(fav, f...)
	}

	return fav, nil
}

func newCreds(f *string) (*secret, error) {
	s := &secret{}

	// Check if env vars are set if no credentials file provided and fail early
	if *f == "" {
		if v := os.Getenv("TW_CONSUMER_KEY"); v != "" {
			s.ConsumerKey = v
		} else {
			return nil, errors.New("Environment variable TW_CONSUMER_KEY not set. Set variable or provide a credentials file (-f)")
		}

		if v := os.Getenv("TW_CONSUMER_SECRET"); v != "" {
			s.ConsumerSecret = v
		} else {
			return nil, errors.New("Environment variable TW_CONSUMER_SECRET not set. Set variable or provide a credentials file (-f)")
		}

		if v := os.Getenv("TW_ACCESS_TOKEN"); v != "" {
			s.AccessToken = v
		} else {
			return nil, errors.New("Environment variable TW_ACCESS_TOKEN not set. Set variable or provide a credentials file (-f)")
		}

		if v := os.Getenv("TW_TOKEN_SECRET"); v != "" {
			s.AccessTokenSecret = v
		} else {
			return nil, errors.New("Environment variable TW_TOKEN_SECRET not set. Set variable or provide a credentials file (-f)")
		}
	} else {
		file, err := os.Open(*f)
		if err != nil {
			return nil, err
		}

		err = json.NewDecoder(file).Decode(&s)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func main() {

	// Flags
	app := kingpin.New("tw", usage()).
		Version(fmt.Sprintf("Version: %s\nCommit: %s\nBuild: %s", version, commit, date)).
		DefaultEnvars().
		ErrorWriter(os.Stderr).
		UsageWriter(os.Stdout)

	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')
	file := app.Flag("file", "Path to a JSON file containing auth credentials").Short('f').String()
	pretty := app.Flag("pretty", "Pretty print instead of JSON (default: JSON)").Short('p').Bool()

	// Sub-commands
	likes := app.Command("likes", "Get your likes")

	// Parse flags
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	// Set up credentials
	c, err := newCreds(file)
	if err != nil {
		app.Fatalf("could not get credentials: %v", err)
	}

	anaconda.SetConsumerKey(c.ConsumerKey)
	anaconda.SetConsumerSecret(c.ConsumerSecret)
	api := anaconda.NewTwitterApi(c.AccessToken, c.AccessTokenSecret)

	var searchResult []anaconda.Tweet

	switch command {
	case likes.FullCommand():
		searchResult, err = getLikes(api, pretty)
		if err != nil {
			app.Fatalf("could not get tweets: %v", err)
		}
	}

	if !*pretty {
		b, err := json.Marshal(searchResult)
		if err != nil {
			app.Fatalf("could not encode data: %v", err)
		}
		fmt.Println(string(b))
	} else {
		for _, tweet := range searchResult {
			fmt.Println("-----------------------")
			fmt.Printf("From: %s\n", tweet.User.ScreenName)
			fmt.Printf("Text: %q\n", tweet.Text)
			for _, u := range tweet.Entities.Urls {
				fmt.Printf("Link: %s\n", u.Expanded_url)
			}

		}
	}

}
