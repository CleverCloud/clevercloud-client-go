package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/adrg/xdg"
)

// Best effort to grep credentials from environment
// Used internally or for extracting credentials
func (c *Client) guessOauth1Config() *OAuth1Config {
	if conf := c.guessOauth1ConfigFromEnv(); conf != nil {
		c.log.Info("Using Oauth1 user env vars")
		return conf
	}
	if conf := c.guessOauth1ConfigFromConfigFile(); conf != nil {
		c.log.Info("Using Oauth1 user config file")
		return conf
	}

	return nil
}

// guessOauth1ConfigFromEnv try to load OAuth1 credentials from user environment variables
func (c *Client) guessOauth1ConfigFromEnv() *OAuth1Config {
	secret := os.Getenv("CC_OAUTH_SECRET")
	token := os.Getenv("CC_OAUTH_TOKEN")
	if secret == "" || token == "" {
		c.log.Debug("Oauth1 user env vars are not set")
		return nil
	}

	return &OAuth1Config{
		AccessSecret:   secret,
		AccessToken:    token,
		ConsumerKey:    os.Getenv("CC_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("CC_CONSUMER_SECRET"),
	}
}

// guessOauth1ConfigFromConfigFile try to load OAuth1 credentials from user files
func (c *Client) guessOauth1ConfigFromConfigFile() *OAuth1Config {
	path := fmt.Sprintf("%s/%s", CONFIG_DIR, CONFIG_FILE_NAME)
	configFilePath, _ := xdg.SearchConfigFile(path)

	if configFilePath == "" {
		c.log.Debug("not user define configuration file")
		return nil
	}
	c.log.Debugf("Trying to get config from '%s'", configFilePath)

	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		c.log.WithError(err).Warn("cannot read user config file")
		return nil
	}

	var conf OAuth1Config
	if err := json.Unmarshal(content, &conf); err != nil {
		c.log.WithError(err).Warn("cannot parse user config file")
		return nil
	}

	if conf.AccessSecret == "" || conf.AccessToken == "" {
		c.log.Debug("Oauth1 user config file vars are not set")
		return nil
	}

	return &conf
}
