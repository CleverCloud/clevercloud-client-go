package client

import "os"

// Best effort to grep credentials from environment
// Used internally or for extracting credentials.
func (c *Client) GuessOauth1Config() *OAuth1Config {
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

// guessOauth1ConfigFromEnv try to load OAuth1 credentials from user environment variables.
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

// guessOauth1ConfigFromConfigFile try to load OAuth1 credentials from user files.
func (c *Client) guessOauth1ConfigFromConfigFile() *OAuth1Config {
	configFilePath := ConfigFilePath()
	if configFilePath == "" {
		c.log.Debug("not user define configuration file")

		return nil
	}

	c.log.Debugf("Trying to get config from '%s'", configFilePath)

	profile, err := ActiveProfile(configFilePath)
	if err != nil {
		c.log.WithError(err).Warn("cannot parse user config file")

		return nil
	}

	if profile == nil {
		c.log.Debug("Oauth1 user config file vars are not set")

		return nil
	}

	consumerKey := OAUTH_CONSUMER_KEY
	consumerSecret := OAUTH_CONSUMER_SECRET
	if profile.Overrides != nil {
		if profile.Overrides.OAuthConsumerKey != "" {
			consumerKey = profile.Overrides.OAuthConsumerKey
		}
		if profile.Overrides.OAuthConsumerSecret != "" {
			consumerSecret = profile.Overrides.OAuthConsumerSecret
		}
	}

	return &OAuth1Config{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		AccessToken:    profile.Token,
		AccessSecret:   profile.Secret,
	}
}

func (c *Client) guessBearerConfigFromEnv() *BearerConfig {
	token := os.Getenv("CLEVER_API_TOKEN")
	if token == "" {
		c.log.Warn("no CLEVER_API_TOKEN set in env")
	}

	return &BearerConfig{Token: token}
}
