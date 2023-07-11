package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/alecthomas/kingpin/v2"

	"github.com/ChimeraCoder/anaconda"
	"github.com/carlmjohnson/versioninfo"
	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
)

type credentials struct {
	ConsumerKey       string `json:"Consumer_Key"`
	ConsumerSecret    string `json:"Consumer_Secret"`
	AccessToken       string `json:"Access_Token"`
	AccessTokenSecret string `json:"Access_Token_Secret"`
}

func main() {
	app := kingpin.New("tw", usage()).
		Version(versioninfo.Short()).
		DefaultEnvars().
		ErrorWriter(os.Stderr).
		UsageWriter(os.Stdout)

	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')
	file := app.Flag("file", "Path to a JSON file containing auth credentials (- for standard input)").Short('f').String()
	pretty := app.Flag("pretty", "Pretty print instead of JSON (default: JSON)").Short('p').Bool()

	// Sub-commands
	likes := app.Command("likes", "Get your likes")

	// Parse flags
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	c, err := readCredentials(file)
	if err != nil {
		app.Fatalf("could not read credentials: %v", err)
	}

	anaconda.SetConsumerKey(c.ConsumerKey)
	anaconda.SetConsumerSecret(c.ConsumerSecret)
	client := anaconda.NewTwitterApi(c.AccessToken, c.AccessTokenSecret)

	var searchResult []anaconda.Tweet

	switch command {
	case likes.FullCommand():
		searchResult, err = getLikes(client, pretty)
		if err != nil {
			app.Fatalf("could not get tweets: %v", err)
		}
	default:
		app.FatalUsage("unknown command")
	}

	if err := printTweets(os.Stdout, searchResult, *pretty); err != nil {
		app.Fatalf("could not print tweets: %v", err)
	}

}

func printTweets(out io.Writer, tweets []anaconda.Tweet, pretty bool) error {
	if !pretty {
		if err := json.NewEncoder(out).Encode(tweets); err != nil {
			return fmt.Errorf("encode tweets to json: %w", err)
		}
		return nil
	}

	for _, tweet := range tweets {
		fmt.Fprintln(out, "-----------------------")
		fmt.Fprintf(out, "From: %s\n", tweet.User.ScreenName)
		fmt.Fprintf(out, "Text: %q\n", tweet.Text)
		for _, u := range tweet.Entities.Urls {
			fmt.Fprintf(out, "Link: %s\n", u.Expanded_url)
		}
	}

	return nil
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

func readCredentials(f *string) (credentials, error) {
	s := credentials{}

	var input io.Reader
	if *f == "" {
		// read credentials from standard input
		input = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open(*f)
		if err != nil {
			return s, fmt.Errorf("open file: %w", err)
		}
		defer file.Close()
		input = file
	}

	if err := json.NewDecoder(input).Decode(&s); err != nil {
		return s, fmt.Errorf("decode json file: %w", err)
	}

	return s, nil
}

func usage() string {
	return `
  ______   __     __    
 /\__  _\ /\ \  _ \ \   
 \/_/\ \/ \ \ \/ ".\ \  
    \ \_\  \ \__/".~\_\ 
     \/_/   \/_/   \/_/ 
						
  A minimal Twitter CLI
	  
  This application uses Oauthv1 to securely authenticate requests.
  You can obtain API credentials from https://apps.twitter.com/.
  Always handle secrets carefully!

  Example:
  tw -f twitter_credentials.json likes
`
}
