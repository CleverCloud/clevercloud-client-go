package client

// Using works of https://github.com/klaidas/go-oauth1/
// use PLAINTEXT signature instead of HMAC-SHA512

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// OAuth1Config own credentials to contact CleverCloud API.
type OAuth1Config struct {
	ConsumerKey    string `json:"-"`
	ConsumerSecret string `json:"-"`
	AccessToken    string `json:"token"`
	AccessSecret   string `json:"secret"`
}

// Sign an HTTP request with the given OAuth1 signature.
func (auth *OAuth1Config) Sign(req *http.Request) {
	if auth == nil {
		return
	}

	authHeader := auth.buildOAuth1Header()
	req.Header.Set("Authorization", authHeader)
}

// Params being any key-value url query parameter pairs.
func (auth OAuth1Config) buildOAuth1Header() string {
	nonce := auth.generateNonce()
	timestamp := strconv.Itoa(int(time.Now().Unix()))

	// PLAINTEXT signature: consumerSecret&tokenSecret
	signature := url.QueryEscape(auth.ConsumerSecret) + "&" + url.QueryEscape(auth.AccessSecret)

	return "OAuth oauth_consumer_key=\"" + url.QueryEscape(auth.ConsumerKey) +
		"\", oauth_nonce=\"" + url.QueryEscape(nonce) +
		"\", oauth_signature=\"" + url.QueryEscape(signature) +
		"\", oauth_signature_method=\"PLAINTEXT" +
		"\", oauth_timestamp=\"" + url.QueryEscape(timestamp) +
		"\", oauth_token=\"" + url.QueryEscape(auth.AccessToken) +
		"\", oauth_version=\"1.0\""
}

const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const NONCE_SIZE = 48

func (auth OAuth1Config) generateNonce() string {
	b := make([]byte, NONCE_SIZE)
	for i := range b {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allowed))))
		b[i] = allowed[r.Int64()]
	}

	return string(b)
}

func (auth OAuth1Config) Oauth1UserCredentials() (string, string) {
	return auth.AccessToken, auth.AccessSecret
}
