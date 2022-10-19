package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
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

const (
	OAuthTokenEnvKeyName       = "GOTWI_ACCESS_TOKEN"
	OAuthTokenSecretEnvKeyName = "GOTWI_ACCESS_TOKEN_SECRET"
	USER                       = "1510513502628331522"
	SPAM                       = "1343113027139284992"
)

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
				if gotwi.StringValue(t.Data.AuthorID) != USER && gotwi.StringValue(t.Data.AuthorID) != SPAM {
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
	lambda.Start(execSearchStream)
}
