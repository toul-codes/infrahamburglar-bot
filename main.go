package main

import (
	"context"
	"fmt"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/tweet/filteredstream"
	"github.com/michimani/gotwi/tweet/filteredstream/types"
	"github.com/michimani/gotwi/tweet/managetweet"
	twt "github.com/michimani/gotwi/tweet/managetweet/types"
	"net/http"
	"os"
	"strings"
	"time"
)

// ListSearchStreamRules lists search stream rules.
func listSearchStreamRules() {
	c, err := newGotwiClientWithTimeout(30)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := &types.ListRulesInput{}
	res, err := filteredstream.ListRules(context.Background(), c, p)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, r := range res.Data {
		fmt.Printf("ID: %s, Value: %s, Tag: %s\n", gotwi.StringValue(r.ID), gotwi.StringValue(r.Value), gotwi.StringValue(r.Tag))
	}
}

func deleteSearchStreamRules(ruleID string) {
	c, err := newGotwiClientWithTimeout(30)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := &types.DeleteRulesInput{
		Delete: &types.DeletingRules{
			IDs: []string{
				ruleID,
			},
		},
	}

	res, err := filteredstream.DeleteRules(context.TODO(), c, p)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, r := range res.Data {
		fmt.Printf("ID: %s, Value: %s, Tag: %s\n", gotwi.StringValue(r.ID), gotwi.StringValue(r.Value), gotwi.StringValue(r.Tag))
	}
}

// createSearchStreamRules creates a search stream rule.
func createSearchStreamRules(keyword string) {
	c, err := newGotwiClientWithTimeout(30)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := &types.CreateRulesInput{
		Add: []types.AddingRule{
			{Value: gotwi.String(keyword), Tag: gotwi.String(keyword)},
		},
	}

	res, err := filteredstream.CreateRules(context.TODO(), c, p)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, r := range res.Data {
		fmt.Printf("ID: %s, Value: %s, Tag: %s\n", gotwi.StringValue(r.ID), gotwi.StringValue(r.Value), gotwi.StringValue(r.Tag))
	}
}

// SimpleTweet posts a tweet with only text, and return posted tweet ID.
func SimpleTweet(c *gotwi.Client, text string) (string, error) {
	p := &twt.CreateInput{
		Text: gotwi.String(text),
	}

	res, err := managetweet.Create(context.Background(), c, p)
	if err != nil {
		return "", err
	}

	return gotwi.StringValue(res.Data.ID), nil
}

const (
	OAuthTokenEnvKeyName       = "GOTWI_ACCESS_TOKEN"
	OAuthTokenSecretEnvKeyName = "GOTWI_ACCESS_TOKEN_SECRET"
	USER                       = "1510513502628331522"
)

func newOAuth1Client() (*gotwi.Client, error) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           os.Getenv(OAuthTokenEnvKeyName),
		OAuthTokenSecret:     os.Getenv(OAuthTokenSecretEnvKeyName),
	}
	return gotwi.NewClient(in)
}

// execSearchStream call GET /2/tweets/search/stream API
// and outputs up to 10 results.
func execSearchStream() {

	c, err := newGotwiClientWithTimeout(120)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := &types.SearchStreamInput{
		TweetFields: fields.TweetFieldList{
			fields.TweetFieldAuthorID,
		},
	}
	s, err := filteredstream.SearchStream(context.Background(), c, p)
	if err != nil {
		fmt.Println(err)
		return
	}
	cnt := 0
	for s.Receive() {
		t, err := s.Read()
		if err != nil {
			fmt.Printf("ERR: ", err)
		} else {
			if t != nil {
				if gotwi.StringValue(t.Data.AuthorID) != USER {
					if !strings.Contains(gotwi.StringValue(t.Data.Text), "RT") {
						cnt++
						oauth1Client, err := newOAuth1Client()
						if err != nil {
							panic(err)
						}
						p := "Robble, Robble, Good Gobble " + "https://twitter.com/" + gotwi.StringValue(t.Data.AuthorID) + "/status/" + gotwi.StringValue(t.Data.ID)
						tweetID, err := SimpleTweet(oauth1Client, p)
						if err != nil {
							panic(err)
						}
						fmt.Println("Posted tweet ID is ", tweetID)
					}

				}
			}
			if cnt > 10 {
				s.Stop()
				break
			}
		}

	}
}

func newGotwiClientWithTimeout(timeout int) (*gotwi.Client, error) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth2BearerToken,
		HTTPClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
	return gotwi.NewClient(in)
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Println("The 1st parameter for command is required. (create|stream)")
		os.Exit(1)
	}
	command := args[1]
	switch command {
	case "list":
		// list search stream rules
		listSearchStreamRules()
	case "delete":
		// delete a specified rule
		if len(args) < 3 {
			fmt.Println("The 2nd parameter for rule ID to delete is required.")
			os.Exit(1)
		}

		ruleID := args[2]
		deleteSearchStreamRules(ruleID)
	case "create":
		// create a search stream rule
		if len(args) < 3 {
			fmt.Println("The 2nd parameter for keyword of search stream rule is required.")
			os.Exit(1)
		}

		keyword := args[2]
		createSearchStreamRules(keyword)
	case "stream":
		// exec filtered stream API
		execSearchStream()
	default:
		fmt.Println("Undefined command. Command should be 'create' or 'stream'.")
		os.Exit(1)
	}
}
