package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	defaultMsBetweenPoll         = 30000
	defaultMaxConcurrentIndexers = 2
	defaultPushEnabled           = false
	defaultPollEnabled           = true
	defaultVcs                   = "git"
	defaultBaseUrl               = "{url}/blob/master/{path}{anchor}"
	defaultAnchor                = "#L{line}"
	defaultHealthChekURI         = "/healthz"
)

type UrlPattern struct {
	BaseUrl string `json:"base-url"`
	Anchor  string `json:"anchor"`
}

type Repo struct {
	Url               string         `json:"url"`
	MsBetweenPolls    int            `json:"ms-between-poll"`
	Vcs               string         `json:"vcs"`
	VcsConfigMessage  *SecretMessage `json:"vcs-config"`
	UrlPattern        *UrlPattern    `json:"url-pattern"`
	ExcludeDotFiles   bool           `json:"exclude-dot-files"`
	EnablePollUpdates *bool          `json:"enable-poll-updates"`
	EnablePushUpdates *bool          `json:"enable-push-updates"`
}

// Used for interpreting the config value for fields that use *bool. If a value
// is present, that value is returned. Otherwise, the default is returned.
func optionToBool(val *bool, def bool) bool {
	if val == nil {
		return def
	}
	return *val
}

// Are polling based updates enabled on this repo?
func (r *Repo) PollUpdatesEnabled() bool {
	return optionToBool(r.EnablePollUpdates, defaultPollEnabled)
}

// Are push based updates enabled on this repo?
func (r *Repo) PushUpdatesEnabled() bool {
	return optionToBool(r.EnablePushUpdates, defaultPushEnabled)
}

type Config struct {
	DbPath                string           `json:"dbpath"`
	Repos                 map[string]*Repo `json:"repos"`
	MaxConcurrentIndexers int              `json:"max-concurrent-indexers"`
	HealthCheckURI        string           `json:"health-check-uri"`
	FullCertFilename      string           `json:"full_cert_filename"`
	PrivCertFilename      string           `json:"priv_cert_filename"`
	JwtLoginURL           string           `json:"jwt_login_url"`
	JwtCookieName         string           `json:"jwt_cookie_name"`
	JwtSecretKey          string           `json:"jwt_secret_key"`
	FaviconURL            string           `json:"favicon_url"`
	TopHTML               string           `json:"top_html"`
}

// SecretMessage is just like json.RawMessage but it will not
// marshal its value as JSON. This is to ensure that vcs-config
// is not marshalled into JSON and send to the UI.
type SecretMessage []byte

// This always marshals to an empty object.
func (s *SecretMessage) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}

// See http://golang.org/pkg/encoding/json/#RawMessage.UnmarshalJSON
func (s *SecretMessage) UnmarshalJSON(b []byte) error {
	if b == nil {
		return errors.New("SecretMessage: UnmarshalJSON on nil pointer")
	}
	*s = append((*s)[0:0], b...)
	return nil
}

// Get the JSON encode vcs-config for this repo. This returns nil if
// the repo doesn't declare a vcs-config.
func (r *Repo) VcsConfig() []byte {
	if r.VcsConfigMessage == nil {
		return nil
	}
	return *r.VcsConfigMessage
}

// Populate missing config values with default values.
func initRepo(r *Repo) {
	if r.MsBetweenPolls == 0 {
		r.MsBetweenPolls = defaultMsBetweenPoll
	}

	if r.Vcs == "" {
		r.Vcs = defaultVcs
	}

	if r.UrlPattern == nil {
		r.UrlPattern = &UrlPattern{
			BaseUrl: defaultBaseUrl,
			Anchor:  defaultAnchor,
		}
	} else {
		if r.UrlPattern.BaseUrl == "" {
			r.UrlPattern.BaseUrl = defaultBaseUrl
		}

		if r.UrlPattern.Anchor == "" {
			r.UrlPattern.Anchor = defaultAnchor
		}
	}
}

// Populate missing config values with default values.
func initConfig(c *Config) {
	if c.MaxConcurrentIndexers == 0 {
		c.MaxConcurrentIndexers = defaultMaxConcurrentIndexers
	}

	if c.HealthCheckURI == "" {
		c.HealthCheckURI = defaultHealthChekURI
	}
}

func (c *Config) LoadFromFile(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := json.NewDecoder(r).Decode(c); err != nil {
		return err
	}

	if c.DbPath != "" && !filepath.IsAbs(c.DbPath) {
		path, err := filepath.Abs(
			filepath.Join(filepath.Dir(filename), c.DbPath))
		if err != nil {
			return err
		}
		c.DbPath = path
	}

	if c.FullCertFilename != "" && !filepath.IsAbs(c.FullCertFilename) {
		path, err := filepath.Abs(
			filepath.Join(filepath.Dir(filename), c.FullCertFilename))
		if err != nil {
			return err
		}
		c.FullCertFilename = path
	}
	if c.PrivCertFilename != "" && !filepath.IsAbs(c.PrivCertFilename) {
		path, err := filepath.Abs(
			filepath.Join(filepath.Dir(filename), c.PrivCertFilename))
		if err != nil {
			return err
		}
		c.PrivCertFilename = path
	}

	for _, repo := range c.Repos {
		initRepo(repo)
	}

	initConfig(c)

	return nil
}

func (c *Config) ToJsonString() (string, error) {
	b, err := json.Marshal(c.Repos)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
