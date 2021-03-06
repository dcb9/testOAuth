package twitter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	goTwitter "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
)

func NewConfig() *oauth1.Config {
	return &oauth1.Config{
		ConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		CallbackURL:    os.Getenv("TWITTER_CALLBACK_URL"),
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}
}

func GenerateTwitterLoginURL(config *oauth1.Config) (string, error) {
	requestToken, _, err := config.RequestToken()
	if err != nil {
		return "", err
	}

	authorizationURL, err := config.AuthorizationURL(requestToken)
	if err != nil {
		return "", err
	}

	return authorizationURL.String(), nil
}

func GetTwitterUser(config *oauth1.Config, request *http.Request) *goTwitter.User {
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(request)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	accessToken, accessSecret, err := config.AccessToken(requestToken, "", verifier)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	httpClient := config.Client(context.Background(), oauth1.NewToken(accessToken, accessSecret))
	twitterClient := goTwitter.NewClient(httpClient)
	accountVerifyParams := &goTwitter.AccountVerifyParams{
		IncludeEntities: goTwitter.Bool(false),
		SkipStatus:      goTwitter.Bool(true),
		IncludeEmail:    goTwitter.Bool(true),
	}
	user, resp, err := twitterClient.Accounts.VerifyCredentials(accountVerifyParams)
	err = validateResponse(user, resp, err)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return user
}

// Twitter login errors
var (
	ErrUnableToGetTwitterUser = errors.New("twitter: unable to get Twitter User")
)

// validateResponse returns an error if the given Twitter user, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *goTwitter.User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetTwitterUser
	}
	if user == nil || user.ID == 0 || user.IDStr == "" {
		return ErrUnableToGetTwitterUser
	}
	return nil
}
