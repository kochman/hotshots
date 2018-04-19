package config

import (
	"os"
	"path"
	"time"
)

// Config contains Hotshots configuration.
type Config struct {
	// Where the server listens
	ListenURL string

	// Where the server stores data
	PhotosDirectory string

	// Where the pusher should expect to find the server
	ServerURL string

	// How often the Pusher should check for new photos on the camera
	RefreshInterval time.Duration
}

// New reads from the environment to determine the configuration.
func New() (*Config, error) {
	c := &Config{
		ListenURL:       "127.0.0.1:8000",
		PhotosDirectory: "/var/hotshots",
		ServerURL:       "http://127.0.0.1:8000",
		RefreshInterval: 5 * time.Second,
	}

	dir, ok := os.LookupEnv("HOTSHOTS_DIR")
	if ok {
		c.PhotosDirectory = dir
	}

	listen, ok := os.LookupEnv("HOTSHOTS_LISTEN_URL")
	if ok {
		c.ListenURL = listen
	}

	server, ok := os.LookupEnv("HOTSHOTS_SERVER_URL")
	if ok {
		c.ServerURL = server
	}

	refresh, ok := os.LookupEnv("HOTSHOTS_REFRESH_INTERVAL")
	if ok {
		duration, err := time.ParseDuration(refresh)
		if err != nil {
			return nil, err
		}
		c.RefreshInterval = duration
	}

	return c, nil
}

func (c *Config) ImgFolder() string {
	return path.Join(c.PhotosDirectory, "/img")
}
func (c *Config) ConfFolder() string {
	return path.Join(c.PhotosDirectory, "/conf.d")
}

func (c *Config) StormFile() string {
	return path.Join(c.ConfFolder(), "/hotshot.db")
}
